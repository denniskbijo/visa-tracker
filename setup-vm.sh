#!/usr/bin/env bash
set -euo pipefail

# Run this ON the Oracle Cloud VM after first SSH login.
# Usage: ssh opc@<VM_IP> 'bash -s' < setup-vm.sh
#
# What it does:
#   1. Opens firewall ports 80/443
#   2. Installs Caddy
#   3. Creates /opt/visa-tracker directory
#   4. Installs systemd service for visa-tracker
#   5. Prompts for DuckDNS domain and configures Caddy

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

echo "==> Installing Caddy..."
if command -v dnf &>/dev/null; then
    sudo dnf install -y 'dnf-command(copr)' 2>/dev/null || true
    sudo dnf copr enable -y @caddy/caddy 2>/dev/null || true
    sudo dnf install -y caddy
elif command -v apt &>/dev/null; then
    sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
    curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
    sudo apt update
    sudo apt install -y caddy
fi
echo "    Caddy installed."

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
sudo systemctl enable caddy
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
