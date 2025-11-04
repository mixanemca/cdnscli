# GitHub Secrets Configuration

This document describes the GitHub Secrets required for CI/CD workflows.

## Required Secrets

### For Release Workflow (`.github/workflows/release.yml`)

**`GITHUB_TOKEN`** - ✅ **Automatically provided and used** (no action needed)

This token is automatically provided by GitHub Actions and is used internally by the GoReleaser action. It has the necessary permissions to create releases and upload assets. No manual configuration or explicit passing is needed - the action handles it securely.

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

