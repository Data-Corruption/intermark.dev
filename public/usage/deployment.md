# Deployment

_Note: This guide assumes the server you're hosting on is linux and the git repo platform github. I can add more example but honestly just show this one to an llm and ask about how do do it on your platform. Like with all llm usage it's probs wrong and outdated but a decent starting point for research_

---

## 1. Clone To Server

After writing content, previewing it locally using `go run .\inter.go edit`, and pushing to your repo, clone the repository to your server. Ensure the server has Git, Go, and Node.js installed.

---

## 2. Generate Auth Keys

### Update Secret

This will be included in update requests sent by the repo workflow on pushes to main. This is so only update requests from the repo can update the site. Here is a quick way to generate one, any random string will do:

<div id="secret_gen"></div>

Here's the big reveal... My evil marketing lied, there is a config file! muhahahaha. Copy that string to the `update-secret` field in `./public/.meta/config.toml`. This is used to verify that update requests are coming from the repo workflow.

Now in your Repository settings, add the secret.

- Go to **Settings** > **Secrets and variables** > **Actions**.
- **New repository secret**, Name it `UPDATE_TOKEN`, paste that shit in there.

While you're here, add the url of your server. If you bought a domain (recommended), use that.

- **New repository secret**, Name it `SERVER_ADDRESS`, paste the url/domain

### Deploy Key

This will allow Intermark to pull from private repos. **Skip this step if your repo is public.**

First, gen a ssh key pair. **When prompted for a password, Do Not Set One, just hit enter.**

<div id="ssh_gen"></div>

Then in your Repository settings, add the key.

- Go to **Settings** > **Deploy keys**, Click **Add deploy key**.
- Enter a title (e.g., "Prod Deployment Key").
- Paste the public key you just generated into the **Key** field. Here is a command to print the pub key for easy copying:

<div id="ssh_copy"></div>

---

## 4. Point Your Domain to Your Server

If you have a domain, log into your domain registrar (e.g., Namecheap, GoDaddy, Cloudflare) and set the **A Record**:

- **Type**: A
- **Name**: @ (or your subdomain, e.g. `www`)
- **Value**: Your server's public IP

Use `dig` or `nslookup` to verify DNS propagation:

<div id="dns_check"></div>

You should see your server's IP returned.

---

## 5. Setup NGINX as a Reverse Proxy with TLS

### Install NGINX

<div id="nginx_1_install"></div>
<div id="nginx_2_install"></div>

### Create Basic HTTP Reverse Proxy

Create a new NGINX configuration file for your site at `/etc/nginx/sites-available/yourdomain.com`:

<div id="nginx_config"></div>

Enable it:

<div id="nginx_1_enable"></div>
<div id="nginx_2_enable"></div>

### Test Your Site

Start Intermark in production mode:

<div id="edit_mode"></div>

Visit `http://yourdomain.com` in your browser. You should see your Intermark site.

---

## 6. Secure Site with Certbot + Let's Encrypt

Follow [these](https://certbot.eff.org/instructions?ws=nginx&os=pip) steps to secure your site with TLS. _Note: this does involve installing python/pip, but it's well worth it for the free auto renewing TLS cert._

---

You're all set! Your Intermark site should now be live and secure. Push changes to your repo and watch them manifest on your site automatically. <3

<script>
  window.addEventListener('load', () => {
    const nginx_config =
`server {
    listen 80;
    server_name yourdomain.com www.yourdomain.com;

    location / {
        proxy_pass http://127.0.0.1:9292;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}`;

    codeBlock('secret_gen', 'openssl rand -base64 32', 'sh');
    codeBlock('ssh_gen', 'ssh-keygen -t ed25519 -f ~/.ssh/id_ed25519_intermark', 'sh');
    codeBlock('ssh_copy', 'cat ~/.ssh/id_ed25519_intermark.pub', 'sh');
    codeBlock('dns_check', 'dig +short yourdomain.com', 'sh');
    codeBlock('nginx_1_install', `sudo apt update`, 'sh');
    codeBlock('nginx_2_install', `sudo apt install nginx`, 'sh');
    codeBlock('nginx_1_enable', `sudo ln -s /etc/nginx/sites-available/yourdomain.com /etc/nginx/sites-enabled/`, 'sh');
    codeBlock('nginx_2_enable', `sudo nginx -t && sudo systemctl reload nginx`, 'sh');
    codeBlock('nginx_config', nginx_config, 'nginx');
    codeBlock('edit_mode', `go run ./inter.go prod`, 'sh');
  });
</script>
