package cmd

import (
	"fmt"
	"net"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"syscall"

	"waypasta/internal/socket"
	"waypasta/internal/storage"
	"waypasta/pkg/config"
)

// RunDaemon starts the waypasta daemon
func RunDaemon() {
	// Check if daemon is already running by attempting to connect to socket
	socketPath := config.GetSocketPath()
	if conn, err := net.Dial("unix", socketPath); err == nil {
		conn.Close()
		fmt.Fprintf(os.Stderr, "Daemon is already running\n")
		os.Exit(1)
	}

	// Daemonize the process
	daemonize()

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

// daemonize forks the process to run in the background
func daemonize() {
	// Check if we're already running as a daemon
	if os.Getenv("WAYPASTA_DAEMON") == "1" {
		// Already daemonized, just return
		return
	}

	// Get the path to the current executable
	exe, err := os.Executable()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get executable path: %v\n", err)
		os.Exit(1)
	}

	// Prepare command to re-execute ourselves
	cmd := exec.Command(exe, "start")

	// Set environment variable to mark as daemonized
	cmd.Env = append(os.Environ(), "WAYPASTA_DAEMON=1")

	// Detach from parent process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
		Pgid:    0,
	}

	// Redirect standard file descriptors to /dev/null
	cmd.Stdin = nil
	cmd.Stdout = nil
	cmd.Stderr = nil

	// Start the daemon process
	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to daemonize: %v\n", err)
		os.Exit(1)
	}

	// Parent process exits
	fmt.Fprintf(os.Stderr, "Daemon started with PID %d\n", cmd.Process.Pid)
	os.Exit(0)
}
