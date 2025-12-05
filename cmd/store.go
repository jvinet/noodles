package cmd

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
	"time"

	"waypasta/internal/protocol"
	"waypasta/internal/socket"
)

// RunStore handles the store command
func RunStore() {
	// Check CLIPBOARD_STATE environment variable
	clipState := os.Getenv("CLIPBOARD_STATE")
	if clipState != "data" {
		// Only store if CLIPBOARD_STATE is "data"
		os.Exit(0)
	}

	// Read clipboard data from stdin
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to read clipboard data: %v\n", err)
		os.Exit(1)
	}

	client := socket.NewClient()

	// Auto-start daemon if socket doesn't exist
	if !client.SocketExists() {
		if err := startDaemon(); err != nil {
			fmt.Fprintf(os.Stderr, "Failed to start daemon: %v\n", err)
			os.Exit(1)
		}

		// Wait for daemon to be ready with retry
		if err := waitForSocket(client); err != nil {
			fmt.Fprintf(os.Stderr, "Daemon did not start in time: %v\n", err)
			os.Exit(1)
		}
	}

	// Send STORE command
	msg := protocol.Message{
		Command: protocol.CmdStore,
		Data:    data,
	}

	resp, err := client.SendCommand(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to store clipboard data: %v\n", err)
		os.Exit(1)
	}

	// Check response status (should be ACK)
	if resp.Status != protocol.StatusACK {
		fmt.Fprintf(os.Stderr, "Store failed: %s\n", resp.Message)
		os.Exit(1)
	}
}

// startDaemon starts the waypasta daemon as a background process
func startDaemon() error {
	// Get the path to the current executable
	exe, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to get executable path: %w", err)
	}

	// Start daemon in background
	cmd := exec.Command(exe, "start")
	cmd.Stdout = nil
	cmd.Stderr = nil
	cmd.Stdin = nil

	// Detach from parent process
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setpgid: true,
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start daemon: %w", err)
	}

	// Don't wait for the daemon to finish
	return nil
}

// waitForSocket waits for the daemon socket to become available
// Retries with exponential backoff: 50ms, 100ms, 200ms
func waitForSocket(client *socket.Client) error {
	delays := []time.Duration{50 * time.Millisecond, 100 * time.Millisecond, 200 * time.Millisecond}

	for _, delay := range delays {
		time.Sleep(delay)
		if client.SocketExists() {
			return nil
		}
	}

	return fmt.Errorf("socket not available after retries")
}
