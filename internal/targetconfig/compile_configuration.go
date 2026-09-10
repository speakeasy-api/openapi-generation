package targetconfig

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/processrunner"
)

// Represents target-specific configuration for compiling generated code.
type CompileConfiguration struct {
	// Dependencies and processes for running generated code compilation.
	Runner *processrunner.Runner
}
