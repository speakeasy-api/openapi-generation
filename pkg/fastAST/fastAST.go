package fastAST

import (
	"context"
	"fmt"
	"regexp"

	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	extensions "github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	internalOpenAPI "github.com/speakeasy-api/openapi-generation/v2/internal/openapi"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
)

type FastAST struct {
	openAPIDocument *openapi.OpenAPI `json:"-"`

	OpenAPIVersion string `json:"openAPIVersion"`

	DocInfo *FastDocInfo `json:"docInfo"`

	Operations []*FastOperation `json:"operations"`

	// MCP context across all operations
	MCP *MCP `json:"mcp,omitempty"`

	// Terraform context across all operations
	Terraform *Terraform `json:"terraform,omitempty"`

	// GroupTree organizes operations into a hierarchical structure
	GroupTree *GroupTree `json:"groupTree"`
}

type FastDocInfo struct {
	Title       string `json:"title"`
	Version     string `json:"version"`
	Description string `json:"description"`
	Summary     string `json:"summary"`
}

type FastOperation struct {
	OperationID string `json:"operationID"`
	Method      string `json:"method"`
	Description string `json:"description"`
	Summary     string `json:"summary"`
	DisplayName string `json:"displayName"`

	// The JSONPath of the operation, if any
	JSONPath string `json:"jsonPath"`

	// The group of the operation, if any
	// This is a computed value based on the owning SDK's group
	// which is set if the operation has a tag, or if the operation
	// is marked with x-speakeasy-group
	Group string `json:"group"`

	// Underlying OpenAPI tags array retrieved from the OpenAPI document
	Tags []string `json:"tags"`
}

// Describes MCP context across all operations.
type MCP struct {
	Prompts   []*MCPPrompt   `json:"prompts,omitempty"`
	Resources []*MCPResource `json:"resources,omitempty"`
	Tools     []*MCPTool     `json:"tools,omitempty"`
}

// Describes a single MCP prompt.
type MCPPrompt struct {
	Description string `json:"description"`
	Name        string `json:"name"`
}

// Describes a single MCP resource.
type MCPResource struct {
	Description string `json:"description"`
	Name        string `json:"name"`
}

// Describes a single MCP tool.
type MCPTool struct {
	// Tool description.
	Description string `json:"description"`

	// Set to true when tool generation is explicitly disabled.
	Disabled bool `json:"disabled"`

	// Tool name.
	Name string `json:"name"`

	// Associated API operation identifiers.
	OperationIDs []string `json:"operation_ids,omitempty"`

	// Scopes assigned to tool.
	Scopes []string `json:"scopes"`
}

// Describes Terraform context across all operations.
type Terraform struct {
	DataResources    []*TerraformDataResource    `json:"data_resources,omitempty"`
	ManagedResources []*TerraformManagedResource `json:"managed_resources,omitempty"`
}

// Describes a single Terraform data resource.
type TerraformDataResource struct {
	// Description of the data resource.
	Description string `json:"description"`

	// Name of the data resource, e.g. examplecloud_thing.
	Name string `json:"name"`

	// Associated API operation identifiers.
	OperationIDs []string `json:"operation_ids,omitempty"`
}

func newTerraformDataResourceFromEntityOperationV1Config(config extensions.EntityOperationV1Config, operation *openapi.Operation) *TerraformDataResource {
	return &TerraformDataResource{
		Description:  config.Entity + " (DataSource)",
		Name:         config.Entity,
		OperationIDs: []string{operation.GetOperationID()},
	}
}

func newTerraformManagedResourceFromEntityOperationV1Config(config extensions.EntityOperationV1Config, operation *openapi.Operation) *TerraformManagedResource {
	return &TerraformManagedResource{
		Description:  config.Entity + " (Resource)",
		Name:         config.Entity,
		OperationIDs: []string{operation.GetOperationID()},
	}
}

// Describes a single Terraform managed resource.
type TerraformManagedResource struct {
	// Description of the managed resource.
	Description string `json:"description"`

	// Name of the managed resource, e.g. examplecloud_thing.
	Name string `json:"name"`

	// Associated API operation identifiers.
	OperationIDs []string `json:"operation_ids,omitempty"`
}

const (
	OperationIdKey = "operationId"
)

