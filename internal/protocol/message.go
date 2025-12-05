package protocol

import "waypasta/internal/types"

// Command constants
const (
	CmdStore = "STORE"
	CmdList  = "LIST"
	CmdGet   = "GET"
	CmdWipe  = "WIPE"
	CmdStop  = "STOP"
)

// Response status constants
const (
	StatusACK   = "ACK"
	StatusError = "ERROR"
)

// Message structure for client-daemon communication
type Message struct {
	Command string // Command type (STORE, LIST, GET, WIPE, STOP)
	Data    []byte // Payload (for STORE command)
	Index   int    // Item index (for GET command)
}

// Response structure from daemon to client
type Response struct {
	Status  string                // "ACK" or "ERROR"
	Message string                // Error message if applicable
	Items   []types.ClipboardItem // For LIST command
	Data    []byte                // For GET command
}
