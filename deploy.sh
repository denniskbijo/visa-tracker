#!/usr/bin/env bash
set -euo pipefail

# Deploy visa-tracker to Oracle Cloud ARM VM.
# Usage: ./deploy.sh <VM_IP> [SSH_USER]
#
# Prerequisites:
#   - SSH key access to the VM
#   - /opt/visa-tracker directory exists on the VM (run setup-vm.sh first)

HOST="${1:?Usage: ./deploy.sh <VM_IP> [SSH_USER]}"
USER="${2:-opc}"
REMOTE_DIR="/opt/visa-tracker"

echo "==> Cross-compiling for linux/arm64..."
GOOS=linux GOARCH=arm64 go build -o visa-tracker-arm64 ./cmd/server

echo "==> Uploading to ${USER}@${HOST}:${REMOTE_DIR}..."
scp visa-tracker-arm64 "${USER}@${HOST}:${REMOTE_DIR}/visa-tracker"
scp -r data/ "${USER}@${HOST}:${REMOTE_DIR}/data/"
scp -r static/ "${USER}@${HOST}:${REMOTE_DIR}/static/"
ssh "${USER}@${HOST}" "mkdir -p ${REMOTE_DIR}/internal"
scp -r internal/templates/ "${USER}@${HOST}:${REMOTE_DIR}/internal/templates/"

echo "==> Restarting visa-tracker service..."
ssh "${USER}@${HOST}" "sudo systemctl restart visa-tracker"

echo "==> Checking service status..."
ssh "${USER}@${HOST}" "sudo systemctl is-active visa-tracker"

echo "==> Deployed successfully."
rm -f visa-tracker-arm64
