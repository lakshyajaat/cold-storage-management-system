# GitHub Repository Setup Guide

Follow these steps to create a new GitHub repository and push your Cold Storage Management System.

---

## Step 1: Create GitHub Repository

### Option A: Using GitHub Website (Recommended)

1. **Go to GitHub:** https://github.com/new

2. **Fill in repository details:**
   - **Repository name:** `cold-storage-management-system`
   - **Description:** `A comprehensive web-based management system for cold storage facilities with role-based access control, inventory tracking, and payment management.`
   - **Visibility:** Choose either:
     - ✅ **Private** (recommended for business projects)
     - ⬜ Public (if you want it open source)
   - **DO NOT** initialize with README, .gitignore, or license (we already have these)

3. **Click "Create repository"**

4. **Copy the repository URL** (it will look like):
   ```
   https://github.com/YOUR_USERNAME/cold-storage-management-system.git
   ```

### Option B: Using GitHub CLI (Alternative)

```bash
# Install GitHub CLI if not already installed
# For Ubuntu/Debian:
sudo apt install gh

# Login to GitHub
gh auth login

# Create the repository (private)
gh repo create cold-storage-management-system --private --source=. --remote=origin

# Or create public repository
gh repo create cold-storage-management-system --public --source=. --remote=origin
```

---

## Step 2: Add GitHub Remote and Push

Once you have your GitHub repository URL, run these commands:

```bash
# Navigate to project directory
cd /home/lakshya/jupyter-/cold/cold-backend

# Remove existing remote if any
git remote remove origin 2>/dev/null || true

# Add your new GitHub repository as remote
# Replace YOUR_USERNAME with your actual GitHub username
git remote add origin https://github.com/YOUR_USERNAME/cold-storage-management-system.git

# Verify remote is added
git remote -v

# Push to GitHub (main branch)
git push -u origin main
```

---

## Step 3: Verify Upload

After pushing, verify by visiting:
```
https://github.com/YOUR_USERNAME/cold-storage-management-system
```

You should see:
- ✅ All your code files
- ✅ Complete documentation (README.md, API_DOCUMENTATION.md, etc.)
- ✅ Migrations folder
- ✅ Templates and static files
- ✅ .gitignore (hiding .env and binaries)

---

## Quick Command Summary

```bash
# 1. Create repo on GitHub (via website or gh cli)

# 2. Add remote and push
git remote add origin https://github.com/YOUR_USERNAME/cold-storage-management-system.git
git push -u origin main

# 3. Done! 🎉
```

---

## Alternative: Using SSH (More Secure)

If you prefer using SSH keys:

### Setup SSH Key (One-time)

```bash
# Generate SSH key
ssh-keygen -t ed25519 -C "your_email@example.com"

# Start SSH agent
eval "$(ssh-agent -s)"

# Add SSH key
ssh-add ~/.ssh/id_ed25519

# Copy public key
cat ~/.ssh/id_ed25519.pub
# Copy the output and add to GitHub: Settings → SSH Keys → New SSH Key
```

### Push Using SSH

```bash
# Add remote using SSH
git remote add origin git@github.com:YOUR_USERNAME/cold-storage-management-system.git

# Push
git push -u origin main
```

---

## Troubleshooting

### Problem: Authentication Failed (HTTPS)

GitHub no longer accepts password authentication. You need a **Personal Access Token**:

