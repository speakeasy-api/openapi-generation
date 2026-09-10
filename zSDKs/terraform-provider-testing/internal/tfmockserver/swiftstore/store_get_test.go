package swiftstore

import (
	"testing"
)

func TestGetValidID(t *testing.T) {
	store := New()

	// Add data to the store
	data := map[string]interface{}{
		"name": "John Doe",
		"age":  30,
		"city": "New York",
	}

	id, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Retrieve the data
	retrievedData, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	// Verify all fields are present
	if retrievedData["name"] != "John Doe" {
		t.Fatalf("Expected name 'John Doe', got %v", retrievedData["name"])
	}

	if retrievedData["age"] != 30 {
		t.Fatalf("Expected age 30, got %v", retrievedData["age"])
	}

	if retrievedData["city"] != "New York" {
		t.Fatalf("Expected city 'New York', got %v", retrievedData["city"])
	}

	// Verify 3 fields total
	if len(retrievedData) != 3 {
		t.Fatalf("Expected 3 fields, got %d", len(retrievedData))
	}
}

func TestGetNestedData(t *testing.T) {
	store := New()

	// Add nested data to the store
	data := map[string]interface{}{
		"user": map[string]interface{}{
			"name": "Jane Doe",
			"address": map[string]interface{}{
				"street": "123 Main St",
				"city":   "Anytown",
				"zip":    "12345",
			},
		},
		"tags": []string{"admin", "user", "premium"},
		"settings": map[string]interface{}{
			"notifications": true,
			"theme":         "dark",
		},
	}

	id, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Retrieve the data
	retrievedData, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	// Verify nested user data
	user, ok := retrievedData["user"].(map[string]interface{})
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

	if address["street"] != "123 Main St" {
		t.Fatalf("Expected street '123 Main St', got %v", address["street"])
	}

	if address["city"] != "Anytown" {
		t.Fatalf("Expected city 'Anytown', got %v", address["city"])
	}

	if address["zip"] != "12345" {
		t.Fatalf("Expected zip '12345', got %v", address["zip"])
	}

	// Verify tags array
	tags, ok := retrievedData["tags"].([]string)
	if !ok {
		t.Fatal("tags field is not a slice")
	}

	if len(tags) != 3 {
		t.Fatalf("Expected 3 tags, got %d", len(tags))
	}

	if tags[0] != "admin" || tags[1] != "user" || tags[2] != "premium" {
		t.Fatalf("Expected tags ['admin', 'user', 'premium'], got %v", tags)
	}

	// Verify settings
	settings, ok := retrievedData["settings"].(map[string]interface{})
	if !ok {
		t.Fatal("settings field is not a map")
	}

	if settings["notifications"] != true {
		t.Fatalf("Expected notifications true, got %v", settings["notifications"])
	}

	if settings["theme"] != "dark" {
		t.Fatalf("Expected theme 'dark', got %v", settings["theme"])
	}
}

func TestGetEmptyData(t *testing.T) {
	store := New()

	// Add empty data to the store
	emptyData := map[string]interface{}{}

	id, err := store.Add(emptyData)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Retrieve the empty data
	retrievedData, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	// Verify empty data is returned
	if len(retrievedData) != 0 {
		t.Fatalf("Expected empty data, got %d fields", len(retrievedData))
	}
}

func TestGetIDNotFound(t *testing.T) {
	store := New()

	// Try to get non-existent ID
	nonExistentID := "12345678-1234-1234-1234-123456789abc"

	retrievedData, err := store.Get(nonExistentID)

	if err == nil {
		t.Fatal("Get() should return error for non-existent ID")
	}

	if retrievedData != nil {
		t.Fatalf("Get() should return nil data for non-existent ID, got %v", retrievedData)
	}

	expectedError := "entry with ID " + nonExistentID + " not found"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s', got '%v'", expectedError, err)
	}
}

func TestGetEmptyID(t *testing.T) {
	store := New()

	retrievedData, err := store.Get("")

	if err == nil {
		t.Fatal("Get() should return error for empty ID")
	}

	if retrievedData != nil {
		t.Fatalf("Get() should return nil data for empty ID, got %v", retrievedData)
	}

	if err.Error() != "cannot get with empty ID" {
		t.Fatalf("Expected error 'cannot get with empty ID', got '%v'", err)
	}
}

func TestGetReturnsCopy(t *testing.T) {
	store := New()

	// Add data with nested structures
	data := map[string]interface{}{
		"nested": map[string]interface{}{
			"value": "original",
			"deep": map[string]interface{}{
				"level": 1,
			},
		},
		"array": []string{"item1", "item2"},
	}

	id, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Retrieve the data
	retrievedData, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	// Modify the retrieved data
	nested, ok := retrievedData["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("nested field is not a map")
	}
	nested["value"] = "modified"

	deep, ok := nested["deep"].(map[string]interface{})
	if !ok {
		t.Fatal("deep field is not a map")
	}
	deep["level"] = 999

	// Modify array
	array, ok := retrievedData["array"].([]string)
	if !ok {
		t.Fatal("array field is not a slice")
	}
	array[0] = "modified_item"

	// Verify stored data is not affected
	storedData := store.Data[id]
	storedNested, ok := storedData["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("stored nested field is not a map")
	}

	if storedNested["value"] != "original" {
		t.Fatal("Stored nested value should not be affected by modifications to retrieved data")
	}

	storedDeep, ok := storedNested["deep"].(map[string]interface{})
	if !ok {
		t.Fatal("stored deep field is not a map")
	}

	if storedDeep["level"] != 1 {
		t.Fatal("Stored deep level should not be affected by modifications to retrieved data")
	}

	storedArray, ok := storedData["array"].([]string)
	if !ok {
		t.Fatal("stored array field is not a slice")
	}

	if storedArray[0] != "item1" {
		t.Fatal("Stored array should not be affected by modifications to retrieved data")
	}
}

