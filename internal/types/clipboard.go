package types

import "time"

// ClipboardItem represents a single clipboard entry
type ClipboardItem struct {
	Index     int       // Sequential index
	Data      []byte    // Raw clipboard data (text or binary)
	IsBinary  bool      // Whether data is binary
	Timestamp time.Time // When item was stored
	Size      int       // Size in bytes
}
