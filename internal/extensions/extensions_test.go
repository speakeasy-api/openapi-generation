package extensions

import (
	"testing"

	"github.com/speakeasy-api/openapi/openapi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtensions_Ignore_Nil_Extensions(t *testing.T) {
	e := Extensions{}

	ignored, err := e.Ignore(nil)
	require.NoError(t, err)
	assert.False(t, ignored)
}

func TestExtensions_GetServerID_Nil_Extensions(t *testing.T) {
	e := Extensions{}

	serverID, err := e.GetServerID(&openapi.Server{})
	require.NoError(t, err)
	assert.Empty(t, serverID)
}
