package generate

import (
	"context"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	oas "github.com/speakeasy-api/openapi/openapi"
)

func (g *Generator) handleGlobalServers(ctx context.Context, oapServers []*oas.Server, a *ast.AST) (*ast.Servers, error) {
	ss, err := g.handleServers(ctx, oapServers, ast.ScopeSDK)
	if err != nil {
		return nil, err
	}

	baseServerURL := g.subsystem.Config.Generation.BaseServerURL
	if baseServerURL == "" {
		baseServerURL = "/"
	}

	if ss == nil {
		ss = &ast.Servers{
			Servers: []*ast.Server{},
		}
	}
	if len(ss.Servers) == 0 {
		baseServerURL, isAbsolute, err := normalizeUrl(baseServerURL, "")
		if err != nil {
			return nil, err
		}
		ss.Servers = append(ss.Servers, &ast.Server{
			URL:        baseServerURL,
			IsRelative: !isAbsolute,
		})
	}

	if a != nil && a.MainSDK != nil {
		a.MainSDK.Servers = ss
	}

	if ss != nil {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureGlobalServerURLs)
	}

	return ss, nil
}

func (g *Generator) handleServers(ctx context.Context, oapServers []*oas.Server, scope ast.Scope) (*ast.Servers, error) {
	if len(oapServers) == 0 {
		return nil, nil
	}

	ss := &ast.Servers{
		Servers: []*ast.Server{},
	}

	isServerMap := false

	for _, server := range oapServers {
		id, err := g.subsystem.Extensions.GetServerID(server)
		if err != nil {
			return nil, err
		}

		// OpenAPI 3.2: Fall back to the native `name` field when x-speakeasy-server-id is absent.
		if id == "" {
			if name := server.GetName(); name != "" {
				id = name
			}
		}

		if id != "" {
			if ss.Default == "" {
				ss.Default = id
			}

			isServerMap = true
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureServerIDs)
		}

		baseServerURL, _, err := normalizeUrl(g.subsystem.Config.Generation.BaseServerURL, "")
		if err != nil {
			return nil, err
		}
		url, isAbsolute, err := normalizeUrl(server.URL, baseServerURL)
		if err != nil {
			// Include the line number of the server URL in the error message
			err = errors.NewUnsupportedError(err.Error(), server.GetCore().URL.GetKeyNodeOrRoot(server.GetRootNode()))
			return nil, err
		}

		var variables []*ast.ServerVariable

		for name, variable := range server.GetVariables().AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
			var typ *ast.TypeDef

			if len(variable.Enum) > 0 {
				typ = ast.NewType(&ast.TypeDef{
					Name:  "Server_" + name,
					Type:  ast.DataTypeEnum,
					Scope: scope,
					Enum: &ast.Enum{
						Type:         ast.NewType(&ast.TypeDef{Type: ast.DataTypeString}, nil),
						Values:       variable.Enum,
						Descriptions: make(map[string]string),
					},
				}, nil)
			} else {
				typ = ast.NewType(&ast.TypeDef{
					Type: ast.DataTypeString,
				}, nil)
			}

			var comment *ast.Comment

			if variable.GetDescription() != "" {
				comment = &ast.Comment{
					Description: variable.GetDescription(),
				}
			}

			typ.Comments = comment

			variables = append(variables, &ast.ServerVariable{
				Name:    name,
				Type:    typ,
				Default: variable.Default,
			})
		}

		var comment *ast.Comment

		if server.GetDescription() != "" {
			comment = &ast.Comment{
				Description: server.GetDescription(),
			}
		}

		ss.Servers = append(ss.Servers, &ast.Server{
			ID:         id,
			URL:        url,
			IsRelative: !isAbsolute,
			Comments:   comment,
			Variables:  variables,
		})
	}

	ss.ServerMap = isServerMap

	return ss, nil
}

// normalizeUrl normalizes the URL and returns a boolean indicating if it is absolute
func normalizeUrl(url string, baseUrl string) (string, bool, error) {
	if url == "" {
		return "", false, nil
	}

	if strings.HasPrefix(url, "ws:") || strings.HasPrefix(url, "wss:") {
		return "", false, errors.New("ws/wss server URLs are not currently supported")
	}

	// hasHttpScheme := strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
	hasHttpScheme := strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://")
	if hasHttpScheme {
		// Leave it alone.
		return url, true, nil
	}

	isRelative := strings.HasPrefix(url, "/")
	specifiesScheme := strings.Contains(url, "://")
	baseUrlPresent := baseUrl != ""

	switch {
	case isRelative && baseUrlPresent:
		// /api/v1  ->  http://example.com/api/v1
		url = baseUrl + url
		return url, true, nil
	case !isRelative && !specifiesScheme:
		// example.com/api/v1  ->  https://example.com/api/v1
		url = "https://" + url
		return url, true, nil
	}

	return url, !isRelative, nil
}
