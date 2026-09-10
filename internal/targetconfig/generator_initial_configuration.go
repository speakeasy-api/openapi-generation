package targetconfig

// Represents the initial generator configuration presented to the target to
// fetch the target configuration.
type GeneratorInitialConfiguration struct {
	// Runtime environment variables, e.g. UNITY_ROOT.
	Env map[string]string

	// Target-specific configuration passed by customer.
	LangCfg map[string]any

	// Generation output directory.
	OutDir string

	// Runtime platform, e.g. GOOS environment variable.
	Platform string

	// Enabled when the generation includes publishing.
	Publish bool

	// URL of the repository for the generated code.
	RepoURL string

	// Enabled when the generation includes testing.
	Tests bool
}
