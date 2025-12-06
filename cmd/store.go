package cmd

import (
	"fmt"
	"io"
	"os"

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

	// Check if daemon is running
	if !client.SocketExists() {
		fmt.Fprintf(os.Stderr, "Daemon is not running\n")
		os.Exit(1)
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
