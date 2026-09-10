package swiftstore

import (
	"testing"

	"github.com/google/uuid"
)

func TestNew(t *testing.T) {
	store := New()

	if store == nil {
		t.Fatal("New() returned nil")
	}

	if store.Data == nil {
		t.Fatal("store.data is nil")
	}

	if len(store.Data) != 0 {
		t.Fatal("store.data should be empty")
	}
}

func TestAddValidData(t *testing.T) {
	store := New()

	data := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	id, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	if id == "" {
		t.Fatal("Add() returned empty ID")
	}

	// Verify data is stored
	if len(store.Data) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(store.Data))
	}

	storedData, exists := store.Data[id]
	if !exists {
		t.Fatal("Data not found with returned ID")
	}

	// Verify data content
	if storedData["name"] != "John Doe" {
		t.Fatalf("Expected name 'John Doe', got %v", storedData["name"])
	}

	if storedData["age"] != 30 {
		t.Fatalf("Expected age 30, got %v", storedData["age"])
	}
}

func TestAddNestedData(t *testing.T) {
	store := New()

	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Jane Doe",
			"address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "Anytown",
			},
		},
		"tags": []string{"admin", "user"},
	}

	id, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Verify nested data is stored correctly
	storedData := store.Data[id]

	user, ok := storedData["user"].(map[string]interface{})
	if !ok {
		t.Fatal("user field is not a map")
	}

	if user["name"] != "Jane Doe" {
		t.Fatalf("Expected user name 'Jane Doe', got %v", user["name"])
	}

	address, ok := user["address"].(map[string]interface{})
	if !ok {
		t.Fatal("address field is not a map")
	}

	if address["city"] != "Anytown" {
		t.Fatalf("Expected city 'Anytown', got %v", address["city"])
	}

	tags, ok := storedData["tags"].([]string)
	if !ok {
		t.Fatal("tags field is not a slice")
	}

	if len(tags) != 2 || tags[0] != "admin" || tags[1] != "user" {
		t.Fatalf("Expected tags ['admin', 'user'], got %v", tags)
	}
}

func TestAddEmptyData(t *testing.T) {
	store := New()

	emptyData := map[string]interface{}{}

	id, err := store.Add(emptyData)
	if err != nil {
		t.Fatalf("Add() returned error for empty data: %v", err)
	}

	if id == "" {
		t.Fatal("Add() returned empty ID for empty data")
	}

	// Verify empty data is stored
	if len(store.Data) != 1 {
		t.Fatalf("Expected 1 entry, got %d", len(store.Data))
	}

	storedData := store.Data[id]
	if len(storedData) != 0 {
		t.Fatalf("Expected empty stored data, got %v", storedData)
	}
}

func TestAddNilData(t *testing.T) {
	store := New()

	id, err := store.Add(nil)
	if err == nil {
		t.Fatal("Add() should return error for nil data")
	}

	if id != "" {
		t.Fatalf("Add() should return empty ID for nil data, got %s", id)
	}

	if err.Error() != "cannot add nil data to store" {
		t.Fatalf("Expected error 'cannot add nil data to store', got %v", err)
	}

	// Verify no data was stored
	if len(store.Data) != 0 {
		t.Fatal("No data should be stored when Add() fails")
	}
}

func TestAddUniqueIDs(t *testing.T) {
	store := New()

	data1 := map[string]interface{}{"key1": "value1"}
	data2 := map[string]interface{}{"key2": "value2"}
	data3 := map[string]interface{}{"key3": "value3"}

	id1, err := store.Add(data1)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	id2, err := store.Add(data2)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	id3, err := store.Add(data3)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Verify all IDs are unique
	if id1 == id2 || id1 == id3 || id2 == id3 {
		t.Fatal("Generated IDs should be unique")
	}

	// Verify all data is stored
	if len(store.Data) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(store.Data))
	}
}

func TestAddUUIDFormat(t *testing.T) {
	store := New()

	data := map[string]interface{}{"test": "value"}

	id, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Verify ID is valid UUID v4
	_, err = uuid.Parse(id)
	if err != nil {
		t.Fatalf("Generated ID is not a valid UUID: %v", err)
	}
}

func TestAddConcurrent(t *testing.T) {
	store := New()
	const numGoroutines = 100

	// Channel to collect results
	results := make(chan struct {
		id   string
		err  error
		data map[string]interface{}
	}, numGoroutines)

	// Launch concurrent goroutines
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			data := map[string]interface{}{
				"index": index,
				"data":  "concurrent test",
			}

			id, err := store.Add(data)
			results <- struct {
				id   string
				err  error
				data map[string]interface{}
			}{id, err, data}
		}(i)
	}

	// Collect all results
	collectedIDs := make(map[string]bool)
	collectedData := make(map[string]map[string]interface{})

	for i := 0; i < numGoroutines; i++ {
		result := <-results

		if result.err != nil {
			t.Fatalf("Goroutine %d returned error: %v", i, result.err)
		}

		if result.id == "" {
			t.Fatalf("Goroutine %d returned empty ID", i)
		}

		// Check for duplicate IDs
		if collectedIDs[result.id] {
			t.Fatalf("Duplicate ID generated: %s", result.id)
		}

		collectedIDs[result.id] = true
		collectedData[result.id] = result.data
	}

	// Verify all data was stored correctly
	if len(store.Data) != numGoroutines {
		t.Fatalf("Expected %d entries, got %d", numGoroutines, len(store.Data))
	}

	// Verify each stored entry matches the collected data
	for id, expectedData := range collectedData {
		storedData, exists := store.Data[id]
		if !exists {
			t.Fatalf("Data not found for ID: %s", id)
		}

		if storedData["index"] != expectedData["index"] {
			t.Fatalf("Data mismatch for ID %s: expected index %v, got %v",
				id, expectedData["index"], storedData["index"])
		}
	}
}
