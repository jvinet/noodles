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
	mu        sync.RWMutex
	items     []types.ClipboardItem
	nextIndex int
}

// NewMemoryStore creates a new MemoryStore
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		items:     make([]types.ClipboardItem, 0),
		nextIndex: 0,
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

	item := types.ClipboardItem{
		Index:     m.nextIndex,
		Data:      data,
		IsBinary:  !utf8.Valid(data),
		Timestamp: time.Now(),
		Size:      len(data),
	}

	m.items = append(m.items, item)
	index := m.nextIndex
	m.nextIndex++

	return index, nil
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
}
