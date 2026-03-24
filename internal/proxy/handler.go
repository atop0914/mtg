package proxy

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"time"
)

// Handler handles bidirectional data forwarding
type Handler struct {
	server *Server
}

// NewHandler creates a new connection handler
func NewHandler(server *Server) *Handler {
	return &Handler{server: server}
}

// HandleConnection handles the full proxy connection lifecycle
func (h *Handler) HandleConnection(clientConn *net.TCPConn) error {
	defer clientConn.Close()

	// Read and validate secret
	if err := h.validateSecret(clientConn); err != nil {
		return fmt.Errorf("secret validation: %w", err)
	}

	// Connect to target (simplified - normally from client request)
	targetConn, err := net.DialTCP("tcp", nil, &net.TCPAddr{
		IP:   []byte{127, 0, 0, 1},
		Port: 443,
	})
	if err != nil {
		return fmt.Errorf("connect to target: %w", err)
	}
	defer targetConn.Close()

	// Bidirectional forwarding
	go func() {
		h.forward(clientConn, targetConn)
	}()
	go func() {
		h.forward(targetConn, clientConn)
	}()

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

	// TODO: implement actual secret validation
	return nil
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
