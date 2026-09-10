package ast

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
)

func TestUsageContext_PopulateServerSelectionScope(t *testing.T) {
	tests := []struct {
		name               string
		sdk                *SDK
		expectedScopeCount int
		expectServerScope  bool
	}{
		{
			name: "Multiple servers should add server_selection scope",
			sdk: &SDK{
				Servers: &Servers{
					Servers: []*Server{
						{ID: "prod-internal-aws-us-east-2", URL: "https://api-east.example.com"},
						{ID: "prod-internal-aws-us-west-2", URL: "https://api-west.example.com"},
					},
				},
			},
			expectedScopeCount: 1,
			expectServerScope:  true,
		},
		{
			name: "Single server should not add server_selection scope",
			sdk: &SDK{
				Servers: &Servers{
					Servers: []*Server{
						{ID: "prod-internal-aws-us-east-2", URL: "https://api.example.com"},
					},
				},
			},
			expectedScopeCount: 0,
			expectServerScope:  false,
		},
		{
			name:               "No servers should not add server_selection scope",
			sdk:                &SDK{Servers: nil},
			expectedScopeCount: 0,
			expectServerScope:  false,
		},
		{
			name:               "Empty servers should not add server_selection scope",
			sdk:                &SDK{Servers: &Servers{Servers: []*Server{}}},
			expectedScopeCount: 0,
			expectServerScope:  false,
		},
		{
			name:               "Nil SDK should not add server_selection scope",
			sdk:                nil,
			expectedScopeCount: 0,
			expectServerScope:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a usage context
			uc := &UsageContext{
				SDK:    tt.sdk,
				Scopes: []UsageExampleScope{},
			}

			// Call the method under test
			uc.populateServerSelectionScope()

			// Verify the results
			if len(uc.Scopes) != tt.expectedScopeCount {
				t.Errorf("Expected %d scopes, got %d", tt.expectedScopeCount, len(uc.Scopes))
			}

			if tt.expectServerScope {
				found := false
				for _, scope := range uc.Scopes {
					if scope.Feature == "server_selection" && scope.IsGlobal {
						found = true
						break
					}
				}
				if !found {
					t.Error("Expected server_selection scope but it was not found")
				}
			} else {
				for _, scope := range uc.Scopes {
					if scope.Feature == "server_selection" {
						t.Error("Did not expect server_selection scope but it was found")
					}
				}
			}
		})
	}
}

func TestUsageContext_PopulateServerSelectionScope_NoDuplicates(t *testing.T) {
	// Create a usage context with multiple servers
	uc := &UsageContext{
		SDK: &SDK{
			Servers: &Servers{
				Servers: []*Server{
					{ID: "prod-internal-aws-us-east-2", URL: "https://api-east.example.com"},
					{ID: "prod-internal-aws-us-west-2", URL: "https://api-west.example.com"},
				},
			},
		},
		Scopes: []UsageExampleScope{
			{Feature: "server_selection", IsGlobal: true}, // Already has server_selection scope
		},
	}

	// Call the method multiple times
	uc.populateServerSelectionScope()
	uc.populateServerSelectionScope()

	// Verify no duplicates are created
	serverScopeCount := 0
	for _, scope := range uc.Scopes {
		if scope.Feature == "server_selection" && scope.IsGlobal {
			serverScopeCount++
		}
	}

	if serverScopeCount != 1 {
		t.Errorf("Expected exactly 1 server_selection scope, got %d", serverScopeCount)
	}
}

func TestUsageContext_PopulateGlobalParameterScopes_IncludesServerSelection(t *testing.T) {
	// Create a usage context with multiple servers
	uc := CreateUsageContext(
		&SDK{
			Servers: &Servers{
				Servers: []*Server{
					{ID: "prod-internal-aws-us-east-2", URL: "https://api-east.example.com"},
					{ID: "prod-internal-aws-us-west-2", URL: "https://api-west.example.com"},
				},
			},
		},
		&Operation{
			BaseOperation: BaseOperation{
				ID: "testOp",
			},
		},
		&extensions.UsageExampleConfig{},
		false,
	)

	// Call PopulateGlobalParameterScopes which should now include server selection
	uc.PopulateGlobalParameterScopes("", true)

	// Verify server_selection scope was added
	found := false
	for _, scope := range uc.Scopes {
		if scope.Feature == "server_selection" && scope.IsGlobal {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected PopulateGlobalParameterScopes to add server_selection scope for multiple servers")
	}
}

func TestUsageContext_PopulateGlobalParameterScopes_NoServerSelectionForSingleServer(t *testing.T) {
	// Create a usage context with single server
	uc := CreateUsageContext(
		&SDK{
			Servers: &Servers{
				Servers: []*Server{
					{ID: "prod-internal-aws-us-east-2", URL: "https://api.example.com"},
				},
			},
		},
		&Operation{
			BaseOperation: BaseOperation{
				ID: "testOp",
			},
		},
		&extensions.UsageExampleConfig{},
		false,
	)

	// Call PopulateGlobalParameterScopes
	uc.PopulateGlobalParameterScopes("", true)

	// Verify server_selection scope was NOT added
	for _, scope := range uc.Scopes {
		if scope.Feature == "server_selection" {
			t.Error("Did not expect server_selection scope for single server")
		}
	}
}

func TestUsageContext_PopulateGlobalParameterScopes_ConditionalServerSelection(t *testing.T) {
	// Create a usage context with multiple servers to enable server selection
	sdk := &SDK{
		Servers: &Servers{
			Servers: []*Server{
				{ID: "server1", URL: "http://server1.com"},
				{ID: "server2", URL: "http://server2.com"},
			},
		},
	}

	op := &Operation{}
	uc := CreateUsageContext(sdk, op, nil, false)

	// Test 1: When shouldIncludeServerSelection = true, server_selection scope should be added
	uc.PopulateGlobalParameterScopes("", true)

	found := false
	for _, scope := range uc.Scopes {
		if scope.Feature == "server_selection" && scope.IsGlobal {
			found = true
			break
		}
	}

	if !found {
		t.Error("Expected PopulateGlobalParameterScopes(true) to add server_selection scope")
	}

	// Test 2: When shouldIncludeServerSelection = false, server_selection scope should NOT be added
	uc2 := CreateUsageContext(sdk, op, nil, false)
	uc2.PopulateGlobalParameterScopes("", false)

	found = false
	for _, scope := range uc2.Scopes {
		if scope.Feature == "server_selection" && scope.IsGlobal {
			found = true
			break
		}
	}

	if found {
		t.Error("Expected PopulateGlobalParameterScopes(false) to NOT add server_selection scope")
	}
}
