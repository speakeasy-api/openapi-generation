package namer

import (
	"bytes"
	"context"
	"testing"

	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/references"
	"github.com/stretchr/testify/require"
)

// TestReferenceChainRegression demonstrates a regression in GetReferenceChain()
// behavior between the old and new openapi library versions.
//
// This test documents the EXPECTED behavior that the openapi library should preserve.
//
// EXPECTED: When accessing a schema through a property of another schema,
// the reference chain should include BOTH the parent schema and the target schema.
//
// Example: ContainerSchema.nested → $ref: SharedSchema
// Expected chain: [ContainerSchema, SharedSchema] (length 2)
// Actual chain (regression): [SharedSchema] (length 1)
//
// This context is critical for emulating libopenapi's nested reference bug behavior.
func TestReferenceChainRegression(t *testing.T) {
	openAPIYAML := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /container:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/ContainerSchema"
components:
  schemas:
    SharedSchema:
      type: object
      properties:
        value:
          type: string
    ContainerSchema:
      type: object
      properties:
        nested:
          $ref: "#/components/schemas/SharedSchema"
`

	ctx := context.Background()
	doc, _, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(openAPIYAML)))
	require.NoError(t, err)

	resolveOpts := references.ResolveOptions{
		RootDocument: doc,
	}

	// Get the ContainerSchema from the path response
	pathItem := doc.GetPaths().GetOrZero("/container")
	_, err = pathItem.Resolve(ctx, resolveOpts)
	require.NoError(t, err)

	response := pathItem.GetResolvedObject().Get().GetResponses().GetOrZero("200")
	_, err = response.Resolve(ctx, resolveOpts)
	require.NoError(t, err)

	containerSchema := response.GetResolvedObject().
		GetContent().GetOrZero("application/json").
		GetSchema()
	require.NotNil(t, containerSchema)

	_, err = containerSchema.Resolve(ctx, resolveOpts)
	require.NoError(t, err)

	// Resolve ContainerSchema and get its 'nested' property
	containerResolved := containerSchema.MustGetResolvedSchema()
	props := containerResolved.GetSchema().GetProperties()
	nestedProp := props.GetOrZero("nested")
	require.NotNil(t, nestedProp)

	// Resolve the nested property
	_, err = nestedProp.Resolve(ctx, resolveOpts)
	require.NoError(t, err)

	// Get the reference chain for the nested property
	// This is where the regression occurs
	nestedResolved := nestedProp.MustGetResolvedSchema()
	chain := nestedResolved.GetReferenceChain()

	// EXPECTED BEHAVIOR (old openapi library):
	// The chain should show the ACCESS PATH through ContainerSchema
	// Chain length: 2
	//   Chain[0]: #/components/schemas/ContainerSchema (parent/access context)
	//   Chain[1]: #/components/schemas/SharedSchema (target)
	//
	// ACTUAL BEHAVIOR (new openapi library with pre-resolved index):
	// The chain only shows the target schema
	// Chain length: 1
	//   Chain[0]: #/components/schemas/SharedSchema (target only)

	// Log the actual chain for debugging BEFORE assertions
	t.Logf("Reference chain length: %d", len(chain))
	for i, entry := range chain {
		t.Logf("  Chain[%d]: %s", i, entry.Reference)
	}

	require.Len(t, chain, 2,
		"Reference chain should include both parent (ContainerSchema) and target (SharedSchema)")

	if len(chain) == 2 {
		require.Equal(t, "#/components/schemas/ContainerSchema", string(chain[0].Reference),
			"First chain entry should be the parent schema (ContainerSchema)")
		require.Equal(t, "#/components/schemas/SharedSchema", string(chain[1].Reference),
			"Second chain entry should be the target schema (SharedSchema)")
	}
}

// TestReferenceChainDirectAccess tests that direct component access
// (not through a property) has a simpler chain.
func TestReferenceChainDirectAccess(t *testing.T) {
	openAPIYAML := `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths:
  /direct:
    get:
      responses:
        '200':
          description: OK
          content:
            application/json:
              schema:
                $ref: "#/components/schemas/SharedSchema"
components:
  schemas:
    SharedSchema:
      type: object
      properties:
        value:
          type: string
`

	ctx := context.Background()
	doc, _, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(openAPIYAML)))
	require.NoError(t, err)

	resolveOpts := references.ResolveOptions{
		RootDocument: doc,
	}

	// Get the schema from the path response (direct reference)
	pathItem := doc.GetPaths().GetOrZero("/direct")
	_, err = pathItem.Resolve(ctx, resolveOpts)
	require.NoError(t, err)

	response := pathItem.GetResolvedObject().Get().GetResponses().GetOrZero("200")
	_, err = response.Resolve(ctx, resolveOpts)
	require.NoError(t, err)

	directSchema := response.GetResolvedObject().
		GetContent().GetOrZero("application/json").
		GetSchema()
	require.NotNil(t, directSchema)

	_, err = directSchema.Resolve(ctx, resolveOpts)
	require.NoError(t, err)

	// Get the reference chain for direct access
	directResolved := directSchema.MustGetResolvedSchema()
	chain := directResolved.GetReferenceChain()

	// For direct access, the chain should just be the target
	// This should be consistent across both old and new library versions
	require.Len(t, chain, 1,
		"Direct access should have chain length 1")

	if len(chain) == 1 {
		require.Equal(t, "#/components/schemas/SharedSchema", string(chain[0].Reference),
			"Chain should contain the target schema")
	}

	t.Logf("Direct access chain length: %d", len(chain))
	for i, entry := range chain {
		t.Logf("  Chain[%d]: %s", i, entry.Reference)
	}
}