func TestGetConcurrent(t *testing.T) {
	store := New()

	// Add data
	data := map[string]interface{}{
		"name":  "concurrent test",
		"value": 42,
	}

	id, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	const numGoroutines = 100
	results := make(chan struct {
		data map[string]interface{}
		err  error
	}, numGoroutines)

	// Get data concurrently
	for i := 0; i < numGoroutines; i++ {
		go func() {
			retrievedData, err := store.Get(id)
			results <- struct {
				data map[string]interface{}
				err  error
			}{retrievedData, err}
		}()
	}

	// Collect results
	for i := 0; i < numGoroutines; i++ {
		result := <-results

		if result.err != nil {
			t.Fatalf("Get() returned error: %v", result.err)
		}

		if result.data == nil {
			t.Fatal("Get() returned nil data")
		}

		if result.data["name"] != "concurrent test" {
			t.Fatalf("Expected name 'concurrent test', got %v", result.data["name"])
		}

		if result.data["value"] != 42 {
			t.Fatalf("Expected value 42, got %v", result.data["value"])
		}
	}
}

func TestGetAfterUpdate(t *testing.T) {
	store := New()

	// Add initial data
	initialData := map[string]interface{}{
		"name": "Original Name",
		"age":  25,
	}

	id, err := store.Add(initialData)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Get initial data
	initialRetrieved, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get() returned error: %v", err)
	}

	if initialRetrieved["name"] != "Original Name" {
		t.Fatalf("Expected initial name 'Original Name', got %v", initialRetrieved["name"])
	}

	// Update the data
	updateData := map[string]interface{}{
		"name": "Updated Name",
		"city": "New City",
	}

	_, err = store.Update(id, updateData)
	if err != nil {
		t.Fatalf("Update() returned error: %v", err)
	}

	// Get updated data
	updatedRetrieved, err := store.Get(id)
	if err != nil {
		t.Fatalf("Get() after update returned error: %v", err)
	}

	if updatedRetrieved["name"] != "Updated Name" {
		t.Fatalf("Expected updated name 'Updated Name', got %v", updatedRetrieved["name"])
	}

	if updatedRetrieved["age"] != 25 {
		t.Fatalf("Expected age 25 to be preserved, got %v", updatedRetrieved["age"])
	}

	if updatedRetrieved["city"] != "New City" {
		t.Fatalf("Expected new city 'New City', got %v", updatedRetrieved["city"])
	}

	if len(updatedRetrieved) != 3 {
		t.Fatalf("Expected 3 fields after update, got %d", len(updatedRetrieved))
	}
}

func TestGetAfterDelete(t *testing.T) {
	store := New()

	// Add data
	data := map[string]interface{}{
		"name":  "Test Data",
		"value": 100,
	}

	id, err := store.Add(data)
	if err != nil {
		t.Fatalf("Add() returned error: %v", err)
	}

	// Verify data exists
	_, err = store.Get(id)
	if err != nil {
		t.Fatalf("Get() before delete returned error: %v", err)
	}

	// Delete the data
	err = store.Delete(id)
	if err != nil {
		t.Fatalf("Delete() returned error: %v", err)
	}

	// Try to get deleted data
	_, err = store.Get(id)
	if err == nil {
		t.Fatal("Get() should return error for deleted ID")
	}

	expectedError := "entry with ID " + id + " not found"
	if err.Error() != expectedError {
		t.Fatalf("Expected error '%s', got '%v'", expectedError, err)
	}
}

func TestGetMultipleEntries(t *testing.T) {
	store := New()

	// Add multiple entries
	data1 := map[string]interface{}{"key1": "value1", "id": 1}
	data2 := map[string]interface{}{"key2": "value2", "id": 2}
	data3 := map[string]interface{}{"key3": "value3", "id": 3}

	id1, err := store.Add(data1)
	if err != nil {
		t.Fatalf("Add() data1 returned error: %v", err)
	}

	id2, err := store.Add(data2)
	if err != nil {
		t.Fatalf("Add() data2 returned error: %v", err)
	}

	id3, err := store.Add(data3)
	if err != nil {
		t.Fatalf("Add() data3 returned error: %v", err)
	}

	// Get all entries
	retrieved1, err := store.Get(id1)
	if err != nil {
		t.Fatalf("Get() data1 returned error: %v", err)
	}

	retrieved2, err := store.Get(id2)
	if err != nil {
		t.Fatalf("Get() data2 returned error: %v", err)
	}

	retrieved3, err := store.Get(id3)
	if err != nil {
		t.Fatalf("Get() data3 returned error: %v", err)
	}

	// Verify each entry has correct data
	if retrieved1["key1"] != "value1" || retrieved1["id"] != 1 {
		t.Fatalf("Data1 mismatch: expected key1='value1', id=1, got key1='%v', id=%v", retrieved1["key1"], retrieved1["id"])
	}

	if retrieved2["key2"] != "value2" || retrieved2["id"] != 2 {
		t.Fatalf("Data2 mismatch: expected key2='value2', id=2, got key2='%v', id=%v", retrieved2["key2"], retrieved2["id"])
	}

	if retrieved3["key3"] != "value3" || retrieved3["id"] != 3 {
		t.Fatalf("Data3 mismatch: expected key3='value3', id=3, got key3='%v', id=%v", retrieved3["key3"], retrieved3["id"])
	}

	// Verify all entries are independent
	if len(retrieved1) != 2 || len(retrieved2) != 2 || len(retrieved3) != 2 {
		t.Fatal("All entries should have exactly 2 fields")
	}
}
