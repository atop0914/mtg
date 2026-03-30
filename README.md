# MTG - MTPROTO Telegram Proxy

A lightweight, high-performance MTPROTO proxy for Telegram, written in Go. MTG helps you bypass network restrictions to access Telegram when it's blocked or throttled.

[![][godoc-shield]](https://pkg.go.dev)
[![][go-shield]](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white)
[![][license-shield]](LICENSE)

[godoc-shield]: https://pkg.go.dev/badge/github.com/atop0914/mtg.svg
[go-shield]: https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat-square&logo=go&logoColor=white
[license-shield]: https://img.shields.io/badge/License-MIT-green?style=flat-square

## Overview

MTG implements the [MTPROTO](https://core.telegram.org/mtproto) proxy protocol natively in Go, providing a fast and reliable way to connect to Telegram through restrictive networks. It features multiple obfuscation techniques to evade detection and deep packet inspection (DPI).

## Features

| Feature | Description |
|---------|-------------|
| **MTPROTO Protocol** | Native implementation of Telegram's custom protocol |
| **Domain Fronting** | Masquerade traffic through legitimate websites (Cloudflare, etc.) |
| **FakeTLS** | Obfuscate traffic to appear as regular HTTPS/TLS connections |
| **IP Blocklist** | Built-in FireHOL blocklist support to block known malicious IPs |
| **Proxy Chaining** | Route traffic through SOCKS5 proxies for additional anonymity |
| **Zero Dependencies** | Static binary, no external runtime required |

## How It Works

```
Client <--[MTPROTO+TLS]--> MTG Proxy <--[Telegram]--> Telegram Servers
         \                        \
          \--[Domain Fronting]----/---> Legitimate CDN (Cloudflare, etc.)
```

When domain fronting is enabled, your traffic appears to connect to a legitimate website (e.g., `www.cloudflare.com`) while actually reaching Telegram's servers. This makes it extremely difficult for censors to detect and block.

## Quick Start

### Prerequisites

- Go 1.21 or higher
- Linux/macOS/Windows

### Build

```bash
git clone https://github.com/atop0914/mtg.git
cd mtg
go build -o mtg ./cmd/mtg
```

### Generate a Secret

```bash
openssl rand -hex 16
```

Save the generated secret — you'll need it to configure the proxy and to share with users.

### Configure

```bash
cp configs/config.yaml.example config.yaml
```

Edit `config.yaml`:

```yaml
# Required: Your secret key
secret: "YOUR_SECRET_HERE"

# Required: Address to listen on
bind-to: ":443"

# Domain for traffic fronting (must be reachable)
domain: "www.cloudflare.com"

# Optional: TLS certificate (for real HTTPS)
# tls:
#   cert-file: "cert.pem"
#   key-file: "key.pem"

# Optional: Blocklist to filter malicious IPs
# blocklist: "https://raw.githubusercontent.com/firehol/blocklist-ipsets/master/firehol_level1.netset"

# Optional: Chain through a SOCKS5 proxy
# socks5: "127.0.0.1:1080"
```

### Run

```bash
./mtg -c config.yaml
```

### Docker

```bash
# Build
docker build -t mtg .

# Run
docker run -d -p 443:443 -v $(pwd)/config.yaml:/config.yaml mtg -c /config.yaml
```

## Usage with Telegram

Once your proxy is running, share the connection link with users:

```
tg://proxy?server=YOUR_SERVER_IP&port=443&secret=YOUR_SECRET
```

Or as a plain link:

```
https://t.me/proxy?server=YOUR_SERVER_IP&port=443&secret=YOUR_SECRET
```

Users can add this proxy directly in Telegram Settings > Data & Storage > Proxy Settings.

## Project Structure

```
mtg/
├── cmd/mtg/              # Application entry point
├── internal/
│   ├── config/           # Configuration parsing (YAML)
│   ├── logging/          # Structured logging
│   ├── proxy/            # Core proxy server implementation
│   ├── mtproto/          # MTPROTO protocol handshake & framing
│   ├── fronting/         # Domain fronting (HTTP/HTTPS disguise)
│   ├── faketls/          # FakeTLS traffic obfuscation
│   └── security/         # IP blocklist management
├── configs/
│   └── config.yaml.example  # Sample configuration
├── go.mod
├── go.sum
└── README.md
```

## Configuration Reference

| Field | Required | Description |
|-------|----------|-------------|
| `secret` | Yes | 32-byte secret in hex (generate with `openssl rand -hex 16`) |
| `bind-to` | Yes | Listen address (e.g., `:443` or `0.0.0.0:443`) |
| `domain` | Yes | Domain for fronting (must resolve to a CDN IP) |
| `tls.cert-file` | No | TLS certificate path |
| `tls.key-file` | No | TLS private key path |
| `blocklist` | No | FireHOL-compatible blocklist URL |
| `socks5` | No | SOCKS5 proxy address for chaining |

## Deployment

### Production Checklist

- [ ] Use a strong secret (`openssl rand -hex 16`)
- [ ] Choose a reliable fronting domain (Cloudflare, Google, etc.)
- [ ] Enable TLS certificate for real HTTPS (recommended)
- [ ] Set up blocklist to filter malicious IPs
- [ ] Configure firewall to allow only needed ports
- [ ] Use systemd for process management
- [ ] Set up log rotation

### Systemd Service Example

```ini
[Unit]
Description=MTG MTPROTO Proxy
After=network.target

[Service]
Type=simple
User=nobody
ExecStart=/opt/mtg/mtg -c /opt/mtg/config.yaml
Restart=on-failure
RestartSec=5

[Install]
WantedBy=multi-user.target
```

### Reverse Proxy (Optional)

For additional layer, put behind nginx with real TLS:

```nginx
server {
    listen 443 ssl;
    server_name your-proxy.domain.com;

    ssl_certificate /path/to/cert.pem;
    ssl_certificate_key /path/to/key.pem;

    location / {
        proxy_pass http://127.0.0.1:8443;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection "upgrade";
    }
}
```

## Security Considerations

- **Use FakeTLS**: Enable domain fronting to make traffic look like regular HTTPS
- **Rotate secrets**: Periodically change your proxy secret
- **Use blocklists**: Enable the FireHOL blocklist to reject known malicious IPs
- **Firewall**: Restrict access to only the ports you need

## Performance

MTG is designed for high throughput and low latency:

- Handles thousands of concurrent connections
- Minimal memory footprint (< 20MB baseline)
- Efficient Goroutine-based concurrency

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- Telegram for the [MTPROTO protocol](https://core.telegram.org/mtproto) specification
- [Official MTPROTO Proxy Documentation](https://core.telegram.org/mtproto/mtproto-proxy)
- Inspired by various open source MTPROTO proxy implementations
