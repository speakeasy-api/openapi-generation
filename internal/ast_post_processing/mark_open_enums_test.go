package ast_post_processing

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/stretchr/testify/assert"
)

func responseEnum(name string, values ...string) *ast.TypeDef {
	return &ast.TypeDef{
		Name:           name,
		Type:           ast.DataTypeEnum,
		UsedInResponse: true,
		Enum: &ast.Enum{
			Type:   ast.NewType(&ast.TypeDef{Type: ast.DataTypeString}, nil),
			Values: values,
		},
	}
}

func enumTaggedMember(name string, tagEnum *ast.TypeDef) *ast.TypeDef {
	return &ast.TypeDef{
		Name: name,
		Type: ast.DataTypeClass,
		Fields: ast.Fields{
			{
				Name:         "type",
				OriginalName: "type",
				Type:         tagEnum,
			},
		},
	}
}

func runMarkOpenEnums(openSingleValue bool, types ...*ast.TypeDef) {
	MarkOpenEnums(types, map[string]any{"forwardCompatibleEnumsByDefault": true}, openSingleValue)
}

func TestMarkOpenEnumsOpensMultiValueResponseEnumRegardlessOfFix(t *testing.T) {
	for _, openSingleValue := range []bool{false, true} {
		pair := responseEnum("Pair", "python", "bash")

		runMarkOpenEnums(openSingleValue, pair)

		assert.True(t, pair.Enum.Open, "openSingleValue=%v", openSingleValue)
	}
}

func TestMarkOpenEnumsKeepsSingleValueResponseEnumClosedWithoutFix(t *testing.T) {
	language := responseEnum("Language", "python")

	runMarkOpenEnums(false, language)

	assert.False(t, language.Enum.Open)
}

func TestMarkOpenEnumsOpensSingleValueResponseEnumWithFix(t *testing.T) {
	language := responseEnum("Language", "python")

	runMarkOpenEnums(true, language)

	assert.True(t, language.Enum.Open)
}

func TestMarkOpenEnumsKeepsSingleValueDiscriminatorTagClosedWithFix(t *testing.T) {
	textTag := responseEnum("TextContentType", "text")
	imageTag := responseEnum("ImageContentType", "image")
	textMember := enumTaggedMember("TextContent", textTag)
	imageMember := enumTaggedMember("ImageContent", imageTag)
	union := &ast.TypeDef{
		Name:            "Content",
		Type:            ast.DataTypeUnion,
		UsedInResponse:  true,
		AssociatedTypes: ast.TypeDefs{textMember, imageMember},
		Discriminator: &ast.Discriminator{
			TypePropertyName: "type",
			Mapping: ast.DiscriminatorMappings{
				{Name: "text", Type: textMember},
				{Name: "image", Type: imageMember},
			},
		},
	}
	language := responseEnum("Language", "python")

	runMarkOpenEnums(true, union, textMember, imageMember, textTag, imageTag, language)

	assert.False(t, textTag.Enum.Open, "discriminator tag enum must stay closed")
	assert.False(t, imageTag.Enum.Open, "discriminator tag enum must stay closed")
	assert.True(t, language.Enum.Open, "unrelated single-value enum opens")
}

func TestMarkOpenEnumsKeepsNestedUnionDiscriminatorTagsClosedWithFix(t *testing.T) {
	audioTag := responseEnum("AudioMediaType", "media")
	videoTag := responseEnum("VideoMediaType", "media")
	audioMember := enumTaggedMember("AudioMedia", audioTag)
	videoMember := enumTaggedMember("VideoMedia", videoTag)
	mediaUnion := &ast.TypeDef{
		Name:            "Media",
		Type:            ast.DataTypeUnion,
		AssociatedTypes: ast.TypeDefs{audioMember, videoMember},
	}
	contentUnion := &ast.TypeDef{
		Name:           "Content",
		Type:           ast.DataTypeUnion,
		UsedInResponse: true,
		Discriminator: &ast.Discriminator{
			TypePropertyName: "type",
			Mapping:          ast.DiscriminatorMappings{{Name: "media", Type: mediaUnion}},
		},
	}

	runMarkOpenEnums(true, contentUnion, mediaUnion, audioMember, videoMember, audioTag, videoTag)

	assert.False(t, audioTag.Enum.Open)
	assert.False(t, videoTag.Enum.Open)
}

func TestMarkOpenEnumsMatchesDiscriminatorTagByOriginalName(t *testing.T) {
	tag := responseEnum("KindTag", "widget")
	member := &ast.TypeDef{
		Name: "Widget",
		Type: ast.DataTypeClass,
		Fields: ast.Fields{
			{Name: "ObjectKind", OriginalName: "object_kind", Type: tag},
		},
	}
	union := &ast.TypeDef{
		Name:            "Item",
		Type:            ast.DataTypeUnion,
		AssociatedTypes: ast.TypeDefs{member},
		Discriminator: &ast.Discriminator{
			TypePropertyName: "object_kind",
			Mapping:          ast.DiscriminatorMappings{{Name: "widget", Type: member}},
		},
	}

	runMarkOpenEnums(true, union, member, tag)

	assert.False(t, tag.Enum.Open)
}

func TestMarkOpenEnumsRespectsExplicitSettingOnSingleValueEnum(t *testing.T) {
	language := responseEnum("Language", "python")
	language.Extensions = &ast.TypeDefExtensions{All: map[string]any{"x-speakeasy-unknown-values": "disallow"}}

	runMarkOpenEnums(true, language)

	assert.False(t, language.Enum.Open)
}

func TestMarkOpenEnumsSkipsRequestOnlyAndEmptyEnums(t *testing.T) {
	requestOnly := responseEnum("Mode", "fast")
	requestOnly.UsedInResponse = false
	empty := responseEnum("Empty")

	runMarkOpenEnums(true, requestOnly, empty)

	assert.False(t, requestOnly.Enum.Open)
	assert.False(t, empty.Enum.Open)
}

func TestMarkOpenEnumsDoesNothingWhenForwardCompatibilityIsOff(t *testing.T) {
	language := responseEnum("Language", "python")
	pair := responseEnum("Pair", "python", "bash")

	MarkOpenEnums([]*ast.TypeDef{language, pair}, map[string]any{"forwardCompatibleEnumsByDefault": false}, true)

	assert.False(t, language.Enum.Open)
	assert.False(t, pair.Enum.Open)
}
