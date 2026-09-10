package fastAST

import (
	"bytes"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFastAST(t *testing.T) {
	tests := []struct {
		name           string
		openAPIContent string
		expectError    bool
		expectOps      int
		expectGroups   []string
	}{
		{
			name: "basic document with operations",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      summary: Get all users
      description: Retrieve a list of users
    post:
      operationId: createUser
      summary: Create user
      description: Create a new user
  /users/{id}:
    get:
      operationId: getUserById
      summary: Get user by ID
      description: Retrieve a specific user`,
			expectError:  false,
			expectOps:    3,
			expectGroups: []string{""},
		},
		{
			name: "document with speakeasy extensions",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      summary: Get all users
      x-speakeasy-group: UserOperations
    post:
      operationId: createUser
      summary: Create user
      x-speakeasy-ignore: true
  /pets:
    get:
      operationId: getPets
      summary: Get all pets
      x-speakeasy-name-override: ListAllPets`,
			expectError:  false,
			expectOps:    2, // createUser is ignored
			expectGroups: []string{"UserOperations", ""},
		},
		{
			name: "document with tags",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      summary: Get all users
      tags:
        - Users
    post:
      operationId: createUser
      summary: Create user
      tags:
        - Users
  /pets:
    get:
      operationId: getPets
      summary: Get all pets
      tags:
        - Pets`,
			expectError:  false,
			expectOps:    3,
			expectGroups: []string{"Users", "Pets"},
		},
		{
			name: "deprecated operations ignored",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      summary: Get all users
    post:
      operationId: createUser
      summary: Create user
      deprecated: true`,
			expectError:  false,
			expectOps:    1, // deprecated operation is ignored
			expectGroups: []string{""},
		},
		{
			name: "empty document",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0`,
			expectError:  false,
			expectOps:    0,
			expectGroups: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _, err := openapi.Unmarshal(t.Context(), bytes.NewBufferString(tt.openAPIContent))
			require.NoError(t, err, "Failed to parse OpenAPI document: %v", err)

			// Create FastAST
			fastAST, err := NewFastAST(t.Context(), doc)
			if tt.expectError {
				require.Error(t, err, "Expected error but got none")
			}
			require.NoError(t, err, "Unexpected error: %v", err)

			// Verify basic structure
			require.NotNil(t, fastAST, "Expected FastAST but got nil")

			assert.Equal(t, "3.0.3", fastAST.OpenAPIVersion, "Expected OpenAPI version 3.0.3, got %s", fastAST.OpenAPIVersion)
			require.NotNil(t, fastAST.DocInfo, "Expected DocInfo but got nil")
			assert.Equal(t, "Test API", fastAST.DocInfo.Title, "Expected title 'Test API', got %s", fastAST.DocInfo.Title)

			// Verify operations count
			assert.Len(t, fastAST.Operations, tt.expectOps, "Expected %d operations, got %d", tt.expectOps, len(fastAST.Operations))

			// Verify MCP tools are created for each operation
			if tt.expectOps > 0 {
				require.NotNil(t, fastAST.MCP, "Expected MCP to be initialized")
				assert.Len(t, fastAST.MCP.Tools, tt.expectOps, "Expected %d MCP tools, got %d", tt.expectOps, len(fastAST.MCP.Tools))
			}

			// Verify GroupTree
			assert.NotNil(t, fastAST.GroupTree, "Expected GroupTree but got nil")
		})
	}
}

func TestFastASTMCPExtensions(t *testing.T) {
	tests := []struct {
		name           string
		openAPIContent string
		expectedTools  []MCPTool
	}{
		{
			name: "operation with MCP extension",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      summary: Get all users
      description: Retrieve a list of users from the system
      x-speakeasy-mcp:
        name: list_users
        description: Lists all users in the system
        disabled: false
        scopes:
          - read
          - admin`,
			expectedTools: []MCPTool{
				{
					Name:         "list_users",
					Description:  "Lists all users in the system",
					Disabled:     false,
					OperationIDs: []string{"getUsers"},
					Scopes:       []string{"read", "admin"},
				},
			},
		},
		{
			name: "operation with partial MCP extension",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      operationId: createUser
      summary: Create user
      description: Create a new user account
      x-speakeasy-mcp:
        name: create_user_account
        disabled: true`,
			expectedTools: []MCPTool{
				{
					Name:         "create_user_account",
					Description:  "Create a new user account", // fallback to operation description
					Disabled:     true,
					OperationIDs: []string{"createUser"},
					Scopes:       []string{}, // defaults to empty
				},
			},
		},
		{
			name: "operation without MCP extension (defaults)",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /pets:
    get:
      operationId: getPets
      summary: Get pets
      description: Retrieve all pets`,
			expectedTools: []MCPTool{
				{
					Name:         "get_pets",          // normalized operationId
					Description:  "Retrieve all pets", // defaults to description
					Disabled:     false,               // defaults to false
					OperationIDs: []string{"getPets"},
					Scopes:       []string{}, // defaults to empty
				},
			},
		},
		{
			name: "document without MCP extensions still generates MCP tools",
			openAPIContent: `openapi: 3.0.3
info:
  title: Pet Store API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: listUsers
      summary: List all users
      description: Returns a list of users
    post:
      operationId: createUser
      summary: Create a user
      description: Creates a new user
  /pets/{id}:
    get:
      operationId: getPetById
      summary: Get pet by ID
      description: Returns a specific pet`,
			expectedTools: []MCPTool{
				{
					Name:         "list_users",
					Description:  "Returns a list of users",
					Disabled:     false,
					OperationIDs: []string{"listUsers"},
					Scopes:       []string{},
				},
				{
					Name:         "create_user",
					Description:  "Creates a new user",
					Disabled:     false,
					OperationIDs: []string{"createUser"},
					Scopes:       []string{},
				},
				{
					Name:         "get_pet_by_id",
					Description:  "Returns a specific pet",
					Disabled:     false,
					OperationIDs: []string{"getPetById"},
					Scopes:       []string{},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, _, err := openapi.Unmarshal(t.Context(), bytes.NewBufferString(tt.openAPIContent))
			require.NoError(t, err, "Failed to parse OpenAPI document: %v", err)

			// Create FastAST
			fastAST, err := NewFastAST(t.Context(), doc)
			require.NoError(t, err, "Failed to create FastAST: %v", err)

			require.Len(t, fastAST.MCP.Tools, len(tt.expectedTools), "Expected %d MCP tools, got %d", len(tt.expectedTools), len(fastAST.MCP.Tools))

			// Verify each tool
			for i, expectedTool := range tt.expectedTools {
				actualTool := fastAST.MCP.Tools[i]

				assert.Equal(t, expectedTool.Name, actualTool.Name, "Tool %d: expected name %s, got %s", i, expectedTool.Name, actualTool.Name)
				assert.Equal(t, expectedTool.Description, actualTool.Description, "Tool %d: expected description %s, got %s", i, expectedTool.Description, actualTool.Description)
				assert.Equal(t, expectedTool.Disabled, actualTool.Disabled, "Tool %d: expected disabled %t, got %t", i, expectedTool.Disabled, actualTool.Disabled)
				assert.Equal(t, expectedTool.OperationIDs, actualTool.OperationIDs, "Tool %d: expected %d operation IDs, got %d", i, len(expectedTool.OperationIDs), len(actualTool.OperationIDs))
				assert.Equal(t, expectedTool.Scopes, actualTool.Scopes, "Tool %d: expected %d scopes, got %d", i, len(expectedTool.Scopes), len(actualTool.Scopes))
			}
		})
	}
}

func TestFastASTMCPInvalidYAML(t *testing.T) {
	openAPIContent := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      summary: Get users
      description: Get all users
      x-speakeasy-mcp: "invalid yaml string"`

	// Parse the OpenAPI document
	doc, _, err := openapi.Unmarshal(t.Context(), bytes.NewBufferString(openAPIContent))
	require.NoError(t, err, "Failed to parse OpenAPI document: %v", err)

	// Create FastAST - should not fail even with invalid MCP YAML
	fastAST, err := NewFastAST(t.Context(), doc)
	require.NoError(t, err, "Failed to create FastAST: %v", err)

	require.Len(t, fastAST.MCP.Tools, 1, "Expected 1 MCP tool, got %d", len(fastAST.MCP.Tools))

	// Should fall back to defaults when YAML decode fails
	tool := fastAST.MCP.Tools[0]
	assert.Equal(t, "get_users", tool.Name, "Expected tool name 'get_users', got %s", tool.Name)
	assert.Equal(t, "Get all users", tool.Description, "Expected tool description 'Get all users', got %s", tool.Description)
	assert.False(t, tool.Disabled, "Expected tool disabled false, got %t", tool.Disabled)
	assert.Empty(t, tool.Scopes, "Expected empty scopes, got %v", tool.Scopes)
}

func TestBuildDisplayName(t *testing.T) {
	tests := []struct {
		name            string
		openAPIContent  string
		operationId     string
		expectedDisplay string
	}{
		{
			name: "operation with name override",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      summary: Get users
      x-speakeasy-name-override: "ListAllUsers"`,
			operationId:     "getUsers",
			expectedDisplay: "ListAllUsers",
		},
		{
			name: "operation without name override",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      summary: Get users`,
			operationId:     "getUsers",
			expectedDisplay: "getUsers",
		},
		{
			name: "operation without operationId",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      summary: Get users`,
			operationId:     "",
			expectedDisplay: "/users/get", // fallback to path/method
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the OpenAPI document
			doc, _, err := openapi.Unmarshal(t.Context(), bytes.NewBufferString(tt.openAPIContent))
			require.NoError(t, err, "Failed to parse OpenAPI document: %v", err)

			// Create FastAST
			fastAST, err := NewFastAST(t.Context(), doc)
			require.NoError(t, err, "Failed to create FastAST: %v", err)

			// Find the operation
			var foundOp *FastOperation
			for _, op := range fastAST.Operations {
				if (tt.operationId == "" && op.OperationID == "") || op.OperationID == tt.operationId {
					foundOp = op
					break
				}
			}
			require.NotNil(t, foundOp, "Expected to find operation")
			assert.Equal(t, tt.expectedDisplay, foundOp.DisplayName, "Expected display name %s, got %s", tt.expectedDisplay, foundOp.DisplayName)
		})
	}
}

