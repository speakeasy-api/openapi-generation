package namer

import (
	"bytes"
	"context"
	"testing"
	"testing/fstest"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/speakeasy-api/openapi/references"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// A property whose value is a whole-document external $ref (no "#" fragment)
// resolves to an inline object, and GetRefName returns "" for the fragment-less
// ref. Under qualified mode the empty name must not be treated as an unnamed
// inline schema and prefixed with the parent chain — $refs are never qualified.
// This exercises the s.GetRef() == "" guard in GetTypeName, which the in-memory
// component harness cannot reach because internal "#/components" refs always
// yield a non-empty name.
func TestGetTypeName_ExternalRefWithoutFragmentNotQualified(t *testing.T) {
	const rootYAML = `openapi: 3.0.0
info:
  title: Test API
  version: 1.0.0
paths: {}
components:
  schemas:
    Parent:
      type: object
      properties:
        external:
          $ref: "external.yaml"
`
	const externalYAML = `type: object
properties:
  value:
    type: string
`

	ctx := context.Background()
	doc, _, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(rootYAML)))
	require.NoError(t, err)

	resolveOpts := references.ResolveOptions{
		RootDocument:   doc,
		TargetLocation: "test.yaml",
		VirtualFS:      fstest.MapFS{"external.yaml": {Data: []byte(externalYAML)}},
	}

	parent := doc.GetComponents().GetSchemas().GetOrZero("Parent")
	require.NotNil(t, parent)
	_, err = parent.Resolve(ctx, resolveOpts)
	require.NoError(t, err)

	externalProp := parent.MustGetResolvedSchema().GetSchema().GetProperties().GetOrZero("external")
	require.NotNil(t, externalProp)
	_, err = externalProp.Resolve(ctx, resolveOpts)
	require.NoError(t, err)

	// Precondition: a fragment-less external ref yields no ref name.
	require.NotEmpty(t, externalProp.GetRef(), "property should carry the external ref")
	refName, _ := GetRefName(externalProp.GetRef())
	require.Empty(t, refName, "fragment-less external ref has no derivable name")

	n := newQualifiedNamer(t)

	// Stack shape the walker produces for Parent.external: a component ancestor
	// (refName + component) with the property leaf on top — exactly what
	// qualifyInlineName folds into ParentExternal for a genuine inline schema.
	contextStack := ast.ContextStack{
		{Type: ast.ContextTypeRefName, Identifier: "Parent", IdentifierForNaming: pointer.From("Parent")},
		{Type: ast.ContextTypeComponent, Identifier: "true", Used: true},
		{Type: ast.ContextTypeProperty, Identifier: "external", IdentifierForNaming: pointer.From("external")},
	}

	name, _, _, err := n.GetTypeName(ctx, externalProp, contextStack, "", false)
	require.NoError(t, err)

	assert.NotEqual(t, "ParentExternal", name, "external $ref must not be qualified with the parent chain")
	assert.Equal(t, "external", name, "external $ref falls back to the unprefixed leaf name")
}

func newQualifiedNamer(t *testing.T) *Namer {
	t.Helper()

	target := types.Target{Target: "go"}
	cfg := configuration.New(&config.Configuration{
		Generation: config.Generation{
			NameResolution: config.NameResolutionQualified,
			Fixes:          &config.Fixes{},
		},
	}, target.Target)

	return New(&subsystem.Subsystem{
		Config:     cfg,
		Target:     target,
		Extensions: extensions.New(target),
	})
}
