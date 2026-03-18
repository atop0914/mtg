# MTG - MTPROTO Telegram Proxy

A lightweight, high-performance MTPROTO proxy for Telegram built in Go.

## Features

- **MTPROTO Protocol**: Native Telegram protocol implementation
- **Domain Fronting**: Fallback to legitimate websites when blocked
- **FakeTLS**: Traffic obfuscation to mimic real TLS connections
- **IP Blocklist**: Native FireHOL blocklist support
- **Proxy Chaining**: SOCKS5 proxy support
- **Single Secret**: Simple yet secure authentication

## Requirements

- Go 1.21 or higher

## Installation

```bash
# Clone the repository
git clone https://github.com/atop0914/mtg.git
cd mtg

# Build
go build -o mtg ./cmd/mtg
```

## Configuration

Copy the sample configuration:

```bash
cp configs/config.yaml config.yaml
```

Edit `config.yaml` and set your secret:

```yaml
secret: "your-secret-here"  # Generate with: openssl rand -hex 16
bind-to: ":443"
domain: "www.cloudflare.com"
```

## Usage

```bash
# Run with config file
./mtg -c config.yaml

# Show version
./mtg -v
```

## Generate Secret

```bash
# Generate a random secret
openssl rand -hex 16
```

## Architecture

See [PRD](../prd/github-trend-mtg.md) for detailed architecture.

## License

MIT License
