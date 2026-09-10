package swiftstore

import (
	"fmt"
	"testing"
)

func TestDeleteValidID(t *testing.T) {
	store := New()

	// Add some data first
	data := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
	}

	id, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Verify data exists
	if len(store.Data) != 1 {
		t.Fatalf("Expected 1 entry before delete, got %d", len(store.Data))
	}

	// Delete the data
	err = store.Delete(id)
	if err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	// Verify data was deleted
	if len(store.Data) != 0 {
		t.Fatalf("Expected 0 entries after delete, got %d", len(store.Data))
	}

	// Verify the specific ID no longer exists
	if _, exists := store.Data[id]; exists {
		t.Fatal("Data should not exist after deletion")
	}
}

func TestDeleteIDNotFound(t *testing.T) {
	store := New()

	// Try to delete a non-existent ID
	nonExistentID := "12345678-1234-1234-1234-123456789abc"
	err := store.Delete(nonExistentID)

	if err == nil {
		t.Fatal("Delete() should return error for non-existent ID")
	}

	expectedError := "entry with ID " + nonExistentID + " not found"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s', got '%v'", expectedError, err)
	}

	// Verify store is still empty
	if len(store.Data) != 0 {
		t.Fatal("Store should remain empty after failed delete")
	}
}

func TestDeleteEmptyID(t *testing.T) {
	store := New()

	// Try to delete with empty ID
	err := store.Delete("")

	if err == nil {
		t.Fatal("Delete() should return error for empty ID")
	}

	if err.Error() != "cannot delete with empty ID" {
		t.Fatalf("Expected error 'cannot delete with empty ID', got '%v'", err)
	}

	// Verify store is still empty
	if len(store.Data) != 0 {
		t.Fatal("Store should remain empty after failed delete")
	}
}

func TestDeleteMultipleEntries(t *testing.T) {
	store := New()

	// Add multiple entries
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

	// Verify all entries exist
	if len(store.Data) != 3 {
		t.Fatalf("Expected 3 entries, got %d", len(store.Data))
	}

	// Delete one entry
	err = store.Delete(id2)
	if err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	// Verify only one entry was deleted
	if len(store.Data) != 2 {
		t.Fatalf("Expected 2 entries after delete, got %d", len(store.Data))
	}

	// Verify the deleted entry no longer exists
	if _, exists := store.Data[id2]; exists {
		t.Fatal("Deleted entry should not exist")
	}

	// Verify other entries still exist
	if _, exists := store.Data[id1]; !exists {
		t.Fatal("First entry should still exist")
	}

	if _, exists := store.Data[id3]; !exists {
		t.Fatal("Third entry should still exist")
	}
}

func TestDeleteConcurrent(t *testing.T) {
	store := New()
	const numEntries = 50

	// Add multiple entries first
	ids := make([]string, numEntries)
	for i := 0; i < numEntries; i++ {
		data := map[string]interface{}{
			"index": i,
			"value": fmt.Sprintf("value_%d", i),
		}

		id, err := store.Add(data)
		if err != nil {
			t.Fatalf("Add() returned error: %v", err)
		}

		ids[i] = id
	}

	// Verify all entries were added
	if len(store.Data) != numEntries {
		t.Fatalf("Expected %d entries before delete, got %d", numEntries, len(store.Data))
	}

	// Delete entries concurrently
	const numGoroutines = 25
	results := make(chan error, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			// Delete every other entry
			if index < len(ids) {
				err := store.Delete(ids[index])
				results <- err
			} else {
				results <- nil
			}
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err != nil {
			t.Fatalf("Delete() returned error: %v", err)
		}
	}

	// Verify entries were deleted
	expectedRemaining := numEntries - numGoroutines
	if len(store.Data) != expectedRemaining {
		t.Fatalf("Expected %d entries after concurrent delete, got %d", expectedRemaining, len(store.Data))
	}
}

func TestDeleteAddDeleteAdd(t *testing.T) {
	store := New()

	// Add data
	data := map[string]interface{}{"test": "value"}
	id, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Delete data
	err = store.Delete(id)
	if err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	// Add data again (should work fine)
	newData := map[string]interface{}{"new": "data"}
	newID, err := store.Add(newData)
	if err != nil {
		t.Fatalf("Add() after delete returned error: %v", err)
	}

	if newID == "" {
		t.Fatal("Add() after delete returned empty ID")
	}

	// Verify the new data is stored
	if len(store.Data) != 1 {
		t.Fatalf("Expected 1 entry after add-delete-add, got %d", len(store.Data))
	}

	// Verify the new ID is different from the deleted one
	if newID == id {
		t.Fatal("New ID should be different from deleted ID")
	}
}
