package oauth2

import (
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	oas "github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/yml"
	"gopkg.in/yaml.v3"
)

// ClientCredentialsAdditionalProperties represents the structure of an additional property
// within the x-speakeasy-token-endpoint-additional-properties extension.
type ClientCredentialsAdditionalProperties struct {
	Type    string
	Example string
}

func NewClientCredentialsFields(scheme *oas.SecurityScheme, flow *oas.OAuthFlow, example *yaml.Node, overridableScopes bool) ast.Fields {
	var clientIDExample []*ast.Example
	var clientSecretExample []*ast.Example

	if example != nil {
		var ex string
		_ = example.Decode(&ex)

		examples := strings.Split(ex, ";")

		if len(examples) == 2 {
			clientIDExample = []*ast.Example{ast.NewExample("", "", yml.CreateStringNode(examples[0]))}
			clientSecretExample = []*ast.Example{ast.NewExample("", "", yml.CreateStringNode(examples[1]))}
		}
	}

	description := scheme.GetDescription()

	if description == "" {
		description = "OAuth2 Client Credentials Flow"
	}

	fields := ast.Fields{
		{
			Name:         "ClientID",
			OriginalName: "clientID",
			Type: ast.NewType(&ast.TypeDef{
				Type: ast.DataTypeString,
				Comments: &ast.Comment{
					Description: description + " client identifier",
				},
				Examples: clientIDExample,
			}, nil),
			Annotations: []ast.Annotation{
				&ast.SecurityAnnotation{FieldName: "clientID"},
				&ast.NeedsCasingAnnotation{},
			},
		},
		{
			Name:         "ClientSecret",
			OriginalName: "clientSecret",
			Type: ast.NewType(&ast.TypeDef{
				Type: ast.DataTypeString,
				Comments: &ast.Comment{
					Description: description + " client secret",
				},
				Examples: clientSecretExample,
			}, nil),
			Annotations: []ast.Annotation{
				&ast.SecurityAnnotation{FieldName: "clientSecret"},
				&ast.NeedsCasingAnnotation{},
			},
		},
		{
			Name:         "TokenURL",
			OriginalName: "tokenURL",
			Type: ast.NewType(&ast.TypeDef{
				Type: ast.DataTypeString,
				Comments: &ast.Comment{
					Description: description + " token URL",
				},
			}, nil),
			Annotations: []ast.Annotation{
				&ast.NeedsCasingAnnotation{},
			},
			Default: &ast.AnyValue{
				Value: flow.GetTokenURL(),
			},
		},
	}

	// Only add the `Scopes` field if end users are expected to be able to override the list
	// of required scopes on any token request (x-speakeasy-overridable-scopes: true).
	if overridableScopes {
		fields = append(fields, &ast.FieldDef{
			Name:         "Scopes",
			OriginalName: "scopes",
			Type: ast.NewType(&ast.TypeDef{
				Type: ast.DataTypeArray,
				ItemType: ast.NewType(&ast.TypeDef{
					Type: ast.DataTypeString,
				}, nil),
				Comments: &ast.Comment{
					Description: description + " scopes override (optional)",
				},
			}, nil),
			Annotations: []ast.Annotation{
				&ast.NeedsCasingAnnotation{},
			},
			Optional: true,
		})
	}

	// Process the x-speakeasy-token-endpoint-additional-properties extension.
	ext, ok := flow.GetExtensions().Get(extensions.ExtTokenEndpointAdditionalPropertiess.Name())
	if ok && ext != nil {
		var additionalProperties map[string]ClientCredentialsAdditionalProperties

		if err := ext.Decode(&additionalProperties); err == nil {
			for name, prop := range additionalProperties {
				var propExample []*ast.Example
				if prop.Example != "" {
					propExample = []*ast.Example{ast.NewExample("", "", yml.CreateStringNode(prop.Example))}
				}

				newField := &ast.FieldDef{
					Name:         strcase.ToPascal(name),
					OriginalName: name,
					Type: ast.NewType(&ast.TypeDef{
						Type: ast.DataTypeString,
						Comments: &ast.Comment{
							Description: description + " " + name,
						},
						Examples: propExample,
					}, nil),
					Annotations: []ast.Annotation{
						&ast.SecurityAnnotation{FieldName: name},
						&ast.NeedsCasingAnnotation{},
					},
				}

				fields = append(fields, newField)
			}
		}
	}

	return fields
}