func TestFastASTGroupTree(t *testing.T) {
	openAPIContent := `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      summary: Get users
      tags:
        - Users
    post:
      operationId: createUser
      summary: Create user
      tags:
        - Users
  /pets:
    get:
      operationId: getPets
      summary: Get pets
      x-speakeasy-group: Animals
  /orders:
    get:
      operationId: getOrders
      summary: Get orders`

	// Parse the OpenAPI document
	doc, _, err := openapi.Unmarshal(t.Context(), bytes.NewBufferString(openAPIContent))
	require.NoError(t, err, "Failed to parse OpenAPI document: %v", err)

	// Create FastAST
	fastAST, err := NewFastAST(t.Context(), doc)
	require.NoError(t, err, "Failed to create FastAST: %v", err)

	// Verify GroupTree is initialized
	require.NotNil(t, fastAST.GroupTree, "Expected GroupTree to be initialized")

	// Verify operations are added to correct groups
	assert.Len(t, fastAST.Operations, 4, "Expected 4 operations, got %d", len(fastAST.Operations))

	// Check that operations have correct groups assigned
	groupCounts := make(map[string]int)
	for _, op := range fastAST.Operations {
		groupCounts[op.Group]++
	}

	expectedGroups := map[string]int{
		"Users":   2, // getUsers, createUser
		"Animals": 1, // getPets (x-speakeasy-group)
		"":        1, // getOrders (no group/tag)
	}

	assert.Equal(t, expectedGroups, groupCounts, "Expected %v, got %v", expectedGroups, groupCounts)
}

