package configuration

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

type ImportOption string

const (
	ImportOptionOpenAPI ImportOption = "openapi"
)

// ImportConfig represents the import path configuration for a particular language specifically where models should be generated
type ImportConfig struct {
	Option   ImportOption      `json:"option" yaml:"option"`
	Paths    map[string]string `json:"paths" yaml:"paths"`
	Relative bool              `json:"relative,omitempty" yaml:"relative,omitempty"`
}

// TODO in other options this won't be a singular path will need to rethink this
func (c ImportConfig) GetOperationsPath() string {
	switch c.Option {
	case ImportOptionOpenAPI:
		return c.Paths[string(ast.ScopeOperations)]
	default:
		panic(fmt.Sprintf("unknown import option %s", c.Option))
	}
}

func (c ImportConfig) GetSharedPath() string {
	switch c.Option {
	case ImportOptionOpenAPI:
		return c.Paths[string(ast.ScopeShared)]
	default:
		panic(fmt.Sprintf("unknown import option %s", c.Option))
	}
}

func (c ImportConfig) GetErrorsPath() string {
	switch c.Option {
	case ImportOptionOpenAPI:
		return c.Paths[string(ast.ScopeErrors)]
	default:
		panic(fmt.Sprintf("unknown import option %s", c.Option))
	}
}

func (c ImportConfig) GetCallbacksPath() string {
	switch c.Option {
	case ImportOptionOpenAPI:
		return c.Paths[string(ast.ScopeCallbacks)]
	default:
		panic(fmt.Sprintf("unknown import option %s", c.Option))
	}
}

func (c ImportConfig) GetWebhooksPath() string {
	switch c.Option {
	case ImportOptionOpenAPI:
		return c.Paths[string(ast.ScopeWebhooks)]
	default:
		panic(fmt.Sprintf("unknown import option %s", c.Option))
	}
}

// GetResourcesPath returns the configured location for the public-export
// resources surface, or "" when unset. Implicit model-namespace exports are
// only rendered when this path is configured; explicit x-speakeasy-exports
// fall back to "resources" when it is not. Unlike the scope accessors this
// never panics: it is queried before the imports config is known to exist.
func (c ImportConfig) GetResourcesPath() string {
	return c.Paths["resources"]
}

func (c ImportConfig) GetImportPath(scope ast.Scope) string {
	switch c.Option {
	case ImportOptionOpenAPI:
		path, ok := c.Paths[string(scope)]
		if !ok {
			panic(fmt.Sprintf("%s import path is not configured in gen.yaml file", scope))
		}
		return path
	default:
		panic(fmt.Sprintf("unknown import option %s", c.Option))
	}
}
