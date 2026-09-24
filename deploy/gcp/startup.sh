#!/usr/bin/env bash
#
# Runs on every boot of the collector VM (ADR 0015).
#
# The machine exists for about three minutes a day: it boots, fetches its
# credentials from Secret Manager, downloads the current collector, runs it, and
# powers itself off. The instance schedule's stop time is the safety net if
# anything here hangs.
#
# Secrets live in tmpfs under /run, which does not survive the shutdown, so
# nothing sensitive is written to the boot disk. The binary cannot live there
# too: /run is mounted noexec on Debian, so it goes to /usr/local/lib, where
# being world-readable costs nothing — it is a public release artefact.

set -uo pipefail

REPO="${REPO:-vinit-churi/tracesarkar}"
RELEASE="${RELEASE:-collector-latest}"
WORKDIR=/run/tracesarkar          # tmpfs: secrets, gone at shutdown
BINDIR=/usr/local/lib/tracesarkar  # /run is noexec, so the binary lives here
LOG_TAG=tracesarkar

log() { echo "[$(date -u +%FT%TZ)] $*" | tee >(logger -t "$LOG_TAG"); }

# Power off whatever happens. A VM left running costs money and, worse, hides a
# failure: the next night's schedule would find it already up.
finish() {
  local code=$?
  log "shutting down (exit ${code})"
  rm -rf "$WORKDIR"
  shutdown -h now
}
trap finish EXIT

mkdir -p "$WORKDIR" "$BINDIR"
cd "$WORKDIR"

log "fetching credentials from Secret Manager"
if ! gcloud secrets versions access latest --secret=tracesarkar-env > "$WORKDIR/.env"; then
  log "FATAL: could not read the tracesarkar-env secret"
  exit 1
fi
if ! gcloud secrets versions access latest --secret=tracesarkar-ca > "$WORKDIR/ca.pem"; then
  log "FATAL: could not read the tracesarkar-ca secret"
  exit 1
fi
chmod 600 "$WORKDIR/.env" "$WORKDIR/ca.pem"

# The CA path inside the secret is relative to the .env file, which is here.
if ! grep -q '^POSTGRES_CA_PATH=' "$WORKDIR/.env"; then
  echo "POSTGRES_CA_PATH=./ca.pem" >> "$WORKDIR/.env"
fi

log "downloading the collector"
BINARY_URL="https://github.com/${REPO}/releases/download/${RELEASE}/ingest-linux-amd64"
if ! curl -fsSL --retry 3 --retry-delay 5 -o "$BINDIR/ingest" "$BINARY_URL"; then
  log "FATAL: could not download ${BINARY_URL}"
  exit 1
fi
chmod +x "$BINDIR/ingest"

# Verify the checksum when it is published alongside the binary.
if curl -fsSL --retry 2 -o "$WORKDIR/ingest.sha256" "${BINARY_URL}.sha256"; then
  expected=$(awk '{print $1}' "$WORKDIR/ingest.sha256")
  actual=$(sha256sum "$BINDIR/ingest" | awk '{print $1}')
  if [ "$expected" != "$actual" ]; then
    log "FATAL: checksum mismatch for the collector binary"
    exit 1
  fi
  log "checksum verified"
fi

# The collector reads the source register at runtime and this machine has no
# checkout, so it comes from the same release as the binary.
log "downloading the source register"
if ! curl -fsSL --retry 3 --retry-delay 5 -o "$WORKDIR/sources.yaml" \
     "https://github.com/${REPO}/releases/download/${RELEASE}/sources.yaml"; then
  log "FATAL: could not download the source register"
  exit 1
fi

export TRACESARKAR_ENV_FILE="$WORKDIR/.env"
export TRACESARKAR_TIER=personal

log "applying migrations"
"$BINDIR/ingest" migrate 2>&1 | tee >(logger -t "$LOG_TAG")

# This VM exists because these sources refuse foreign IPs. It collects
# everything; GitHub Actions separately covers the sources reachable from
# anywhere, so a failure here does not stop the whole record.
log "collecting"
"$BINDIR/ingest" run --register "$WORKDIR/sources.yaml" 2>&1 | tee >(logger -t "$LOG_TAG")
run_status=${PIPESTATUS[0]}

log "watching for new government resolutions"
"$BINDIR/ingest" watch --limit 40 --register "$WORKDIR/sources.yaml" 2>&1 | tee >(logger -t "$LOG_TAG")

log "collection status"
"$BINDIR/ingest" status 2>&1 | tee >(logger -t "$LOG_TAG")

if [ "$run_status" -ne 0 ]; then
  log "collection reported errors (exit ${run_status})"
  exit "$run_status"
fi
log "done"