func TestFastASTEmptyDocument(t *testing.T) {
	openAPIContent := `openapi: 3.0.3
info:
  title: Empty API
  version: 1.0.0`

	// Parse the OpenAPI document
	doc, _, err := openapi.Unmarshal(t.Context(), bytes.NewBufferString(openAPIContent))
	require.NoError(t, err, "Failed to parse OpenAPI document: %v", err)

	// Create FastAST
	fastAST, err := NewFastAST(t.Context(), doc)
	require.NoError(t, err, "Failed to create FastAST: %v", err)
	require.NotNil(t, fastAST, "Expected FastAST but got nil")

	// Verify basic structure
	assert.Equal(t, "3.0.3", fastAST.OpenAPIVersion, "Expected OpenAPI version 3.0.3, got %s", fastAST.OpenAPIVersion)
	assert.Equal(t, "Empty API", fastAST.DocInfo.Title, "Expected title 'Empty API', got %s", fastAST.DocInfo.Title)

	// Should have no operations
	assert.Empty(t, fastAST.Operations, "Expected 0 operations, got %d", len(fastAST.Operations))

	// GroupTree should still be initialized
	require.NotNil(t, fastAST.GroupTree, "Expected GroupTree to be initialized even for empty document")
}

func TestMCPToolNameNormalization(t *testing.T) {
	caser := casing.New()

	tests := []struct {
		operationId  string
		expectedName string
	}{
		{"getPets", "get_pets"},
		{"createUser", "create_user"},
		{"getUsers", "get_users"},
		{"updateUserById", "update_user_by_id"},
		{"deleteOrderItem", "delete_order_item"},
		{"getHTTPResponse", "get_http_response"},
		{"createAPIKey", "create_api_key"},
	}

	for _, tt := range tests {
		t.Run(tt.operationId, func(t *testing.T) {
			result := caser.ToSnake(tt.operationId)
			assert.Equal(t, tt.expectedName, result, "Expected %s -> %s, got %s", tt.operationId, tt.expectedName, result)
		})
	}
}

