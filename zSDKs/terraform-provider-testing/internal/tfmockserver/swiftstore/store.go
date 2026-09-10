package swiftstore

import (
	"fmt"
	"sync"

	"github.com/google/uuid"
)

// Store represents an in-memory data store for testing purposes
type Store struct {
	Data  map[string]map[string]interface{}
	mutex sync.RWMutex
}

// New creates and returns a new empty Store instance
func New() *Store {
	return &Store{
		Data: make(map[string]map[string]interface{}),
	}
}

// Add - stores the provided data and returns a unique ID
// Returns an error if the input data is invalid (e.g., nil)
func (s *Store) Add(data map[string]interface{}) (string, error) {
	if data == nil {
		return "", fmt.Errorf("cannot add nil data to store")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if store is destroyed
	if s.Data == nil {
		return "", fmt.Errorf("store is destroyed")
	}

	// Generate a unique UUID v4
	id := uuid.New().String()

	// Create a deep copy of the data to store
	dataCopy := make(map[string]interface{})
	for k, v := range data {
		dataCopy[k] = deepCopyValue(v)
	}

	// Store the data
	s.Data[id] = dataCopy

	return id, nil
}

// Get retrieves data from the store by ID
// Returns a copy of the data and an error if the ID is not found
func (s *Store) Get(id string) (map[string]interface{}, error) {
	if id == "" {
		return nil, fmt.Errorf("cannot get with empty ID")
	}

	s.mutex.RLock()
	defer s.mutex.RUnlock()

	// Check if store is destroyed
	if s.Data == nil {
		return nil, fmt.Errorf("store is destroyed")
	}

	// Check if the ID exists
	data, exists := s.Data[id]
	if !exists {
		return nil, fmt.Errorf("entry with ID %s not found", id)
	}

	// Return a deep copy of the data
	result := make(map[string]interface{})
	for k, v := range data {
		result[k] = deepCopyValue(v)
	}

	return result, nil
}

// Update performs a partial update (merge) on existing data in the store
// Returns the final merged object and an error if the ID is not found
func (s *Store) Update(id string, data map[string]interface{}) (map[string]interface{}, error) {
	if id == "" {
		return nil, fmt.Errorf("cannot update with empty ID")
	}

	if data == nil {
		return nil, fmt.Errorf("cannot update with nil data")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if store is destroyed
	if s.Data == nil {
		return nil, fmt.Errorf("store is destroyed")
	}

	// Check if the ID exists
	existingData, exists := s.Data[id]
	if !exists {
		return nil, fmt.Errorf("entry with ID %s not found", id)
	}

	// Create a deep copy of existing data to merge into
	mergedData := make(map[string]interface{})
	for k, v := range existingData {
		mergedData[k] = deepCopyValue(v)
	}

	// Merge the update data (overwrite existing fields, add new ones)
	for k, v := range data {
		mergedData[k] = deepCopyValue(v)
	}

	// Store the merged data
	s.Data[id] = mergedData

	// Return a deep copy of the merged data
	result := make(map[string]interface{})
	for k, v := range mergedData {
		result[k] = deepCopyValue(v)
	}

	return result, nil
}

// deepCopyValue creates a deep copy of interface{} values for nested structures
func deepCopyValue(value interface{}) interface{} {
	switch v := value.(type) {
	case map[string]interface{}:
		copy := make(map[string]interface{})
		for key, val := range v {
			copy[key] = deepCopyValue(val)
		}
		return copy
	case []interface{}:
		localCopy := make([]interface{}, len(v))
		for i, val := range v {
			localCopy[i] = deepCopyValue(val)
		}
		return localCopy
	case []string:
		localCopy := make([]string, len(v))
		copy(localCopy, v)
		return localCopy
	case []int:
		localCopy := make([]int, len(v))
		copy(localCopy, v)
		return localCopy
	case []float64:
		localCopy := make([]float64, len(v))
		copy(localCopy, v)
		return localCopy
	case []bool:
		localCopy := make([]bool, len(v))
		copy(localCopy, v)
		return localCopy
	default:
		return v
	}
}

// Delete removes an entry from the store by ID
// Returns an error if the ID is not found
func (s *Store) Delete(id string) error {
	if id == "" {
		return fmt.Errorf("cannot delete with empty ID")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if store is destroyed
	if s.Data == nil {
		return fmt.Errorf("store is destroyed")
	}

	// Check if the ID exists
	if _, exists := s.Data[id]; !exists {
		return fmt.Errorf("entry with ID %s not found", id)
	}

	// Delete the entry
	delete(s.Data, id)

	return nil
}

// AddWithID stores the provided data with a custom ID
// Returns an error if the input data is invalid or if the ID conflicts with an existing entry
func (s *Store) AddWithID(id string, data map[string]interface{}) error {
	if id == "" {
		return fmt.Errorf("cannot add with empty ID")
	}

	if data == nil {
		return fmt.Errorf("cannot add nil data to store")
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if store is destroyed
	if s.Data == nil {
		return fmt.Errorf("store is destroyed")
	}

	// Check if the ID already exists
	if _, exists := s.Data[id]; exists {
		return fmt.Errorf("entry with ID %s already exists", id)
	}

	// Create a deep copy of the data to store
	dataCopy := make(map[string]interface{})
	for k, v := range data {
		dataCopy[k] = deepCopyValue(v)
	}

	// Store the data with the custom ID
	s.Data[id] = dataCopy

	return nil
}

// Destroy cleans up the store and releases all resources
// Returns an error if called more than once
func (s *Store) Destroy() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Check if already destroyed
	if s.Data == nil {
		return fmt.Errorf("store already destroyed")
	}

	// Clear all data
	s.Data = nil

	return nil
}
