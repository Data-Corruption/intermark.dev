# Deployment

_Note: This guide assumes your server is linux and your repo is on github._

_Other Note: I highly recommend using Cloudflare for your domain, you can buy one, point it to your server, and get TLS no setup needed. Perfect for public sites that don't transmit personal data. If you need free end to end encryption, see the TLS section at the bottom of this page._

---

## Clone To Server

Clone the repository to your server. Ensure the server has Git, Go, and Node installed. Run the setup command, same as when you first made your fork and set it up locally.

---

## Continuous Deployment

### 1. Workflow Variables

The `IM_UPDATE_SECRET` var is used by the workflow on main pushes to authenticate update requests to your server.

- **1.1 - Generate a random string, here’s an easy way to do it**:

  ```sh
  openssl rand -base64 32
  ```

- **1.2 - Make it a permanent env**:

  First, check which shell you’re using:

  ```sh
  echo $SHELL
  ```

  Then run one of these:

  - **For Bash** (`/bin/bash`):

    ```sh
    echo 'export INTERMARK_UPDATE_SECRET="your-generated-secret-here"' >> ~/.bashrc
    source ~/.bashrc
    ```

  - **For Zsh** (`/bin/zsh`):

    ```sh
    echo 'export INTERMARK_UPDATE_SECRET="your-generated-secret-here"' >> ~/.zshrc
    source ~/.zshrc
    ```

- **1.3 - In your Repository settings, add the update secret and server url**:

  - Go to **Settings** > **Secrets and variables** > **Actions**.
  - **New repository secret**, Name it `IM_UPDATE_SECRET`, paste the generated secret.
  - **New repository secret**, Name it `IM_SERVER_ADDRESS`, paste your server url/domain

---

### 2. SSH Deploy Key

!!! **Skip the following if your repo is public** !!!

- **2.1 - Generate a ssh key pair**:
  
  **When prompted for a password, Do Not Set One, just hit enter.**
  
  <div id="ssh_gen"></div>

- **2.2 - In your Repository settings, add the key**:

  - Go to **Settings** > **Deploy keys**.
  - **Add deploy key**
    - Title it something like "Intermark Deployment Key".
    - Paste the public key you just generated into the **Key** field. Here is a command to print the pub key for easy copying:

      <div id="ssh_copy"></div>

---

## Point Domain To Server

If you have a domain, log into your domain registrar (e.g., Namecheap, GoDaddy, Cloudflare) and set the **A Record**:

- **Type**: A
- **Name**: @ (or your subdomain, e.g. `www`)
- **Value**: Your server's public IP

Use `dig` or `nslookup` to verify DNS propagation:

<div id="dns_check"></div>

You should see your server's IP returned.

---

## Setup NGINX

### Install

<div id="nginx_1_install"></div>
<div id="nginx_2_install"></div>

### Create Site Configuration

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

## TLS (https)

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
