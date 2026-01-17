package server

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fooling/6-color-editor/internal/web/handlers"
	"github.com/spf13/cobra"
)

var (
	port        int
	host        string
	remoteURL   string
)

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Start the web server for E-Ink image processing",
	Long: `Start a web server with an interactive UI for E-Ink image processing.

Features:
- Split view: Original vs E-Ink preview
- Live adjustments: Saturation, Contrast, Brightness sliders
- Step visualization: See intermediate processing results
- Direct upload to E-Ink display`,
	RunE: runServer,
}

func init() {
	serverCmd.Flags().IntVarP(&port, "port", "p", 3000, "Server port")
	serverCmd.Flags().StringVarP(&host, "host", "H", "0.0.0.0", "Server host")
	serverCmd.Flags().StringVar(&remoteURL, "remote-url", "http://127.0.0.1:8080/esp/dataUP", "Default remote display URL")
}

func runServer(cmd *cobra.Command, args []string) error {
	addr := fmt.Sprintf("%s:%d", host, port)

	// Create handler registry with config
	registry := handlers.NewRegistry(&handlers.Config{
		RemoteURL: remoteURL,
	})

	// Create HTTP server
	srv := &http.Server{
		Addr:         addr,
		Handler:      registry.Routes(),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in background
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		log.Printf("Starting server on http://%s\n", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server error: %v", err)
		}
	}()

	// Wait for interrupt signal
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	log.Println("Press Ctrl+C to stop the server")
	<-ctx.Done()

	log.Println("Shutting down server...")

	// Graceful shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("Shutdown error: %v", err)
		return err
	}

	wg.Wait()
	log.Println("Server stopped")

	return nil
}

// GetCommand returns the server command for registration
func GetCommand() *cobra.Command {
	return serverCmd
}