1. Go to: https://github.com/settings/tokens
2. Click "Generate new token (classic)"
3. Give it a name: "Cold Storage Project"
4. Select scopes: `repo` (all)
5. Click "Generate token"
6. **Copy the token** (you won't see it again!)
7. When pushing, use the token as password:
   ```bash
   Username: YOUR_USERNAME
   Password: ghp_xxxxxxxxxxxxxxxxxxxx (paste your token)
   ```

### Problem: Remote Already Exists

```bash
# Remove existing remote
git remote remove origin

# Add new remote
git remote add origin https://github.com/YOUR_USERNAME/cold-storage-management-system.git
```

### Problem: Permission Denied (SSH)

```bash
# Test SSH connection
ssh -T git@github.com

# If it fails, check if SSH key is added to GitHub
cat ~/.ssh/id_ed25519.pub
# Copy and add to: https://github.com/settings/keys
```

---

## Post-Setup: Add Repository Description

After creating the repository, add a description and topics:

1. Go to your repository page
2. Click ⚙️ (Settings) or "About" section
3. Add description:
   ```
   A comprehensive web-based management system for cold storage facilities with role-based access control, inventory tracking, and payment management.
   ```
4. Add topics (tags):
   ```
   go, golang, postgresql, cold-storage, inventory-management,
   role-based-access-control, payment-system, web-application,
   jwt-authentication, rest-api
   ```
5. Set website URL (if deployed):
   ```
   http://your-deployment-url.com
   ```

---

## Recommended: Add GitHub Repository Badges

Add these badges to the top of your README.md after pushing:

```markdown
![Go Version](https://img.shields.io/badge/Go-1.22-00ADD8?logo=go)
![PostgreSQL](https://img.shields.io/badge/PostgreSQL-15-316192?logo=postgresql)
![License](https://img.shields.io/badge/license-MIT-green)
![Tests](https://img.shields.io/badge/tests-12%2F12%20passing-brightgreen)
```

---

## Next Steps After Pushing

1. ✅ **Enable GitHub Actions** (for CI/CD)
   - Create `.github/workflows/test.yml` for automated testing

2. ✅ **Add Branch Protection**
   - Go to Settings → Branches
   - Add rule for `main` branch
   - Require pull request reviews

3. ✅ **Create Development Branch**
   ```bash
   git checkout -b develop
   git push -u origin develop
   ```

4. ✅ **Add Collaborators** (if team project)
   - Settings → Collaborators → Add people

5. ✅ **Enable Issues and Wiki**
   - Settings → Features → Enable Issues and Wiki

---

## Git Workflow Going Forward

### Making Changes

```bash
# 1. Create feature branch
git checkout -b feature/new-feature

# 2. Make changes and commit
git add .
git commit -m "Add new feature"

# 3. Push to GitHub
git push -u origin feature/new-feature

# 4. Create Pull Request on GitHub
# Go to GitHub and create PR from feature branch to main

# 5. After PR is merged, update main
git checkout main
git pull origin main
```

### Keeping Repository Updated

```bash
# Pull latest changes
git pull origin main

# Check status
git status

# View commit history
git log --oneline --graph --all

# View changes
git diff
```

---

## Repository Statistics

**Your Cold Storage Management System:**

- **Language:** Go 1.22
- **Database:** PostgreSQL 15
- **Files:** 50+ source files
- **Documentation:** 7 comprehensive docs (~24,500 words)
- **Tests:** 12/12 passing (100%)
- **Lines of Code:** ~8,000+ (excluding documentation)
- **Features:**
  - Role-based access control
  - Entry management
  - Payment processing
  - Room allocation
  - System settings
  - Complete REST API

---

## Security Notes

### Files Already Ignored (in .gitignore):

✅ `.env` - Environment variables (passwords, secrets)
✅ `cold-backend` - Compiled binary
✅ Build artifacts (*.exe, *.dll, etc.)

### Never Commit These:

❌ Database passwords
❌ JWT secrets
❌ API keys
❌ Personal access tokens
❌ SSL certificates/private keys

### If You Accidentally Commit Secrets:

```bash
# Remove file from git but keep locally
git rm --cached .env

# Commit the removal
git commit -m "Remove .env from tracking"

# Push
git push origin main

# IMPORTANT: Change all exposed secrets immediately!
```

---

## Helpful Git Commands

```bash
# View all branches
git branch -a

# Switch branches
git checkout branch-name

# Delete branch
git branch -d branch-name

# View remote info
git remote -v

# View commit history
git log --oneline

# Undo last commit (keep changes)
git reset --soft HEAD~1

# View what changed in last commit
git show

# View repository stats
git log --stat
```

---

## Done! 🎉

Your Cold Storage Management System is now on GitHub!

**Repository URL:**
```
https://github.com/YOUR_USERNAME/cold-storage-management-system
```

Share this URL with your team or deploy from GitHub!

---

**Last Updated:** 2025-12-13
