#!/bin/sh
# CLOUDSDK_CORE_PROJECT=<project> BILLING_ACCOUNT=<id> sh deploy/gcp/setup.sh
#
# Skips what already exists, so it can run again (docs/15-despliegue.md). The
# roles of each database are created through the Cloud SQL Auth Proxy, so it
# needs cloud-sql-proxy and psql.
set -eu

project=${CLOUDSDK_CORE_PROJECT:?set CLOUDSDK_CORE_PROJECT}
billing_account=${BILLING_ACCOUNT:?set BILLING_ACCOUNT}
region=us-east1
deployer=sigta-deploy@$project.iam.gserviceaccount.com
# Ids instead of names, so a recreated repository with the same name gets
# nothing.
repository_id=1404675991
owner_id=337795182
number=$(gcloud projects describe "$project" --format='value(projectNumber)')
here=$(cd "$(dirname "$0")" && pwd)

exists() { "$@" >/dev/null 2>&1; }
secret() { printf %s "$2" | gcloud secrets create "$1" --replication-policy=automatic --data-file=- >/dev/null; }

gcloud services enable run.googleapis.com sqladmin.googleapis.com secretmanager.googleapis.com \
    artifactregistry.googleapis.com iam.googleapis.com iamcredentials.googleapis.com sts.googleapis.com \
    billingbudgets.googleapis.com cloudresourcemanager.googleapis.com

exists gcloud artifacts repositories describe sigta --location="$region" ||
    gcloud artifacts repositories create sigta --location="$region" --repository-format=docker
gcloud artifacts repositories set-cleanup-policies sigta --location="$region" --no-dry-run \
    --policy=/dev/stdin >/dev/null <<'POLICY'
[{"name": "keep-recent", "action": {"type": "Keep"}, "mostRecentVersions": {"keepCount": 10}},
 {"name": "delete-old", "action": {"type": "Delete"}, "condition": {"olderThan": "7d"}}]
POLICY

for name in google-client-id google-client-secret; do
    exists gcloud secrets describe "sigta-$name" || {
        printf 'sigta-%s: ' "$name"
        read -r value
        secret "sigta-$name" "$value"
    }
done

for environment in prod dev; do
    if [ "$environment" = prod ]; then instance=sigta; else instance=sigta-dev; fi
    connection=$project:$region:$instance
    runner=sigta-$environment-run@$project.iam.gserviceaccount.com
    bucket=$project-sigta-$environment-documents

    exists gcloud sql instances describe "$instance" ||
        gcloud sql instances create "$instance" --region="$region" --database-version=POSTGRES_18 \
            --edition=enterprise --tier=db-f1-micro --storage-type=HDD --backup-start-time=06:00 \
            --deletion-protection
    exists gcloud sql databases describe sigta --instance="$instance" ||
        gcloud sql databases create sigta --instance="$instance"

    # The roles and their passwords are made once, together with the secrets
    # that carry them.
    if ! exists gcloud secrets describe "sigta-$environment-database-url"; then
        postgres=$(openssl rand -hex 24)
        backend=$(openssl rand -hex 24)
        migrations=$(openssl rand -hex 24)
        gcloud sql users set-password postgres --instance="$instance" --password="$postgres"
        cloud-sql-proxy --port 5439 --token "$(gcloud auth print-access-token)" "$connection" >/dev/null 2>&1 &
        proxy=$!
        sleep 5
        PGPASSWORD=$postgres psql -v ON_ERROR_STOP=1 -h 127.0.0.1 -p 5439 -U postgres -d sigta \
            -v backend_password="$backend" -v migrations_password="$migrations" \
            -v ai_password="$(openssl rand -hex 24)" -f "$here/../postgres/sql/roles.sql"
        kill "$proxy"
        socket=host=/cloudsql/$connection
        secret "sigta-$environment-migrations-database-url" "postgres://backend_migrations:$migrations@/sigta?$socket"
        secret "sigta-$environment-database-url" "postgres://backend:$backend@/sigta?$socket"
    fi

    exists gcloud storage buckets describe "gs://$bucket" ||
        gcloud storage buckets create "gs://$bucket" --location="$region" --uniform-bucket-level-access \
            --public-access-prevention

    # Each environment runs as its own account: dev cannot read the secrets,
    # the database or the documents of prod.
    exists gcloud iam service-accounts describe "$runner" ||
        gcloud iam service-accounts create "sigta-$environment-run" --display-name="SIGTA $environment"
    gcloud projects add-iam-policy-binding "$project" --member="serviceAccount:$runner" --role=roles/cloudsql.client \
        --condition="expression=resource.name == 'projects/$project/instances/$instance',title=$instance" >/dev/null
    gcloud storage buckets add-iam-policy-binding "gs://$bucket" --member="serviceAccount:$runner" \
        --role=roles/storage.objectUser >/dev/null
    for name in "$environment-database-url" "$environment-migrations-database-url" google-client-id google-client-secret; do
        gcloud secrets add-iam-policy-binding "sigta-$name" --member="serviceAccount:$runner" \
            --role=roles/secretmanager.secretAccessor >/dev/null
    done
    gcloud iam service-accounts add-iam-policy-binding "$runner" --member="serviceAccount:$deployer" \
        --role=roles/iam.serviceAccountUser >/dev/null
done

exists gcloud iam service-accounts describe "$deployer" ||
    gcloud iam service-accounts create sigta-deploy --display-name="SIGTA deploy from GitHub Actions"
exists gcloud iam workload-identity-pools describe github --location=global ||
    gcloud iam workload-identity-pools create github --location=global --display-name=GitHub
condition="assertion.repository_id == '$repository_id' && assertion.repository_owner_id == '$owner_id' && assertion.ref in ['refs/heads/main', 'refs/heads/develop']"
if exists gcloud iam workload-identity-pools providers describe sigta --location=global --workload-identity-pool=github; then
    gcloud iam workload-identity-pools providers update-oidc sigta --location=global \
        --workload-identity-pool=github --attribute-condition="$condition"
else
    gcloud iam workload-identity-pools providers create-oidc sigta --location=global \
        --workload-identity-pool=github --issuer-uri=https://token.actions.githubusercontent.com \
        --attribute-mapping=google.subject=assertion.sub,attribute.repository_id=assertion.repository_id \
        --attribute-condition="$condition"
fi
gcloud iam service-accounts add-iam-policy-binding "$deployer" --role=roles/iam.workloadIdentityUser \
    --member="principalSet://iam.googleapis.com/projects/$number/locations/global/workloadIdentityPools/github/attribute.repository_id/$repository_id" >/dev/null
# run.admin and not run.developer: the deploy makes the service public.
gcloud projects add-iam-policy-binding "$project" --member="serviceAccount:$deployer" \
    --role=roles/run.admin --condition=None >/dev/null
gcloud artifacts repositories add-iam-policy-binding sigta --location="$region" \
    --member="serviceAccount:$deployer" --role=roles/artifactregistry.writer >/dev/null

budget=$(gcloud billing budgets list --billing-account="$billing_account" --filter='displayName="SIGTA"' --format='value(name)')
if [ -z "$budget" ]; then
    gcloud billing budgets create --billing-account="$billing_account" --display-name=SIGTA \
        --budget-amount=60USD --filter-projects="projects/$project" \
        --threshold-rule=percent=0.5 --threshold-rule=percent=0.9 --threshold-rule=percent=1.0
else
    gcloud billing budgets update "$budget" --budget-amount=60USD
fi
