package rugs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

// Storage provides JSON-based persistent storage for commands
type Storage struct {
	path string
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewStorage creates a new storage instance for a command
func NewStorage(commandName string) *Storage {
	return &Storage{
		path: filepath.Join("rugs", commandName+".json"),
		data: make(map[string]interface{}),
	}
}

// Load reads the storage file from disk
func (s *Storage) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if file exists
	if _, err := os.Stat(s.path); os.IsNotExist(err) {
		// Create empty storage file
		s.data = make(map[string]interface{})
		return s.saveUnsafe()
	}

	bytes, err := os.ReadFile(s.path)
	if err != nil {
		return fmt.Errorf("failed to read storage file: %w", err)
	}

	if err := json.Unmarshal(bytes, &s.data); err != nil {
		return fmt.Errorf("failed to parse storage file: %w", err)
	}

	return nil
}

// Save writes the storage to disk
func (s *Storage) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.saveUnsafe()
}

// saveUnsafe writes to disk without locking (caller must hold lock)
func (s *Storage) saveUnsafe() error {
	bytes, err := json.MarshalIndent(s.data, "", "\t")
	if err != nil {
		return fmt.Errorf("failed to marshal storage: %w", err)
	}

	if err := os.WriteFile(s.path, bytes, 0644); err != nil {
		return fmt.Errorf("failed to write storage file: %w", err)
	}

	return nil
}

// Get retrieves a value from storage
func (s *Storage) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	val, ok := s.data[key]
	return val, ok
}

// GetString retrieves a string value
func (s *Storage) GetString(key string) string {
	val, ok := s.Get(key)
	if !ok {
		return ""
	}
	if str, ok := val.(string); ok {
		return str
	}
	return ""
}

// GetMap retrieves a map value
func (s *Storage) GetMap(key string) map[string]interface{} {
	val, ok := s.Get(key)
	if !ok {
		return nil
	}
	if m, ok := val.(map[string]interface{}); ok {
		return m
	}
	return nil
}

// GetSlice retrieves a slice value
func (s *Storage) GetSlice(key string) []interface{} {
	val, ok := s.Get(key)
	if !ok {
		return nil
	}
	if slice, ok := val.([]interface{}); ok {
		return slice
	}
	return nil
}

// Set stores a value
func (s *Storage) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
}

// SetAndSave stores a value and immediately saves to disk
func (s *Storage) SetAndSave(key string, value interface{}) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.data[key] = value
	return s.saveUnsafe()
}

// Delete removes a key from storage
func (s *Storage) Delete(key string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.data, key)
}

// All returns all data (read-only copy)
func (s *Storage) All() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to prevent external modification
	copy := make(map[string]interface{})
	for k, v := range s.data {
		copy[k] = v
	}
	return copy
}

// Raw returns the raw data reference (use with caution)
func (s *Storage) Raw() map[string]interface{} {
	return s.data
}

// GetNested retrieves a nested value using dot notation
// e.g., GetNested("settings", "api_key")
func (s *Storage) GetNested(keys ...string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var current interface{} = s.data
	for _, key := range keys {
		m, ok := current.(map[string]interface{})
		if !ok {
			return nil, false
		}
		current, ok = m[key]
		if !ok {
			return nil, false
		}
	}
	return current, true
}

// SetNested sets a nested value, creating intermediate maps as needed
func (s *Storage) SetNested(value interface{}, keys ...string) {
	if len(keys) == 0 {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	current := s.data
	for i := 0; i < len(keys)-1; i++ {
		key := keys[i]
		if _, ok := current[key]; !ok {
			current[key] = make(map[string]interface{})
		}
		if m, ok := current[key].(map[string]interface{}); ok {
			current = m
		} else {
			current[key] = make(map[string]interface{})
			current = current[key].(map[string]interface{})
		}
	}
	current[keys[len(keys)-1]] = value
}
