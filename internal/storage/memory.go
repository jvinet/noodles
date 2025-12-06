package storage

import (
	"fmt"
	"sync"
	"time"
	"unicode/utf8"

	"waypasta/internal/types"
)

const (
	// MaxClipboardSize is the maximum size of a clipboard item (10 MB)
	MaxClipboardSize = 10 * 1024 * 1024
)

// MemoryStore provides thread-safe in-memory storage for clipboard items
type MemoryStore struct {
	mu             sync.RWMutex
	items          []types.ClipboardItem
	nextIndex      int
	maxItems       int
	maxMemoryBytes int
	totalBytes     int
}

// NewMemoryStore creates a new MemoryStore with specified limits
func NewMemoryStore(maxItems, maxMemoryBytes int) *MemoryStore {
	return &MemoryStore{
		items:          make([]types.ClipboardItem, 0),
		nextIndex:      0,
		maxItems:       maxItems,
		maxMemoryBytes: maxMemoryBytes,
		totalBytes:     0,
	}
}

// Store adds a new clipboard item to storage
// Returns the index of the stored item or an error
func (m *MemoryStore) Store(data []byte) (int, error) {
	if len(data) > MaxClipboardSize {
		return -1, fmt.Errorf("clipboard data exceeds maximum size of %d bytes", MaxClipboardSize)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	dataSize := len(data)

	// Check if adding this item would exceed max memory
	if m.totalBytes+dataSize > m.maxMemoryBytes {
		return -1, fmt.Errorf("clipboard data would exceed maximum memory of %d bytes", m.maxMemoryBytes)
	}

	// If at max items, remove oldest item (first in slice)
	if len(m.items) >= m.maxItems {
		m.evictOldest()
	}

	item := types.ClipboardItem{
		Index:     m.nextIndex,
		Data:      data,
		IsBinary:  !utf8.Valid(data),
		Timestamp: time.Now(),
		Size:      dataSize,
	}

	m.items = append(m.items, item)
	m.totalBytes += dataSize
	index := m.nextIndex
	m.nextIndex++

	return index, nil
}

// evictOldest removes the oldest item and shifts all indexes down
// Must be called with lock held
func (m *MemoryStore) evictOldest() {
	if len(m.items) == 0 {
		return
	}

	// Remove first item and update total bytes
	oldestItem := m.items[0]
	m.totalBytes -= oldestItem.Size
	m.items = m.items[1:]

	// Shift all indexes down by 1
	for i := range m.items {
		m.items[i].Index--
	}

	// Decrement nextIndex since we removed an item
	m.nextIndex--
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
	m.nextIndex = 0
	m.totalBytes = 0
}
