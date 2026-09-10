package extensions

import (
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
)

// PublicExportRepresentation selects which rendering of the target type an
// export alias refers to in languages that generate more than one.
type PublicExportRepresentation string

const (
	// PublicExportRepresentationModel aliases the type's primary rendering
	// (e.g. the Pydantic model class in Python). This is the default.
	PublicExportRepresentationModel PublicExportRepresentation = "model"
	// PublicExportRepresentationInput aliases the type's request-input
	// rendering where the language generates a separate one — e.g. the
	// TypedDict companion in Python, used when callers pass plain dicts.
	// Languages without a separate input rendering, and targets without an
	// input companion (enums, errors), fall back to the model rendering.
	PublicExportRepresentationInput PublicExportRepresentation = "input"
)

type PublicExport struct {
	Group string `json:"group" yaml:"group"`
	Name  string `json:"name" yaml:"name"`
	// Representation selects the rendering the alias refers to: "model"
	// (default) or "input". See PublicExportRepresentation.
	Representation PublicExportRepresentation `json:"representation,omitempty" yaml:"representation,omitempty"`
}

// Input reports whether the export aliases the request-input rendering.
func (p PublicExport) Input() bool {
	return p.Representation == PublicExportRepresentationInput
}

func (p PublicExport) IsZero() bool {
	return strings.TrimSpace(p.Group) == "" &&
		strings.TrimSpace(p.Name) == "" &&
		p.Representation == ""
}

func (p PublicExport) Normalize() PublicExport {
	representation := PublicExportRepresentation(strings.TrimSpace(string(p.Representation)))
	if representation == PublicExportRepresentationModel {
		representation = ""
	}
	return PublicExport{
		Group:          strings.TrimSpace(p.Group),
		Name:           strings.TrimSpace(p.Name),
		Representation: representation,
	}
}

func (e *Extensions) HandlePublicExportsExtension(extensions OAExtensions) ([]PublicExport, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	node, ok := e.findExtension(extensions, ExtPublicExports)
	if !ok {
		return nil, nil
	}

	var exports []PublicExport
	if err := node.Decode(&exports); err != nil {
		var single PublicExport
		if singleErr := node.Decode(&single); singleErr != nil {
			return nil, errors.NewValidationError("failed to unmarshal "+ExtPublicExports.Name(), node, ErrUnmarshal.Wrap(err))
		}
		exports = []PublicExport{single}
	}

	normalized := make([]PublicExport, 0, len(exports))
	seen := map[string]bool{}
	for _, export := range exports {
		export = export.Normalize()
		if export.IsZero() {
			continue
		}
		if export.Group == "" {
			return nil, errors.NewValidationError(ExtPublicExports.Name()+".group is required", node, nil)
		}
		if export.Name == "" {
			return nil, errors.NewValidationError(ExtPublicExports.Name()+".name is required", node, nil)
		}
		if export.Representation != "" && export.Representation != PublicExportRepresentationInput {
			return nil, errors.NewValidationError(ExtPublicExports.Name()+`.representation must be "model" or "input"`, node, nil)
		}
		key := export.Key()
		if seen[key] {
			continue
		}
		seen[key] = true
		normalized = append(normalized, export)
	}

	return normalized, nil
}

func (p PublicExport) Key() string {
	return strings.Join([]string{
		p.Group,
		p.Name,
		string(p.Representation),
	}, "\x01")
}
