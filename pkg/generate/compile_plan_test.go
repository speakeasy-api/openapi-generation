package generate

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCompilePlan(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		compileEnabled bool
		pristineMode   bool
		patchesEnabled bool
		wantPristine   bool
		wantFinal      bool
	}{
		// Compile disabled - nothing runs
		{"disabled", false, true, true, false, false},

		// compilePristine=true (default behavior)
		{"pristine=true patches=false", true, true, false, true, false},
		{"pristine=true patches=true", true, true, true, true, true},

		// compilePristine=false (skip normalization, useful when SDK doesn't compile)
		{"pristine=false patches=false", true, false, false, false, true},
		{"pristine=false patches=true", true, false, true, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotPristine, gotFinal := compilePlan(tt.compileEnabled, tt.pristineMode, tt.patchesEnabled)

			assert.Equal(t, tt.wantPristine, gotPristine, "runPristine")
			assert.Equal(t, tt.wantFinal, gotFinal, "runFinal")
		})
	}
}
