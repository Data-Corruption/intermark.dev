# Getting Started

---

## Installation

---

### Prerequisites

Before you begin, ensure you have the following installed on your system:

[**Git**](https://git-scm.com/downloads), [**Go**](https://go.dev/doc/install), and [**Node.js**](https://nodejs.org/en/download)

_For Node.js on windows, I recommend using the **Windows Installer (.msi)**, not the powershell script. Say yes to npm and adding to path if prompted_

---

### Forking the Repository

On your preferred Git platform (GitHub, GitLab, etc.) fork the Intermark repository (On Github use the template feature), clone it to your local machine, and open it in a terminal.

If you want / need to manually fork for whatever reason:

<div id="fork_code"></div>

---

### Dependencies

Run the following command to install dependencies (go modules and tail/daisyui):

<div id="deps_code"></div>

---

## Running

Intermark runs in two modes: **Edit Mode** and **Production Mode**.

### Edit Mode

Use this mode locally to preview your site as you write, and edit the sidebar with a GUI. It rebuilds pages from source on every request. To start Intermark in Edit Mode, run:

<div id="edit_code"></div>

### Production Mode

This mode builds everything with various optimizations during:

- Initial startup
- On updates when you push to the main branch of your repo

To start Intermark in Production Mode, run:

<div id="prod_code"></div>

---

<div class="flex flex-row justify-between mt-10">
  <a href="/p/introduction" class="btn btn-primary">Previous: Introduction</a>
  <a href="/p/usage/writing-content" class="btn btn-secondary">Next: Writing Content</a>
</div>

<script>
  window.addEventListener('load', () => {
    const fork_code =
`git clone https://github.com/Data-Corruption/Intermark.git
cd intermark
# After creating a repo on your preferred platform, e.g. Github,
# set it as the remote and push your changes
git remote add origin "your-repo-url"
git push -u origin main`;

    codeBlock('fork_code', fork_code, 'sh');
    codeBlock('deps_code', 'go run inter.go deps', 'sh');
    codeBlock('edit_code', 'go run inter.go edit', 'sh');
    codeBlock('prod_code', 'go run inter.go prod', 'sh');
  });
</script>
