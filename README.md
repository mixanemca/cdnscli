# cdnscli

[![Go Version](https://img.shields.io/badge/go-1.24+-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg?style=flat-square)](LICENSE)
[![Build Status](https://img.shields.io/github/actions/workflow/status/mixanemca/cdnscli/ci.yml?branch=main&style=flat-square&label=build)](https://github.com/mixanemca/cdnscli/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/mixanemca/cdnscli?style=flat-square)](https://goreportcard.com/report/github.com/mixanemca/cdnscli)

> **Note**: This tool is under active development.

Cloud DNS CLI - manage DNS records across multiple providers!

## What is it?

`cdnscli` is a powerful cross-platform utility for managing DNS records and zones across multiple DNS providers, written in Go.
It provides convenient tools for both task automation and manual management through the terminal.

The utility supports two modes of operation:

- Classic CLI: Perfect for use in scripts, automation, and executing standalone commands.
- TUI (Text User Interface): An interactive text-based interface for more convenient management.

#### Key Features

- Manage zones and resource records: add, modify, and delete them.
- Retrieve detailed information about zones and records.
- Search DNS records by various parameters.
- Export and import resource record sets in a convenient format.
- Easy-to-use interface, ideal for administrators and developers.

## Supported Providers

| Provider | Authentication | Features | Status |
|----------|---------------|----------|--------|
| [Cloudflare](https://www.cloudflare.com/) | API Token<br>API Key + Email | ✅ Add/Update/Delete records<br>✅ List zones and records<br>✅ Search records<br>✅ Multiple accounts support<br>✅ Custom display names | ✅ Fully Supported |

> **Note**: More providers are planned for future releases. If you'd like to see support for a specific provider, please [open an issue](https://github.com/mixanemca/cdnscli/issues).

## Usage

```bash
cdnscli help
```

#### Receiving a token

Login to Cloudflare [dash](https://dash.cloudflare.com/login).  
Go to `My Account` -> `API Tokens` and create a new token.

## Configuration

Create a configuration file `~/.cdnscli.yaml` in your home directory:

> **Tip**: You can copy the example configuration file from the repository:
> ```bash
> cp cdnscli.yaml.example ~/.cdnscli.yaml
> ```

Then edit the file with your credentials:

```yaml
default-provider: cloudflare
client-timeout: 10s
output-format: text
debug: false

providers:
  cloudflare:
    type: cloudflare
    # display-name: Cloudflare  # Optional: custom display name for the provider (defaults to "Cloudflare" for cloudflare type)
    credentials:
      api-token: your-cloudflare-api-token-here
    # Alternative authentication (api-key + email):
    # credentials:
    #   api-key: your-api-key
    #   email: your-email@example.com
```

#### Multiple Providers

You can configure multiple providers of the same type (e.g., multiple Cloudflare accounts) by giving them different names:

```yaml
default-provider: cf-production
client-timeout: 10s
output-format: text
debug: false

providers:
  cf-production:
    type: cloudflare
    display-name: Cloudflare Production  # Optional: custom display name
    credentials:
      api-token: production-account-token
  cf-staging:
    type: cloudflare
    display-name: Cloudflare Staging  # Optional: custom display name
    credentials:
      api-token: staging-account-token
  cf-personal:
    type: cloudflare
    # display-name is optional - if not specified, defaults to "Cloudflare"
    credentials:
      api-token: personal-account-token
```

To switch between providers, change the `default-provider` value in the config file, or use the default provider specified in the config.

You can also use environment variables instead of a config file:

```bash
export CLOUDFLARE_API_TOKEN=your-cloudflare-api-token-here
# or
export CDNSCLI_PROVIDERS_CLOUDFLARE_CREDENTIALS_API_TOKEN=your-cloudflare-api-token-here
# Note: environment variables use underscores, not dashes
```

#### Multiple Providers with Environment Variables

For multiple providers, you can use environment variables with the provider name:

```bash
# Set default provider
export CDNSCLI_DEFAULT_PROVIDER=cf-production

# Configure each provider
export CDNSCLI_PROVIDERS_CF_PRODUCTION_CREDENTIALS_API_TOKEN=production-account-token
export CDNSCLI_PROVIDERS_CF_STAGING_CREDENTIALS_API_TOKEN=staging-account-token
export CDNSCLI_PROVIDERS_CF_PERSONAL_CREDENTIALS_API_TOKEN=personal-account-token
```

Environment variables follow the pattern: `CDNSCLI_PROVIDERS_<PROVIDER_NAME>_CREDENTIALS_<CREDENTIAL_KEY>`

## Examples

### Managing Zones

List all zones:
```bash
cdnscli zone list
```

List zones with JSON output:
```bash
cdnscli zone list --output-format json
```

Get information about a specific zone:
```bash
cdnscli zone info example.com
```

### Managing DNS Records

Add a new A record:
```bash
cdnscli rr add -t A -n www -z example.com -c 192.0.2.2
```

Add a CNAME record:
```bash
cdnscli rr add -t CNAME -n blog -z example.com -c example.github.io
```

Add an MX record with priority:
```bash
cdnscli rr add -t MX -n example.com -z example.com -c "10 mail.example.com"
```

Update an existing record:
```bash
cdnscli rr update -t A -n www -z example.com -c 192.0.2.3
```

Change a record (with full SOA example):
```bash
cdnscli rr change --name example.com --zone example.com --type SOA --content "ns1.example.com. admins.example.com. 1970010100 1800 900 604800 86400"
```

Delete a record:
```bash
cdnscli rr del -t A -n www -z example.com
```

List all records in a zone:
```bash
cdnscli rr list -z example.com
```

List records with JSON output:
```bash
cdnscli rr list -z example.com --output-format json
```

Get detailed information about a specific record:
```bash
cdnscli rr info -t A -n www -z example.com
```

### Searching Records

Search for records by name:
```bash
cdnscli search -n www
```

Search for records by type:
```bash
cdnscli search -t A
```

Search for records in a specific zone:
```bash
cdnscli search -z example.com
```

### Using Different Providers

If you have multiple providers configured, switch between them by changing `default-provider` in your config file, or specify the provider in commands (if supported).

### Output Formats

Use JSON output for scripting:
```bash
cdnscli zone list --output-format json | jq '.[] | select(.name == "example.com")'
```

Use text output (default):
```bash
cdnscli zone list --output-format text
```

## Installation

### Quick Install (Recommended)

Download the latest release from [GitHub Releases](https://github.com/mixanemca/cdnscli/releases) and extract the binary for your platform.

**macOS (Intel):**
```bash
curl -L https://github.com/mixanemca/cdnscli/releases/latest/download/cdnscli_Darwin_x86_64.tar.gz | tar -xz
sudo mv cdnscli /usr/local/bin/
```

**macOS (Apple Silicon):**
```bash
curl -L https://github.com/mixanemca/cdnscli/releases/latest/download/cdnscli_Darwin_arm64.tar.gz | tar -xz
sudo mv cdnscli /usr/local/bin/
```

**Linux:**
```bash
# For amd64
curl -L https://github.com/mixanemca/cdnscli/releases/latest/download/cdnscli_Linux_x86_64.tar.gz | tar -xz
# For arm64
curl -L https://github.com/mixanemca/cdnscli/releases/latest/download/cdnscli_Linux_arm64.tar.gz | tar -xz
sudo mv cdnscli /usr/local/bin/
```

**Windows:**
Download the appropriate `cdnscli_Windows_x86_64.zip` or `cdnscli_Windows_arm64.zip` from the releases page and extract the `cdnscli.exe` file.

### Homebrew (macOS/Linux)

Install using Homebrew:

```bash
brew install mixanemca/tap/cdnscli
```

### Go Install

Install directly from source:

```bash
go install github.com/mixanemca/cdnscli@latest
```

> **Note**: You can also install a specific version by replacing `@latest` with `@v0.99.0` (or any other version tag).

### Build from Source

```bash
git clone https://github.com/mixanemca/cdnscli.git
cd cdnscli
make
make install
```

## Testing (WIP)

```bash
make test
```

## License

[Apache 2.0](https://github.com/mixanemca/cdnscli/raw/main/LICENSE)