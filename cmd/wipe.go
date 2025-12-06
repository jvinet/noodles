package cmd

import (
	"fmt"
	"os"

	"noodles/internal/protocol"
	"noodles/internal/socket"
)

// RunWipe handles the wipe command
func RunWipe() {
	client := socket.NewClient()

	// Check if daemon is running
	if !client.SocketExists() {
		fmt.Fprintf(os.Stderr, "Daemon is not running\n")
		os.Exit(1)
	}

	// Send WIPE command
	msg := protocol.Message{
		Command: protocol.CmdWipe,
	}

	_, err := client.SendCommand(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to wipe clipboard items: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("All clipboard items cleared")
}
