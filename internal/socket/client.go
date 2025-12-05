package socket

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"

	"waypasta/internal/protocol"
	"waypasta/pkg/config"
)

// Client provides methods to communicate with the daemon via Unix socket
type Client struct {
	socketPath string
}

// NewClient creates a new socket client
func NewClient() *Client {
	return &Client{
		socketPath: config.GetSocketPath(),
	}
}

// SocketExists checks if the daemon socket file exists
func (c *Client) SocketExists() bool {
	_, err := os.Stat(c.socketPath)
	return err == nil
}

// SendCommand sends a command to the daemon and returns the response
func (c *Client) SendCommand(msg protocol.Message) (protocol.Response, error) {
	// Connect to Unix socket
	conn, err := net.Dial("unix", c.socketPath)
	if err != nil {
		return protocol.Response{}, fmt.Errorf("failed to connect to daemon: %w", err)
	}
	defer conn.Close()

	// Encode and send message
	encoder := gob.NewEncoder(conn)
	if err := encoder.Encode(msg); err != nil {
		return protocol.Response{}, fmt.Errorf("failed to send message: %w", err)
	}

	// Decode response
	var resp protocol.Response
	decoder := gob.NewDecoder(conn)
	if err := decoder.Decode(&resp); err != nil {
		return protocol.Response{}, fmt.Errorf("failed to decode response: %w", err)
	}

	// Check for error response
	if resp.Status == protocol.StatusError {
		return resp, fmt.Errorf("daemon error: %s", resp.Message)
	}

	return resp, nil
}