func NewFastAST(ctx context.Context, openAPIDocument *openapi.OpenAPI) (*FastAST, error) {
	fastAST := &FastAST{
		openAPIDocument: openAPIDocument,
		GroupTree:       NewGroupTree(),
		Operations:      []*FastOperation{},
	}

	fastAST.build(ctx)

	return fastAST, nil
}

func (d *FastAST) build(ctx context.Context) {
	// We need to know the OpenAPI version in order to know which
	// JSONSchema to use to validate it
	d.OpenAPIVersion = d.openAPIDocument.OpenAPI

	d.DocInfo = &FastDocInfo{
		Title:       d.openAPIDocument.GetInfo().GetTitle(),
		Version:     d.openAPIDocument.GetInfo().GetVersion(),
		Description: d.openAPIDocument.GetInfo().GetDescription(),
		Summary:     d.openAPIDocument.GetInfo().GetSummary(),
	}

	// Build OpenAPI 3.2 tag hierarchy map for parent resolution
	tagPathMap := internalOpenAPI.BuildTagPathMap(d.openAPIDocument.Tags)

	for item := range openapi.Walk(ctx, d.openAPIDocument) {
		_ = item.Match(openapi.Matcher{
			Operation: func(operation *openapi.Operation) error {
				extensions := operation.GetExtensions()

				if _, ok := extensions.Get("x-speakeasy-ignore"); ok {
					return nil
				}

				// We don't include deprecated operations when generating usage snippets
				// so we should ignore them here, otherwise the snippet for deprecated
				// operations will be stuck in loading state indefinitely.
				if operation.GetDeprecated() {
					return nil
				}

				method, path := extractMethodAndPath(item.Location)

				if method == "" || path == "" {
					return nil
				}

				group := ""

				if groupExt, ok := extensions.Get("x-speakeasy-group"); ok {
					group = groupExt.Value
				} else if len(operation.GetTags()) > 0 {
					// Use the first tag as the group, resolving parent hierarchy if present
					tag := operation.GetTags()[0]
					if resolvedPath, ok := tagPathMap[tag]; ok {
						group = resolvedPath
					} else {
						group = tag
					}
				}

				jsonPath := operation.GetCore().GetJSONPath(d.openAPIDocument.GetRootNode())

				fastOp := &FastOperation{
					OperationID: operation.GetOperationID(),
					Description: operation.GetDescription(),
					Summary:     operation.GetSummary(),
					Tags:        operation.GetTags(),
					Method:      method,
					Group:       group,
					JSONPath:    jsonPath,
					DisplayName: d.buildDisplayName(operation, jsonPath),
				}

				d.Operations = append(d.Operations, fastOp)

				mcpTool := d.buildMCPTool(operation)
				// Initialize MCP if this is our first MCP item
				if d.MCP == nil {
					d.MCP = &MCP{
						Tools: []*MCPTool{},
					}
				}
				d.MCP.Tools = append(d.MCP.Tools, mcpTool)

				// Check for Terraform entity operations
				terraformResources := d.buildTerraformResources(operation)
				if terraformResources != nil && (len(terraformResources.TerraformDataResources) > 0 || len(terraformResources.TerraformManagedResources) > 0) {
					// Initialize Terraform if this is our first Terraform item
					if d.Terraform == nil {
						d.Terraform = &Terraform{
							DataResources:    []*TerraformDataResource{},
							ManagedResources: []*TerraformManagedResource{},
						}
					}

					for _, dataResource := range terraformResources.TerraformDataResources {
						// Check if a data resource with this entity already exists
						found := false
						for _, existingResource := range d.Terraform.DataResources {
							if existingResource.Name == dataResource.Entity {
								// Add operation ID to existing resource
								existingResource.OperationIDs = append(existingResource.OperationIDs, operation.GetOperationID())
								found = true
								break
							}
						}
						if !found {
							// Create new data resource
							d.Terraform.DataResources = append(d.Terraform.DataResources, newTerraformDataResourceFromEntityOperationV1Config(dataResource, operation))
						}
					}
					for _, managedResource := range terraformResources.TerraformManagedResources {
						// Check if a managed resource with this entity already exists
						found := false
						for _, existingResource := range d.Terraform.ManagedResources {
							if existingResource.Name == managedResource.Entity {
								// Add operation ID to existing resource
								existingResource.OperationIDs = append(existingResource.OperationIDs, operation.GetOperationID())
								found = true
								break
							}
						}
						if !found {
							// Create new managed resource
							d.Terraform.ManagedResources = append(d.Terraform.ManagedResources, newTerraformManagedResourceFromEntityOperationV1Config(managedResource, operation))
						}
					}
				}

				// Add operation to the group tree
				// If no group is specified, add to root
				if group == "" {
					group = ""
				}
				d.GroupTree.AddOperation(group, fastOp)

				return nil
			},
		})
	}
}

