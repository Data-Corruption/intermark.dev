# Deployment

_Note: This guide assumes your server is linux and your repo is on github._

_Other Note: I highly recommend using Cloudflare for your domain, you can buy one, point it to your server, and get TLS no setup needed. Perfect for public sites that don't transmit personal data. If you need free end to end encryption, see the TLS section at the bottom of this page._

---

## Clone To Server

Clone the repository to your server. Ensure the server has Git, Go, and Node installed. Run the setup command, same as when you first made your fork and set it up locally.

---

## Configuration

Intermark uses environment variables for configuration. Here is a full list of the variables you can set:

- **IM_ADDRESS**: The port your Intermark server will run on. Default is `:9292`.
- **IM_PAGE_CACHE_MB**: The size of the page cache in megabytes. Default is `1024` (1GB).
- **IM_ASSET_CACHE_MB**: The size of the asset cache in megabytes. Default is `1024` (1GB).
- **IM_LOG_LEVEL**: The log level. `debug`, `info`, `warn`, `error`, `none`. Default is `warn`.
- **IM_UPDATE_SECRET**: A secret string used to authenticate update requests from your GitHub Actions workflow. This is explained in the Continuous Deployment section below.

You can also set minute based timeouts for actions:

- **IM_GIT_M**: Fetch, pull, etc. Default is `5`.
- **IM_LFS_M**: LFS operations. Default is `5`.
- **IM_TAIL_M**: Tailwindcss built. Default is `1`.
- **IM_LUNR_M**: Lunr.js index build. Default is `1`.

You might need to change LFS if you have huge files, and Tailwind/Lunr if you have large sites. Otherwise this is mainly just to prevent the server from getting stuck if something goes wrong.

### Setting Environment Variables

For an example, we'll change the address. First, check which shell you’re using:

<div id="check_shell"></div>

Then run one of these:

- **For Bash** (`/bin/bash`):

  <div id="set_bash_env"></div>

- **For Zsh** (`/bin/zsh`):

  <div id="set_zsh_env"></div>

## Continuous Deployment

### 1. Workflow Setup

Generate a random string, here's an easy method:

<div id="secret_gen"></div>

Set the `IM_UPDATE_SECRET` environment variable to this string. Then, go to your repository settings and set the workflows vars:

- Go to **Settings** > **Secrets and variables** > **Actions**.
- **New repository secret**, name: `IM_UPDATE_SECRET`, value: secret string.
- **New repository variable**, name: `IM_SERVER_ADDRESS`, value: server url/domain

---

### 2. SSH Deploy Key

!!! **Skip the following if your repo is public** !!!

Generate a key pair:
  
**When prompted for a password, Do Not Set One, just hit enter.**
  
<div id="ssh_gen"></div>

In your Repository settings, add the key:

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

Use `dig` or `nslookup` to verify it's working:

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

As stated above, if you're using Cloudflare and not transmitting sensitive data, you can skip this step. Otherwise, if you want to secure your site with end to end TLS, follow [these](https://certbot.eff.org/instructions?ws=nginx&os=pip) steps.

---

## Updating Intermark

To merge changes from the main Intermark repository into your fork you can run:

<div id="update_intermark"></div>

This will fetch the latest changes from the Intermark repository and merge them into your fork. You can then address conflicts with your editor of choice, commit the changes, and push them to your fork.

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

    codeBlock('check_shell', 'echo $SHELL', 'sh');
    codeBlock('set_bash_env', 'echo \'export IM_ADDRESS=":9393"\' >> ~/.bashrc && source ~/.bashrc', 'sh');
    codeBlock('set_zsh_env', 'echo \'export IM_ADDRESS=":9393"\' >> ~/.zshrc && source ~/.zshrc', 'sh');
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
    codeBlock('update_intermark', 'go run inter.go update_intermark', 'sh');
  });
</script>
