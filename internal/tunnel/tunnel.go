package tunnel

import (
	"encoding/json"
	"fmt"
	"io"
	"net"

	"github.com/hashicorp/yamux"
)

// Handshake is the initial message a client sends to register a tunnel.
type Handshake struct {
	Subdomain string `json:"subdomain,omitempty"`
}

// HandshakeResp is the server's response after registering the tunnel.
type HandshakeResp struct {
	Subdomain string `json:"subdomain"`
	URL       string `json:"url"`
	Error     string `json:"error,omitempty"`
}

// Tunnel represents an active tunnel connection between the server and a client.
type Tunnel struct {
	ID        string
	Subdomain string
	Conn      net.Conn
	Session   *yamux.Session
}

// SendHandshake writes a handshake message to the connection.
func SendHandshake(conn net.Conn, h Handshake) error {
	return json.NewEncoder(conn).Encode(h)
}

// ReadHandshake reads a handshake message from the connection.
func ReadHandshake(conn net.Conn) (Handshake, error) {
	var h Handshake
	if err := json.NewDecoder(conn).Decode(&h); err != nil {
		return h, fmt.Errorf("read handshake: %w", err)
	}
	return h, nil
}

// SendHandshakeResp writes a handshake response to the connection.
func SendHandshakeResp(conn net.Conn, resp HandshakeResp) error {
	return json.NewEncoder(conn).Encode(resp)
}

// ReadHandshakeResp reads a handshake response from the connection.
func ReadHandshakeResp(conn net.Conn) (HandshakeResp, error) {
	var resp HandshakeResp
	if err := json.NewDecoder(conn).Decode(&resp); err != nil {
		return resp, fmt.Errorf("read handshake response: %w", err)
	}
	return resp, nil
}

// Relay copies data bidirectionally between two connections.
func Relay(a, b io.ReadWriteCloser) error {
	errc := make(chan error, 2)

	go func() {
		_, err := io.Copy(a, b)
		errc <- err
	}()
	go func() {
		_, err := io.Copy(b, a)
		errc <- err
	}()

	return <-errc
}
