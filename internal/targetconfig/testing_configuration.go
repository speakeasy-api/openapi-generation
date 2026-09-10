package targetconfig

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/processrunner"
)

// Represents target-specific testing configuration.
type TestingConfiguration struct {
	// Processes and dependencies for compiling generated target testing.
	Compile *CompileConfiguration

	// Processes and dependencies for running generated target testing.
	Runner *processrunner.Runner

	// Enabled if all internal testing should be skipped for this target when
	// reporting internal test coverage.
	//
	// TODO: Migrate from features.ts.
	// SkipAllInternalTests bool

	// Set of internal test names from internal/features/tests.go which should
	// be skipped for this target when reporting internal test coverage.
	//
	// TODO: Migrate from features.ts.
	// SkipInternalTests []string
}
