# Terraform Provider Testing Guidelines

This document outlines the important guidelines and patterns for writing tests for this Terraform provider, based on established practices and examples.

## Table of Contents
- [Test Execution](#test-execution)
- [Test Structure and Patterns](#test-structure-and-patterns)
- [Endpoint Discovery](#endpoint-discovery)
- [Mock Server Configuration](#mock-server-configuration)
- [Response Overlays](#response-overlays)
- [Test Configuration Files](#test-configuration-files)
- [Common Test Types](#common-test-types)
- [Best Practices](#best-practices)

## Test Execution

### Running Tests

All Terraform provider tests require the `TF_ACC=1` environment variable to enable acceptance testing:

```bash
# Run a specific test
TF_ACC=1 go test -C ./internal/provider -v -run TestBasicResourceUpdate

# Run all tests for a resource
TF_ACC=1 go test -C ./internal/provider -v -run TestBasicResource

# Run all tests in the provider package
TF_ACC=1 go test -C ./internal/provider -v
```

### Debugging Tests

For debugging failing tests, use the `TF_LOG=DEBUG` environment variable to get detailed Terraform logs:

```bash
# Debug a specific test
TF_LOG=DEBUG TF_ACC=1 go test -C ./internal/provider -v -run TestBasicResourceUpdate

# You can also combine with other log levels
TF_LOG=TRACE TF_ACC=1 go test -C ./internal/provider -v -run TestBasicResourceUpdate
```

### Test Validation

**CRITICAL**: Always run your tests to ensure they pass before considering them complete. Tests that don't run successfully are not valid tests.

## Test Structure and Patterns

### Basic Test Structure
All tests should follow this basic pattern:

```go
func TestResourceNameLifecycle(t *testing.T) {
    t.Parallel()

    endpoints := tfmockserver.ResourceEndpoints{
        Create: tfmockserver.Endpoints{
            {
                Endpoint: "POST /v0/endpoint",
            },
        },
        Get: tfmockserver.Endpoints{
            {
                Endpoint: "GET /v0/endpoint/{id}",
            },
        },
        Update: tfmockserver.Endpoints{  // Include if resource supports updates
            {
                Endpoint: "PATCH /v0/endpoint/{id}",
            },
        },
        Delete: tfmockserver.Endpoints{
            {
                Endpoint: "DELETE /v0/endpoint/{id}",
            },
        },
    }

    mockServer := tfmockserver.StartServer(endpoints, t)
    defer mockServer.Close()

    resourceAddress := "testing_resource_name.my_resource"

    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                ConfigDirectory: config.TestNameDirectory(),
                ProtoV6ProviderFactories: provider.GetTestProviders(),
                ConfigVariables: config.Variables{
                    "server_url": config.StringVariable(mockServer.URL),
                },
                ConfigStateChecks: []statecheck.StateCheck{
                    // Add state checks here
                },
            },
        },
    })
}
```

## Endpoint Discovery

### ⚠️ CRITICAL: Always Verify Exact Endpoints from SDK

**DO NOT GUESS OR ASSUME ENDPOINT PATTERNS!**

Before writing any test, you MUST find the exact HTTP endpoints from the SDK. Common mistakes include:
- ❌ Adding or removing 's' for pluralization (e.g., `/v0/basic` vs `/v0/basics`)
- ❌ Wrong hyphenation (e.g., `/v0/nameshadowing` vs `/v0/name-shadowing`)
- ❌ Case differences (e.g., `/v0/OASDefault` vs `/v0/oasdefault`)
- ❌ Assuming patterns from similar resources

### Step-by-Step Endpoint Discovery Process

#### Step 1: Find SDK Method Names

Examine the resource implementation file (e.g., `nameshadowing_resource.go`):

1. **Create Operation**: Look for `r.client.CreateXxx(ctx, *request)`
2. **Read Operation**: Look for `r.client.GetXxx(ctx, ...)`
3. **Update Operation**: Look for `r.client.UpdateXxx(ctx, ...)`
4. **Delete Operation**: Look for `r.client.DeleteXxx(ctx, ...)`

Example from `nameshadowing_resource.go`:
```go
res, err := r.client.CreateNameShadowing(ctx, *request)  // Create method
res, err := r.client.GetNameShadowing(ctx, *request)     // Read method
res, err := r.client.DeleteNameShadowing(ctx, *request)  // Delete method
```

#### Step 2: Find Exact HTTP Endpoints in SDK

Open `internal/sdk/sdk.go` and search for each method:

**For Create endpoints**, search for `func (s *SDK) CreateXxx`:
```bash
# Example search in your editor or terminal
grep -A 20 "func (s \*SDK) CreateNameShadowing" internal/sdk/sdk.go
```

Look for the `url.JoinPath` line:
```go
opURL, err := url.JoinPath(baseURL, "/v0/name-shadowing")
// This is your CREATE endpoint: POST /v0/name-shadowing
```

**For Read endpoints**, search for `func (s *SDK) GetXxx`:
```bash
grep -A 20 "func (s \*SDK) GetNameShadowing" internal/sdk/sdk.go
```

Look for the `utils.GenerateURL` line:
```go
opURL, err := utils.GenerateURL(ctx, baseURL, "/v0/name-shadowing/{id}", request, nil)
// This is your GET endpoint: GET /v0/name-shadowing/{id}
```

**For Update endpoints**, search for `func (s *SDK) UpdateXxx`:
```bash
grep -A 20 "func (s \*SDK) UpdateNameShadowing" internal/sdk/sdk.go
```

**For Delete endpoints**, search for `func (s *SDK) DeleteXxx`:
```bash
grep -A 20 "func (s \*SDK) DeleteNameShadowing" internal/sdk/sdk.go
```

#### Step 3: Verify HTTP Methods

After finding the URL paths, look for the HTTP method in the same function:
```go
req, err := http.NewRequestWithContext(ctx, "POST", opURL, bodyReader)  // POST
req, err := http.NewRequestWithContext(ctx, "GET", opURL, nil)          // GET
req, err := http.NewRequestWithContext(ctx, "PATCH", opURL, bodyReader) // PATCH
req, err := http.NewRequestWithContext(ctx, "DELETE", opURL, nil)       // DELETE
```

### Example Endpoint Discovery Walkthrough

**Resource**: `nameshadowing_resource.go`

1. **Found SDK methods**:
   - `r.client.CreateNameShadowing(ctx, *request)`
   - `r.client.GetNameShadowing(ctx, *request)`
   - `r.client.DeleteNameShadowing(ctx, *request)`

2. **Searched SDK for CreateNameShadowing**:
```go
func (s *SDK) CreateNameShadowing(ctx context.Context, ...) {
    // ...
    opURL, err := url.JoinPath(baseURL, "/v0/name-shadowing")
    // Found: POST /v0/name-shadowing
}
```

3. **Searched SDK for GetNameShadowing**:
```go
func (s *SDK) GetNameShadowing(ctx context.Context, ...) {
    // ...
    opURL, err := utils.GenerateURL(ctx, baseURL, "/v0/name-shadowing/{id}", ...)
    // Found: GET /v0/name-shadowing/{id}
}
```

4. **Searched SDK for DeleteNameShadowing**:
```go
func (s *SDK) DeleteNameShadowing(ctx context.Context, ...) {
    // ...
    opURL, err := utils.GenerateURL(ctx, baseURL, "/v0/name-shadowing/{id}", ...)
    // Found: DELETE /v0/name-shadowing/{id}
}
```

5. **Mock server configuration**:
```go
endpointsBuilder := tfmockserver.ResourceEndpointsBuilder{
    Create: []tfmockserver.ResourceEndpoint{
        {
            Endpoint: "POST /v0/name-shadowing",  // Exact match!
        },
    },
    Get: []tfmockserver.ResourceEndpoint{
        {
            Endpoint: "GET /v0/name-shadowing/{id}",  // Exact match!
        },
    },
    Delete: []tfmockserver.ResourceEndpoint{
        {
            Endpoint: "DELETE /v0/name-shadowing/{id}",  // Exact match!
        },
    },
}
```

### Common Endpoint Patterns (Reference Only - Always Verify!)

These patterns are COMMON but NOT GUARANTEED. Always verify in the SDK!

- **Create**: Often uses plural form → `POST /v0/resources`
- **Read**: Often uses singular form with ID → `GET /v0/resource/{id}`
- **Update**: Usually matches Read → `PATCH /v0/resource/{id}`
- **Delete**: Usually matches Read → `DELETE /v0/resource/{id}`

### Example Endpoint Mappings (Verified from SDK)

**basic_resource.go**:
- **Create**: `CreateBasic` → `POST /v0/basics`
- **Read**: `GetBasic` → `GET /v0/basic/{id}`
- **Update**: `UpdateBasic` → `PATCH /v0/basic/{id}`
- **Delete**: `DeleteBasic` → `DELETE /v0/basic/{id}`

**nameshadowing_resource.go**:
- **Create**: `CreateNameShadowing` → `POST /v0/name-shadowing`
- **Read**: `GetNameShadowing` → `GET /v0/name-shadowing/{id}`
- **Delete**: `DeleteNameShadowing` → `DELETE /v0/name-shadowing/{id}`

**oasdeprecated_resource.go**:
- **Create**: `CreateOasDeprecated` → `POST /v0/oasdeprecated` (NOT `/v0/oasdeprecateds`)
- **Read**: `GetOasDeprecated` → `GET /v0/oasdeprecated/{id}`
- **Delete**: `DeleteOasDeprecated` → `DELETE /v0/oasdeprecated/{id}`

### Debugging Endpoint Mismatches

If you see a `404 page not found` error in tests:

1. **Check the test logs** for the actual URL being called:
```
tf_http_req_uri=/v0/name-shadowing
```

2. **Compare with your mock server configuration**:
```go
Endpoint: "POST /v0/nameshadowings",  // ❌ Wrong - has extra 's'
```

3. **Go back to SDK and verify** the exact endpoint
4. **Update your mock server configuration** to match exactly

## Mock Server Configuration

### ResourceEndpoints Fields

```go
type ResourceEndpoints struct {
    // Create endpoint (POST)
    Create Endpoints
    
    // Get endpoint (GET) - should include {id} placeholder
    Get Endpoints
    
    // Update endpoint (PATCH) - should include {id} placeholder
    Update Endpoints
    
    // Delete endpoint (DELETE) - should include {id} placeholder
    Delete Endpoints
}

type ResourceEndpoint struct {
    // The HTTP method and path. For example: "GET /v0/resource/{id}"
    Endpoint string

    // Full response. If set, this will be returned as the full JSON response,
    // bypassing underlying store data. Conflicts with ResponseOverlay.
    Response map[string]any

    // Optional overlay to apply to the store data in the response. Conflicts
    // with Response.
    ResponseOverlay map[string]any
}
```

**Key Points**:
- **All endpoints use []ResourceEndpoint**: Consistent structure for Create, Get, Update, and Delete
- **Per-endpoint overlays**: Each endpoint can have its own `ResponseOverlay` configuration
- **Multiple endpoints per operation**: Can define multiple endpoints for the same operation type if needed
- **Include {id} placeholders** in Get, Update, and Delete endpoints
- **Use appropriate HTTP methods**: POST for Create, GET for Read, PATCH for Update, DELETE for Delete

**Response Overlay Changes**:
- **No more global overlays**: `CreateResponseOverlay`, `GetResponseOverlay`, etc. are no longer used
- **Per-endpoint overlays only**: Use `ResponseOverlay` within each `ResourceEndpoint`
- **More flexible**: Different endpoints of the same type can have different overlays

### Mock Server Setup

```go
endpoints := tfmockserver.ResourceEndpoints{
    Create: tfmockserver.Endpoints{
        {
            Endpoint: "POST /v0/resources",
        },
    },
    Get: tfmockserver.Endpoints{
        {
            Endpoint: "GET /v0/resource/{id}",
        },
    },
    Update: tfmockserver.Endpoints{
        {
            Endpoint: "PATCH /v0/resource/{id}",
        },
    },
    Delete: tfmockserver.Endpoints{
        {
            Endpoint: "DELETE /v0/resource/{id}",
        },
    },
}

mockServer := tfmockserver.BuildServer(endpointsBuilder, t)
defer mockServer.Close()
```

## Response Overlays

### When to Use Response Overlays

Add response overlays **ONLY** when the resource schema has response-only fields (computed fields that are not sent in requests but returned in responses).

Use response overlays when:

1. **Response-only fields**: Fields marked as `Computed: true` in the schema
2. **Drift detection**: Simulating external changes to resources
3. **Dynamic responses**: Testing different response scenarios

### Response Overlay Examples

#### For Response-Only Fields

If your resource has computed fields like `id`, `created_at`, or other server-generated values:

```go
// Declare repeated values as variables to avoid duplication
testID := "123"
createdAt := "2023-01-01T00:00:00Z"

endpoints := tfmockserver.ResourceEndpoints{
    Create: tfmockserver.Endpoints{
        {
            Endpoint: "POST /v0/resources",
            ResponseOverlay: map[string]any{
                "id": testID,
                "created_at": createdAt,
            },
        },
    },
    Get: tfmockserver.Endpoints{
        {
            Endpoint: "GET /v0/resource/{id}",
            ResponseOverlay: map[string]any{
                "id": testID,
                "created_at": createdAt,
            },
        },
    },
    Delete: tfmockserver.Endpoints{
        {
            Endpoint: "DELETE /v0/resource/{id}",
        },
    },
}
```

#### For Drift Detection

For testing drift detection (when resource values change outside of Terraform):

```go
endpoints := tfmockserver.ResourceEndpoints{
    Create: tfmockserver.Endpoints{
        {
            Endpoint: "POST /v0/resources",
            ResponseOverlay: map[string]any{
                "name": "original-value",
            },
        },
    },
    Get: tfmockserver.Endpoints{
        {
            Endpoint: "GET /v0/resource/{id}",
            ResponseOverlay: map[string]any{
                "name": "drifted-value", // Different from create - simulates drift
            },
        },
    },
    Delete: tfmockserver.Endpoints{
        {
            Endpoint: "DELETE /v0/resource/{id}",
        },
    },
}
```

## Test Configuration Files

### Directory Structure

Create test configuration directories under `testdata/`:

```
internal/provider/testdata/
├── TestBasicResourceLifecycle/
│   └── main.tf
├── TestBasicResourceUpdate/
│   └── main.tf
└── TestBasicResourceValidation/
    └── main.tf
```

The directory name **must** match the test function name exactly.

### Configuration Templates

#### Basic Configuration
```hcl
variable "server_url" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_basic" "my_basic" {
  name = "test-resource"
}
```

#### Configuration with Variables
```hcl
variable "server_url" {
  type = string
}

variable "name" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_basic" "my_basic" {
  name = var.name
}
```

## Best Practices

### Test Organization
- **One test file per resource** (e.g., `basic_resource_test.go`)
- **Group related tests** with consistent naming
- **Use descriptive test names** that explain what is being tested
- **Consolidate import tests**: Add import testing as a step in lifecycle tests rather than creating separate import test functions to avoid code duplication
- **Remove auto-generated headers**: Tests are hand-written, not auto-generated - remove the "Code generated by Speakeasy" comment from test files
- Use `t.Parallel()` for all tests

### Import Test Configuration
- **Always include `ConfigDirectory: config.TestNameDirectory()` in import steps** - Without this line, the import step may lose provider configuration during the destroy phase, causing 404 errors when Terraform tries to delete resources using default provider settings instead of your mock server.

### State Verification
- **Make sure to declare Variables for static values and use them across the test** in state checks
- **Verify all important fields** in state checks
- **Use appropriate knownvalue matchers** for different types
- **Test both request and response fields** when applicable

**Basic state checks:**
```go
ConfigStateChecks: []statecheck.StateCheck{
    // String fields
    statecheck.ExpectKnownValue(
        resourceAddress,
        tfjsonpath.New("field_name"),
        knownvalue.StringExact("expected_value"),
    ),
    // Numeric fields
    statecheck.ExpectKnownValue(
        resourceAddress,
        tfjsonpath.New("count"),
        knownvalue.Int32Exact(42),
    ),
}
```

**Complex nested structures:**
```go
ConfigStateChecks: []statecheck.StateCheck{
    // Nested objects
    statecheck.ExpectKnownValue(
        resourceAddress,
        tfjsonpath.New("nested_object"),
        knownvalue.ObjectExact(map[string]knownvalue.Check{
            "field1": knownvalue.StringExact("value1"),
            "field2": knownvalue.Int64Exact(42),
        }),
    ),
    // Lists
    statecheck.ExpectKnownValue(
        resourceAddress,
        tfjsonpath.New("list_field"),
        knownvalue.ListExact([]knownvalue.Check{
            knownvalue.StringExact("item1"),
            knownvalue.StringExact("item2"),
        }),
    ),
}
```

### Mock Server Usage
- **⚠️ ALWAYS verify exact endpoints from SDK** - Never guess or assume endpoint patterns
- **Include all necessary endpoints** (Create, Get, Update, Delete)
- **Use appropriate HTTP methods** for each operation
- **Match endpoints exactly** - Including pluralization, hyphenation, and case
- **Add response overlays only when needed** for response-only fields
- **Declare repeated values as variables** - If the same response overlay values are used multiple times in a test, declare them as variables at the top of the test function and reuse them
- Always use `defer mockServer.Close()` to clean up
- **Debug 404 errors** by comparing test logs with mock server configuration

### Error Handling
- **Test validation errors** with `ExpectError`
- **Use appropriate error message patterns** with regex
- **Test edge cases** and boundary conditions
- Use `regexp.MustCompile()` for error pattern matching

### Configuration Management
- **Use variables** for dynamic values in test configs
- **Keep configurations minimal** but complete
- **Reuse configurations** where possible with variables
- Use `config.TestNameDirectory()` for automatic configuration discovery

### Test Execution
- **Always run tests** before considering them complete
- **Use TF_LOG=DEBUG** for debugging failing tests
- **Test in parallel** with `t.Parallel()` for faster execution
- **Clean up resources** with proper defer statements

### Documentation
- **Document complex test scenarios** with comments
- **Explain overlay usage** when implementing drift detection
- **Reference related resources** and dependencies

### Schema Analysis
Before writing tests, analyze the resource schema:
1. **Request fields**: What fields are required/optional for input?
2. **Response fields**: What fields are computed/response-only?
3. **Update behavior**: Does the resource support updates?
4. **Delete behavior**: Is delete implemented or not?

For detailed guidance on which tests to write based on your resource's characteristics, see the **Test Selection Guide** section below.

## Test Selection Guide

This section helps you determine which tests to write for your resource based on its characteristics.

### Step 1: Analyze Your Resource

Before writing tests, analyze your resource file to understand:

1. **Schema Fields**:
   - What fields exist?
   - Which are `Required`, `Optional`, or `Computed`?
   - Are there any with custom validators (e.g., `stringvalidator.OneOf()`, `int32validator.Between()`)?
   - Are there nested objects, lists, or sets?

2. **Operations Supported**:
   - Does `Create()` make an API call or is it not implemented?
   - Does `Update()` make an API call or just say "Not Implemented"?
   - Does `Delete()` make an API call?
   - Does `ImportState()` have custom logic or use default framework behavior?

3. **Special Characteristics**:
   - Are there enum fields with specific allowed values?
   - Are there computed/response-only fields?
   - Are there fields with default values?
   - Is there custom import logic (e.g., parsing multiple IDs, type conversion)?

### Step 2: Required Tests (ALL Resources)

Every resource MUST have:

#### ✅ Lifecycle Test
**Always required** - Tests basic CRUD operations

```go
func TestResourceNameLifecycle(t *testing.T)
```

**What it tests**:
- Resource creation
- State verification after creation
- Reading resource state
- Resource deletion (implicit via framework)
- **Import functionality** (as an additional step)

**When to include import step**:
- If resource has `ImportState()` implemented → Add import test step
- If resource uses default import → Add import test step with `tfmockserver.StoreKey`
- If resource has custom import logic → Add import test step with appropriate ID format

### Step 3: Conditional Tests (Based on Resource Characteristics)

Use this decision tree to determine which additional tests to write:

#### ✅ Update Test
**Required if**: Resource has `Update()` implemented with actual API calls

**Skip if**: 
- Update method contains only `// Not Implemented; all attributes marked as RequiresReplace`
- All fields have `RequiresReplace: true` in plan modifiers

```go
func TestResourceNameUpdate(t *testing.T)
```

**What it tests**:
- Creating resource with initial values
- Updating one or more fields
- Verifying updated values in state

---

#### ✅ Boundary Value Tests
**Required if**: Resource has numeric fields with specific ranges OR int32/int64 type conversions

**Examples**:
- Int32 fields (test min: -2147483648, max: 2147483647)
- Int64 fields (test min/max values)
- Enum integers with specific allowed values
- Custom import logic that validates numeric types

```go
func TestResourceNameBoundaryValues(t *testing.T)
func TestResourceNameMinBoundaryValue(t *testing.T)
```

**What it tests**:
- Minimum valid value
- Maximum valid value
- Values just outside the valid range (for validation tests)

---

#### ✅ Import Validation Test
**Required if**: Resource has custom `ImportState()` logic with validation

**Examples**:
- Custom ID parsing (multiple IDs, JSON format)
- Type conversion (string to int32/int64)
- Format validation (UUIDs, specific patterns)
- Range validation in import

```go
func TestResourceNameImportValidation(t *testing.T)
```

**What it tests**:
- Valid import IDs work correctly
- Invalid formats are rejected with appropriate errors
- Edge cases (empty strings, wrong types, out of range)

---

#### ✅ Validation Test
**Required if**: Resource has custom validators (NOT basic required field validation)

**Examples of custom validators**:
- `stringvalidator.OneOf("value1", "value2")` - enum validation
- `stringvalidator.RegexMatches()` - format validation
- `int32validator.Between()` - range validation
- Custom validators for cross-field validation

**Skip if**: Only basic framework validation (`Required: true`, type checking)

```go
func TestResourceNameValidation(t *testing.T)
```

**What it tests**:
- Invalid enum values trigger errors
- Invalid formats trigger errors
- Cross-field validation rules
- Nested object validation

---

#### ✅ Nested Object Validation Test
**Required if**: Resource has nested objects with custom validators

```go
func TestResourceNameNestedValidation(t *testing.T)
```

**What it tests**:
- Validation in nested object fields
- Enum validation in nested structures
- Complex nested object scenarios

---

#### ⚠️ Drift Detection Test
**Usually SKIP** - Only add if there's a specific need

**Consider adding if**:
- Testing framework behavior for external changes
- Resource has multiple mutable fields that could drift
- Specific business requirement to test drift handling

**Skip if**:
- Resource has only an ID field
- Testing framework functionality rather than provider logic
- No realistic drift scenarios

```go
func TestResourceNameWhenResourceValueDriftsOutsideOfTerraform(t *testing.T)
```

**What it tests**:
- Resource state changes outside Terraform
- Terraform detects drift on refresh
- Plan shows expected changes

---

#### ✅ Response-Only Fields Test
**Required if**: Resource has computed fields returned only in responses

**Examples**:
- `created_at`, `updated_at` timestamps
- Server-generated IDs
- Computed status fields
- Default values set by the API

**Note**: Usually integrated into lifecycle test with `ResponseOverlay`, not a separate test

**What it tests**:
- Response-only fields are populated correctly
- Computed values are stored in state

---

#### ✅ Complex State Verification Tests
**Required if**: Resource has complex nested structures (objects, lists, sets)

**What it tests**:
- Nested objects are correctly stored
- Lists and sets maintain order/uniqueness
- Complex types serialize/deserialize properly

### Step 4: Test Selection Examples

#### Example 1: Simple Resource (ID only, no updates)
**Resource**: `importidint32_resource.go`
- Single computed field: `id` (int32)
- No update support
- Custom import with int32 validation

**Required Tests**:
1. ✅ `TestImportIDInt32ResourceLifecycle` - Basic lifecycle + import
2. ✅ `TestImportIDInt32ResourceBoundaryValues` - Int32 max boundary
3. ✅ `TestImportIDInt32ResourceMinBoundaryValue` - Int32 min boundary
4. ✅ `TestImportIDInt32ResourceImportValidation` - Custom import validation

**Skipped Tests**:
- ❌ Update test (no update support)
- ❌ Validation test (no custom validators in schema)
- ❌ Drift detection (only ID field, not realistic)
- ❌ Nested validation (no nested objects)

---

#### Example 2: Resource with Enums and Updates
**Resource**: `oasenum_resource.go`
- Multiple enum fields (string, int32, int64)
- Update support
- Nested objects with enums
- Response-only fields

**Required Tests**:
1. ✅ `TestOASEnumResourceLifecycle` - Basic lifecycle + import
2. ✅ `TestOASEnumResourceUpdate` - Update different enum values
3. ✅ `TestOASEnumResourceValidation` - Invalid enum values
4. ✅ `TestOASEnumResourceNestedValidation` - Nested object enum validation
5. ✅ `TestOASEnumResourceBoundaryValues` - Enum edge cases
6. ✅ `TestOASEnumResourceWithResponseFields` - Response-only fields

**Skipped Tests**:
- ❌ Import validation (uses default import, no custom logic)
- ❌ Drift detection (not a specific requirement)

---

#### Example 3: Resource with Multiple ID Import
**Resource**: `importmultipleid_resource.go`
- Multiple path parameters (param1, param2)
- Custom JSON import format
- No update support

**Required Tests**:
1. ✅ `TestImportMultipleIDResourceLifecycle` - Lifecycle with custom import
2. ✅ `TestImportMultipleIDImportValidation` - JSON format validation (if implemented)

**Skipped Tests**:
- ❌ Update test (no update support)
- ❌ Validation test (no custom validators)
- ❌ Boundary value test (string fields, no numeric ranges)

---

#### Example 4: Resource with Defaults
**Resource**: `oasdefault_resource.go`
- Multiple fields with OAS default values
- No update support
- Various types (string, int32, float, lists, sets)

**Required Tests**:
1. ✅ `TestOASDefaultResourceLifecycle` - Default values applied
2. ✅ `TestOASDefaultResourceWithCustomValues` - Override defaults
3. ✅ `TestOASDefaultResourceWithComputedResponse` - Response-only fields

**Skipped Tests**:
- ❌ Update test (no update support)
- ❌ Validation test (no custom validators)
- ⚠️ Drift detection (currently skipped, not applicable)

### Step 5: Quick Checklist

Use this checklist when writing tests for a new resource:

- [ ] **Lifecycle test** with import step (ALWAYS REQUIRED)
- [ ] **Update test** (if Update() is implemented)
- [ ] **Boundary value tests** (if int32/int64 fields or numeric enums)
- [ ] **Import validation test** (if custom ImportState() with validation)
- [ ] **Validation test** (if custom validators like OneOf, RegexMatches)
- [ ] **Nested validation test** (if nested objects with validators)
- [ ] **Response-only fields** (if computed fields - usually in lifecycle test)
- [ ] **Complex state verification** (if nested objects, lists, sets)
- [ ] **Drift detection** (RARELY - only if specific requirement)

### Common Mistakes to Avoid

❌ **DON'T** guess or assume endpoint URLs - always verify from SDK  
❌ **DON'T** add/remove 's' for pluralization without checking SDK  
❌ **DON'T** assume hyphenation patterns (name-shadowing vs nameshadowing)  
❌ **DON'T** test basic required field validation - framework handles this  
❌ **DON'T** create separate import test functions - add as step in lifecycle
❌ **DON'T** forget `ConfigDirectory: config.TestNameDirectory()` in import steps - this causes 404 errors during destroy
❌ **DON'T** add drift tests for resources with only ID fields  
❌ **DON'T** test framework functionality - focus on provider logic  
❌ **DON'T** add validation tests if there are no custom validators  

✅ **DO** verify exact endpoints from `internal/sdk/sdk.go` before writing tests  
✅ **DO** run all tests before considering them complete  
✅ **DO** test custom logic in ImportState()  
✅ **DO** test boundary values for numeric types  
✅ **DO** test enum validation when present  
✅ **DO** consolidate related test scenarios to avoid duplication  
✅ **DO** check test logs for actual URLs when debugging 404 errors

## Data Source Testing

### ⚠️ CRITICAL: Verify Endpoints Before Writing Data Source Tests

Data source tests use the SAME endpoints as resource tests. You MUST verify the exact endpoints from the SDK (see [Endpoint Discovery](#endpoint-discovery) section above).

**Common mistake**: Assuming endpoint names without checking the SDK!

### Basic Data Source Test Pattern

Data source tests verify that data sources correctly retrieve information from managed resources.

```go
func TestResourceNameDataSource(t *testing.T) {
    t.Parallel()

    // Define only the minimum required variables
    param1 := "test-value"

    // ⚠️ IMPORTANT: Get these endpoints from internal/sdk/sdk.go
    // DO NOT guess based on resource name!
    // See "Endpoint Discovery" section above for how to find exact endpoints
    endpointsBuilder := tfmockserver.ResourceEndpointsBuilder{
        Create: []tfmockserver.ResourceEndpoint{
            {
                Endpoint: "POST /v0/resources",  // Verify this in SDK!
            },
        },
        Get: []tfmockserver.ResourceEndpoint{
            {
                Endpoint: "GET /v0/resource/{id}",  // Verify this in SDK!
            },
        },
        Delete: []tfmockserver.ResourceEndpoint{
            {
                Endpoint: "DELETE /v0/resource/{id}",  // Verify this in SDK!
            },
        },
    }

    mockServer := tfmockserver.BuildServer(endpointsBuilder, t)
    defer mockServer.Close()

    dataResourceAddress := "data.testing_resource_name.test"
    managedResourceAddress := "testing_resource_name.test"

    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                ConfigDirectory: config.TestNameDirectory(),
                ProtoV6ProviderFactories: provider.GetTestProviders(),
                ConfigVariables: config.Variables{
                    "server_url": config.StringVariable(mockServer.URL),
                    "param1":     config.StringVariable(param1),
                },
                // Use AllModelFieldCompareChecks for comparing all fields
                ConfigStateChecks: provider.AllModelFieldCompareChecks(
                    provider.ResourceNameDataSourceModel{},
                    managedResourceAddress,
                    dataResourceAddress,
                    []string{"id"}, // Exclude ID as Terraform validates it automatically
                ),
            },
        },
    })
}
```

### Key Principles

1. **Use AllModelFieldCompareChecks**: Always use the helper function to compare all fields between managed resource and data source
   ```go
   ConfigStateChecks: provider.AllModelFieldCompareChecks(
       provider.ResourceNameDataSourceModel{},
       managedResourceAddress,
       dataResourceAddress,
       []string{"id"}, // Fields to exclude
   )
   ```

2. **Minimum Configuration**: Only configure fields that are **required** for the resource and data source
   - ❌ Don't include request-only fields in ConfigVariables if not used by data source
   - ❌ Don't set computed/response-only fields in Terraform config
   - ✅ Only include fields needed to create and read the resource

3. **Field Exclusion**: Exclude fields from comparison when:
   - Field is `id` (Terraform validates this automatically)
   - Field is request-only (not returned in responses)
   - Field is computed-only in data source but not in resource

### Data Source Configuration Example

**Minimal Configuration** (Recommended):
```hcl
variable "server_url" {
  type = string
}

variable "name" {
  type = string
}

provider "testing" {
  server_url = var.server_url
}

resource "testing_resource" "test" {
    name = var.name
}

data "testing_resource" "test" {
    id = testing_resource.test.id
}
```

**What NOT to do**:
```hcl
# ❌ Don't add unnecessary variables
variable "computed_field" {  # Bad: This is computed
  type = string
}

resource "testing_resource" "test" {
    name = var.name
    # computed_field is response-only, don't set it
}

data "testing_resource" "test" {
    id = testing_resource.test.id
    # ❌ Don't try to set computed fields
}
```

### Common Patterns

**Simple Data Source** (ID lookup only):
```go
ConfigStateChecks: provider.AllModelFieldCompareChecks(
    provider.ResourceNameDataSourceModel{},
    managedResourceAddress,
    dataResourceAddress,
    []string{"id"},
)
```

**Data Source with Multiple IDs**:
```go
// For resources with multiple path parameters
ConfigVariables: config.Variables{
    "server_url": config.StringVariable(mockServer.URL),
    "param1":     config.StringVariable(param1),
    "param2":     config.StringVariable(param2),
}

// Exclude IDs from comparison as Terraform validates them
ConfigStateChecks: provider.AllModelFieldCompareChecks(
    provider.ResourceNameDataSourceModel{},
    managedResourceAddress,
    dataResourceAddress,
    []string{}, // Empty if no fields to exclude
)
```

**Data Source with Acronym Fields** (apiId, portalId):
```go
// Mock server uses camelCase for JSON
endpointsBuilder := tfmockserver.ResourceEndpointsBuilder{
    Create: []tfmockserver.ResourceEndpoint{
        {
            Endpoint: "POST /v0/resources/{apiId}/{portalId}",
        },
    },
}
```

## Example Reference
For complete examples, refer to existing test files:

**Resource Tests**:
- `basic_resource_test.go` - **Recommended pattern**: Comprehensive test coverage with updates and import testing consolidated into lifecycle tests
- `importidint32_resource_test.go` - Boundary value testing and import validation for int32 types
- `oasenum_resource_test.go` - Complex enum validation testing with nested objects and boundary values
- `importmultipleid_resource_test.go` - Custom import with multiple IDs (JSON format)
- `frameworktype_resource_test.go` - Basic lifecycle pattern
- `oasdefault_resource_test.go` - Advanced scenarios with overlays and default values
- `apicreateandupdate_resource_test.go` - Drift detection pattern

**Data Source Tests**:
- `importmultipleid_data_source_test.go` - **Recommended pattern**: Using AllModelFieldCompareChecks with minimal configuration
- `importmultipleidacronym_data_source_test.go` - Acronym fields (apiId, portalId) pattern
- `oasenum_data_source_test.go` - Complex data source with many fields
