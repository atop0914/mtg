# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/).

## [0.1.0] - 2024-03-XX

### Added
- MTPROTO protocol implementation
- Domain fronting support (Cloudflare, etc.)
- FakeTLS traffic obfuscation
- IP blocklist (FireHOL format)
- SOCKS5 proxy chaining
- Structured logging
- Comprehensive README documentation

### Fixed
- Fix `proxy.NewServer` call signature mismatch
- Implement actual secret validation
- Fix RSA key generation API compatibility
- Fix AES encryption with proper PKCS7 padding
- Fix binary.LittleEndian usage in mtproto package
- Fix faketls ClientHello SNI extension length

### Improved
- Integrate fronting and faketls into proxy handler
- Add Telegram DC fallback addresses
- Better error handling and logging
