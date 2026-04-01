#!/usr/bin/env bash
set -euo pipefail

# Run this ON the Oracle Cloud VM after first SSH login.
# Copy with: scp setup-vm.sh opc@<VM_IP>:~/
# Then:       chmod +x ~/setup-vm.sh && ~/setup-vm.sh
#
# Caddy is installed from the official GitHub release (amd64 tarball) instead of
# dnf copr -- copr often appears to "hang" on small VMs while downloading metadata.

CADDY_VERSION="${CADDY_VERSION:-2.11.2}"

echo "==> Opening firewall ports 80 and 443..."
if command -v firewall-cmd &>/dev/null; then
    sudo firewall-cmd --permanent --add-port=80/tcp
    sudo firewall-cmd --permanent --add-port=443/tcp
    sudo firewall-cmd --reload
elif command -v iptables &>/dev/null; then
    sudo iptables -I INPUT -m state --state NEW -p tcp --dport 80 -j ACCEPT
    sudo iptables -I INPUT -m state --state NEW -p tcp --dport 443 -j ACCEPT
    sudo netfilter-persistent save 2>/dev/null || true
fi
echo "    Firewall ports opened."

echo "==> Installing Caddy (${CADDY_VERSION}, official binary)..."
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64) CADDY_ARCH="amd64" ;;
    aarch64) CADDY_ARCH="arm64" ;;
    *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac

TMP="$(mktemp -d)"
trap 'rm -rf "$TMP"' EXIT
cd "$TMP"
curl -fsSL -o caddy.tar.gz \
    "https://github.com/caddyserver/caddy/releases/download/v${CADDY_VERSION}/caddy_${CADDY_VERSION}_linux_${CADDY_ARCH}.tar.gz"
tar xzf caddy.tar.gz caddy
sudo mv caddy /usr/local/bin/caddy
sudo chmod 755 /usr/local/bin/caddy
sudo chown root:root /usr/local/bin/caddy
if command -v restorecon &>/dev/null; then
    sudo restorecon -v /usr/local/bin/caddy || true
fi

sudo mkdir -p /etc/caddy /var/lib/caddy
sudo tee /etc/caddy/Caddyfile > /dev/null << 'CADDYSTUB'
# Replaced below after DuckDNS prompts
:80 {
    respond "visa-tracker: configure DuckDNS in setup-vm.sh"
}
CADDYSTUB

sudo tee /etc/systemd/system/caddy.service > /dev/null << 'UNIT'
[Unit]
Description=Caddy reverse proxy
After=network.target

[Service]
Type=simple
User=root
ExecStart=/usr/local/bin/caddy run --environ --config /etc/caddy/Caddyfile
ExecReload=/usr/local/bin/caddy reload --config /etc/caddy/Caddyfile --force
TimeoutStopSec=5

[Install]
WantedBy=multi-user.target
UNIT

sudo systemctl daemon-reload
sudo systemctl enable caddy
sudo systemctl restart caddy
echo "    Caddy installed and running."

echo "==> Creating app directory..."
sudo mkdir -p /opt/visa-tracker/internal
sudo chown -R "$(whoami):$(id -gn)" /opt/visa-tracker
echo "    /opt/visa-tracker ready."

echo "==> Installing systemd service..."
sudo tee /etc/systemd/system/visa-tracker.service > /dev/null << 'UNIT'
[Unit]
Description=UK Visa Tracker
After=network.target

[Service]
Type=simple
User=opc
WorkingDirectory=/opt/visa-tracker
ExecStart=/opt/visa-tracker/visa-tracker
Restart=always
RestartSec=5
Environment=PORT=8080
Environment=DB_PATH=/opt/visa-tracker/visa-tracker.db
Environment=DATA_DIR=/opt/visa-tracker/data
Environment=STATIC_DIR=/opt/visa-tracker/static
Environment=TEMPLATES_DIR=/opt/visa-tracker/internal/templates

[Install]
WantedBy=multi-user.target
UNIT
sudo systemctl daemon-reload
sudo systemctl enable visa-tracker
echo "    systemd service installed (will start after first deploy)."

echo ""
read -rp "Enter your DuckDNS subdomain (e.g. visa-tracker): " DUCKDNS_DOMAIN
read -rp "Enter your DuckDNS token: " DUCKDNS_TOKEN

FQDN="${DUCKDNS_DOMAIN}.duckdns.org"

echo "==> Configuring Caddy for ${FQDN}..."
sudo tee /etc/caddy/Caddyfile > /dev/null << EOF
${FQDN} {
    reverse_proxy localhost:8080
}
EOF
sudo systemctl restart caddy
echo "    Caddy configured for https://${FQDN}"

echo "==> Setting up DuckDNS cron (updates IP every 5 minutes)..."
(crontab -l 2>/dev/null | grep -v duckdns; echo "*/5 * * * * curl -s 'https://www.duckdns.org/update?domains=${DUCKDNS_DOMAIN}&token=${DUCKDNS_TOKEN}&ip=' > /dev/null") | crontab -
echo "    DuckDNS cron installed."

echo ""
echo "============================================"
echo "  VM setup complete."
echo ""
echo "  Next steps:"
echo "    1. From your local machine, run:"
echo "       ./deploy.sh <THIS_VM_IP>"
echo ""
echo "    2. Your site will be live at:"
echo "       https://${FQDN}"
echo "============================================"
