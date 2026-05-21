package main

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/parkerscobey/nuvei-dmn-simulator/internal/server"
	"github.com/spf13/cobra"
)

func newServerCommand() *cobra.Command {
	var configPath string
	var host string
	var port int

	cmd := &cobra.Command{
		Use:   "server",
		Short: "Run the local web UI",
		RunE: func(cmd *cobra.Command, args []string) error {
			path, err := resolveConfigPath(configPath)
			if err != nil {
				return err
			}

			handler, err := server.NewHandler(path, verifyMerchantProfile, sendDMNPayload)
			if err != nil {
				return err
			}

			addr := net.JoinHostPort(host, strconv.Itoa(port))
			httpServer := &http.Server{
				Addr:              addr,
				Handler:           handler.Routes(),
				ReadHeaderTimeout: 10 * time.Second,
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Server listening on http://%s\n", addr)
			if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				return err
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&configPath, "config", "", "config file path (defaults to user config directory)")
	cmd.Flags().StringVar(&host, "host", "127.0.0.1", "server bind host")
	cmd.Flags().IntVar(&port, "port", 8080, "server bind port")

	return cmd
}
