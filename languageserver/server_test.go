package languageserver

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerValidation(t *testing.T) {
	t.Skip() // Not an automated test just here for manual testing
	s := NewServer("dev", nil, nil)

	path := "../tests/specs/review.yaml"
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	diag := s.Validate(t.Context(), content)
	assert.NotNil(t, diag)
}
