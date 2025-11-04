# cdnscli

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

## Examples

Add or change resource records

```bash
cdnscli rr add -t A -n www -z example.com -c 192.0.2.2
cdnscli rr change --name example.com --zone example.com --type SOA --content "ns1.example.com. admins.example.com. 1970010100 1800 900 604800 86400"
```

Delete resourse record

```bash
cdnscli rr del -t A -n www -z example.com
```

## Instalation

### Go

```bash
go install github.com/mixanemca/cdnscli@latest
```

### Build

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
