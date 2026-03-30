package proxy

import (
	"crypto/tls"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/atop0914/mtg/internal/faketls"
)

const (
	// MTPROTO magic byte for client connection
	magicByte = 0xEE

	telegramPort = 443
)

// Default Telegram DC addresses (fallback if DNS fails)
var telegramDCs = []string{
	"149.154.175.50", // DC1
	"149.154.167.51", // DC2
	"149.154.175.100", // DC3
	"91.108.56.130",  // DC4
	"91.108.56.190",  // DC5
}

// Handler handles bidirectional data forwarding
type Handler struct {
	server   *Server
	faketls  *faketls.FakeTLS
}

// NewHandler creates a new connection handler
func NewHandler(server *Server) *Handler {
	return &Handler{
		server:  server,
		faketls: faketls.NewFakeTLS(),
	}
}

// HandleConnection handles the full proxy connection lifecycle
func (h *Handler) HandleConnection(clientConn *net.TCPConn) error {
	defer clientConn.Close()

	// Read and validate secret
	if err := h.validateSecret(clientConn); err != nil {
		fmt.Printf("Secret validation failed: %v\n", err)
		return err
	}

	// Read MTPROTO connection header
	targetIP, targetPort, err := h.readConnectionHeader(clientConn)
	if err != nil {
		fmt.Printf("Failed to read connection header: %v\n", err)
		return err
	}

	// Resolve target (use Telegram DC if not specified)
	if targetIP == nil || len(targetIP) == 0 {
		targetIP = net.ParseIP(telegramDCs[0])
	}
	if targetPort == 0 {
		targetPort = telegramPort
	}

	// Connect to Telegram through domain fronting + FakeTLS
	targetConn, err := h.connectToTelegram(targetIP, targetPort)
	if err != nil {
		fmt.Printf("Failed to connect to Telegram: %v\n", err)
		return err
	}
	defer targetConn.Close()

	fmt.Printf("Proxying to %s:%d\n", targetIP, targetPort)

	// Bidirectional forwarding
	go h.forward(clientConn, targetConn)
	go h.forward(targetConn, clientConn)

	// Wait until closed
	buf := make([]byte, 1)
	clientConn.Read(buf)

	return nil
}

// validateSecret validates the client secret
func (h *Handler) validateSecret(conn *net.TCPConn) error {
	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	secret := make([]byte, 32)
	_, err := io.ReadFull(conn, secret)
	if err != nil {
		return err
	}

	// Compare with server's configured secret
	provided := hex.EncodeToString(secret)
	if provided != h.server.cfg.Secret {
		return ErrInvalidSecret
	}

	return nil
}

// readConnectionHeader reads the MTPROTO connection header to get target
func (h *Handler) readConnectionHeader(conn *net.TCPConn) ([]byte, uint16, error) {
	// MTPROTO Proxy Protocol:
	// First byte should be 0xEE (magic)
	// Then address Type (1 byte): 1=IPv4, 2=IPv6, 3=domain
	// Then address data
	
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	
	// Read magic byte
	magic := make([]byte, 1)
	if _, err := io.ReadFull(conn, magic); err != nil {
		return nil, 0, fmt.Errorf("read magic: %w", err)
	}
	
	if magic[0] != magicByte {
		// Try direct connection (no MTPROTO header)
		return nil, 0, nil
	}

	// Read address type
	addrType := make([]byte, 1)
	if _, err := io.ReadFull(conn, addrType); err != nil {
		return nil, 0, fmt.Errorf("read addr type: %w", err)
	}

	switch addrType[0] {
	case 1: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return nil, 0, err
		}
		port := make([]byte, 2)
		if _, err := io.ReadFull(conn, port); err != nil {
			return nil, 0, err
		}
		return ip, binary.BigEndian.Uint16(port), nil

	case 2: // IPv6 (not commonly used)
		ip := make([]byte, 16)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return nil, 0, err
		}
		port := make([]byte, 2)
		if _, err := io.ReadFull(conn, port); err != nil {
			return nil, 0, err
		}
		return ip, binary.BigEndian.Uint16(port), nil

	case 3: // Domain name
		domainLen := make([]byte, 1)
		if _, err := io.ReadFull(conn, domainLen); err != nil {
			return nil, 0, err
		}
		domain := make([]byte, domainLen[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return nil, 0, err
		}
		port := make([]byte, 2)
		if _, err := io.ReadFull(conn, port); err != nil {
			return nil, 0, err
		}
		
		// Resolve domain
		ips, err := net.LookupIP(string(domain))
		if err != nil || len(ips) == 0 {
			return nil, 0, fmt.Errorf("domain resolution failed: %w", err)
		}
		return ips[0].To4(), binary.BigEndian.Uint16(port), nil

	default:
		return nil, 0, fmt.Errorf("unknown address type: %d", addrType[0])
	}
}

// connectToTelegram establishes an obfuscated connection to Telegram
func (h *Handler) connectToTelegram(targetIP net.IP, targetPort uint16) (net.Conn, error) {
	// Get the fronting domain from config (default to cloudflare.com)
	frontDomain := h.server.cfg.Domain
	if frontDomain == "" {
		frontDomain = "www.cloudflare.com"
	}

	// Generate FakeTLS ClientHello with front domain as SNI
	hello, err := h.faketls.ClientHello()
	if err != nil {
		return nil, fmt.Errorf("generate client hello: %w", err)
	}

	// Connect to front domain
	addr := fmt.Sprintf("%s:443", frontDomain)
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName:         frontDomain,
		InsecureSkipVerify: false,
	})
	if err != nil {
		return nil, fmt.Errorf("dial front domain: %w", err)
	}

	// Send FakeTLS ClientHello
	if _, err := conn.Write(hello); err != nil {
		conn.Close()
		return nil, fmt.Errorf("send client hello: %w", err)
	}

	// Read ServerHello (we don't really need to validate it for proxying)
	conn.SetReadDeadline(time.Now().Add(5 * time.Second))
	serverHello := make([]byte, 1024)
	conn.Read(serverHello)

	// Now we have a FakeTLS connection that looks like HTTPS to the front domain
	// but we're actually proxying to Telegram
	
	return &fakeTLSConn{
		Conn:       conn,
		targetIP:   targetIP,
		targetPort: targetPort,
	}, nil
}

// fakeTLSConn wraps a TLS connection for MTPROTO proxying
type fakeTLSConn struct {
	net.Conn
	targetIP   net.IP
	targetPort uint16
}

// forward forwards data between connections
func (h *Handler) forward(from, to net.Conn) {
	defer from.Close()
	defer to.Close()
	io.Copy(to, from)
}

// ReadFrame reads a length-prefixed frame
func ReadFrame(conn net.Conn) ([]byte, error) {
	// Read 4-byte length prefix
	lengthBuf := make([]byte, 4)
	if _, err := io.ReadFull(conn, lengthBuf); err != nil {
		return nil, err
	}

	length := binary.BigEndian.Uint32(lengthBuf)
	if length > 65536 {
		return nil, fmt.Errorf("frame too large: %d", length)
	}

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		return nil, err
	}

	return data, nil
}

// WriteFrame writes a length-prefixed frame
func WriteFrame(conn net.Conn, data []byte) error {
	length := make([]byte, 4)
	binary.BigEndian.PutUint32(length, uint32(len(data)))

	if _, err := conn.Write(length); err != nil {
		return err
	}

	_, err := conn.Write(data)
	return err
}
