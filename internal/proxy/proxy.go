package proxy

import (
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
)

// NewHTTPReverseProxy returns a reverse proxy that forwards requests to the given target.
func NewHTTPReverseProxy(target string) (*httputil.ReverseProxy, error) {
	u, err := url.Parse("http://" + target)
	if err != nil {
		return nil, err
	}
	return httputil.NewSingleHostReverseProxy(u), nil
}

// ProxyTCP bidirectionally copies data between two TCP connections.
func ProxyTCP(client, backend net.Conn) error {
	errc := make(chan error, 2)

	go func() {
		_, err := io.Copy(backend, client)
		errc <- err
	}()
	go func() {
		_, err := io.Copy(client, backend)
		errc <- err
	}()

	return <-errc
}

// HijackAndProxy upgrades an HTTP connection and proxies it as raw TCP.
func HijackAndProxy(w http.ResponseWriter, backend net.Conn) error {
	hj, ok := w.(http.Hijacker)
	if !ok {
		return io.ErrUnexpectedEOF
	}

	clientConn, _, err := hj.Hijack()
	if err != nil {
		return err
	}
	defer clientConn.Close()

	return ProxyTCP(clientConn, backend)
}
