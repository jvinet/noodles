package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"waypasta/internal/socket"
	"waypasta/internal/storage"
)

// RunDaemon starts the waypasta daemon
func RunDaemon() {
	// Initialize storage
	store := storage.NewMemoryStore()

	// Create and start socket server
	server, err := socket.NewServer(store)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to create server: %v\n", err)
		os.Exit(1)
	}

	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start server in a goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Serve()
	}()

	// Wait for shutdown signal or error
	select {
	case sig := <-sigChan:
		fmt.Fprintf(os.Stderr, "Received signal %v, shutting down...\n", sig)
	case err := <-errChan:
		// Ignore "use of closed network connection" error (graceful shutdown)
		if err != nil && !strings.Contains(err.Error(), "use of closed network connection") {
			fmt.Fprintf(os.Stderr, "Server error: %v\n", err)
		}
	}

	// Graceful shutdown
	if err := server.Stop(); err != nil {
		fmt.Fprintf(os.Stderr, "Error during shutdown: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Daemon stopped successfully\n")
}
