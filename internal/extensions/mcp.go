package extensions

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
)

var scopeSegmentRE = regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_-]*$`)

type MCP struct {
	Disabled        bool     `json:"disabled" yaml:"disabled"`
	Name            string   `json:"name" yaml:"name"`
	Scopes          []string `json:"scopes" yaml:"scopes"`
	Description     string   `json:"description" yaml:"description"`
	Title           string   `json:"title" yaml:"title"`
	DestructiveHint bool     `json:"destructiveHint" yaml:"destructiveHint"`
	IdempotentHint  bool     `json:"idempotentHint" yaml:"idempotentHint"`
	OpenWorldHint   bool     `json:"openWorldHint" yaml:"openWorldHint"`
	ReadOnlyHint    bool     `json:"readOnlyHint" yaml:"readOnlyHint"`
}

func (e *Extensions) HandleMCPExtension(operation *openapi.Operation) (*MCP, error) {
	if operation.GetExtensions().Len() == 0 {
		return nil, nil
	}

	extension, ok := e.findExtension(operation.GetExtensions(), ExtMCP)
	if !ok {
		return nil, nil
	}

	var mcp MCP
	if err := extension.Decode(&mcp); err != nil {
		return nil, errors.NewValidationError("failed to unmarshal "+ExtMCP.Name(), extension, err)
	}

	for _, scope := range mcp.Scopes {
		segments := strings.Split(scope, ".")
		for i := range segments {
			if !scopeSegmentRE.MatchString(segments[i]) {
				title := ExtMCP.Name() + " error"
				return nil, errors.NewValidationError(
					title,
					extension,
					fmt.Errorf("%s: scope segment '%s' contains invalid characters (must match %s)", scope, segments[i], scopeSegmentRE.String()),
				)
			}
		}
	}

	return &mcp, nil
}
