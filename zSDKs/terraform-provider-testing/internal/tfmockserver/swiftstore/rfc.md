# RFC: In-Memory Store for HTTP Tests

## Summary

This RFC defines the requirements for creating a Go package that provides an in-memory data store (called swift-store). The package will be used primarily as a stateful backend to http servers. The http server will handle CRUD operations. 

## Motivation

When testing HTTP services, there's often a need for a lightweight, stateful backend that can store and retrieve data without the complexity of setting up a real database. This package will provide a simple in-memory store that can be easily integrated with Go's `net/http/httptest` package for testing purposes.

## Requirements

### Core Data Structure
- The package must provide a struct that acts as an in-memory data store
- Data entries are stored as key-value pairs (similar to JSON objects)
- Each entry must have a unique identifier (ID) generated automatically
- The store must be thread-safe for concurrent access

### Exposed methods

The package must support the following methods

swift-store should allow the user to add data

swift-store should allow the user to retrieve data
swift-store should allow the user to retrieve data
  - Overwrite existing fields if they exist in the swift-store already
  - Keep existing fields not mentioned in the request body
  - Return the final merged object

swift-store should allow the user to retrieve data
-  Delete the entry from the store
-  returns a success type to convey that the item was removed
-  returns an error type if the item to be deleted was not found
- returns an error type if there was an error deleting the item



### Technical Requirements

#### ID Generation
- each value item that is inserted into swift-store should have its own ID
- IDs must be unique across all entries
- Recommended format: UUID v4

#### Thread Safety
- The store must be safe for concurrent access
- Multiple goroutines should be able to read/write simultaneously without data races

#### Data Format
- All data should be stored as JSON-compatible structures
- Support for nested objects and arrays
- Preserve data types from JSON input

## Implementation Architecture

### Core Components

1. **Store Struct**: Main data structure holding the in-memory data
2. **Entry Type**: Flexible type to represent stored data (likely `map[string]interface{}`)
3. **ID Generator**: Utility for generating unique identifiers
4. **Thread Safety**: Mutex-based locking for concurrent access

### Package Interface

```go
type Store struct {
    // internal fields
}

func New() *Store
func (s *Store) ServeHTTP(w http.ResponseWriter, r *http.Request)
// Additional methods for direct programmatic access if needed
```

## Usage Example

```go
// Create a new store
store := New()

// Create an HTTP test server
server := httptest.NewServer(store)
defer server.Close()

// Test endpoints
// POST to server.URL + "/stateful/simple" should store the data in the store
// GET from server.URL + "/stateful/simple/{id}" should return the data that was stored in the POST request
// etc.
```

## Success Criteria

1. Data persistence within the lifetime of the store instance
2. Thread-safe concurrent operations
3. Easy integration with httptest
4. JSON request/response handling

## Test Structure and Terraform Integration

### Test File Structure

Terraform provider tests using swift-store follow a consistent pattern:

### Creating a Terraform Test with Swift-Store

#### 1. Basic Test Structure (Using TestHelpers)

```go
func TestResourceName_lifecycle_mockserver(t *testing.T) {
    t.Parallel()

    // Define endpoints for the resource
    endpoints := testhelpers.ResourceEndpoints{
        Create: "POST /v0/resource",
        Get:    "GET /v0/resource",
        Update: "PUT /v0/resource",
        Delete: "DELETE /v0/resource",
    }

    // Start mock server with custom endpoints
    mockServer := testhelpers.CreateMockServerWithEndpoints(endpoints)
    defer mockServer.Close()

    // Print server details
    t.Logf("🚀 [TEST] Mock server started at: %s", mockServer.URL)

    resourceAddress := "testing_resource_name.my_resource"

    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            // Create and Read test
            {
                ConfigDirectory: config.TestNameDirectory(),
                ProtoV6ProviderFactories: testhelpers.TestProviders(),
                ConfigVariables: config.Variables{
                    "server_url": config.StringVariable(mockServer.URL),
                },
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        resourceAddress,
                        tfjsonpath.New("id"),
                        knownvalue.StringRegexp(regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)),
                    ),
                    statecheck.ExpectKnownValue(
                        resourceAddress,
                        tfjsonpath.New("name"),
                        knownvalue.StringExact("test-resource"),
                    ),
                },
            },
            // Import test
            {
                ConfigDirectory: config.TestNameDirectory(),
                ProtoV6ProviderFactories: testhelpers.TestProviders(),
                ConfigVariables: config.Variables{
                    "server_url": config.StringVariable(mockServer.URL),
                },
                ResourceName: resourceAddress,
                ImportState:  true,
                ImportStateIdFunc: func(s *terraform.State) (string, error) {
                    // Custom import ID logic
                    return s.RootModule().Resources[resourceAddress].Primary.ID, nil
                },
                ImportStateVerify: true,
            },
        },
    })
}
```


#### 2. Key Test Components

**Test Structure:**
- Use `testhelpers.CreateMockServerWithEndpoints()` for mock server setup
- Pass server URL via `ConfigVariables` instead of environment variables
- Use full resource address format: `"resource_type.resource_name"`
- Include `ConfigStateChecks` for state verification

**Test Configuration:**
- Create `testdata/TestName_lifecycle_mockserver/main.tf` with resource definition

#### 3. Test Configuration Files

Create a `testdata/TestResourceName_lifecycle_mockserver/main.tf` file:

```hcl
resource "testing_resource_name" "my_resource" {
  name = "test-resource"
  # server_url is passed via ConfigVariables in the test
}

# Note: server_url is passed via ConfigVariables, not as a variable
# The provider will use the server_url from the test configuration
```

#### 4. Provider Configuration

```go
func testProviders() map[string]func() (tfprotov6.ProviderServer, error) {
    return map[string]func() (tfprotov6.ProviderServer, error){
        "testing": providerserver.NewProtocol6WithError(provider.New("test")()),
    }
}
```

### Best Practices

1. **Server URL Configuration**: Pass the mock server URL directly via `ConfigVariables` instead of using environment variables
2. **Resource Cleanup**: Use `defer` statements to ensure proper cleanup of mock servers
3. **State Verification**: Use `ConfigStateChecks` to verify computed values and expected resource state
4. **TestHelpers**: Use `testhelpers.CreateMockServerWithEndpoints()` for cleaner, more maintainable test code

## Non-Goals

- Persistent storage (data is lost when process ends)
- Complex query capabilities
- Authentication/authorization
- Data validation beyond basic JSON parsing
- Production-ready database replacement
