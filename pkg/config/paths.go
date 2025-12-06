package config

import (
	"os"
	"path/filepath"
)

// GetSocketPath returns the Unix socket path for noodles
// Uses XDG_RUNTIME_DIR if available, otherwise falls back to /tmp
func GetSocketPath() string {
	runtime := os.Getenv("XDG_RUNTIME_DIR")
	if runtime == "" {
		runtime = "/tmp"
	}
	return filepath.Join(runtime, "noodles.sock")
}
