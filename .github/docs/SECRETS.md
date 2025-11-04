# GitHub Secrets Configuration

This document describes the GitHub Secrets required for CI/CD workflows.

## Required Secrets

### For Release Workflow (`.github/workflows/release.yml`)

**`GITHUB_TOKEN`** - ✅ **Automatically provided** (for main repository)

This token is automatically provided by GitHub Actions and has permissions for the current repository (cdnscli). It can create releases and upload assets.

**`HOMEBREW_TAP_TOKEN`** - ⚠️ **Required for Homebrew tap updates**

For GoReleaser to update the Homebrew tap repository (`mixanemca/homebrew-tap`), you need a Personal Access Token (PAT) with write access to that repository.

**How to create and add the token:**

1. Go to GitHub → Click your profile icon (top right) → **Settings**
2. In the left sidebar, scroll down to **Developer settings** (at the bottom)
3. Click **Personal access tokens** → **Tokens (classic)**
4. Click **Generate new token (classic)**
5. **Note:** `homebrew-tap-update`
6. **Expiration:** Choose expiration (e.g., 90 days, 1 year, or no expiration)
7. **Select scopes:** (required permissions)
   - ✅ **`repo`** - This is the only scope needed. It provides:
     - Full control of private repositories
     - Read/write access to public repositories (including `homebrew-tap`)
     - Ability to create, update, and delete files in repositories
     - Access to repository contents, commits, and pull requests
   
   > **Note:** The `repo` scope is sufficient for updating the Homebrew tap. You don't need additional scopes like `workflow`, `write:packages`, etc.
8. Click **Generate token** at the bottom
9. **Copy the token immediately** (you won't see it again!)
10. Go to your `cdnscli` repository → **Settings** → **Secrets and variables** → **Actions**
11. Click **New repository secret**
12. **Name:** `HOMEBREW_TAP_TOKEN`
13. **Secret:** paste the token
14. Click **Add secret**

**Alternative path (if Developer settings not visible):**
- Direct link: https://github.com/settings/tokens?type=beta
- Or: https://github.com/settings/tokens (for classic tokens)

## Optional Secrets

Currently, no additional secrets are required. The workflows use:
- `GITHUB_TOKEN` - automatically provided by GitHub Actions
- No external API keys or tokens needed for CI/CD

## How to Add Secrets (if needed in the future)

If you need to add secrets in the future:

1. Go to your repository on GitHub
2. Navigate to **Settings** → **Secrets and variables** → **Actions**
3. Click **New repository secret**
4. Enter:
   - **Name**: The secret name (e.g., `CLOUDFLARE_API_TOKEN`)
   - **Value**: The secret value
5. Click **Add secret**

## Using Secrets in Workflows

Secrets are accessed in workflow files using:

```yaml
env:
  MY_SECRET: ${{ secrets.MY_SECRET }}
```

Or in a step (without exposing the value):

```yaml
- name: Use secret
  run: |
    # Use the secret value - it will be automatically masked in logs
    some-command --api-key "$MY_SECRET"
  env:
    MY_SECRET: ${{ secrets.MY_SECRET }}
```

⚠️ **Important**: Never use `echo` or `print` with secrets, as they will be exposed in logs. GitHub Actions automatically masks secrets in logs, but only if they're not explicitly printed.

## Security Best Practices

- ✅ Never commit secrets to the repository
- ✅ Use GitHub Secrets for sensitive data
- ✅ Use `GITHUB_TOKEN` for repository operations (automatically scoped)
- ✅ Rotate secrets regularly
- ✅ Use least-privilege principle

## Current Status

✅ **No secrets exposed** - The release workflow uses the automatically provided `GITHUB_TOKEN` internally through the GoReleaser action. The token is:
- Automatically masked in logs (GitHub Actions security feature)
- Not explicitly passed through environment variables
- Used securely by the action itself

The workflow has `contents: write` permission which allows it to create releases without exposing any tokens.

