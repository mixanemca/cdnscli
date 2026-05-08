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

| Provider | Authentication | Status |
|----------|---------------|--------|
| [Cloudflare](https://www.cloudflare.com/) | API Token or API Key + Email | ✅ Fully Supported |
| [Reg.ru](https://www.reg.ru/) | Username + Password | ✅ Fully Supported |
| [Namecheap](https://www.namecheap.com/) | API User + API Key | ✅ Fully Supported |

All providers support: add / update / delete records, list zones and records, search records, multiple accounts, custom display names.

## Usage

```bash
cdnscli help
```

## Configuration

Create a configuration file `~/.cdnscli.yaml` in your home directory:

> **Tip**: You can copy the example configuration file from the repository:
> ```bash
> cp cdnscli.yaml.example ~/.cdnscli.yaml
> ```

### Cloudflare

Login to [Cloudflare dashboard](https://dash.cloudflare.com/login), go to `My Account` → `API Tokens` and create a new token.

```yaml
default-provider: cloudflare
client-timeout: 10s
output-format: text

providers:
  cloudflare:
    type: cloudflare
    credentials:
      api_token: your-cloudflare-api-token-here
    # Alternative authentication (api_key + email):
    # credentials:
    #   api_key: your-api-key
    #   email: your-email@example.com
```

### Reg.ru

Login to [Reg.ru](https://www.reg.ru/), go to account settings and enable API access. Use your account username and password.

```yaml
default-provider: regru
client-timeout: 10s
output-format: text

providers:
  regru:
    type: regru
    credentials:
      username: your-regru-username
      password: your-regru-password
```

### Namecheap

Log in to [Namecheap](https://www.namecheap.com/), go to `Profile` → `Tools` → `Business & Dev Tools` and enable API access. Whitelist your public IP address in the API settings.

```yaml
default-provider: namecheap
client-timeout: 10s
output-format: text

providers:
  namecheap:
    type: namecheap
    credentials:
      api_user: your-namecheap-username
      api_key: your-namecheap-api-key
      # username: your-namecheap-username  # optional, defaults to api_user
      # sandbox: false                     # optional, use sandbox API for testing
    options:
      # ip_provider_url: https://api.ipify.org  # optional, URL to detect your public IP
      #                                           # response can be plain-text or JSON {"ip":"..."}
```

> **Note**: Namecheap requires your public IP to be whitelisted in the API settings. The tool detects your current IP automatically using `https://api.ipify.org` by default. You can override this with the `ip_provider_url` option.

### Multiple Providers

You can configure multiple providers of the same or different types simultaneously:

```yaml
default-provider: cf-production
client-timeout: 10s
output-format: text

providers:
  cf-production:
    type: cloudflare
    display-name: Cloudflare Production
    credentials:
      api_token: production-account-token
  cf-staging:
    type: cloudflare
    display-name: Cloudflare Staging
    credentials:
      api_token: staging-account-token
  regru:
    type: regru
    credentials:
      username: your-regru-username
      password: your-regru-password
  namecheap:
    type: namecheap
    credentials:
      api_user: your-namecheap-username
      api_key: your-namecheap-api-key
```

To switch between providers, change the `default-provider` value in the config file.

### Environment Variables

You can also use environment variables instead of a config file:

```bash
# Cloudflare
export CDNSCLI_PROVIDERS_CLOUDFLARE_CREDENTIALS_API_TOKEN=your-token

# Reg.ru
export CDNSCLI_PROVIDERS_REGRU_CREDENTIALS_USERNAME=your-username
export CDNSCLI_PROVIDERS_REGRU_CREDENTIALS_PASSWORD=your-password

# Namecheap
export CDNSCLI_PROVIDERS_NAMECHEAP_CREDENTIALS_API_USER=your-username
export CDNSCLI_PROVIDERS_NAMECHEAP_CREDENTIALS_API_KEY=your-api-key
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

```bash
brew install mixanemca/tap/cdnscli
```

### Go Install

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

## Testing

```bash
make test
```

## License

[Apache 2.0](https://github.com/mixanemca/cdnscli/raw/main/LICENSE)
