package targetconfig

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/processrunner"
)

// Represents target-specific configuration for performing static analysis of
// generated code.
type LintConfiguration struct {
	// Dependencies and processes for running generated code linting.
	Runner *processrunner.Runner
}
