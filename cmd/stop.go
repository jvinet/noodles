package cmd

import (
	"fmt"
	"os"

	"waypasta/internal/protocol"
	"waypasta/internal/socket"
)

// RunStop handles the stop command
func RunStop() {
	client := socket.NewClient()

	// Check if daemon is running
	if !client.SocketExists() {
		fmt.Fprintf(os.Stderr, "Daemon is not running\n")
		os.Exit(1)
	}

	// Send STOP command
	msg := protocol.Message{
		Command: protocol.CmdStop,
	}

	_, err := client.SendCommand(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to stop daemon: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("Daemon stopped")
}
