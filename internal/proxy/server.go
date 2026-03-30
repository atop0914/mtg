package proxy

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

// Config holds proxy server configuration
type Config struct {
	BindAddr      string
	Secret        string
	Domain        string // Fronting domain for domain fronting
	MaxConns      int
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	BufferSize    int
}

// Server represents the MTG proxy server
type Server struct {
	cfg    Config
	srv    *net.TCPListener
	conns  map[string]*ClientConn
	mu     sync.RWMutex
	ctx    context.Context
	cancel context.CancelFunc
}

// ClientConn represents an active client connection
type ClientConn struct {
	ID        string
	RemoteAddr net.Addr
	Created   time.Time
	AuthKey   []byte
}

// NewServer creates a new proxy server
func NewServer(cfg Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	return &Server{
		cfg:   cfg,
		conns: make(map[string]*ClientConn),
		ctx:   ctx,
		cancel: cancel,
	}
}

// Start starts the proxy server
func (s *Server) Start() error {
	addr, err := net.ResolveTCPAddr("tcp", s.cfg.BindAddr)
	if err != nil {
		return fmt.Errorf("resolve address: %w", err)
	}

	s.srv, err = net.ListenTCP("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen: %w", err)
	}

	fmt.Printf("Proxy server listening on %s\n", s.cfg.BindAddr)

	go s.acceptLoop()
	return nil
}

// acceptLoop accepts new connections
func (s *Server) acceptLoop() {
	for {
		conn, err := s.srv.AcceptTCP()
		if err != nil {
			select {
			case <-s.ctx.Done():
				return
			default:
				fmt.Printf("Accept error: %v\n", err)
				continue
			}
		}

		go s.handleConn(conn)
	}
}

// handleConn handles a single client connection
func (s *Server) handleConn(conn *net.TCPConn) {
	defer conn.Close()

	// Set timeouts
	conn.SetReadDeadline(time.Now().Add(s.cfg.ReadTimeout))
	conn.SetWriteDeadline(time.Now().Add(s.cfg.WriteTimeout))

	// Read secret
	secret := make([]byte, 32)
	n, err := io.ReadFull(conn, secret)
	if err != nil || n != 32 {
		fmt.Println("Failed to read secret")
		return
	}

	// Validate secret
	if hex.EncodeToString(secret) != s.cfg.Secret {
		fmt.Println("Invalid secret")
		return
	}

	fmt.Println("Client connected:", conn.RemoteAddr())
}

// Stop stops the proxy server
func (s *Server) Stop() {
	s.cancel()
	if s.srv != nil {
		s.srv.Close()
	}
}

// Stats returns connection statistics
func (s *Server) Stats() (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.conns), nil
}

var ErrInvalidSecret = errors.New("invalid secret")
var ErrMaxConnsReached = errors.New("maximum connections reached")
