# cfdnscli

> **Note**: This tool is under active development.

Work with CloudFlare DNS easily from CLI!

## What is it?

This utility will allow you to work with CloudFlare DNS from the CLI:
- add, change and delete zones and resource records
- get more information
- search
- export and import resource record sets

## Usage

```bash
cfdnscli help
```

#### Receiving a token

Login to CloudFlare [dash](https://dash.cloudflare.com/login).  
Go to `My Account` -> `API Tokens` and create a new token.

## Examples

Add or change resource records

```bash
cfdnscli rr add -t A -n www -z example.com -c 192.0.2.2
cfdnscli rr change --name example.com --zone example.com --type SOA --content "ns1.example.com. admins.example.com. 1970010100 1800 900 604800 86400"
```

Delete resourse record

```bash
cfdnscli rr del -t A -n www -z example.com
```

## Instalation

### Go

```bash
go install github.com/mixanemca/cfdnscli@latest
```

### Build

```bash
git clone https://github.com/mixanemca/cfdnscli.git
cd cfdnscli
make
make install
```

## Testing

```bash
make test
```

## License

[Apache 2.0](https://github.com/mixanemca/cfdnscli/raw/main/LICENSE)
