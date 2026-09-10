package oauth2

import (
	"context"
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/register"
	oas "github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/yml"
	"gopkg.in/yaml.v3"
)

func NewPasswordFields(ctx context.Context, register *register.Register, schemeKey string, scheme *oas.SecurityScheme, flow *oas.OAuthFlow, example *yaml.Node) ast.Fields {
	var usernameExample []*ast.Example
	var passwordExample []*ast.Example
	var clientIDExample []*ast.Example
	var clientSecretExample []*ast.Example

	ex := ""
	if example != nil {
		_ = example.Decode(&ex)
	}

	fieldExamples := strings.Split(ex, ";")
	if len(fieldExamples) >= 2 {
		usernameExample = []*ast.Example{ast.NewExample("", "", yml.CreateStringNode(fieldExamples[0]))}
		passwordExample = []*ast.Example{ast.NewExample("", "", yml.CreateStringNode(fieldExamples[1]))}
	}
	if len(fieldExamples) >= 3 {
		clientIDExample = []*ast.Example{ast.NewExample("", "", yml.CreateStringNode(fieldExamples[2]))}
	}
	if len(fieldExamples) >= 4 {
		clientSecretExample = []*ast.Example{ast.NewExample("", "", yml.CreateStringNode(fieldExamples[3]))}
	}

	scopes := make([]string, 0, flow.GetScopes().Len())
	for scope := range flow.GetScopes().All() {
		scopes = append(scopes, scope)
	}

	description := scheme.GetDescription()

	if description == "" {
		description = "OAuth2 Resource Owner Password Flow"
	}

	fields := ast.Fields{
		{
			Name:         "Username",
			OriginalName: "username",
			Type: ast.NewType(&ast.TypeDef{
				Type: ast.DataTypeString,
				Comments: &ast.Comment{
					Description: description + " username",
				},
				Examples: usernameExample,
			}, nil),
			Annotations: ast.Annotations{
				&ast.SecurityAnnotation{FieldName: "username"},
				&ast.NeedsCasingAnnotation{},
			},
		},
		{
			Name:         "Password",
			OriginalName: "password",
			Type: ast.NewType(&ast.TypeDef{
				Type: ast.DataTypeString,
				Comments: &ast.Comment{
					Description: description + " password",
				},
				Examples: passwordExample,
			}, nil),
			Annotations: ast.Annotations{
				&ast.SecurityAnnotation{FieldName: "password"},
				&ast.NeedsCasingAnnotation{},
			},
		},
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
			Optional: true,
			Annotations: ast.Annotations{
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
			Optional: true,
			Annotations: ast.Annotations{
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
			Default: &ast.AnyValue{Value: flow.GetTokenURL()},
			Annotations: ast.Annotations{
				&ast.SecurityAnnotation{FieldName: "tokenURL"},
				&ast.NeedsCasingAnnotation{},
			},
		},
	}

	formName := fmt.Sprintf("%s_%s", schemeKey, "Credentials")
	formType := ast.NewType(&ast.TypeDef{
		Name:         formName,
		OriginalName: "Credentials",
		Type:         ast.DataTypeClass,
		Fields:       fields,
		Scope:        ast.ScopeShared,
		IsComponent:  true,
		Comments: &ast.Comment{
			Description: description + " credentials",
		},
	}, ast.ContextStack{})
	formType = register.RegisterType(ctx, formType, true)

	tokenName := fmt.Sprintf("%s_%s", schemeKey, "Token")
	tokenType := ast.NewType(&ast.TypeDef{
		Name:         tokenName,
		OriginalName: "Token",
		Type:         ast.DataTypeString,
		Comments: &ast.Comment{
			Description: description + " token",
		},
	}, nil)

	unionName := fmt.Sprintf("%s_%s", schemeKey, "Input")
	union := ast.NewType(&ast.TypeDef{
		Name:            unionName,
		Type:            ast.DataTypeUnion,
		AssociatedTypes: ast.TypeDefs{formType, tokenType},
		Scope:           ast.ScopeShared,
		IsComponent:     true,
		Comments: &ast.Comment{
			Description: description,
		},
	}, ast.ContextStack{})
	union = register.RegisterType(ctx, union, true)

	field := &ast.FieldDef{
		Name: unionName,
		Type: union,
		Annotations: ast.Annotations{
			&ast.SecurityAnnotation{FieldName: "Authorization"},
		},
	}

	return ast.Fields{field}
}
