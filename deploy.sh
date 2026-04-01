#!/usr/bin/env bash
set -euo pipefail

# Deploy visa-tracker to an Oracle Cloud VM.
# Usage: ./deploy.sh <VM_IP> [SSH_USER]
#
# Prerequisites:
#   - SSH key access to the VM
#   - /opt/visa-tracker directory exists on the VM (run setup-vm.sh first)
#
# The service is stopped before overwriting the binary. Linux returns
# "Text file busy" if you scp over a file that is the running executable.

HOST="${1:?Usage: ./deploy.sh <VM_IP> [SSH_USER]}"
USER="${2:-opc}"
REMOTE_DIR="/opt/visa-tracker"

echo "==> Cross-compiling for linux/amd64..."
GOOS=linux GOARCH=amd64 go build -o visa-tracker-linux ./cmd/server

echo "==> Stopping visa-tracker (required to replace running binary)..."
ssh "${USER}@${HOST}" "sudo systemctl stop visa-tracker" || true

echo "==> Uploading to ${USER}@${HOST}:${REMOTE_DIR}..."
scp visa-tracker-linux "${USER}@${HOST}:${REMOTE_DIR}/visa-tracker"
scp -r data/ "${USER}@${HOST}:${REMOTE_DIR}/data/"
scp -r static/ "${USER}@${HOST}:${REMOTE_DIR}/static/"
ssh "${USER}@${HOST}" "mkdir -p ${REMOTE_DIR}/internal"
scp -r internal/templates/ "${USER}@${HOST}:${REMOTE_DIR}/internal/templates/"

echo "==> Starting visa-tracker..."
ssh "${USER}@${HOST}" "sudo systemctl start visa-tracker"

echo "==> Checking service status..."
sleep 2
ssh "${USER}@${HOST}" "sudo systemctl is-active visa-tracker"

echo "==> Deployed successfully."
rm -f visa-tracker-linux
