package cmd

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"

	"noodles/internal/protocol"
	"noodles/internal/socket"
)

const (
	maxTextDisplayLength = 40
)

// RunList handles the list command
func RunList() {
	client := socket.NewClient()

	// Check if daemon is running
	if !client.SocketExists() {
		fmt.Fprintf(os.Stderr, "Daemon is not running\n")
		os.Exit(1)
	}

	// Send LIST command
	msg := protocol.Message{
		Command: protocol.CmdList,
	}

	resp, err := client.SendCommand(msg)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to list clipboard items: %v\n", err)
		os.Exit(1)
	}

	// Format and output items
	for _, item := range resp.Items {
		content := formatContent(item.Data, item.IsBinary, item.Size)
		fmt.Printf("%d\t%s\n", item.Index, content)
	}
}

// formatContent formats clipboard content for display
func formatContent(data []byte, isBinary bool, size int) string {
	if isBinary {
		// Format binary data with size in KiB
		sizeKiB := float64(size) / 1024.0
		return fmt.Sprintf("[[ binary data %.1f KiB ]]", sizeKiB)
	}

	// Format text data
	text := string(data)

	// Replace newlines and tabs with spaces for clean display
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	text = strings.ReplaceAll(text, "\r", " ")

	// Collapse multiple spaces into one
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	// Trim whitespace
	text = strings.TrimSpace(text)

	// Truncate to max length
	if utf8.RuneCountInString(text) > maxTextDisplayLength {
		runes := []rune(text)
		text = string(runes[:maxTextDisplayLength]) + "..."
	}

	return text
}
