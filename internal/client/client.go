package client

import (
	"bufio"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/hashicorp/yamux"
	"github.com/nitintf/openport/internal/tunnel"
)

var (
	ErrLocalNotReachable  = errors.New("local not reachable")
	ErrServerUnreachable  = errors.New("server unreachable")
	ErrSubdomainTaken     = errors.New("subdomain taken")
	ErrConnectionLost     = errors.New("connection lost")
)

// ConnectError wraps an error with human-readable context.
type ConnectError struct {
	Kind    error
	Addr    string
	Detail  string
}

func (e *ConnectError) Error() string {
	return fmt.Sprintf("%s: %s", e.Kind, e.Detail)
}

func (e *ConnectError) Unwrap() error {
	return e.Kind
}

// Config holds client configuration.
type Config struct {
	ServerAddr  string
	LocalAddr   string
	Subdomain   string
	OnConnected func(tunnelURL string)
	OnRequest   func(RequestLog)
}

// Client connects to the openport server and forwards traffic to a local service.
type Client struct {
	cfg       Config
	conn      net.Conn
	session   *yamux.Session
	TunnelURL string
}

// New creates a new Client.
func New(cfg Config) (*Client, error) {
	return &Client{cfg: cfg}, nil
}

// Port extracts the port number from the local address.
func (c *Client) Port() string {
	_, port, _ := net.SplitHostPort(c.cfg.LocalAddr)
	return port
}

// Connect establishes a tunnel with the server and begins forwarding traffic.
func (c *Client) Connect() error {
	localConn, err := net.DialTimeout("tcp", c.cfg.LocalAddr, 2*time.Second)
	if err != nil {
		return &ConnectError{
			Kind: ErrLocalNotReachable,
			Addr: c.cfg.LocalAddr,
			Detail: fmt.Sprintf("port %s", c.Port()),
		}
	}
	localConn.Close()

	c.conn, err = net.Dial("tcp", c.cfg.ServerAddr)
	if err != nil {
		return &ConnectError{
			Kind:   ErrServerUnreachable,
			Addr:   c.cfg.ServerAddr,
			Detail: c.cfg.ServerAddr,
		}
	}

	err = tunnel.SendHandshake(c.conn, tunnel.Handshake{
		Subdomain: c.cfg.Subdomain,
	})
	if err != nil {
		return &ConnectError{
			Kind:   ErrServerUnreachable,
			Addr:   c.cfg.ServerAddr,
			Detail: "handshake failed",
		}
	}

	resp, err := tunnel.ReadHandshakeResp(c.conn)
	if err != nil {
		return &ConnectError{
			Kind:   ErrServerUnreachable,
			Addr:   c.cfg.ServerAddr,
			Detail: "no response from server",
		}
	}
	if resp.Error != "" {
		if strings.Contains(resp.Error, "already in use") {
			return &ConnectError{
				Kind:   ErrSubdomainTaken,
				Addr:   c.cfg.Subdomain,
				Detail: c.cfg.Subdomain,
			}
		}
		return &ConnectError{
			Kind:   ErrServerUnreachable,
			Addr:   c.cfg.ServerAddr,
			Detail: resp.Error,
		}
	}

	c.TunnelURL = resp.URL

	if c.cfg.OnConnected != nil {
		c.cfg.OnConnected(resp.URL)
	}

	c.session, err = yamux.Server(c.conn, nil)
	if err != nil {
		return &ConnectError{
			Kind:   ErrConnectionLost,
			Addr:   c.cfg.ServerAddr,
			Detail: "failed to establish multiplexed session",
		}
	}

	for {
		stream, err := c.session.Accept()
		if err != nil {
			return &ConnectError{
				Kind:   ErrConnectionLost,
				Addr:   c.cfg.ServerAddr,
				Detail: "tunnel disconnected",
			}
		}
		go c.handleStream(stream)
	}
}

func (c *Client) handleStream(stream net.Conn) {
	defer stream.Close()

	req, err := http.ReadRequest(bufio.NewReader(stream))
	if err != nil {
		return
	}

	method := req.Method
	path := req.URL.Path

	req.URL.Scheme = "http"
	req.URL.Host = c.cfg.LocalAddr
	req.RequestURI = ""

	start := time.Now()
	resp, err := http.DefaultTransport.RoundTrip(req)
	duration := time.Since(start)

	if err != nil {
		errResp := &http.Response{
			StatusCode: http.StatusBadGateway,
			ProtoMajor: 1,
			ProtoMinor: 1,
			Header:     make(http.Header),
		}
		errResp.Write(stream)

		if c.cfg.OnRequest != nil {
			c.cfg.OnRequest(RequestLog{
				Method:     method,
				Path:       path,
				StatusCode: http.StatusBadGateway,
				Duration:   duration,
				Timestamp:  start,
			})
		}
		return
	}
	defer resp.Body.Close()

	resp.Write(stream)

	if c.cfg.OnRequest != nil {
		c.cfg.OnRequest(RequestLog{
			Method:     method,
			Path:       path,
			StatusCode: resp.StatusCode,
			Duration:   duration,
			Timestamp:  start,
		})
	}
}

// Close tears down the tunnel connection.
func (c *Client) Close() {
	if c.session != nil {
		c.session.Close()
	}
	if c.conn != nil {
		c.conn.Close()
	}
}
