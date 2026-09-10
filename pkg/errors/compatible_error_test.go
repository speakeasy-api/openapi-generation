package errors

import (
	stderrors "errors"
	"fmt"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors/testtypes"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type incompatibleError string

func (incompatibleError) Error() string {
	return "incompatible"
}

func TestErrorCompatibility(t *testing.T) {
	const sentinel = Error("sentinel")
	const sameText = Error("sentinel")
	cause := stderrors.New("cause")

	assert.Equal(t, "sentinel", sentinel.Error())
	assert.Equal(t, sentinel, sameText)
	assert.Equal(t, "value", map[Error]string{sentinel: "value"}[sameText])
	require.ErrorIs(t, sentinel, stderrors.New("sentinel"))
	require.NotErrorIs(t, stderrors.New("sentinel"), sentinel)

	wrapped := sentinel.Wrap(cause)
	require.ErrorIs(t, sentinel, wrapped)
	require.NotErrorIs(t, wrapped, Error("other"))
	assert.Equal(t, "sentinel -- cause", wrapped.Error())
	require.ErrorIs(t, wrapped, sentinel)
	require.ErrorIs(t, wrapped, cause)
	require.ErrorIs(t, fmt.Errorf("outer: %w", wrapped), sentinel)
	require.ErrorIs(t, fmt.Errorf("outer: %w", wrapped), cause)

	var asTarget Error
	require.ErrorAs(t, fmt.Errorf("outer: %w", wrapped), &asTarget)
	assert.Equal(t, sentinel, asTarget)

	var externalErrorTarget testtypes.Error
	require.ErrorAs(t, fmt.Errorf("outer: %w", wrapped), &externalErrorTarget)
	assert.Equal(t, testtypes.Error(sentinel), externalErrorTarget)

	var nonErrorTarget incompatibleError
	assert.NotErrorAs(t, wrapped, &nonErrorTarget)

	wrappedNil := sentinel.Wrap(nil)
	assert.Equal(t, "sentinel", wrappedNil.Error())
	assert.NoError(t, stderrors.Unwrap(wrappedNil))
}
