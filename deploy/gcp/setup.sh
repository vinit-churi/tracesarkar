#!/usr/bin/env bash
#
# One-time setup of the Indian collector VM (ADR 0015).
#
# Creates: a service account, two secrets, a small VM in asia-south1, and an
# instance schedule that starts it at 02:30 IST. The VM powers itself off when
# the night's collection finishes.
#
# Safe to re-run: every step is skipped if it already exists.
#
# Before running, from the repository root:
#   gcloud auth login
#   gcloud config set project <your-project-id>
# and make sure .env and ca.pem exist locally, since their contents become the
# secrets this VM reads at boot.

set -euo pipefail

PROJECT="${PROJECT:-$(gcloud config get-value project 2>/dev/null)}"
REGION="${REGION:-asia-south1}"
ZONE="${ZONE:-asia-south1-c}"
VM="${VM:-tracesarkar-collector}"
SA="${SA:-tracesarkar-collector}"
SCHEDULE="${SCHEDULE:-tracesarkar-nightly}"
START_CRON="${START_CRON:-30 2 * * *}"   # 02:30 IST
STOP_CRON="${STOP_CRON:-0 3 * * *}"      # safety net if the script hangs
TIMEZONE="${TIMEZONE:-Asia/Kolkata}"
MACHINE="${MACHINE:-e2-micro}"

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ROOT="$(cd "$HERE/../.." && pwd)"

if [ -z "$PROJECT" ]; then
  echo "No project set. Run: gcloud config set project <project-id>" >&2
  exit 1
fi
for f in "$ROOT/.env" "$ROOT/ca.pem"; do
  [ -f "$f" ] || { echo "Missing $f — the VM reads these as secrets." >&2; exit 1; }
done

say() { printf '\n== %s\n' "$*"; }
exists() { "$@" >/dev/null 2>&1; }

say "Project ${PROJECT}, region ${REGION}, zone ${ZONE}"

say "Enabling the APIs this needs"
gcloud services enable compute.googleapis.com secretmanager.googleapis.com --project "$PROJECT"

say "Service account"
if exists gcloud iam service-accounts describe "${SA}@${PROJECT}.iam.gserviceaccount.com" --project "$PROJECT"; then
  echo "   already exists"
else
  gcloud iam service-accounts create "$SA" \
    --description="Runs the nightly TraceSarkar collector" \
    --display-name="TraceSarkar collector" --project "$PROJECT"
fi

say "Secrets (contents of .env and ca.pem)"
for pair in "tracesarkar-env:$ROOT/.env" "tracesarkar-ca:$ROOT/ca.pem"; do
  name="${pair%%:*}"; file="${pair#*:}"
  if exists gcloud secrets describe "$name" --project "$PROJECT"; then
    gcloud secrets versions add "$name" --data-file="$file" --project "$PROJECT" >/dev/null
    echo "   ${name}: new version added"
  else
    gcloud secrets create "$name" --data-file="$file" --replication-policy=automatic --project "$PROJECT"
  fi
  gcloud secrets add-iam-policy-binding "$name" \
    --member="serviceAccount:${SA}@${PROJECT}.iam.gserviceaccount.com" \
    --role="roles/secretmanager.secretAccessor" --project "$PROJECT" >/dev/null
done
echo "   the collector service account may read both"

say "VM (${MACHINE}, ${ZONE}, on-demand — not spot, see ADR 0015)"
if exists gcloud compute instances describe "$VM" --zone "$ZONE" --project "$PROJECT"; then
  echo "   already exists; updating its startup script"
  gcloud compute instances add-metadata "$VM" --zone "$ZONE" --project "$PROJECT" \
    --metadata-from-file startup-script="$HERE/startup.sh"
else
  gcloud compute instances create "$VM" \
    --project "$PROJECT" --zone "$ZONE" --machine-type "$MACHINE" \
    --image-family=debian-12 --image-project=debian-cloud \
    --boot-disk-size=10GB --boot-disk-type=pd-standard \
    --service-account="${SA}@${PROJECT}.iam.gserviceaccount.com" \
    --scopes=https://www.googleapis.com/auth/cloud-platform \
    --metadata-from-file startup-script="$HERE/startup.sh" \
    --labels=purpose=collector,phase=phase0
fi

say "Instance schedule (start ${START_CRON}, stop ${STOP_CRON}, ${TIMEZONE})"
if exists gcloud compute resource-policies describe "$SCHEDULE" --region "$REGION" --project "$PROJECT"; then
  echo "   already exists"
else
  # The Compute Engine service agent needs permission to start and stop the VM.
  PROJECT_NUMBER=$(gcloud projects describe "$PROJECT" --format='value(projectNumber)')
  gcloud projects add-iam-policy-binding "$PROJECT" \
    --member="serviceAccount:service-${PROJECT_NUMBER}@compute-system.iam.gserviceaccount.com" \
    --role="roles/compute.instanceAdmin.v1" >/dev/null
  gcloud compute resource-policies create instance-schedule "$SCHEDULE" \
    --project "$PROJECT" --region "$REGION" \
    --vm-start-schedule="$START_CRON" --vm-stop-schedule="$STOP_CRON" \
    --timezone="$TIMEZONE"
fi

say "Attaching the schedule"
if gcloud compute instances describe "$VM" --zone "$ZONE" --project "$PROJECT" \
     --format='value(resourcePolicies)' | grep -q "$SCHEDULE"; then
  echo "   already attached"
else
  gcloud compute instances add-resource-policies "$VM" \
    --zone "$ZONE" --project "$PROJECT" --resource-policies "$SCHEDULE"
fi

cat <<EOF

Done.

The VM starts at 02:30 IST, collects, and powers itself off. To test it now
without waiting for the schedule:

  gcloud compute instances start ${VM} --zone ${ZONE}
  gcloud compute instances tail-serial-port-output ${VM} --zone ${ZONE}   # Ctrl-C to stop watching

The first boot answers the open question in ADR 0015: whether BMC accepts an
Indian datacenter IP, or only residential ones. Look for a line reporting
"snapshot ... endpoint=publicdashboard". If instead it reports an i/o timeout,
the filter is by network type and collection has to move to a machine on a home
connection.

Afterwards, from your laptop:

  make status      # per-endpoint collection health
  make changes     # what changed in BMC's published data
EOF
