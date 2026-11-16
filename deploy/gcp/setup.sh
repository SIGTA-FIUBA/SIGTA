#!/bin/sh
# CLOUDSDK_CORE_PROJECT=<project> BILLING_ACCOUNT=<id> sh deploy/gcp/setup.sh
#
# Skips what already exists, so it can run again (docs/15-despliegue.md).
set -eu

project=${CLOUDSDK_CORE_PROJECT:?set CLOUDSDK_CORE_PROJECT}
billing_account=${BILLING_ACCOUNT:?set BILLING_ACCOUNT}
region=southamerica-east1
zone=$region-b
vm=sigta
deployer=sigta-deploy
deployer_email=$deployer@$project.iam.gserviceaccount.com
# Ids instead of names, so a recreated repository with the same name gets
# nothing.
repository_id=1404675991
owner_id=337795182
number=$(gcloud projects describe "$project" --format='value(projectNumber)')

exists() { "$@" >/dev/null 2>&1; }

gcloud services enable compute.googleapis.com iap.googleapis.com oslogin.googleapis.com \
    iam.googleapis.com iamcredentials.googleapis.com sts.googleapis.com \
    billingbudgets.googleapis.com cloudresourcemanager.googleapis.com

exists gcloud compute networks describe sigta ||
    gcloud compute networks create sigta --subnet-mode=custom
exists gcloud compute networks subnets describe sigta --region="$region" ||
    gcloud compute networks subnets create sigta --network=sigta --region="$region" --range=10.10.0.0/24
exists gcloud compute firewall-rules describe sigta-allow-iap-ssh ||
    gcloud compute firewall-rules create sigta-allow-iap-ssh --network=sigta \
        --direction=INGRESS --action=allow --rules=tcp:22 --source-ranges=35.235.240.0/20
exists gcloud compute firewall-rules describe sigta-allow-web ||
    gcloud compute firewall-rules create sigta-allow-web --network=sigta \
        --direction=INGRESS --action=allow --rules=tcp:80,tcp:443 --source-ranges=0.0.0.0/0
exists gcloud compute addresses describe sigta --region="$region" ||
    gcloud compute addresses create sigta --region="$region"

startup=$(mktemp)
trap 'rm -f "$startup"' EXIT
cat >"$startup" <<'SCRIPT'
#!/bin/sh
set -eu
if ! command -v docker >/dev/null; then
    apt-get update
    DEBIAN_FRONTEND=noninteractive apt-get install --yes docker.io
fi
if [ "$(docker info --format '{{.Swarm.LocalNodeState}}')" != active ]; then
    docker swarm init --advertise-addr "$(hostname -I | cut -d' ' -f1)"
fi
SCRIPT
exists gcloud compute instances describe "$vm" --zone="$zone" ||
    gcloud compute instances create "$vm" --zone="$zone" --machine-type=e2-medium \
        --subnet=sigta --address=sigta --no-service-account --no-scopes \
        --image-family=debian-13 --image-project=debian-cloud \
        --boot-disk-size=30GB --boot-disk-type=pd-balanced \
        --shielded-secure-boot --shielded-vtpm --shielded-integrity-monitoring \
        --metadata=enable-oslogin=TRUE --metadata-from-file=startup-script="$startup"

exists gcloud iam service-accounts describe "$deployer_email" ||
    gcloud iam service-accounts create "$deployer" --display-name="SIGTA deploy from GitHub Actions"
exists gcloud iam workload-identity-pools describe github --location=global ||
    gcloud iam workload-identity-pools create github --location=global --display-name=GitHub
condition="assertion.repository_id == '$repository_id' && assertion.repository_owner_id == '$owner_id' && assertion.ref in ['refs/heads/main', 'refs/heads/develop']"
if exists gcloud iam workload-identity-pools providers describe sigta --location=global --workload-identity-pool=github; then
    gcloud iam workload-identity-pools providers update-oidc sigta --location=global \
        --workload-identity-pool=github --attribute-condition="$condition"
else
    gcloud iam workload-identity-pools providers create-oidc sigta --location=global \
        --workload-identity-pool=github --issuer-uri=https://token.actions.githubusercontent.com \
        --attribute-mapping=google.subject=assertion.sub,attribute.repository_id=assertion.repository_id,attribute.repository_owner_id=assertion.repository_owner_id,attribute.ref=assertion.ref \
        --attribute-condition="$condition"
fi
gcloud iam service-accounts add-iam-policy-binding "$deployer_email" --role=roles/iam.workloadIdentityUser \
    --member="principalSet://iam.googleapis.com/projects/$number/locations/global/workloadIdentityPools/github/attribute.repository_id/$repository_id" >/dev/null
for role in roles/compute.viewer roles/iap.tunnelResourceAccessor roles/compute.osAdminLogin; do
    gcloud projects add-iam-policy-binding "$project" --member="serviceAccount:$deployer_email" \
        --role="$role" --condition=None >/dev/null
done

budget=$(gcloud billing budgets list --billing-account="$billing_account" --filter='displayName="SIGTA"' --format='value(name)')
if [ -z "$budget" ]; then
    gcloud billing budgets create --billing-account="$billing_account" --display-name=SIGTA \
        --budget-amount=60USD --filter-projects="projects/$project" \
        --threshold-rule=percent=0.5 --threshold-rule=percent=0.9 --threshold-rule=percent=1.0
else
    gcloud billing budgets update "$budget" --budget-amount=60USD
fi

echo "IP: $(gcloud compute addresses describe sigta --region="$region" --format='value(address)')"
