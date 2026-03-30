package security

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
)

// SOCKS5 represents a SOCKS5 proxy
type SOCKS5 struct {
	// AuthMethods supported
	AuthMethods []byte
	// Commands supported
	Commands []byte
}

// NewSOCKS5 creates a new SOCKS5 handler
func NewSOCKS5() *SOCKS5 {
	return &SOCKS5{
		AuthMethods: []byte{0x00, 0x02}, // No auth, GSSAPI
		Commands:   []byte{0x01, 0x02}, // Connect, UDP Associate
	}
}

// Handle handles a SOCKS5 connection
func (s *SOCKS5) Handle(conn net.Conn) error {
	// Greeting
	if err := s.greeting(conn); err != nil {
		return fmt.Errorf("greeting: %w", err)
	}

	// Request
	return s.request(conn)
}

// greeting performs the SOCKS5 greeting
func (s *SOCKS5) greeting(conn net.Conn) error {
	// Read greeting
	header := make([]byte, 2)
	if _, err := io.ReadFull(conn, header); err != nil {
		return err
	}

	if header[0] != 0x05 {
		return fmt.Errorf("unsupported version: %d", header[0])
	}

	// Respond with auth method
	conn.Write([]byte{0x05, 0x00}) // No auth
	return nil
}

// request handles the SOCKS5 request
func (s *SOCKS5) request(conn net.Conn) error {
	// Read request
	req := make([]byte, 4)
	if _, err := io.ReadFull(conn, req); err != nil {
		return err
	}

	if req[0] != 0x05 {
		return fmt.Errorf("unsupported version: %d", req[0])
	}

	cmd := req[1]
	addrType := req[3]

	var addr string
	switch addrType {
	case 0x01: // IPv4
		ip := make([]byte, 4)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return err
		}
		addr = net.IP(ip).String()
	case 0x03: // Domain
		length := make([]byte, 1)
		if _, err := io.ReadFull(conn, length); err != nil {
			return err
		}
		domain := make([]byte, length[0])
		if _, err := io.ReadFull(conn, domain); err != nil {
			return err
		}
		addr = string(domain)
	case 0x04: // IPv6
		ip := make([]byte, 16)
		if _, err := io.ReadFull(conn, ip); err != nil {
			return err
		}
		addr = net.IP(ip).String()
	default:
		return fmt.Errorf("unsupported address type: %d", addrType)
	}

	// Read port
	port := make([]byte, 2)
	if _, err := io.ReadFull(conn, port); err != nil {
		return err
	}

	dstPort := binary.BigEndian.Uint16(port)

	// Handle command
	switch cmd {
	case 0x01: // Connect
		return s.connect(conn, addr, dstPort)
	default:
		conn.Write([]byte{0x05, 0x07, 0x00, 0x01}) // Command not supported
		return fmt.Errorf("unsupported command: %d", cmd)
	}
}

// connect establishes a connection to the target
func (s *SOCKS5) connect(conn net.Conn, addr string, port uint16) error {
	dstAddr := fmt.Sprintf("%s:%d", addr, port)

	dialer := &net.Dialer{}
	dialConn, err := dialer.Dial("tcp", dstAddr)
	if err != nil {
		conn.Write([]byte{0x05, 0x01, 0x00, 0x01}) // General failure
		return err
	}
	defer dialConn.Close()

	// Send success response
	conn.Write([]byte{0x05, 0x00, 0x00, 0x01})
	conn.Write([]byte{0x00, 0x00, 0x00, 0x00})
	conn.Write([]byte{0x00, 0x00}) // Port 0 (not used)

	// Bidirectional copy
	go io.Copy(dialConn, conn)
	io.Copy(conn, dialConn)

	return nil
}

// Client connects to a SOCKS5 proxy
type SOCKS5Client struct {
	ProxyAddr string
}

// Dial connects through SOCKS5 proxy
func (c *SOCKS5Client) Dial(network, addr string) (net.Conn, error) {
	conn, err := net.Dial("tcp", c.ProxyAddr)
	if err != nil {
		return nil, err
	}

	// Greeting
	conn.Write([]byte{0x05, 0x01, 0x00}) // Version 5, 1 method, no auth
	resp := make([]byte, 2)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return nil, err
	}

	if resp[1] != 0x00 {
		return nil, fmt.Errorf("auth failed")
	}

	// Request
	var req []byte
	if ip := net.ParseIP(addr); ip != nil {
		if ip4 := ip.To4(); ip4 != nil {
			req = []byte{0x05, 0x01, 0x00, 0x01}
			req = append(req, ip4...)
		} else {
			req = []byte{0x05, 0x01, 0x00, 0x04}
			req = append(req, ip.To16()...)
		}
	} else {
		req = []byte{0x05, 0x01, 0x00, 0x03, byte(len(addr))}
		req = append(req, addr...)
	}

	port := make([]byte, 2)
	binary.BigEndian.PutUint16(port, 0) // Will be fixed
	// Actually parse port from addr
	fmt.Sscanf(addr, "%*s:%d", &port)

	req = append(req, port...)
	conn.Write(req)

	// Read response
	resp = make([]byte, 10)
	if _, err := io.ReadFull(conn, resp); err != nil {
		return nil, err
	}

	if resp[1] != 0x00 {
		return nil, fmt.Errorf("connection failed: %d", resp[1])
	}

	return conn, nil
}
