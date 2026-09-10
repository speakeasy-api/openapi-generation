package ast_post_processing

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func stringType() *ast.TypeDef {
	return &ast.TypeDef{Type: ast.DataTypeString}
}

func taggedMember(name, tagProp, tagValue string) *ast.TypeDef {
	return &ast.TypeDef{
		Name: name,
		Type: ast.DataTypeClass,
		Fields: ast.Fields{
			{
				Name:         tagProp,
				OriginalName: tagProp,
				Type:         stringType(),
				Const:        &ast.AnyValue{Value: tagValue},
			},
		},
	}
}

func untaggedMember(name string) *ast.TypeDef {
	return &ast.TypeDef{
		Name: name,
		Type: ast.DataTypeClass,
		Fields: ast.Fields{
			{
				Name:         "payload",
				OriginalName: "payload",
				Type:         stringType(),
			},
		},
	}
}

func runMarkUnionsOpen(t *testing.T, union *ast.TypeDef) []error {
	t.Helper()
	warnings := logging.NewWarningLogger(false, nil)
	ctx := logging.WithWarningLogger(context.Background(), warnings)
	MarkUnionsOpen(ctx, []*ast.TypeDef{union}, map[string]interface{}{
		"forwardCompatibleUnionsByDefault": "tagged-only",
	})
	return warnings.GetWarnings()
}

func TestMarkUnionsOpenWarnsWhenMixedTagUnionStaysClosed(t *testing.T) {
	union := &ast.TypeDef{
		Name:           "MixedTagRecord",
		Type:           ast.DataTypeUnion,
		UsedInResponse: true,
		AssociatedTypes: ast.TypeDefs{
			taggedMember("DraftOptions", "type", "draft"),
			taggedMember("PublishedOptions", "type", "published"),
			taggedMember("LocalOptions", "request", "local"),
			taggedMember("RemoteOptions", "request", "remote"),
		},
	}

	warnings := runMarkUnionsOpen(t, union)

	assert.False(t, union.IsUnionOpen, "mixed-tag union cannot be opened")
	require.Len(t, warnings, 1, "expected a warning for the silent forward-compat downgrade")
	assert.Contains(t, warnings[0].Error(), "MixedTagRecord")
	assert.Contains(t, warnings[0].Error(), "type")
	assert.Contains(t, warnings[0].Error(), "request")
}

// A union where two or more members are tagged but another has no tag at all
// is also a silent downgrade and must warn.
func TestMarkUnionsOpenWarnsWhenPartiallyTaggedUnionStaysClosed(t *testing.T) {
	union := &ast.TypeDef{
		Name:           "PartiallyTagged",
		Type:           ast.DataTypeUnion,
		UsedInResponse: true,
		AssociatedTypes: ast.TypeDefs{
			taggedMember("TaggedA", "type", "a"),
			taggedMember("TaggedB", "type", "b"),
			untaggedMember("Untagged"),
		},
	}

	warnings := runMarkUnionsOpen(t, union)

	assert.False(t, union.IsUnionOpen)
	require.Len(t, warnings, 1)
	assert.Contains(t, warnings[0].Error(), "PartiallyTagged")
	assert.Contains(t, warnings[0].Error(), "Untagged")
}

func TestMarkUnionsOpenDoesNotWarnForSingleTaggedMember(t *testing.T) {
	union := &ast.TypeDef{
		Name:           "SingleTagged",
		Type:           ast.DataTypeUnion,
		UsedInResponse: true,
		AssociatedTypes: ast.TypeDefs{
			taggedMember("Record", "status", "completed"),
			untaggedMember("EventStream"),
		},
	}

	warnings := runMarkUnionsOpen(t, union)

	assert.False(t, union.IsUnionOpen)
	assert.Empty(t, warnings)
}

func TestMarkUnionsOpenDoesNotWarnWhenMapMemberAbsorbsUnknowns(t *testing.T) {
	union := &ast.TypeDef{
		Name:           "RecordPayload",
		Type:           ast.DataTypeUnion,
		UsedInResponse: true,
		AssociatedTypes: ast.TypeDefs{
			taggedMember("CompactPayload", "type", "audio"),
			taggedMember("DetailedPayload", "type", "image"),
			taggedMember("SummaryPayload", "type", "text"),
			{Name: "FreeForm", Type: ast.DataTypeMap},
		},
	}

	warnings := runMarkUnionsOpen(t, union)

	assert.False(t, union.IsUnionOpen)
	assert.Empty(t, warnings)
}

func openEnumMember(name, prop string, values ...string) *ast.TypeDef {
	return &ast.TypeDef{
		Name: name,
		Type: ast.DataTypeClass,
		Fields: ast.Fields{
			{
				Name:         prop,
				OriginalName: prop,
				Type: &ast.TypeDef{
					Type: ast.DataTypeEnum,
					Enum: &ast.Enum{
						Type:   stringType(),
						Values: values,
						Open:   true,
					},
				},
			},
		},
	}
}

// A required open enum accepts arbitrary values, so it can never be a
// discriminator (mirrors parsePrimitiveDataType in the inference code).
// Members whose only constrained property is an open enum are untagged;
// a union of such members shows no discriminated-union intent and must
// NOT warn.
func TestMarkUnionsOpenDoesNotWarnForOpenEnumMembers(t *testing.T) {
	union := &ast.TypeDef{
		Name:           "OpenEnumTagged",
		Type:           ast.DataTypeUnion,
		UsedInResponse: true,
		AssociatedTypes: ast.TypeDefs{
			openEnumMember("ModelA", "model", "a-1", "a-2"),
			openEnumMember("ModelB", "model", "b-1"),
			untaggedMember("Other"),
		},
	}

	warnings := runMarkUnionsOpen(t, union)

	assert.False(t, union.IsUnionOpen)
	assert.Empty(t, warnings)
}