// buildMCPTool builds MCP information from the AST given a specific operation node
func (*FastAST) buildMCPTool(operation *openapi.Operation) *MCPTool {
	extension := extensions.New(types.Target{})
	caser := casing.New()
	normalizedName := operation.GetOperationID()
	if normalizedName != "" {
		normalizedName = caser.ToSnake(operation.GetOperationID())
	}

	// Sensible defaults based on operation
	mcpNode := &extensions.MCP{
		Name:        normalizedName,
		Description: operation.GetDescription(),
		Disabled:    false,
		Scopes:      []string{},
	}

	// Lookup values from MCP extension if it exists
	determinedValues, err := extension.HandleMCPExtension(operation)
	if err != nil {
		// Log error but continue with defaults - don't fail MCP tool creation
		fmt.Printf("Error parsing MCP extension, using defaults: %v\n", err)
		determinedValues = nil
	}

	if determinedValues != nil {
		if determinedValues.Name != "" {
			mcpNode.Name = determinedValues.Name
		}
		if determinedValues.Description != "" {
			mcpNode.Description = determinedValues.Description
		}
		if len(determinedValues.Scopes) > 0 {
			mcpNode.Scopes = determinedValues.Scopes
		}
		if determinedValues.Disabled {
			mcpNode.Disabled = determinedValues.Disabled
		}
	}

	mcpTool := &MCPTool{
		Name:         mcpNode.Name,
		Description:  mcpNode.Description,
		Disabled:     mcpNode.Disabled,
		OperationIDs: []string{operation.GetOperationID()},
		Scopes:       mcpNode.Scopes,
	}
	return mcpTool
}

// buildTerraformResources builds Terraform information from the AST given a specific operation node
func (d *FastAST) buildTerraformResources(operation *openapi.Operation) *extensions.EntityOperationV1 {
	extension := extensions.New(types.Target{})
	collection, err := extension.HandleEntityOperationExtension(operation)
	if err != nil {
		// Log the error but continue processing other operations
		fmt.Printf("Error handling x-speakeasy-entity-operation extension: %v\n", err)
	}

	return collection
}

func (d *FastAST) buildDisplayName(operation *openapi.Operation, jsonPath string) string {
	displayName := operation.GetOperationID()

	if nameOverride, ok := operation.GetExtensions().Get("x-speakeasy-name-override"); ok {
		displayName = nameOverride.Value
	}

	if displayName == "" {
		// humanized jsonpath.
		// go from $.paths["/pets"].get to /pets/get
		// regex to remove $.paths[' and '].
		re := regexp.MustCompile(`\$\.paths\['(.*)'\].(.*)`)
		matches := re.FindStringSubmatch(jsonPath)
		if len(matches) == 3 {
			displayName = fmt.Sprintf("%s/%s", matches[1], matches[2])
		}
	}

	if displayName == "" {
		displayName = jsonPath
	}

	return displayName
}

func extractMethodAndPath(locations openapi.Locations) (string, string) {
	if len(locations) == 0 {
		return "", ""
	}

	var method, path string

	for i := len(locations) - 1; i >= 0; i-- {
		switch getParentType(locations[i]) {
		case "Paths":
			path = pointer.Value(locations[i].ParentKey)
		case "PathItem":
			method = pointer.Value(locations[i].ParentKey)
		case "OpenAPI":
		default:
			// Matched something unexpected so not likely an operation in paths
			return "", ""
		}
	}

	return method, path
}

func getParentType(location openapi.LocationContext) string {
	parentType := ""
	_ = location.ParentMatchFunc(openapi.Matcher{
		Any: func(a any) error {
			switch a.(type) {
			case *openapi.Paths:
				parentType = "Paths"
			case *openapi.ReferencedPathItem:
				parentType = "PathItem"
			case *openapi.OpenAPI:
				parentType = "OpenAPI"
			default:
				parentType = "Unknown"
			}
			return nil
		},
	})
	return parentType
}
