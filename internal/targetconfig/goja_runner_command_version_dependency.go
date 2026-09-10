package targetconfig

import (
	"context"
	"fmt"
	"regexp"

	"github.com/hashicorp/go-version"
	"github.com/speakeasy-api/openapi-generation/v2/internal/processrunner"
)

// Represents a [processrunner.CommandVersionDependency] configuration in
// primitive Go types for simplifying the conversion from goja. Use the Convert
// method to finish the conversion for generator usage.
type GojaRunnerCommandDependencyVersion struct {
	Args       []string
	Regex      string
	MinVersion string
}

// Converts the GojaRunnerCommandDependencyVersion into its
// [processrunner.CommandVersionDependency] form for the generator.
func (g *GojaRunnerCommandDependencyVersion) Convert(ctx context.Context) (*processrunner.CommandVersionDependency, error) {
	if g == nil {
		return nil, nil
	}

	result := &processrunner.CommandVersionDependency{
		Args: g.Args,
	}

	minimumVersion, err := version.NewVersion(g.MinVersion)
	if err != nil {
		return nil, fmt.Errorf("error parsing dependency minimum version %s from target: %w", g.MinVersion, err)
	}

	result.MinimumVersion = minimumVersion

	pattern, err := regexp.Compile(g.Regex)
	if err != nil {
		return nil, fmt.Errorf("error parsing dependency version regular expression %s from target: %w", g.Regex, err)
	}

	result.Pattern = pattern

	return result, nil
}
