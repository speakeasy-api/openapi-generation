package validation

import (
	"context"
	stderrors "errors"
	"fmt"
	"net/url"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"gopkg.in/yaml.v3"
)

type ValidateServers struct {
	BaseServerURL string
}

var _ Rule = (*ValidateServers)(nil)

func (r *ValidateServers) ID() string {
	return "generator-validate-servers"
}

func (r *ValidateServers) Category() string {
	return "validation"
}

func (r *ValidateServers) Summary() string {
	return "Validate servers, server variables, and server ID extensions."
}

func (r *ValidateServers) HowToFix() string {
	return "Define valid server URLs, ensure required server variables are consistent, and add unique x-speakeasy-server-id values when using that extension."
}

func (r *ValidateServers) Description() string {
	return "Validate servers, their variables, and the " + extensions.ExtServerID.Name() + " extension. This keeps server URLs and IDs consistent for generation."
}

func (r *ValidateServers) Link() string {
	return ""
}

func (r *ValidateServers) DefaultSeverity() validation.Severity {
	return validation.SeverityHint
}

func (r *ValidateServers) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateServers) SetConfig(cfg *config.Configuration) {
	if cfg == nil {
		return
	}

	r.BaseServerURL = cfg.Generation.BaseServerURL
}

func (r *ValidateServers) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil {
		return nil
	}

	var validationErrors []error

	doc := docInfo.Document

	// Track if servers are found at different levels
	globalServersFound := false
	serversFound := false

	// Check if base server URL is configured (counts as global server)
	if r.BaseServerURL != "" {
		globalServersFound = true
		serversFound = true
	}

	// Check global servers - only count explicitly defined servers, not the implicit default
	// OpenAPI 3.0 spec says if no servers are defined, there's an implicit default server with URL "/"
	// Use the Present field to check if servers were explicitly defined in the YAML
	hasExplicitServers := doc.GetCore().Servers.Present

	globalServers := doc.GetServers()
	if hasExplicitServers && len(globalServers) > 0 {
		globalServersFound = true
		serversFound = true
	}

	// Track server ID extensions and variables for global servers
	serverIDExts := map[string]bool{}
	var missingServerIDExts []*yaml.Node
	nonUniqueServerIDExts := map[string]*yaml.Node{}
	uniqueVariables := map[string]variableInfo{}

	// Validate global servers
	for _, server := range globalServers {
		serverNode := server.GetRootNode()

		// Validate server URL and variables
		if errs := r.validateServer(server, serverNode, uniqueVariables); len(errs) > 0 {
			validationErrors = append(validationErrors, errs...)
		}

		// Resolve server ID: prefer x-speakeasy-server-id, fall back to OpenAPI 3.2 name field
		serverID := ""
		if exts := server.GetExtensions(); exts.Len() > 0 {
			if serverIDExt, ok := exts.Get(extensions.ExtServerID.Name()); ok {
				if err := serverIDExt.Decode(&serverID); err != nil {
					validationErrors = append(validationErrors, &validation.Error{
						Rule:            r.ID(),
						Severity:        validation.SeverityError,
						Node:            serverNode,
						UnderlyingError: fmt.Errorf("failed to decode %s extension: %w", extensions.ExtServerID.Name(), err),
					})
				}
			}
		}
		if serverID == "" {
			if name := server.GetName(); name != "" {
				serverID = name
			}
		}

		if serverID != "" {
			if _, exists := serverIDExts[serverID]; exists {
				nonUniqueServerIDExts[serverID] = serverNode
			} else {
				serverIDExts[serverID] = true
			}
		} else {
			missingServerIDExts = append(missingServerIDExts, serverNode)
		}
	}

	// Report server ID issues for global servers
	if len(serverIDExts) > 0 && len(serverIDExts) != len(globalServers) {
		// Get the servers key node
		serversKeyNode := doc.GetCore().Servers.GetKeyNodeOrRoot(doc.GetRootNode())

		isOpenAPI32OrLater := doc.OpenAPI != "" && strings.Compare(doc.OpenAPI, "3.2") >= 0

		idSourceHint := fmt.Sprintf("`%s` extension", extensions.ExtServerID.Name())
		if isOpenAPI32OrLater {
			idSourceHint += " or `name` field"
		}

		validationErrors = append(validationErrors, &validation.Error{
			Rule:            r.ID(),
			Severity:        validation.SeverityError,
			Node:            serversKeyNode,
			UnderlyingError: fmt.Errorf("if using server IDs (via %s), all servers must have a unique ID", idSourceHint),
		})

		for name, node := range nonUniqueServerIDExts {
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        validation.SeverityError,
				Node:            node,
				UnderlyingError: fmt.Errorf("server ID `%s` is not unique", name),
			})
		}

		for _, node := range missingServerIDExts {
			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        validation.SeverityError,
				Node:            node,
				UnderlyingError: fmt.Errorf("server is missing an ID (add %s)", idSourceHint),
			})
		}
	}

	// Track which paths have servers
	pathServersMap := make(map[string]bool)

	// Check path-level servers using Index
	for _, pathItemNode := range docInfo.Index.InlinePathItems {
		pathStr := pathItemNode.Location.ParentKey()
		pathItem := pathItemNode.Node.GetObject()

		if pathItem == nil {
			continue
		}

		// Check path-level servers
		pathServers := pathItem.GetServers()
		if len(pathServers) > 0 {
			pathServersMap[pathStr] = true
			serversFound = true

			// Validate path-level servers
			for _, server := range pathServers {
				serverNode := server.GetRootNode()

				if errs := r.validateServer(server, serverNode, nil); len(errs) > 0 {
					validationErrors = append(validationErrors, errs...)
				}
			}
		}
	}

	// Check operation-level servers using Index
	for _, opNode := range docInfo.Index.Operations {
		operation := opNode.Node
		if operation == nil {
			continue
		}

		// Get path and method from index node location
		httpMethod, pathStr := openapi.ExtractMethodAndPath(opNode.Location)

		// Check operation-level servers
		opServers := operation.GetServers()
		opServersFound := len(opServers) > 0
		if opServersFound {
			serversFound = true

			// Validate operation-level servers
			for _, server := range opServers {
				serverNode := server.GetRootNode()

				if errs := r.validateServer(server, serverNode, nil); len(errs) > 0 {
					validationErrors = append(validationErrors, errs...)
				}
			}
		}

		// Check if operation has servers (either locally, path-level, or globally)
		pathServersFound := pathServersMap[pathStr]
		if !globalServersFound && !pathServersFound && !opServersFound {
			// Get the operation key node (the "get", "post", etc key) for better line number
			// PathItem is a map[httpMethod]*Operation, so we use GetMapKeyNodeOrRoot
			opKeyNode := operation.GetRootNode() // Default to root node

			// Get the PathItem from the location stack
			// Location stack: [0]=paths, [1]=pathKey, [2]=operationMethod
			const pathItemParentIdx = 2
			if len(opNode.Location) > pathItemParentIdx {
				pathItemLoc := opNode.Location[pathItemParentIdx]

				var parentPathItem *openapi.PathItem
				matcher := openapi.Matcher{
					ReferencedPathItem: func(pi *openapi.ReferencedPathItem) error {
						if pi != nil {
							parentPathItem = pi.GetObject()
						}
						return nil
					},
				}
				_ = pathItemLoc.ParentMatchFunc(matcher)

				if parentPathItem != nil {
					// PathItem embeds a map, use GetMapKeyNodeOrRoot to get the operation key node
					opKeyNode = parentPathItem.GetCore().GetMapKeyNodeOrRoot(httpMethod, operation.GetRootNode())
				}
			}

			validationErrors = append(validationErrors, &validation.Error{
				Rule:            r.ID(),
				Severity:        r.DefaultSeverity(),
				Node:            opKeyNode,
				UnderlyingError: fmt.Errorf("no servers found for operation `%s` `%s` either locally, on the path or globally", httpMethod, pathStr),
			})
		}
	}

	// If no servers found anywhere, report it
	if !serversFound {
		validationErrors = append(validationErrors, &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            doc.GetRootNode(),
			UnderlyingError: stderrors.New("no servers found in document, either add servers to the document or set a `baseServerUrl` in the `gen.yaml` config file"),
		})
	}

	return validationErrors
}

