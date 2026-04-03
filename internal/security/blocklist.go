package security

import (
	"fmt"
	"net"
	"sync"
	"time"
)

// Blocklist manages IP blocklist
type Blocklist struct {
	ips     map[string]time.Time
	domains map[string]time.Time
	mu      sync.RWMutex
}

// NewBlocklist creates a new blocklist
func NewBlocklist() *Blocklist {
	return &Blocklist{
		ips:     make(map[string]time.Time),
		domains: make(map[string]time.Time),
	}
}

// AddIP adds an IP to blocklist
func (b *Blocklist) AddIP(ip string, duration time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.ips[ip] = time.Now().Add(duration)
}

// RemoveIP removes an IP from blocklist
func (b *Blocklist) RemoveIP(ip string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.ips, ip)
}

// IsBlocked checks if an IP is blocked
func (b *Blocklist) IsBlocked(ip string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if expiry, ok := b.ips[ip]; ok {
		if time.Now().After(expiry) {
			return false
		}
		return true
	}
	return false
}

// AddDomain adds a domain to blocklist
func (b *Blocklist) AddDomain(domain string, duration time.Duration) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.domains[domain] = time.Now().Add(duration)
}

// IsDomainBlocked checks if a domain is blocked
func (b *Blocklist) IsDomainBlocked(domain string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()

	if expiry, ok := b.domains[domain]; ok {
		if time.Now().After(expiry) {
			return false
		}
		return true
	}
	return false
}

// Cleanup removes expired entries
func (b *Blocklist) Cleanup() {
	b.mu.Lock()
	defer b.mu.Unlock()

	now := time.Now()
	for ip, expiry := range b.ips {
		if now.After(expiry) {
			delete(b.ips, ip)
		}
	}
	for domain, expiry := range b.domains {
		if now.After(expiry) {
			delete(b.domains, domain)
		}
	}
}

// BlockedError is returned when IP is blocked
type BlockedError struct {
	IP   string
	Until time.Time
}

func (e *BlockedError) Error() string {
	return fmt.Sprintf("IP %s blocked until %s", e.IP, e.Until)
}
