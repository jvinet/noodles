package storage

import (
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"noodles/internal/types"
)

const (
	// MaxClipboardSize is the maximum size of a clipboard item (10 MB)
	MaxClipboardSize = 10 * 1024 * 1024
)

// MemoryStore provides thread-safe in-memory storage for clipboard items
type MemoryStore struct {
	mu             sync.RWMutex
	items          []types.ClipboardItem
	maxItems       int
	maxMemoryBytes int
	totalBytes     int
}

// NewMemoryStore creates a new MemoryStore with specified limits
func NewMemoryStore(maxItems, maxMemoryBytes int) *MemoryStore {
	return &MemoryStore{
		items:          make([]types.ClipboardItem, 0),
		maxItems:       maxItems,
		maxMemoryBytes: maxMemoryBytes,
		totalBytes:     0,
	}
}

// Store adds a new clipboard item to storage
// Returns the index of the stored item or an error
func (m *MemoryStore) Store(data []byte) (int, error) {
	if len(data) < 1 {
		// This is a silent no-op; we don't want to store empty strings.
		return 0, nil
	}
	if len(data) > MaxClipboardSize {
		return -1, fmt.Errorf("clipboard data exceeds maximum size of %d bytes", MaxClipboardSize)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	dataSize := len(data)

	// Check if this single item exceeds max memory
	if dataSize > m.maxMemoryBytes {
		return -1, fmt.Errorf("clipboard item size (%d bytes) exceeds maximum memory limit (%d bytes)", dataSize, m.maxMemoryBytes)
	}

	// Evict oldest items until there's enough memory for the new item
	for m.totalBytes+dataSize > m.maxMemoryBytes && len(m.items) > 0 {
		m.evictOldest()
	}

	// If at max items, remove oldest item (last in slice)
	if len(m.items) >= m.maxItems {
		m.evictOldest()
	}

	// Increment all existing item indexes by 1
	for i := range m.items {
		m.items[i].Index++
	}

	// New item always gets index 0
	item := types.ClipboardItem{
		Index:     0,
		Data:      data,
		IsBinary:  !utf8.Valid(data),
		Timestamp: time.Now(),
		Size:      dataSize,
	}

	// Prepend new item to the front of the slice
	m.items = append([]types.ClipboardItem{item}, m.items...)
	m.totalBytes += dataSize

	return 0, nil
}

// evictOldest removes the oldest item (last in slice, highest index)
// Must be called with lock held
func (m *MemoryStore) evictOldest() {
	if len(m.items) == 0 {
		return
	}

	// Remove last item (oldest, highest index) and update total bytes
	lastIndex := len(m.items) - 1
	oldestItem := m.items[lastIndex]
	m.totalBytes -= oldestItem.Size
	m.items = m.items[:lastIndex]
}

// List returns all stored clipboard items
func (m *MemoryStore) List() []types.ClipboardItem {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a copy to avoid external modification
	items := make([]types.ClipboardItem, len(m.items))
	copy(items, m.items)
	return items
}

// Get retrieves a clipboard item by index
func (m *MemoryStore) Get(index int) (types.ClipboardItem, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	for _, item := range m.items {
		if item.Index == index {
			return item, nil
		}
	}

	return types.ClipboardItem{}, fmt.Errorf("clipboard item with index %d not found", index)
}

// Wipe deletes all stored clipboard items
func (m *MemoryStore) Wipe() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.items = make([]types.ClipboardItem, 0)
	m.totalBytes = 0
}
