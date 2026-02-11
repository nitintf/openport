package server

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"

	"github.com/hashicorp/yamux"
	"github.com/nitintf/openport/internal/tunnel"
)

// Config holds server configuration.
type Config struct {
	Addr       string // public HTTP address
	TunnelAddr string // address for client tunnel connections
	Domain     string // base domain for subdomains
}

// Server manages tunnel registrations and proxies HTTP traffic to connected clients.
type Server struct {
	cfg      Config
	tunnels  map[string]*tunnel.Tunnel
	mu       sync.RWMutex
	listener net.Listener
	httpSrv  *http.Server
}

// New creates a new Server.
func New(cfg Config) (*Server, error) {
	s := &Server{
		cfg:     cfg,
		tunnels: make(map[string]*tunnel.Tunnel),
	}
	return s, nil
}

// Start begins listening for both tunnel client connections and public HTTP traffic.
func (s *Server) Start() error {
	var err error
	s.listener, err = net.Listen("tcp", s.cfg.TunnelAddr)
	if err != nil {
		return fmt.Errorf("tunnel listen: %w", err)
	}

	go s.acceptTunnels()

	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleHTTP)

	s.httpSrv = &http.Server{
		Addr:    s.cfg.Addr,
		Handler: mux,
	}
	return s.httpSrv.ListenAndServe()
}

// Stop shuts down the server.
func (s *Server) Stop() {
	if s.listener != nil {
		s.listener.Close()
	}
	if s.httpSrv != nil {
		s.httpSrv.Close()
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	for _, t := range s.tunnels {
		t.Session.Close()
		t.Conn.Close()
	}
}

func (s *Server) acceptTunnels() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			log.Printf("tunnel accept error: %v", err)
			return
		}
		go s.handleNewTunnel(conn)
	}
}

func (s *Server) handleNewTunnel(conn net.Conn) {
	hs, err := tunnel.ReadHandshake(conn)
	if err != nil {
		log.Printf("handshake error: %v", err)
		conn.Close()
		return
	}

	subdomain := hs.Subdomain
	if subdomain == "" {
		subdomain = randomSubdomain()
	}

	s.mu.Lock()
	if _, exists := s.tunnels[subdomain]; exists {
		s.mu.Unlock()
		tunnel.SendHandshakeResp(conn, tunnel.HandshakeResp{
			Error: fmt.Sprintf("subdomain %q is already in use", subdomain),
		})
		conn.Close()
		return
	}
	s.mu.Unlock()

	url := fmt.Sprintf("http://%s.%s%s", subdomain, s.cfg.Domain, s.cfg.Addr)

	tunnel.SendHandshakeResp(conn, tunnel.HandshakeResp{
		Subdomain: subdomain,
		URL:       url,
	})

	// Server is the yamux client (opens streams TO the tunnel client).
	// The tunnel client is the yamux server (accepts streams).
	session, err := yamux.Client(conn, nil)
	if err != nil {
		log.Printf("yamux session error: %v", err)
		conn.Close()
		return
	}

	t := &tunnel.Tunnel{
		ID:        randomID(),
		Subdomain: subdomain,
		Conn:      conn,
		Session:   session,
	}

	s.mu.Lock()
	s.tunnels[subdomain] = t
	s.mu.Unlock()

	log.Printf("tunnel registered: %s -> %s (%s)", subdomain, t.ID, url)

	// Block until the session is closed (client disconnected).
	<-session.CloseChan()

	s.mu.Lock()
	delete(s.tunnels, subdomain)
	s.mu.Unlock()
	log.Printf("tunnel unregistered: %s", subdomain)
}

func (s *Server) handleHTTP(w http.ResponseWriter, r *http.Request) {
	subdomain := extractSubdomain(r.Host, s.cfg.Domain)
	if subdomain == "" {
		http.Error(w, "openport: no tunnel specified", http.StatusBadRequest)
		return
	}

	s.mu.RLock()
	t, ok := s.tunnels[subdomain]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, fmt.Sprintf("openport: tunnel %q not found", subdomain), http.StatusNotFound)
		return
	}

	// Open a new yamux stream to the client for this request.
	stream, err := t.Session.Open()
	if err != nil {
		http.Error(w, "openport: failed to reach tunnel client", http.StatusBadGateway)
		log.Printf("yamux open stream error for %s: %v", subdomain, err)
		return
	}
	defer stream.Close()

	// Write the incoming HTTP request into the stream.
	if err := r.Write(stream); err != nil {
		http.Error(w, "openport: failed to forward request", http.StatusBadGateway)
		return
	}

	// Read the HTTP response back from the stream.
	resp, err := http.ReadResponse(bufio.NewReader(stream), r)
	if err != nil {
		http.Error(w, "openport: failed to read response from tunnel", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	// Copy response headers.
	for k, vv := range resp.Header {
		for _, v := range vv {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func extractSubdomain(host, baseDomain string) string {
	if idx := strings.LastIndex(host, ":"); idx != -1 {
		host = host[:idx]
	}
	if !strings.HasSuffix(host, "."+baseDomain) {
		return ""
	}
	return strings.TrimSuffix(host, "."+baseDomain)
}

func randomSubdomain() string {
	b := make([]byte, 4)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func randomID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}