type variableInfo struct {
	Type string
	Node *yaml.Node
}

func (r *ValidateServers) validateServer(server *openapi.Server, serverNode *yaml.Node, uniqueVariables map[string]variableInfo) []error {
	var errors []error

	// Check URL exists
	serverURL := server.GetURL()
	if serverURL == "" {
		errors = append(errors, &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            serverNode,
			UnderlyingError: stderrors.New("server definition is missing a URL"),
		})
		return errors
	}

	// Note: Double curly braces validation is handled by built-in validation-invalid-syntax rule

	// Validate URL if it doesn't contain variables
	if !strings.Contains(serverURL, "{") || !strings.Contains(serverURL, "}") {
		// Add protocol if missing
		urlToValidate := serverURL
		if !strings.HasPrefix(urlToValidate, "http") {
			if !strings.HasPrefix(urlToValidate, "/") {
				urlToValidate = "https://" + urlToValidate
			}
		}

		if urlToValidate != "" && strings.HasPrefix(urlToValidate, "http") {
			parsed, err := url.Parse(urlToValidate)
			if err != nil {
				errors = append(errors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            serverNode,
					UnderlyingError: fmt.Errorf("server URL cannot be parsed: `%s`", err.Error()),
				})
			} else if parsed.Host == "" && parsed.Path == "" {
				errors = append(errors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            serverNode,
					UnderlyingError: stderrors.New("server URL is not valid: no hostname or path provided"),
				})
			}
		}
	}

	// Check variables
	variables := server.GetVariables()
	if variables != nil && variables.Len() > 0 {
		for varName, variable := range variables.All() {

			node := server.GetCore().Variables.GetMapKeyNodeOrRoot(varName, serverNode)

			// Get enum values
			enumValues := variable.GetEnum()
			varType := "string"
			if len(enumValues) > 0 {
				varType = "enum"

				// Note: Default value validation is handled by built-in validation-allowed-values rule
			}

			// Check for variable type conflicts across servers
			if uniqueVariables != nil {
				if existing, ok := uniqueVariables[varName]; ok {
					if existing.Type != varType {
						errors = append(errors, &validation.Error{
							Rule:            r.ID(),
							Severity:        r.DefaultSeverity(),
							Node:            node,
							UnderlyingError: fmt.Errorf("variable `%s` (type: `%s`) has different types in different servers, conflicts with variable (type: `%s`) at line `%d`", varName, varType, existing.Type, existing.Node.Line),
						})
					}
				} else {
					uniqueVariables[varName] = variableInfo{
						Type: varType,
						Node: node,
					}
				}
			}
		}
	}

	return errors
}
