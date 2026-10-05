package extensions

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	oasextensions "github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestGetEnumGroups(t *testing.T) {
	var enumNode yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("[red, null, blue, blue]"), &enumNode))
	schema := &oas3.Schema{
		Enum:       enumNode.Content[0].Content,
		Extensions: oasextensions.New(),
	}
	var groupNode yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("[ ' Primary ', '', Secondary, 'Ignored duplicate' ]"), &groupNode))
	schema.Extensions.Set(ExtEnumGroups.Name(), groupNode.Content[0])

	groups, groupMap, err := New(types.NewTargetFromTemplate("go")).GetEnumGroups(schema)
	require.NoError(t, err)
	require.Equal(t, []string{"Primary", "", "Secondary", "Ignored duplicate"}, groups)
	require.Nil(t, groupMap)

	var mapNode yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("red: ' First '\nblue: ''"), &mapNode))
	schema.Extensions.Set(ExtEnumGroups.Name(), mapNode.Content[0])
	groups, groupMap, err = New(types.NewTargetFromTemplate("go")).GetEnumGroups(schema)
	require.NoError(t, err)
	require.Nil(t, groups)
	require.Equal(t, map[string]string{"red": "First"}, groupMap)
}

func TestGetEnumGroupsScalarKeys(t *testing.T) {
	var enumNode, groupNode yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("[1, 2]"), &enumNode))
	require.NoError(t, yaml.Unmarshal([]byte("1: ' Primary '\n'2': Secondary"), &groupNode))
	schema := &oas3.Schema{Enum: enumNode.Content[0].Content, Extensions: oasextensions.New()}
	schema.Extensions.Set(ExtEnumGroups.Name(), groupNode.Content[0])
	groups, groupMap, err := New(types.NewTargetFromTemplate("go")).GetEnumGroups(schema)
	require.NoError(t, err)
	require.Nil(t, groups)
	require.Equal(t, map[string]string{"1": "Primary", "2": "Secondary"}, groupMap)
}

func TestGetEnumGroupsRejectsInvalidValues(t *testing.T) {
	var enumNode yaml.Node
	require.NoError(t, yaml.Unmarshal([]byte("[red, blue]"), &enumNode))
	cases := []struct {
		name string
		raw  string
	}{
		{name: "non-string list member", raw: "[Primary, 1]"},
		{name: "wrong list length", raw: "[Primary]"},
		{name: "non-string map value", raw: "red: 1"},
		{name: "unknown map key", raw: "green: Other"},
		{name: "duplicate map key", raw: "red: Primary\nred: Secondary"},
		{name: "scalar", raw: "Primary"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var groupNode yaml.Node
			require.NoError(t, yaml.Unmarshal([]byte(tc.raw), &groupNode))
			extensions := oasextensions.New()
			extensions.Set(ExtEnumGroups.Name(), groupNode.Content[0])
			schema := &oas3.Schema{Enum: enumNode.Content[0].Content, Extensions: extensions}
			_, _, err := New(types.NewTargetFromTemplate("go")).GetEnumGroups(schema)
			require.Error(t, err)
		})
	}
}
