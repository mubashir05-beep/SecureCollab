# Prerequisites Setup Guide

This guide helps you set up the SecureCollab development environment from scratch.

## Option 1: Install Devbox (Recommended)

Devbox automatically provides all tools at pinned versions.

### Windows Installation

**Via WinGet (Recommended):**
```powershell
winget install jetpack-io.devbox
```

**Via Scoop:**
```powershell
scoop bucket add jetpack-io https://github.com/jetpack-io/scoop-bucket.git
scoop install devbox
```

**Via Manual Download:**
Download from: https://www.jetpack.io/devbox/docs/installing_devbox/

### After Installation
```powershell
# Restart your terminal
cd C:\Users\Mubashir\Documents\GitHub\SecureCollab
devbox shell
```

---

## Option 2: Install Tools Manually

If you don't want to use Devbox, install these tools:

### Required Tools

**1. Go 1.22**
```powershell
winget install GoLang.Go.1.22
```

**2. Task (go-task)**
```powershell
# Via Chocolatey
choco install go-task

# Via Scoop
scoop install task

# Via Go
go install github.com/go-task/task/v3/cmd/task@latest
```

**3. Docker Desktop**
```powershell
winget install Docker.DockerDesktop
```

### Optional (for full dev experience)
- Rust: `winget install Rustlang.Rust.MSVC`
- Node.js 20: `winget install OpenJS.NodeJS.LTS`
- Terraform: `winget install Hashicorp.Terraform`
- Helm: `choco install kubernetes-helm`
- kubectl: `choco install kubernetes-cli`

---

## Fix WSL Task Command Errors

Your WSL Ubuntu has task aliased but not installed. Fix it:

```bash
# Option A: Install task in WSL
wsl -d ubuntu
sudo snap install task --classic

# Option B: Remove the alias from WSL if not needed
wsl -d ubuntu
echo "# Disabled task alias" >> ~/.bashrc
sed -i 's/alias t=task/#alias t=task/' ~/.bashrc
source ~/.bashrc
```

---

## Verify Installation

```powershell
# Check tools are available
go version          # Should show 1.22.x
task --version      # Should show task v3.x
docker --version    # Should show Docker version

# Enter project
cd C:\Users\Mubashir\Documents\GitHub\SecureCollab

# If using Devbox
devbox shell
task --list

# If using manual install
task --list
```

---

## Quick Start After Setup

```powershell
# Create your local env file from the template (first time only)
Copy-Item .env.example .env

# Start all services
task dev

# In another terminal, run tests
task test

# Check everything is working
task smoke:gateway
```

---

## Troubleshooting

**"devbox: command not found"**
- Restart your terminal after installing Devbox
- Check PATH includes Devbox binary location

**"task: command not found"**
- If using Devbox, make sure you're inside `devbox shell`
- If manual install, make sure task is in your PATH

**WSL spamming "Taskfile.yml: command not found"**
- Your WSL shell is trying to execute task but it's not installed
- Install task in WSL or remove the alias from `~/.bashrc`

**Docker Compose errors**
- Make sure Docker Desktop is running
- Run `docker --version` to verify installation
