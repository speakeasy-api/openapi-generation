package swiftstore

import (
	"fmt"
	"testing"
)

func TestUpdateValidID(t *testing.T) {
	store := New()

	// Add initial data
	initialData := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
		"city": "New York",
	}

	id, err := store.Add(initialData)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Update some fields
	updateData := map[string]interface{}{
		"age":  31,
		"city": "Boston",
	}

	mergedData, err := store.Update(id, updateData)
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	// Verify merged data contains updated fields
	if mergedData["age"] != 31 {
		t.Fatalf("Expected age 31, got %v", mergedData["age"])
	}

	if mergedData["city"] != "Boston" {
		t.Fatalf("Expected city 'Boston', got %v", mergedData["city"])
	}

	// Verify unchanged field is preserved
	if mergedData["name"] != "John Doe" {
		t.Fatalf("Expected name 'John Doe', got %v", mergedData["name"])
	}

	// Verify data is actually stored in the store
	storedData := store.Data[id]
	if storedData["age"] != 31 {
		t.Fatalf("Expected stored age 31, got %v", storedData["age"])
	}

	if storedData["name"] != "John Doe" {
		t.Fatalf("Expected stored name 'John Doe', got %v", storedData["name"])
	}
}

func TestUpdateAddNewFields(t *testing.T) {
	store := New()

	// Add initial data
	initialData := map[string]interface{}{
		"name": "Jane Doe",
		"age":  25,
	}

	id, err := store.Add(initialData)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Update with new fields
	updateData := map[string]interface{}{
		"city":  "Seattle",
		"email": "jane@example.com",
	}

	mergedData, err := store.Update(id, updateData)
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	// Verify all fields are present
	if mergedData["name"] != "Jane Doe" {
		t.Fatalf("Expected name 'Jane Doe', got %v", mergedData["name"])
	}

	if mergedData["age"] != 25 {
		t.Fatalf("Expected age 25, got %v", mergedData["age"])
	}

	if mergedData["city"] != "Seattle" {
		t.Fatalf("Expected city 'Seattle', got %v", mergedData["city"])
	}

	if mergedData["email"] != "jane@example.com" {
		t.Fatalf("Expected email 'jane@example.com', got %v", mergedData["email"])
	}

	// Verify 4 fields total
	if len(mergedData) != 4 {
		t.Fatalf("Expected 4 fields, got %d", len(mergedData))
	}
}

func TestUpdateNestedData(t *testing.T) {
	store := New()

	// Add initial nested data
	initialData := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Bob Smith",
			"age":  35,
		},
		"preferences": map[string]interface{}{
			"theme": "dark",
			"lang":  "en",
		},
	}

	id, err := store.Add(initialData)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Update nested data
	updateData := map[string]interface{}{
		"user": map[string]interface{}{
			"age":   36,
			"email": "bob@example.com",
		},
		"preferences": map[string]interface{}{
			"theme": "light",
		},
	}

	mergedData, err := store.Update(id, updateData)
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	// Verify nested updates - note: this completely replaces nested objects
	user, ok := mergedData["user"].(map[string]interface{})
	if !ok {
		t.Fatal("user field is not a map")
	}

	if user["age"] != 36 {
		t.Fatalf("Expected user age 36, got %v", user["age"])
	}

	if user["email"] != "bob@example.com" {
		t.Fatalf("Expected user email 'bob@example.com', got %v", user["email"])
	}

	// Original name should be gone since we replaced the entire user object
	if _, exists := user["name"]; exists {
		t.Fatal("Original user name should not exist after nested update")
	}

	preferences, ok := mergedData["preferences"].(map[string]interface{})
	if !ok {
		t.Fatal("preferences field is not a map")
	}

	if preferences["theme"] != "light" {
		t.Fatalf("Expected theme 'light', got %v", preferences["theme"])
	}

	// Original lang should be gone since we replaced the entire preferences object
	if _, exists := preferences["lang"]; exists {
		t.Fatal("Original lang should not exist after nested update")
	}
}

func TestUpdateEmptyUpdate(t *testing.T) {
	store := New()

	// Add initial data
	initialData := map[string]interface{}{
		"name": "Alice",
		"age":  28,
	}

	id, err := store.Add(initialData)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Update with empty data
	updateData := map[string]interface{}{}

	mergedData, err := store.Update(id, updateData)
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	// Verify original data is unchanged
	if mergedData["name"] != "Alice" {
		t.Fatalf("Expected name 'Alice', got %v", mergedData["name"])
	}

	if mergedData["age"] != 28 {
		t.Fatalf("Expected age 28, got %v", mergedData["age"])
	}

	if len(mergedData) != 2 {
		t.Fatalf("Expected 2 fields, got %d", len(mergedData))
	}
}

func TestUpdateIDNotFound(t *testing.T) {
	store := New()

	// Try to update non-existent ID
	nonExistentID := "12345678-1234-1234-1234-123456789abc"
	updateData := map[string]interface{}{"name": "Test"}

	mergedData, err := store.Update(nonExistentID, updateData)

	if err == nil {
		t.Fatal("Update() should return error for non-existent ID")
	}

	if mergedData != nil {
		t.Fatalf("Update() should return nil data for non-existent ID, got %v", mergedData)
	}

	expectedError := "entry with ID " + nonExistentID + " not found"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s', got '%v'", expectedError, err)
	}
}

