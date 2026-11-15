#!/bin/sh
# Creates the staging environment of SIGTA in Google Cloud: a small VM with
# Docker Swarm in a network that accepts no connection from the internet, and
# the identity GitHub Actions uses to deploy on it, without long-lived keys.
# Each step skips what already exists, so it can run again.
#
#   CLOUDSDK_CORE_PROJECT=<project> BILLING_ACCOUNT=<id> sh deploy/gcp/setup.sh
#
# What it creates and why is in docs/15-despliegue.md.
set -eu

project=${CLOUDSDK_CORE_PROJECT:?set CLOUDSDK_CORE_PROJECT}
billing_account=${BILLING_ACCOUNT:?set BILLING_ACCOUNT}
region=southamerica-east1
zone=$region-b
vm=sigta-staging
deployer=sigta-deploy
deployer_email=$deployer@$project.iam.gserviceaccount.com
# GitHub ids instead of names: a renamed or recreated repository with the
# same name gets nothing.
repository_id=1404675991
owner_id=337795182
here=$(cd "$(dirname "$0")" && pwd)
number=$(gcloud projects describe "$project" --format='value(projectNumber)')

exists() { "$@" >/dev/null 2>&1; }

gcloud services enable compute.googleapis.com iap.googleapis.com oslogin.googleapis.com \
    iam.googleapis.com iamcredentials.googleapis.com sts.googleapis.com \
    billingbudgets.googleapis.com cloudresourcemanager.googleapis.com

# Network: no external addresses. SSH only through Identity-Aware Proxy, and
# the VM goes out to the internet (images, Google, Cloudflare) through NAT.
exists gcloud compute networks describe sigta ||
    gcloud compute networks create sigta --subnet-mode=custom
exists gcloud compute networks subnets describe sigta --region="$region" ||
    gcloud compute networks subnets create sigta --network=sigta --region="$region" \
        --range=10.10.0.0/24 --enable-private-ip-google-access
exists gcloud compute firewall-rules describe sigta-allow-iap-ssh ||
    gcloud compute firewall-rules create sigta-allow-iap-ssh --network=sigta \
        --direction=INGRESS --action=allow --rules=tcp:22 --source-ranges=35.235.240.0/20
exists gcloud compute routers describe sigta --region="$region" ||
    gcloud compute routers create sigta --network=sigta --region="$region"
exists gcloud compute routers nats describe sigta --router=sigta --region="$region" ||
    gcloud compute routers nats create sigta --router=sigta --region="$region" \
        --auto-allocate-nat-external-ips --nat-all-subnet-ip-ranges

# The VM has no service account: it does not call Google Cloud.
exists gcloud compute instances describe "$vm" --zone="$zone" ||
    gcloud compute instances create "$vm" --zone="$zone" --machine-type=e2-small \
        --subnet=sigta --no-address --no-service-account --no-scopes \
        --image-family=debian-13 --image-project=debian-cloud \
        --boot-disk-size=20GB --boot-disk-type=pd-balanced \
        --shielded-secure-boot --shielded-vtpm --shielded-integrity-monitoring \
        --metadata=enable-oslogin=TRUE \
        --metadata-from-file=startup-script="$here/startup.sh"

# GitHub Actions gets a token for the deployer only from a run of the
# develop branch of the repository, through Workload Identity Federation.
exists gcloud iam service-accounts describe "$deployer_email" ||
    gcloud iam service-accounts create "$deployer" --display-name="SIGTA deploy from GitHub Actions"
exists gcloud iam workload-identity-pools describe github --location=global ||
    gcloud iam workload-identity-pools create github --location=global --display-name=GitHub
exists gcloud iam workload-identity-pools providers describe sigta --location=global --workload-identity-pool=github ||
    gcloud iam workload-identity-pools providers create-oidc sigta --location=global \
        --workload-identity-pool=github --issuer-uri=https://token.actions.githubusercontent.com \
        --attribute-mapping=google.subject=assertion.sub,attribute.repository_id=assertion.repository_id,attribute.repository_owner_id=assertion.repository_owner_id,attribute.ref=assertion.ref \
        --attribute-condition="assertion.repository_id == '$repository_id' && assertion.repository_owner_id == '$owner_id' && assertion.ref == 'refs/heads/develop'"
gcloud iam service-accounts add-iam-policy-binding "$deployer_email" --role=roles/iam.workloadIdentityUser \
    --member="principalSet://iam.googleapis.com/projects/$number/locations/global/workloadIdentityPools/github/attribute.repository_id/$repository_id" >/dev/null

# The deployer reaches the VM by SSH through the proxy and runs the deploy
# with sudo; it can do nothing else in the project.
for role in roles/compute.viewer roles/iap.tunnelResourceAccessor roles/compute.osAdminLogin; do
    gcloud projects add-iam-policy-binding "$project" --member="serviceAccount:$deployer_email" \
        --role="$role" --condition=None >/dev/null
done

# Budget alert: emails the billing administrators at half, 90% and all of it.
if [ -z "$(gcloud billing budgets list --billing-account="$billing_account" --filter='displayName="SIGTA staging"' --format='value(name)')" ]; then
    gcloud billing budgets create --billing-account="$billing_account" --display-name="SIGTA staging" \
        --budget-amount=25USD --filter-projects="projects/$project" \
        --threshold-rule=percent=0.5 --threshold-rule=percent=0.9 --threshold-rule=percent=1.0
fi

echo "Workload identity provider: projects/$number/locations/global/workloadIdentityPools/github/providers/sigta"
echo "Service account: $deployer_email"