// An open-enum property must also not be listed as a tag candidate in the
// warning summary when genuinely const-tagged members trigger it.
func TestMarkUnionsOpenSummaryExcludesOpenEnumProperties(t *testing.T) {
	mixed := taggedMember("TaggedC", "type", "c")
	mixed.Fields = append(mixed.Fields, &ast.FieldDef{
		Name:         "model",
		OriginalName: "model",
		Type: &ast.TypeDef{
			Type: ast.DataTypeEnum,
			Enum: &ast.Enum{
				Type:   stringType(),
				Values: []string{"m-1"},
				Open:   true,
			},
		},
	})

	union := &ast.TypeDef{
		Name:           "MixedWithOpenEnum",
		Type:           ast.DataTypeUnion,
		UsedInResponse: true,
		AssociatedTypes: ast.TypeDefs{
			taggedMember("TaggedA", "type", "a"),
			taggedMember("TaggedB", "request", "b"),
			mixed,
		},
	}

	warnings := runMarkUnionsOpen(t, union)

	require.Len(t, warnings, 1)
	assert.NotContains(t, warnings[0].Error(), "model")
}

// Unions with no tagged members at all (e.g. string | int) show no intent to
// be discriminated; they must NOT warn.
func TestMarkUnionsOpenDoesNotWarnForFullyUntaggedUnion(t *testing.T) {
	union := &ast.TypeDef{
		Name:           "FullyUntagged",
		Type:           ast.DataTypeUnion,
		UsedInResponse: true,
		AssociatedTypes: ast.TypeDefs{
			untaggedMember("A"),
			untaggedMember("B"),
		},
	}

	warnings := runMarkUnionsOpen(t, union)

	assert.False(t, union.IsUnionOpen)
	assert.Empty(t, warnings)
}

// Request-only unions never get the open-union treatment, so no warning.
func TestMarkUnionsOpenDoesNotWarnForRequestOnlyUnion(t *testing.T) {
	union := &ast.TypeDef{
		Name:           "RequestOnly",
		Type:           ast.DataTypeUnion,
		UsedInResponse: false,
		AssociatedTypes: ast.TypeDefs{
			taggedMember("A", "type", "a"),
			taggedMember("B", "request", "b"),
		},
	}

	warnings := runMarkUnionsOpen(t, union)

	assert.False(t, union.IsUnionOpen)
	assert.Empty(t, warnings)
}

// A union that successfully opened (discriminator inferred upstream) must not warn.
func TestMarkUnionsOpenDoesNotWarnForOpenedUnion(t *testing.T) {
	union := &ast.TypeDef{
		Name:           "FullyTagged",
		Type:           ast.DataTypeUnion,
		UsedInResponse: true,
		Discriminator: &ast.Discriminator{
			TypePropertyName: "type",
			Inferred:         true,
		},
		AssociatedTypes: ast.TypeDefs{
			taggedMember("A", "type", "a"),
			taggedMember("B", "type", "b"),
		},
	}

	warnings := runMarkUnionsOpen(t, union)

	assert.True(t, union.IsUnionOpen)
	assert.Empty(t, warnings)
}

// Explicit x-speakeasy-unknown-values: disallow is a deliberate opt-out; no warning.
func TestMarkUnionsOpenDoesNotWarnForExplicitOptOut(t *testing.T) {
	union := &ast.TypeDef{
		Name:           "OptedOut",
		Type:           ast.DataTypeUnion,
		UsedInResponse: true,
		Extensions: &ast.TypeDefExtensions{
			All: map[string]any{
				"x-speakeasy-unknown-values": "disallow",
			},
		},
		AssociatedTypes: ast.TypeDefs{
			taggedMember("A", "type", "a"),
			taggedMember("B", "request", "b"),
		},
	}

	warnings := runMarkUnionsOpen(t, union)

	assert.False(t, union.IsUnionOpen)
	assert.Empty(t, warnings)
}

// Java's config parameter is canonicalized to a boolean; true must behave as
// tagged-and-untagged, and false (or any other value) must leave unions closed.
func TestMarkUnionsOpenBooleanConfig(t *testing.T) {
	for _, tc := range []struct {
		name   string
		config interface{}
		open   bool
	}{
		{"boolean true opens", true, true},
		{"raw string true opens", "true", true},
		{"boolean false stays closed", false, false},
		{"legacy string false stays closed", "false", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			union := &ast.TypeDef{
				Name:           "Pet",
				Type:           ast.DataTypeUnion,
				UsedInResponse: true,
				AssociatedTypes: []*ast.TypeDef{
					taggedMember("Dog", "kind", "dog"),
					taggedMember("Cat", "kind", "cat"),
				},
			}

			warnings := logging.NewWarningLogger(false, nil)
			ctx := logging.WithWarningLogger(context.Background(), warnings)
			MarkUnionsOpen(ctx, []*ast.TypeDef{union}, map[string]interface{}{
				"forwardCompatibleUnionsByDefault": tc.config,
			})

			assert.Equal(t, tc.open, union.IsUnionOpen)
		})
	}
}
