package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/nitintf/openport/internal/client"
	"github.com/nitintf/openport/internal/ui"
	"github.com/nitintf/openport/internal/version"
)

func main() {
	var serverAddr string
	var subdomain string

	rootCmd := &cobra.Command{
		Use:     "op <port>",
		Short:   "Expose a local port to the internet",
		Long:    "openport (op) creates a secure tunnel to expose a local service to the public internet.",
		Version: version.Full(),
		Example: `  op 3000
  op 8080 --server tunnel.example.com:9090
  op 4000 --subdomain myapp`,
		Args:          cobra.ExactArgs(1),
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			localAddr := "localhost:" + args[0]

			cfg := client.Config{
				ServerAddr: serverAddr,
				LocalAddr:  localAddr,
				Subdomain:  subdomain,
				OnConnected: func(tunnelURL string) {
					ui.PrintBanner(tunnelURL, localAddr)
				},
				OnRequest: ui.PrintRequestLog,
			}

			c, err := client.New(cfg)
			if err != nil {
				ui.PrintError(err)
				return err
			}

			quit := make(chan os.Signal, 1)
			signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

			errCh := make(chan error, 1)
			go func() {
				errCh <- c.Connect()
			}()

			select {
			case err := <-errCh:
				ui.PrintError(err)
				return err
			case <-quit:
				fmt.Println()
				ui.PrintShutdown()
				c.Close()
				return nil
			}
		},
	}

	rootCmd.Flags().StringVarP(&serverAddr, "server", "s", "localhost:9090", "openport server address")
	rootCmd.Flags().StringVarP(&subdomain, "subdomain", "d", "", "request a specific subdomain")

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
