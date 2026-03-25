package fronting

import (
	"crypto/tls"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// HTTPFronting provides domain fronting capability
type HTTPFronting struct {
	// FrontDomain is the domain that appears in TLS SNI
	FrontDomain string
	// BackendDomain is the actual target
	BackendDomain string
	// Hosts maps front domains to backends
	Hosts map[string]string
}

// NewHTTPFronting creates a new domain fronting handler
func NewHTTPFronting() *HTTPFronting {
	return &HTTPFronting{
		Hosts: make(map[string]string),
	}
}

// Request represents a fronted HTTP request
type Request struct {
	Method  string
	Path    string
	Host    string
	Headers http.Header
	Body    []byte
}

// Response represents a fronted HTTP response
type Response struct {
	StatusCode int
	Headers    http.Header
	Body       []byte
}

// Dial connects to the front domain but routes to backend
func (f *HTTPFronting) Dial(frontDomain, backendDomain string) (net.Conn, error) {
	// Connect to front domain
	addr := fmt.Sprintf("%s:443", frontDomain)
	conn, err := tls.Dial("tcp", addr, &tls.Config{
		ServerName:         frontDomain,
		InsecureSkipVerify: false,
	})
	if err != nil {
		return nil, fmt.Errorf("dial front: %w", err)
	}

	// Store backend for request routing
	f.FrontDomain = frontDomain
	f.BackendDomain = backendDomain
	f.Hosts[frontDomain] = backendDomain

	return conn, nil
}

// BuildRequest builds an HTTP request with fronting
func (f *HTTPFronting) BuildRequest(method, path, host string, body []byte) *http.Request {
	req := &http.Request{
		Method:     method,
		URL:        &url.URL{Path: path},
		Host:       host,
		Header:     make(http.Header),
		Body:       nil,
		BodyLength: len(body),
	}

	if len(body) > 0 {
		req.Body = io.NopCloser(strings.NewReader(string(body)))
	}

	// Add common headers for domain fronting
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	return req
}

// GetBackend returns the actual backend for a front domain
func (f *HTTPFronting) GetBackend(frontDomain string) string {
	if backend, ok := f.Hosts[frontDomain]; ok {
		return backend
	}
	return frontDomain
}

// HTTPFallback provides HTTP fallback when direct connection fails
type HTTPFallback struct {
	Transport *http.Transport
	Timeout   time.Duration
}

// NewHTTPFallback creates a new HTTP fallback handler
func NewHTTPFallback() *HTTPFallback {
	return &HTTPFallback{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				// Customize TLS to mimic browser
				MinVersion:         tls.VersionTLS12,
				MaxVersion:         tls.VersionTLS13,
				InsecureSkipVerify: false,
			},
		},
		Timeout: 30 * time.Second,
	}
}

// Fetch performs an HTTP request through fallback
func (h *HTTPFallback) Fetch(req *Request) (*Response, error) {
	client := &http.Client{
		Transport: h.Transport,
		Timeout:   h.Timeout,
	}

	httpReq, err := http.NewRequest(req.Method, req.Path, strings.NewReader(string(req.Body)))
	if err != nil {
		return nil, err
	}

	httpReq.Host = req.Host
	httpReq.Header = req.Headers

	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Response{
		StatusCode: resp.StatusCode,
		Headers:    resp.Header,
		Body:       body,
	}, nil
}
