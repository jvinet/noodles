package cmd

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"waypasta/internal/protocol"
	"waypasta/internal/socket"
)

// RunGet handles the get command
func RunGet(args []string) {
	// Parse index from args or stdin
	index, err := parseIndex(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse index: %v\n", err)
		os.Exit(1)
	}

	client := socket.NewClient()

	// Check if daemon is running
	if !client.SocketExists() {
		fmt.Fprintf(os.Stderr, "Daemon is not running\n")
		os.Exit(1)
	}

	// Send GET command
	msg := protocol.Message{
		Command: protocol.CmdGet,
		Index:   index,
	}

	resp, err := client.SendCommand(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to get clipboard item: %v\n", err)
		os.Exit(1)
	}

	// Output raw clipboard data to stdout
	os.Stdout.Write(resp.Data)
}

// parseIndex parses the index from command-line args or stdin
func parseIndex(args []string) (int, error) {
	var indexStr string

	if len(args) > 0 {
		// Index provided as command-line argument
		indexStr = args[0]
	} else {
		// Read index from stdin
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return 0, fmt.Errorf("failed to read from stdin: %w", err)
		}
		indexStr = strings.TrimSpace(string(data))
	}

	// Parse as integer
	index, err := strconv.Atoi(indexStr)
	if err != nil {
		return 0, fmt.Errorf("invalid index '%s': %w", indexStr, err)
	}

	if index < 0 {
		return 0, fmt.Errorf("index must be non-negative")
	}

	return index, nil
}