func TestFastASTTerraformExtensions(t *testing.T) {
	tests := []struct {
		name                     string
		openAPIContent           string
		expectedDataResources    []TerraformDataResource
		expectedManagedResources []TerraformManagedResource
	}{
		{
			name: "GET operation with entity extension creates both data and managed resources",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      operationId: getUserById
      summary: Get user by ID
      description: Retrieve a specific user
      x-speakeasy-entity-operation: User#read`,
			expectedDataResources: []TerraformDataResource{
				{
					Name:         "User",
					Description:  "User (DataSource)",
					OperationIDs: []string{"getUserById"},
				},
			},
			expectedManagedResources: []TerraformManagedResource{
				{
					Name:         "User",
					Description:  "User (Resource)",
					OperationIDs: []string{"getUserById"},
				},
			},
		},
		{
			name: "POST operation with entity extension creates managed resource",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    post:
      operationId: createUser
      summary: Create user
      description: Create a new user
      x-speakeasy-entity-operation: UserAccount#create`,
			expectedDataResources: []TerraformDataResource{},
			expectedManagedResources: []TerraformManagedResource{
				{
					Name:         "UserAccount",
					Description:  "UserAccount (Resource)",
					OperationIDs: []string{"createUser"},
				},
			},
		},
		{
			name: "operation without entity extension creates no resources",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users:
    get:
      operationId: getUsers
      summary: Get users
      description: List all users`,
			expectedDataResources:    []TerraformDataResource{},
			expectedManagedResources: []TerraformManagedResource{},
		},
		{
			name: "Terraform resources with custom entity descriptions",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
components:
  schemas:
    Order:
      description: An order helps you make coffee
      x-speakeasy-entity: Order
      x-speakeasy-entity-description: |
        Manage a coffee order.
    User:
      description: A user of the system
      x-speakeasy-entity-description: |
        Handle user management operations.
paths:
  /orders:
    post:
      operationId: createOrder
      summary: Create order
      description: Create a new coffee order
      x-speakeasy-entity-operation: Order#create
  /users:
    get:
      operationId: getUsers
      summary: Get users  
      description: List all users
      x-speakeasy-entity-operation: User#read`,
			expectedDataResources: []TerraformDataResource{
				{
					Name:         "User",
					Description:  "User (DataSource)",
					OperationIDs: []string{"getUsers"},
				},
			},
			expectedManagedResources: []TerraformManagedResource{
				{
					Name:         "Order",
					Description:  "Order (Resource)",
					OperationIDs: []string{"createOrder"},
				},
				{
					Name:         "User",
					Description:  "User (Resource)",
					OperationIDs: []string{"getUsers"},
				},
			},
		},
		{
			name: "entity operation with fragment value like 'Apps#read'",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /apps/{id}:
    get:
      operationId: getAppById
      summary: Get app by ID
      description: Retrieve a specific app
      x-speakeasy-entity-operation: Apps#read
  /mobile-apps:
    post:
      operationId: createMobileApp
      summary: Create mobile app
      description: Create a new mobile app
      x-speakeasy-entity-operation: MobileApps#create`,
			expectedDataResources: []TerraformDataResource{
				{
					Name:         "Apps",
					Description:  "Apps (DataSource)",
					OperationIDs: []string{"getAppById"},
				},
			},
			expectedManagedResources: []TerraformManagedResource{
				{
					Name:         "Apps",
					Description:  "Apps (Resource)",
					OperationIDs: []string{"getAppById"},
				},
				{
					Name:         "MobileApps",
					Description:  "MobileApps (Resource)",
					OperationIDs: []string{"createMobileApp"},
				},
			},
		},
		{
			name: "multiple operations for same entity should be consolidated",
			openAPIContent: `openapi: 3.0.3
info:
  title: Test API
  version: 1.0.0
paths:
  /users/{id}:
    get:
      operationId: getUserById
      summary: Get user by ID
      description: Retrieve a specific user
      x-speakeasy-entity-operation: User#read
  /users:
    post:
      operationId: createUser
      summary: Create user
      description: Create a new user
      x-speakeasy-entity-operation: User#create
    put:
      operationId: updateUser
      summary: Update user
      description: Update an existing user
      x-speakeasy-entity-operation: User#update`,
			expectedDataResources: []TerraformDataResource{
				{
					Name:         "User",
					Description:  "User (DataSource)",
					OperationIDs: []string{"getUserById"},
				},
			},
			expectedManagedResources: []TerraformManagedResource{
				{
					Name:         "User",
					Description:  "User (Resource)",
					OperationIDs: []string{"getUserById", "createUser", "updateUser"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Parse the OpenAPI document
			doc, _, err := openapi.Unmarshal(t.Context(), bytes.NewBufferString(tt.openAPIContent))
			require.NoError(t, err, "Failed to parse OpenAPI document: %v", err)

			// Create FastAST
			fastAST, err := NewFastAST(t.Context(), doc)
			require.NoError(t, err, "Failed to create FastAST: %v", err)

			// Check Terraform resources
			if len(tt.expectedDataResources) > 0 || len(tt.expectedManagedResources) > 0 {
				require.NotNil(t, fastAST.Terraform, "Expected Terraform to be initialized")

				// Verify data resources
				assert.Len(t, fastAST.Terraform.DataResources, len(tt.expectedDataResources), "Expected %d data resources, got %d", len(tt.expectedDataResources), len(fastAST.Terraform.DataResources))

				for i, expected := range tt.expectedDataResources {
					if i >= len(fastAST.Terraform.DataResources) {
						continue
					}
					actual := fastAST.Terraform.DataResources[i]
					assert.Equal(t, expected.Name, actual.Name, "Data resource %d: expected name %s, got %s", i, expected.Name, actual.Name)
					assert.Equal(t, expected.Description, actual.Description, "Data resource %d: expected description %s, got %s", i, expected.Description, actual.Description)
					assert.Len(t, expected.OperationIDs, len(actual.OperationIDs), "Data resource %d: expected %d operation IDs, got %d", i, len(expected.OperationIDs), len(actual.OperationIDs))
				}

				// Verify managed resources
				assert.Len(t, fastAST.Terraform.ManagedResources, len(tt.expectedManagedResources), "Expected %d managed resources, got %d", len(tt.expectedManagedResources), len(fastAST.Terraform.ManagedResources))

				for i, expected := range tt.expectedManagedResources {
					if i >= len(fastAST.Terraform.ManagedResources) {
						continue
					}
					actual := fastAST.Terraform.ManagedResources[i]
					assert.Equal(t, expected.Name, actual.Name, "Managed resource %d: expected name %s, got %s", i, expected.Name, actual.Name)
					assert.Equal(t, expected.Description, actual.Description, "Managed resource %d: expected description %s, got %s", i, expected.Description, actual.Description)
					assert.Len(t, expected.OperationIDs, len(actual.OperationIDs), "Managed resource %d: expected %d operation IDs, got %d", i, len(expected.OperationIDs), len(actual.OperationIDs))
				}
			} else if fastAST.Terraform != nil {
				// Should have no Terraform resources
				assert.Empty(t, fastAST.Terraform.DataResources, "Expected no Terraform resources, but got some")
				assert.Empty(t, fastAST.Terraform.ManagedResources, "Expected no Terraform resources, but got some")
			}
		})
	}
}
