package targetconfig

import (
	"context"

	"github.com/mitchellh/mapstructure"
)

// Creates a GojaConfiguration from a (goja.Value).Export() value.
func NewGojaConfiguration(value any) (*GojaConfiguration, error) {
	result := new(GojaConfiguration)
	decoderConfig := &mapstructure.DecoderConfig{
		ErrorUnused: true,
		Result:      result,
	}

	decoder, err := mapstructure.NewDecoder(decoderConfig)

	if err != nil {
		return nil, err
	}

	if err := decoder.Decode(value); err != nil {
		return nil, err
	}

	return result, nil
}

// Represents the entire target configuration in primitive Go types for
// simplifying the conversion from goja. Use the Convert method to finish the
// conversion for generator usage.
type GojaConfiguration struct {
	Compile *GojaCompileConfiguration
	Lint    *GojaLintConfiguration
	Testing *GojaTestingConfiguration
}

// Converts the GojaConfiguration into its Configuration form for the generator.
func (c *GojaConfiguration) Convert(ctx context.Context) (*Configuration, error) {
	if c == nil {
		return nil, nil
	}

	compile, err := c.Compile.Convert(ctx)

	if err != nil {
		return nil, err
	}

	lint, err := c.Lint.Convert(ctx)

	if err != nil {
		return nil, err
	}

	testing, err := c.Testing.Convert(ctx)

	if err != nil {
		return nil, err
	}

	result := &Configuration{
		Compile: compile,
		Lint:    lint,
		Testing: testing,
	}

	return result, nil
}
