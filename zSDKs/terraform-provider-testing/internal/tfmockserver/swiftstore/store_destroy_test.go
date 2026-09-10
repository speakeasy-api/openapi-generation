package swiftstore

import (
	"testing"
)

func TestDestroyEmptyStore(t *testing.T) {
	store := New()

	// Verify store is initially empty
	if len(store.Data) != 0 {
		t.Fatalf("Expected empty store, got %d entries", len(store.Data))
	}

	// Destroy the store
	err := store.Destroy()
	if err != nil {
		t.Fatalf("Destroy() returned error: %v", err)
	}

	// Verify store data is nil
	if store.Data != nil {
		t.Fatal("Store data should be nil after destroy")
	}
}

func TestDestroyStoreWithData(t *testing.T) {
	store := New()

	// Add some data to the store
	data1 := map[string]interface{}{"name": "John Doe"}
	data2 := map[string]interface{}{"name": "Jane Doe"}

	_, err := store.Add(data1)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	err = store.AddWithID("custom-123", data2)
	if err != nil {
		t.Fatalf("AddWithID() returned error: %v", err)
	}

	// Verify data is stored
	if len(store.Data) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(store.Data))
	}

	// Destroy the store
	err = store.Destroy()
	if err != nil {
		t.Fatalf("Destroy() returned error: %v", err)
	}

	// Verify store data is nil
	if store.Data != nil {
		t.Fatal("Store data should be nil after destroy")
	}
}

func TestDestroyMultipleCalls(t *testing.T) {
	store := New()

	// Add some data
	data := map[string]interface{}{"name": "Test User"}
	_, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// First destroy should succeed
	err = store.Destroy()
	if err != nil {
		t.Fatalf("First Destroy() returned error: %v", err)
	}

	// Second destroy should return error
	err = store.Destroy()
	if err == nil {
		t.Fatal("Second Destroy() should return error")
	}

	if err.Error() != "store already destroyed" {
		t.Fatalf("Expected error 'store already destroyed', got %v", err)
	}
}

func TestDestroyAfterOperations(t *testing.T) {
	store := New()

	// Add data with auto-generated ID
	data1 := map[string]interface{}{"name": "Auto User"}
	id1, err := store.Add(data1)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Add data with custom ID
	data2 := map[string]interface{}{"name": "Custom User"}
	err = store.AddWithID("custom-456", data2)
	if err != nil {
		t.Fatalf("AddWithID() returned error: %v", err)
	}

	// Update data
	updateData := map[string]interface{}{"age": 25}
	_, err = store.Update(id1, updateData)
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	// Get data to verify it exists
	_, err = store.Get(id1)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	// Verify store has data
	if len(store.Data) != 2 {
		t.Fatalf("Expected 2 entries, got %d", len(store.Data))
	}

	// Destroy the store
	err = store.Destroy()
	if err != nil {
		t.Fatalf("Destroy() returned error: %v", err)
	}

	// Verify store data is nil
	if store.Data != nil {
		t.Fatal("Store data should be nil after destroy")
	}

	// Verify operations fail after destroy
	_, err = store.Get(id1)
	if err == nil {
		t.Fatal("Get() should fail after destroy")
	}

	_, err = store.Update(id1, updateData)
	if err == nil {
		t.Fatal("Update() should fail after destroy")
	}

	err = store.Delete(id1)
	if err == nil {
		t.Fatal("Delete() should fail after destroy")
	}

	_, err = store.Add(data1)
	if err == nil {
		t.Fatal("Add() should fail after destroy")
	}

	err = store.AddWithID("new-id", data1)
	if err == nil {
		t.Fatal("AddWithID() should fail after destroy")
	}
}

func TestDestroyConcurrent(t *testing.T) {
	store := New()

	// Add some data
	data := map[string]interface{}{"name": "Concurrent User"}
	_, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Launch goroutine to destroy store
	destroyDone := make(chan error, 1)
	go func() {
		destroyDone <- store.Destroy()
	}()

	// Wait for destroy to complete
	err = <-destroyDone
	if err != nil {
		t.Fatalf("Destroy() returned error: %v", err)
	}

	// Verify store is destroyed
	if store.Data != nil {
		t.Fatal("Store data should be nil after destroy")
	}
}
