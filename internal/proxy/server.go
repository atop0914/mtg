package proxy

import (
	"crypto/tls"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/atop0914/mtg/internal/config"
	"github.com/rs/zerolog"
)

type Server struct {
	cfg    *config.Config
	logger zerolog.Logger
	server *http.Server
}

func NewServer(cfg *config.Config, logger zerolog.Logger) *Server {
	return &Server{
		cfg:    cfg,
		logger: logger,
	}
}

func (s *Server) Start() error {
	addr := s.cfg.BindTo
	
	// Create TLS config if certs are provided
	tlsConfig := &tls.Config{}
	if s.cfg.TLS.CertFile != "" && s.cfg.TLS.KeyFile != "" {
		cert, err := tls.LoadX509KeyPair(s.cfg.TLS.CertFile, s.cfg.TLS.KeyFile)
		if err != nil {
			return fmt.Errorf("failed to load TLS certificates: %w", err)
		}
		tlsConfig.Certificates = []tls.Certificate{cert}
	}

	// Create HTTP server with MTPROTO handler
	s.server = &http.Server{
		Addr:      addr,
		TLSConfig: tlsConfig,
		Handler:   s.handleRequest(),
	}

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		s.logger.Info().Msg("Shutting down server...")
		s.server.Close()
	}()

	// Start server
	if tlsConfig.Certificates != nil {
		return s.server.ListenAndServeTLS("", "")
	}
	return s.server.ListenAndServe()
}

func (s *Server) handleRequest() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.logger.Debug().
			Str("remote", r.RemoteAddr).
			Str("method", r.Method).
			Str("url", r.URL.String()).
			Msg("Received request")

		// For MTPROTO, we handle the connection at TCP level
		// This is a placeholder - actual MTPROTO handling in Day 3-4
		conn, _, err := w.(http.Hijacker).Hijack()
		if err != nil {
			s.logger.Error().Err(err).Msg("Failed to hijack connection")
			return
		}
		defer conn.Close()

		s.handleMTProtoConnection(conn)
	})
}

func (s *Server) handleMTProtoConnection(conn net.Conn) {
	// Placeholder for MTPROTO protocol handling
	// Will be implemented in Day 2-3
	s.logger.Debug().Msg("New MTPROTO connection")
}
