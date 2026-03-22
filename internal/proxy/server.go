package proxy

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/atop0914/mtg/internal/config"
	"github.com/rs/zerolog"
)

const (
	// ReadTimeout is the maximum duration for reading the entire request
	ReadTimeout = 10 * time.Second
	// WriteTimeout is the maximum duration for writing the response
	WriteTimeout = 10 * time.Second
	// IdleTimeout is the maximum duration for waiting for the next request
	IdleTimeout = 60 * time.Second
)

// Server represents the MTPROTO proxy server
type Server struct {
	cfg    *config.Config
	logger zerolog.Logger
	server *http.Server
}

// NewServer creates a new proxy server instance
func NewServer(cfg *config.Config, logger zerolog.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
	}
}

// Start starts the MTPROTO proxy server
func (s *Server) Start() error {
	addr := s.cfg.BindTo

	tlsConfig := s.buildTLSConfig()

	s.server = &http.Server{
		Addr:         addr,
		TLSConfig:    tlsConfig,
		Handler:      s.handleRequest(),
		ReadTimeout:  ReadTimeout,
		WriteTimeout: WriteTimeout,
		IdleTimeout:  IdleTimeout,
	}

	// Start graceful shutdown handler
	go s.handleShutdown()

	s.logger.Info().Str("bind", addr).Msg("Starting MTPROTO proxy server")

	if tlsConfig != nil {
		return s.server.ListenAndServeTLS("", "")
	}
	return s.server.ListenAndServe()
}

// buildTLSConfig builds TLS configuration from config
func (s *Server) buildTLSConfig() *tls.Config {
	if s.cfg.TLS.CertFile == "" || s.cfg.TLS.KeyFile == "" {
		return nil
	}

	cert, err := tls.LoadX509KeyPair(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile)
	if err != nil {
		s.logger.Fatal().Err(err).Msg("Failed to load TLS certificates")
		return nil
	}

	return &tls.Config{
		Certificates: []tls.Certificate{cert},
		MinVersion:   tls.VersionTLS12,
	}
}

// handleRequest returns the HTTP handler for MTPROTO connections
func (s *Server) handleRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug().
			Str("remote", r.RemoteAddr).
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("Received request")

		// For MTPROTO, we handle the connection at TCP level
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to hijack connection")
			return
		}
		defer conn.Close()

		s.handleMTProtoConnection(conn)
	})
}

// handleMTProtoConnection handles the MTPROTO protocol connection
func (s *Server) handleMTProtoConnection(conn net.Conn) {
	defer conn.Close()

	s.logger.Debug().Msg("New MTPROTO connection")

	// TODO: Implement MTPROTO protocol handling
	// 1. Read and validate secret
	// 2. Perform MTPROTO handshake
	// 3. Handle Telegram protocol messages
}

// handleShutdown handles graceful server shutdown
func (s *Server) handleShutdown() {
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	s.logger.Info().Msg("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := s.server.Shutdown(ctx); err != nil {
		s.logger.Error().Err(err).Msg("Server shutdown failed")
	}
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	if s.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		return s.server.Shutdown(ctx)
	}
	return nil
}