func TestUpdateEmptyID(t *testing.T) {
	store := New()

	updateData := map[string]interface{}{"name": "Test"}

	mergedData, err := store.Update("", updateData)

	if err == nil {
		t.Fatal("Update() should return error for empty ID")
	}

	if mergedData != nil {
		t.Fatalf("Update() should return nil data for empty ID, got %v", mergedData)
	}

	if err.Error() != "cannot update with empty ID" {
		t.Fatalf("Expected error 'cannot update with empty ID', got '%v'", err)
	}
}

func TestUpdateNilData(t *testing.T) {
	store := New()

	// Add initial data
	initialData := map[string]interface{}{"name": "Test"}
	id, err := store.Add(initialData)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Try to update with nil data
	mergedData, err := store.Update(id, nil)

	if err == nil {
		t.Fatal("Update() should return error for nil data")
	}

	if mergedData != nil {
		t.Fatalf("Update() should return nil data for nil update data, got %v", mergedData)
	}

	if err.Error() != "cannot update with nil data" {
		t.Fatalf("Expected error 'cannot update with nil data', got '%v'", err)
	}

	// Verify original data is unchanged
	storedData := store.Data[id]
	if storedData["name"] != "Test" {
		t.Fatal("Original data should be unchanged after failed update")
	}
}

func TestUpdateConcurrent(t *testing.T) {
	store := New()

	// Add initial data
	initialData := map[string]interface{}{
		"counter": 0,
		"name":    "concurrent test",
	}

	id, err := store.Add(initialData)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	const numGoroutines = 50
	results := make(chan error, numGoroutines)

	// Update concurrently
	for i := 0; i < numGoroutines; i++ {
		go func(index int) {
			updateData := map[string]interface{}{
				"counter": index,
				"updated": fmt.Sprintf("goroutine_%d", index),
			}

			_, err := store.Update(id, updateData)
			results <- err
		}(i)
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		err := <-results
		if err != nil {
			t.Fatalf("Update() returned error: %v", err)
		}
	}

	// Verify data was updated (some goroutine won)
	storedData := store.Data[id]
	if storedData["name"] != "concurrent test" {
		t.Fatal("Original name should be preserved")
	}

	// Counter should be one of the values from 0 to numGoroutines-1
	counter, ok := storedData["counter"].(int)
	if !ok {
		t.Fatal("Counter should be an integer")
	}

	if counter < 0 || counter >= numGoroutines {
		t.Fatalf("Counter should be between 0 and %d, got %d", numGoroutines-1, counter)
	}

	// Updated field should exist
	if _, exists := storedData["updated"]; !exists {
		t.Fatal("Updated field should exist")
	}
}

func TestUpdateReturnsCopy(t *testing.T) {
	store := New()

	// Add initial data
	initialData := map[string]interface{}{
		"nested": map[string]interface{}{
			"value": "original",
		},
	}

	id, err := store.Add(initialData)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Update data
	updateData := map[string]interface{}{
		"new_field": "new_value",
	}

	mergedData, err := store.Update(id, updateData)
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	// Modify the returned data
	nested, ok := mergedData["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("nested field is not a map")
	}
	nested["value"] = "modified"

	// Verify stored data is not affected
	storedData := store.Data[id]
	storedNested, ok := storedData["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("stored nested field is not a map")
	}

	if storedNested["value"] != "original" {
		t.Fatal("Stored data should not be affected by modifications to returned data")
	}
}

func TestUpdateMultipleUpdates(t *testing.T) {
	store := New()

	// Add initial data
	initialData := map[string]interface{}{
		"version": 1,
		"name":    "test",
	}

	id, err := store.Add(initialData)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// First update
	update1 := map[string]interface{}{
		"version": 2,
		"status":  "updated",
	}

	merged1, err := store.Update(id, update1)
	if err != nil {
		t.Fatalf("First Update() returned error: %v", err)
	}

	if merged1["version"] != 2 {
		t.Fatalf("Expected version 2, got %v", merged1["version"])
	}

	if merged1["name"] != "test" {
		t.Fatalf("Expected name 'test', got %v", merged1["name"])
	}

	if merged1["status"] != "updated" {
		t.Fatalf("Expected status 'updated', got %v", merged1["status"])
	}

	// Second update
	update2 := map[string]interface{}{
		"version": 3,
		"name":    "updated_test",
	}

	merged2, err := store.Update(id, update2)
	if err != nil {
		t.Fatalf("Second Update() returned error: %v", err)
	}

	if merged2["version"] != 3 {
		t.Fatalf("Expected version 3, got %v", merged2["version"])
	}

	if merged2["name"] != "updated_test" {
		t.Fatalf("Expected name 'updated_test', got %v", merged2["name"])
	}

	if merged2["status"] != "updated" {
		t.Fatalf("Expected status 'updated', got %v", merged2["status"])
	}

	if len(merged2) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(merged2))
	}
}
