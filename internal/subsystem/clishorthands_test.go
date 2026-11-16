package subsystem

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCLIShorthandsRecord(t *testing.T) {
	t.Parallel()

	letters := map[string]any{"flags": map[string]any{"names": "n"}}
	empty := map[string]any{"flags": map[string]any{}, "retired": map[string]any{}}

	var unstarted CLIShorthands
	unstarted.Record("CreatePet", "json", letters, false)
	result, recorded := unstarted.Result()
	assert.False(t, recorded)
	assert.Nil(t, result)

	var c CLIShorthands
	c.Start()
	c.Record("CreatePet", "json", empty, true)
	c.Record("CreatePet", "multipart", letters, true)
	c.Record("ListPets", "json", letters, false)
	c.Record("ListPets", "json", empty, false)
	c.Record("UploadPhoto", "json", empty, true)
	c.Record("UploadPhoto", "multipart", empty, true)

	result, recorded = c.Result()
	assert.True(t, recorded)
	assert.Equal(t, map[string]any{
		"CreatePet": map[string]any{
			"json":      map[string]any{},
			"multipart": letters,
		},
	}, result)
}
