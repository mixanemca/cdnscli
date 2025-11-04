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

## Usage

```bash
cdnscli help
```

#### Receiving a token

Login to CloudFlare [dash](https://dash.cloudflare.com/login).  
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
    credentials:
      api-token: your-cloudflare-api-token-here
    # Alternative authentication (api-key + email):
    # credentials:
    #   api-key: your-api-key
    #   email: your-email@example.com
```

#### Multiple Providers

You can configure multiple providers of the same type (e.g., multiple CloudFlare accounts) by giving them different names:

```yaml
default-provider: cf-production
client-timeout: 10s
output-format: text
debug: false

providers:
  cf-production:
    type: cloudflare
    credentials:
      api-token: production-account-token
  cf-staging:
    type: cloudflare
    credentials:
      api-token: staging-account-token
  cf-personal:
    type: cloudflare
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

Add or change resource records

```bash
cdnscli rr add -t A -n www -z example.com -c 192.0.2.2
cdnscli rr change --name example.com --zone example.com --type SOA --content "ns1.example.com. admins.example.com. 1970010100 1800 900 604800 86400"
```

Delete resource record

```bash
cdnscli rr del -t A -n www -z example.com
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