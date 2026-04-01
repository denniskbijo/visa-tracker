# Hosting on Oracle Cloud (Free Tier)

Step-by-step guide to deploy visa-tracker on an Oracle Cloud VM with HTTPS and a free subdomain.

## What you get

- Always-on VM running the Go binary
- HTTPS with auto-renewing Let's Encrypt certificate
- Free subdomain via DuckDNS (e.g. `visa-tracker.duckdns.org`)
- Auto-restart on crash (systemd)
- Auto-deploy on `git push` (GitHub Actions)

## Prerequisites

- An Oracle Cloud account (sign up at https://cloud.oracle.com)
- An SSH key pair on your local machine (`~/.ssh/id_ed25519` or `~/.ssh/id_rsa`)
- Go 1.22+ installed locally

---

## Step 1: Create the VM

1. Log in to **https://cloud.oracle.com**
2. Go to **Compute > Instances > Create Instance**
3. Configure:

| Setting | Value |
|---|---|
| **Name** | `visa-tracker` |
| **Region** | Your nearest (e.g. UK South - London) |
| **Image** | Oracle Linux 9 |
| **Shape** | VM.Standard.E2.1.Micro (AMD) -- 1 OCPU, 1 GB RAM (Always Free) |
| **Networking** | Create new VCN or use default. **Tick "Assign public IPv4 address"** |
| **SSH key** | Upload your public key (`~/.ssh/id_ed25519.pub`) |
| **Boot volume** | 50 GB (default) |

4. Click **Create** -- the VM provisions in about 60 seconds
5. Once running, note the **Public IP address** from the instance details page

### Shape notes

| Shape | Arch | Free tier? | Specs |
|---|---|---|---|
| VM.Standard.E2.1.Micro | AMD (x86_64) | Yes -- 1 OCPU / 1 GB forever | Recommended. App uses ~30 MB idle |
| VM.Standard.A1.Flex | ARM (aarch64) | Yes -- 4 OCPU / 24 GB total forever | Best specs, but unavailable in some regions (e.g. London) |
| VM.Standard.E5.Flex | AMD (x86_64) | No -- paid (~$22/mo) | Not needed for this app |

If you use A1.Flex (ARM), change `GOARCH=amd64` to `GOARCH=arm64` in `deploy.sh` and `.github/workflows/deploy.yml`.

---

## Step 2: Open firewall ports

Oracle has **two firewalls** -- the cloud security list AND the OS-level firewall. Both must allow ports 80 and 443.

### 2a. Cloud Security List (Oracle Console)

1. Go to **Networking > Virtual Cloud Networks**
2. Click your VCN (created with the instance)
3. Click **Security Lists > Default Security List**
4. Click **Add Ingress Rules** and add two rules:

| Source CIDR | Protocol | Dest Port | Description |
|---|---|---|---|
| `0.0.0.0/0` | TCP | 80 | HTTP |
| `0.0.0.0/0` | TCP | 443 | HTTPS |

### 2b. OS-level firewall (done in Step 3 via setup script)

The `setup-vm.sh` script handles this automatically.

---

## Step 3: Set up the VM

First, register a free subdomain:

1. Go to **https://www.duckdns.org** and sign in with GitHub
2. Create a subdomain -- e.g. `visa-tracker` (gives you `visa-tracker.duckdns.org`)
3. Set the IP to your VM's public IP address
4. Copy your **DuckDNS token** (shown at the top of the page)

Copy the script to the VM and run it **on the instance** (do not pipe `bash -s` over SSH -- long `dnf` installs can drop the connection and look hung):

```bash
scp setup-vm.sh opc@<YOUR_VM_IP>:~/
ssh opc@<YOUR_VM_IP>
chmod +x ~/setup-vm.sh
~/setup-vm.sh
```

`setup-vm.sh` installs Caddy from the official GitHub release tarball (not `dnf copr`), which avoids slow or stuck COPR metadata on small VMs.

The script will:
- Open firewall ports 80/443
- Install Caddy (reverse proxy with auto-HTTPS)
- Create `/opt/visa-tracker` directory
- Install a systemd service for visa-tracker
- Ask for your DuckDNS subdomain and token
- Configure Caddy for HTTPS
- Set up a cron job to keep the DuckDNS IP updated

When prompted:
- **DuckDNS subdomain**: enter just the subdomain part (e.g. `visa-tracker`)
- **DuckDNS token**: paste the token from duckdns.org

**Oracle Linux and Caddy:** the image uses **SELinux** in enforcing mode. The Caddy binary must be owned by **`root`** and labeled so systemd (running as root) may execute it under `/usr/local/bin`. Current **`setup-vm.sh`** sets **`chown root:root`**, **`chmod 755`**, and **`restorecon`** on `/usr/local/bin/caddy` after install. If you installed Caddy manually or copied the binary as a normal user, **`systemctl status caddy`** may show **`status=203/EXEC`** even when **`file /usr/local/bin/caddy`** looks correct — fix with ownership + **`restorecon`** (see **Caddy failed** under Troubleshooting below).

---

## Step 4: Deploy the application

From your local machine, in the project root:

```bash
./deploy.sh <YOUR_VM_IP>
```

This will:
1. Cross-compile the Go binary for linux/amd64
2. Upload the binary, data files, templates, and static assets via SCP
3. Restart the **`visa-tracker`** systemd service (not Caddy)
4. Verify the service is running

**Caddy** (HTTPS reverse proxy) is installed in Step 3 only. **`deploy.sh`** and **GitHub Actions** deploy do not modify Caddy. If the app deploy succeeds but the site is unreachable over HTTPS, check **`sudo systemctl status caddy`** on the VM first.

Your site should now be live at **https://visa-tracker.duckdns.org** (or whatever subdomain you chose).

### Verify manually

```bash
# check service logs on the VM
ssh opc@<YOUR_VM_IP> "sudo journalctl -u visa-tracker --no-pager -n 30"

# test from your machine
curl -I https://visa-tracker.duckdns.org
```

---

## Step 5: Set up auto-deploy (GitHub Actions)

So that every `git push` to `main` automatically builds and deploys:

1. Go to **https://github.com/denniskbijo/visa-tracker/settings/secrets/actions**
2. Click **New repository secret** and add these secrets:

| Secret name | Value |
|---|---|
| `OCI_HOST` | Your VM's public IP address (e.g. `129.151.xxx.xxx`) |
| `OCI_USER` | `opc` (default for Oracle Linux) |
| `OCI_SSH_KEY` | Your SSH **private** key (the full content of `~/.ssh/id_ed25519`) |
| `OCI_SSH_KEY_PASSPHRASE` | Passphrase for `OCI_SSH_KEY` if that key is encrypted. **Skip this secret** if the private key has no passphrase. |

To copy your private key:

```bash
cat ~/.ssh/id_ed25519
```

Paste the entire output (including the `-----BEGIN` and `-----END` lines) as the value of `OCI_SSH_KEY`.

If your key prompts for a passphrase when you use `ssh`, create **`OCI_SSH_KEY_PASSPHRASE`** and paste the passphrase only (no surrounding quotes or newlines). For a **dedicated CI key with no passphrase**, do not add this secret.

3. Test it by pushing a change:

```bash
git add -A && git commit -m "test deploy" && git push origin master:main
```

4. Go to **Actions** tab on GitHub to watch the deploy run

---

## Troubleshooting

### "Connection refused" on ports 80/443

Both firewalls need to be open. Check:
```bash
# on the VM -- is Caddy running?
sudo systemctl status caddy

# is the app running?
sudo systemctl status visa-tracker

# are firewall ports open?
sudo firewall-cmd --list-ports
```

### Caddy failed (`status=203/EXEC`)

If `systemctl status caddy` shows **Active: failed** and **`code=exited, status=203/EXEC`**, systemd never started Caddy.

**1. Oracle Linux: ownership and SELinux (common when the binary is fine)**

On Oracle Linux / RHEL, **`203/EXEC`** often appears when **`/usr/local/bin/caddy`** is the **correct ELF** for your CPU but has the **wrong owner** (e.g. `opc:opc`) or **wrong SELinux context**, so **`execve`** is denied. **`file`** and **`uname -m`** look normal.

Run on the VM:

```bash
ls -la /usr/local/bin/caddy
getenforce
sudo ls -Z /usr/local/bin/caddy
```

Fix (order matters: ownership, then default label for that path):

```bash
sudo chown root:root /usr/local/bin/caddy
sudo chmod 755 /usr/local/bin/caddy
sudo restorecon -v /usr/local/bin/caddy
sudo /usr/local/bin/caddy version
sudo systemctl restart caddy
sudo systemctl status caddy
```

If **`sudo /usr/local/bin/caddy version`** still fails, check SELinux denials: **`sudo ausearch -m avc -ts recent`**, and mounts: **`mount | grep -E 'noexec|/usr'`**.

**2. Wrong architecture or missing binary**

If **`file /usr/local/bin/caddy`** is not an ELF for your VM, reinstall. Oracle **ARM** (`aarch64`) needs **arm64**; **AMD** (`x86_64`) needs **amd64**:

```bash
uname -m
ls -la /usr/local/bin/caddy
file /usr/local/bin/caddy
```

Reinstall Caddy to match **`uname -m`**, then apply the same **root** ownership and **`restorecon`** as in `setup-vm.sh`:

```bash
CADDY_VERSION="${CADDY_VERSION:-2.11.2}"
ARCH="$(uname -m)"
case "$ARCH" in
  x86_64)  CADDY_ARCH="amd64" ;;
  aarch64) CADDY_ARCH="arm64" ;;
  *) echo "Unsupported arch: $ARCH"; exit 1 ;;
esac
cd "$(mktemp -d)"
curl -fsSL -o caddy.tar.gz \
  "https://github.com/caddyserver/caddy/releases/download/v${CADDY_VERSION}/caddy_${CADDY_VERSION}_linux_${CADDY_ARCH}.tar.gz"
tar xzf caddy.tar.gz caddy
sudo mv caddy /usr/local/bin/caddy
sudo chmod 755 /usr/local/bin/caddy
sudo chown root:root /usr/local/bin/caddy
command -v restorecon &>/dev/null && sudo restorecon -v /usr/local/bin/caddy || true
sudo /usr/local/bin/caddy version
sudo systemctl restart caddy
sudo systemctl status caddy
```

Then from your laptop, **`curl -I https://your-name.duckdns.org`** should get a response (often **200** or **502** if the app is down — **502** still proves Caddy is listening).

### App starts but pages don't load

Check that Caddy is proxying correctly:
```bash
# test the app directly
curl http://localhost:8080/

# check Caddy logs
sudo journalctl -u caddy --no-pager -n 20
```

### DuckDNS not resolving

```bash
# force update DuckDNS
curl "https://www.duckdns.org/update?domains=visa-tracker&token=YOUR_TOKEN&ip="

# check DNS
nslookup visa-tracker.duckdns.org
```

### Deploy workflow fails

- Verify the GitHub secrets **`OCI_HOST`**, **`OCI_USER`**, and **`OCI_SSH_KEY`** are set correctly
- Ensure `OCI_SSH_KEY` contains the **private** key, not the public key
- If the log shows **`ssh.ParsePrivateKey: ssh: this private key is passphrase protected`** or **`unable to authenticate`**, set **`OCI_SSH_KEY_PASSPHRASE`** to the key’s passphrase, or switch to a passphrase-less deploy key and update `OCI_SSH_KEY`
- Check that the VM allows SSH from GitHub’s runners (outbound SSH from the internet to your VM on port 22 — same as your laptop)

### `scp: ... Text file busy` when running `./deploy.sh`

Linux cannot overwrite a running executable. The **`deploy.sh`** script stops `visa-tracker` before `scp`, then starts it again. If you deploy manually, run:

```bash
ssh opc@<VM_IP> 'sudo systemctl stop visa-tracker'
# ... scp files ...
ssh opc@<VM_IP> 'sudo systemctl start visa-tracker'
```

### Caddy install stuck on "Installing Caddy..."

Older versions of `setup-vm.sh` used `dnf copr enable @caddy/caddy`, which can sit for many minutes on a 1 GB VM. **Pull the latest `setup-vm.sh`** from the repo -- it downloads the official Caddy binary from GitHub instead. If a run is stuck, press **Ctrl+C**, then run the updated script again.

### Oracle Cloud metrics and logs

**Console metrics** (CPU, memory, network) update **every few minutes**, not in real time. If you are under heavy load, charts can lag.

**Live logs on the VM** (always authoritative):

```bash
# kernel and systemd
sudo journalctl -f

# your app
sudo journalctl -u visa-tracker -f

# Caddy
sudo journalctl -u caddy -f

# last 100 lines
sudo journalctl -u visa-tracker -n 100 --no-pager
```

**OCI Console:** **Compute > your instance > Metrics** for CPU / memory / network. **Monitoring > Service health** is account-wide, not per-VM detail.

---

## Other "free forever" or long-term free hosting options

| Provider | What is free | Caveat |
|---|---|---|
| **Oracle Cloud Always Free** | E2.1.Micro, A1 ARM, 200 GB storage, 10 TB egress | Regions/shapes sometimes unavailable; setup is manual |
| **Fly.io** | Small shared VMs, 3 VMs on free allowance | Not truly "forever" unlimited; check current limits |
| **Render** | Web service with cold starts | Free tier sleeps; not ideal for always-on SQLite |
| **Google Cloud** | $300 credit for 90 days, then paid | Always Free tier is small (e.g. `e2-micro` US regions only) |
| **AWS** | 12-month free trial | After that, paid |
| **Azure** | $200 credit for 30 days | Free tier is limited and time-limited |

For a **SQLite file on disk** and **always-on** app, **Oracle Always Free** or a **small VPS you control** (including Fly.io with a volume, if within free limits) are the practical fit. Serverless platforms (Cloud Run, Lambda) are a poor match unless you move the DB to a managed service.

---

## Updating

After the initial setup, deploying new code is just:

```bash
git push origin master:main
```

GitHub Actions will automatically build the binary and deploy it to the VM.

For manual deploys (without CI):

```bash
./deploy.sh <YOUR_VM_IP>
```

## Cost

| Component | Cost |
|---|---|
| VM.Standard.E2.1.Micro (1 OCPU, 1 GB) | Always Free forever |
| VM.Standard.A1.Flex ARM (if available) | Always Free forever |
| Network egress (10 TB/month) | Free |
| Boot volume (50 GB) | Free |
| DuckDNS subdomain | Free |
| Let's Encrypt TLS | Free |
| Caddy | Free (open source) |
