package socket

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"

	"waypasta/internal/protocol"
	"waypasta/internal/storage"
	"waypasta/pkg/config"
)

// Server represents a Unix socket server
type Server struct {
	listener net.Listener
	store    *storage.MemoryStore
	stopChan chan bool
}

// NewServer creates a new Unix socket server
func NewServer(store *storage.MemoryStore) (*Server, error) {
	socketPath := config.GetSocketPath()

	// Remove stale socket file if it exists
	if err := removeStaleSocket(socketPath); err != nil {
		return nil, err
	}

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create socket: %w", err)
	}

	// Set socket permissions to owner-only (0600)
	if err := os.Chmod(socketPath, 0600); err != nil {
		listener.Close()
		os.Remove(socketPath)
		return nil, fmt.Errorf("failed to set socket permissions: %w", err)
	}

	return &Server{
		listener: listener,
		store:    store,
		stopChan: make(chan bool),
	}, nil
}

// removeStaleSocket attempts to remove a stale socket file
func removeStaleSocket(socketPath string) error {
	// Check if socket file exists
	if _, err := os.Stat(socketPath); err != nil {
		if os.IsNotExist(err) {
			return nil // Socket doesn't exist, nothing to do
		}
		return err
	}

	// Try to connect to see if it's stale
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		// Connection failed, assume stale and remove
		return os.Remove(socketPath)
	}

	// Socket is active, don't remove it
	conn.Close()
	return fmt.Errorf("socket already in use")
}

// Serve starts accepting connections and handling requests
func (s *Server) Serve() error {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			select {
			case <-s.stopChan:
				return nil // Graceful shutdown
			default:
				return fmt.Errorf("accept error: %w", err)
			}
		}

		go s.handleConnection(conn)
	}
}

// handleConnection processes a single client connection
func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	// Decode message from client
	var msg protocol.Message
	decoder := gob.NewDecoder(conn)
	if err := decoder.Decode(&msg); err != nil {
		s.sendError(conn, fmt.Sprintf("decode error: %v", err))
		return
	}

	// Process command
	resp := s.processCommand(&msg)

	// Send response
	encoder := gob.NewEncoder(conn)
	if err := encoder.Encode(resp); err != nil {
		// Can't send error response if encoding fails
		return
	}
}

// processCommand handles different command types
func (s *Server) processCommand(msg *protocol.Message) protocol.Response {
	switch msg.Command {
	case protocol.CmdStore:
		index, err := s.store.Store(msg.Data)
		if err != nil {
			return protocol.Response{
				Status:  protocol.StatusError,
				Message: err.Error(),
			}
		}
		return protocol.Response{
			Status:  protocol.StatusACK,
			Message: fmt.Sprintf("stored at index %d", index),
		}

	case protocol.CmdList:
		items := s.store.List()
		return protocol.Response{
			Status: protocol.StatusACK,
			Items:  items,
		}

	case protocol.CmdGet:
		item, err := s.store.Get(msg.Index)
		if err != nil {
			return protocol.Response{
				Status:  protocol.StatusError,
				Message: err.Error(),
			}
		}
		return protocol.Response{
			Status: protocol.StatusACK,
			Data:   item.Data,
		}

	case protocol.CmdWipe:
		s.store.Wipe()
		return protocol.Response{
			Status: protocol.StatusACK,
		}

	case protocol.CmdStop:
		// Signal shutdown and close listener to unblock Accept()
		go func() {
			close(s.stopChan)
			s.listener.Close()
		}()
		return protocol.Response{
			Status: protocol.StatusACK,
		}

	default:
		return protocol.Response{
			Status:  protocol.StatusError,
			Message: fmt.Sprintf("unknown command: %s", msg.Command),
		}
	}
}

// sendError sends an error response
func (s *Server) sendError(conn net.Conn, message string) {
	resp := protocol.Response{
		Status:  protocol.StatusError,
		Message: message,
	}
	encoder := gob.NewEncoder(conn)
	encoder.Encode(resp)
}

// Stop gracefully stops the server
func (s *Server) Stop() error {
	// Close stopChan if not already closed
	select {
	case <-s.stopChan:
		// Already closed
	default:
		close(s.stopChan)
	}

	// Close listener (ignore error if already closed)
	s.listener.Close()

	// Remove socket file (ignore error if already removed)
	socketPath := config.GetSocketPath()
	err := os.Remove(socketPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
