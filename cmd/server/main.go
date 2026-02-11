package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nitintf/openport/internal/server"
	"github.com/nitintf/openport/internal/version"
)

func main() {
	showVersion := flag.Bool("version", false, "print version and exit")
	addr := flag.String("addr", "", "public HTTP address to listen on")
	tunnelAddr := flag.String("tunnel-addr", ":9090", "address for tunnel client connections")
	domain := flag.String("domain", "localhost", "base domain for subdomain routing")
	flag.Parse()

	if *showVersion {
		fmt.Printf("openport-server %s\n", version.Full())
		return
	}

	// Railway / cloud platforms set PORT env var.
	if *addr == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		*addr = ":" + port
	}

	if env := os.Getenv("DOMAIN"); env != "" && *domain == "localhost" {
		*domain = env
	}

	cfg := server.Config{
		Addr:       *addr,
		TunnelAddr: *tunnelAddr,
		Domain:     *domain,
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("failed to create server: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("openport-server %s starting on %s (tunnels on %s)", version.Full(), cfg.Addr, cfg.TunnelAddr)
		if err := srv.Start(); err != nil {
			log.Fatalf("server error: %v", err)
		}
	}()

	<-quit
	log.Println("shutting down server...")
	srv.Stop()
}
