package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/nitintf/openport/internal/client"
)

func main() {
	serverAddr := flag.String("server", "localhost:9090", "openport server address")
	localAddr := flag.String("local", "localhost:3000", "local service address to expose")
	subdomain := flag.String("subdomain", "", "requested subdomain (optional)")
	flag.Parse()

	cfg := client.Config{
		ServerAddr: *serverAddr,
		LocalAddr:  *localAddr,
		Subdomain:  *subdomain,
	}

	c, err := client.New(cfg)
	if err != nil {
		log.Fatalf("failed to create client: %v", err)
	}

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("connecting to %s, exposing %s", cfg.ServerAddr, cfg.LocalAddr)
		if err := c.Connect(); err != nil {
			log.Fatalf("client error: %v", err)
		}
	}()

	<-quit
	log.Println("disconnecting...")
	c.Close()
}
