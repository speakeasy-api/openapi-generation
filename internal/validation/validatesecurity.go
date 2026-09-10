package validation

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/openapi/core"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

type ValidateSecurity struct {
	target types.Target
}

type schemeOverrideInfo struct {
	originalName string
	namespace    string
}

var _ Rule = (*ValidateSecurity)(nil)

func (r *ValidateSecurity) SetTarget(target types.Target) {
	r.target = target
}

func (r *ValidateSecurity) ID() string {
	return "generator-validate-security"
}

func (r *ValidateSecurity) Category() string {
	return "validation"
}

func (r *ValidateSecurity) Summary() string {
	return "Validate security schemes and security references are supported."
}

func (r *ValidateSecurity) HowToFix() string {
	return "Define referenced security schemes in components, and configure supported scheme types, flows, and required fields."
}

func (r *ValidateSecurity) Description() string {
	return "Validate security schemes have supported types, flows, and required fields. Also ensure every referenced scheme is defined in the document."
}

func (r *ValidateSecurity) Link() string {
	return ""
}

func (r *ValidateSecurity) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidateSecurity) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidateSecurity) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil {
		return nil
	}

	var validationErrors []error

	doc := docInfo.Document
	components := doc.GetComponents()
	if components == nil {
		return nil
	}

	// Collect all security schemes for reference validation
	securitySchemes := components.GetSecuritySchemes()
	schemeMap := make(map[string]*openapi.ReferencedSecurityScheme)
	overrideToOriginal := map[string][]schemeOverrideInfo{}

	if securitySchemes != nil {
		// Initialize extensions for name override resolution
		exts := extensions.New(r.target)
		_ = exts.HandleRewriteExtension(extensions.WithDocumentExtensions(doc.GetExtensions()))

		// Build the set of all scheme keys for conflict detection
		schemeKeySet := map[string]struct{}{}
		for schemeKey := range securitySchemes.Keys() {
			schemeKeySet[schemeKey] = struct{}{}
		}

		for schemeName, schemeRef := range securitySchemes.All() {
			schemeMap[schemeName] = schemeRef

			// Validate the security scheme itself
			scheme := schemeRef.GetObject()
			if scheme == nil {
				continue
			}

			schemeNode := schemeRef.GetRootNode()

			// Check for name override collisions
			if scheme.GetExtensions().Len() > 0 {
				override, err := exts.GetResolvedSchemaName(scheme.GetExtensions(), schemeName)
				if err == nil && override != "" && override != schemeName {
					modelNamespace, _ := exts.GetModelNamespace(scheme.GetExtensions())

					// Check for collision with existing overrides in the same namespace
					collision := false
					if existingInfos, exists := overrideToOriginal[override]; exists {
						for _, info := range existingInfos {
							if info.namespace == modelNamespace && info.originalName != schemeName {
								validationErrors = append(validationErrors, &validation.Error{
									Rule:            r.ID(),
									Severity:        r.DefaultSeverity(),
									Node:            schemeNode,
									UnderlyingError: fmt.Errorf("x-speakeasy-name-override collision: %q is already used by security scheme %q", override, info.originalName),
								})
								collision = true
								break
							}
						}
					}

					if !collision {
						if _, exists := schemeKeySet[override]; exists {
							// Override conflicts with an existing security scheme key
							validationErrors = append(validationErrors, &validation.Error{
								Rule:            r.ID(),
								Severity:        r.DefaultSeverity(),
								Node:            schemeNode,
								UnderlyingError: fmt.Errorf("x-speakeasy-name-override collision: %q conflicts with existing security scheme key", override),
							})
						} else {
							overrideToOriginal[override] = append(overrideToOriginal[override], schemeOverrideInfo{
								originalName: schemeName,
								namespace:    modelNamespace,
							})
						}
					}
				}
			}

			schemeCore := scheme.GetCore()

			// Skip if type is not present - built-in validation handles this
			if !schemeCore.Type.Present || schemeCore.Type.Value == "" {
				continue
			}

			schemeType := schemeCore.Type.Value

			switch schemeType {
			case "http":
				if err := r.validateHTTPAuth(schemeName, scheme, schemeNode); err != nil {
					validationErrors = append(validationErrors, err)
				}
			case "apiKey":
				if err := r.validateAPIKeyAuthorizationHeaderHint(schemeName, scheme, schemeNode); err != nil {
					validationErrors = append(validationErrors, err)
				}
			case "oauth2":
				validationErrors = append(validationErrors, r.validateOAuth2Auth(schemeName, scheme, schemeNode)...)
				// openIdConnect: built-in validation handles all required checks
			}
		}
	}

	// Validate global security references
	globalSecurity := doc.GetSecurity()
	if len(globalSecurity) > 0 {
		globalSecNode := doc.GetCore().Security.ValueNode
		if globalSecNode == nil {
			globalSecNode = doc.GetRootNode()
		}
		validationErrors = append(validationErrors, r.validateSecurityRefs("$", globalSecurity, schemeMap, globalSecNode, overrideToOriginal)...)
	}

	// Validate operation-level security references
	for _, opIndexNode := range docInfo.Index.Operations {
		operation := opIndexNode.Node
		if operation == nil {
			continue
		}

		opSecurity := operation.GetSecurity()
		if len(opSecurity) > 0 {
			operationID := operation.GetOperationID()
			pathStr := "$.paths"
			if operationID != "" {
				pathStr = fmt.Sprintf("$.paths[operationId=%s]", operationID)
			}

			securityNode := operation.GetCore().Security.ValueNode
			if securityNode == nil {
				securityNode = operation.GetRootNode()
			}
			validationErrors = append(validationErrors, r.validateSecurityRefs(pathStr, opSecurity, schemeMap, securityNode, overrideToOriginal)...)
		}
	}

	return validationErrors
}

func (r *ValidateSecurity) validateHTTPAuth(schemeName string, scheme *openapi.SecurityScheme, schemeNode *yaml.Node) error {
	schemeCore := scheme.GetCore()

	// Built-in validation handles missing scheme field, we just validate the value
	if !schemeCore.Scheme.Present || schemeCore.Scheme.Value == nil || *schemeCore.Scheme.Value == "" {
		return nil
	}

	schemeValue := *schemeCore.Scheme.Value
	acceptedSchemes := []string{"basic", "bearer", "custom"}

	// Case-insensitive validation (our added value over built-in validation)
	if !slices.ContainsFunc(acceptedSchemes, func(s string) bool {
		return strings.EqualFold(s, schemeValue)
	}) {
		schemeValueNode := schemeCore.Scheme.ValueNode
		if schemeValueNode == nil {
			schemeValueNode = schemeNode
		}
		return &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            schemeValueNode,
			UnderlyingError: fmt.Errorf("http security scheme `%s` has invalid scheme `%s` requires `basic`, `bearer`, or `custom`", schemeName, schemeValue),
		}
	}

	return nil
}

func (r *ValidateSecurity) validateOAuth2Auth(schemeName string, scheme *openapi.SecurityScheme, schemeNode *yaml.Node) []error {
	var validationErrors []error
	schemeCore := scheme.GetCore()

	// Built-in validation handles missing flows, we validate flow content
	if !schemeCore.Flows.Present || schemeCore.Flows.Value == nil {
		return validationErrors
	}

	flows := schemeCore.Flows.Value
	hasFlows := false
	flowsNode := schemeCore.Flows.ValueNode
	if flowsNode == nil {
		flowsNode = schemeNode
	}

	// Validate implicit flow
	if flows.Implicit.Present && flows.Implicit.Value != nil {
		hasFlows = true
		implicit := flows.Implicit.Value
		implicitNode := flows.Implicit.ValueNode
		if implicitNode == nil {
			implicitNode = flowsNode
		}

		// Validate Speakeasy-specific extensions only (built-in handles URL validation)
		if err := r.validateOverridableScopesExtension(schemeName, implicit, implicitNode, "implicit"); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	// Validate password flow
	if flows.Password.Present && flows.Password.Value != nil {
		hasFlows = true
		password := flows.Password.Value
		passwordNode := flows.Password.ValueNode
		if passwordNode == nil {
			passwordNode = flowsNode
		}

		// Validate Speakeasy-specific extensions only (built-in handles URL validation)
		if err := r.validateOverridableScopesExtension(schemeName, password, passwordNode, "password"); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	// Validate clientCredentials flow
	if flows.ClientCredentials.Present && flows.ClientCredentials.Value != nil {
		hasFlows = true
		// No additional validation needed beyond built-in (clientCredentials doesn't use overridable-scopes extension)
	}

	// Validate authorizationCode flow
	if flows.AuthorizationCode.Present && flows.AuthorizationCode.Value != nil {
		hasFlows = true
		authCode := flows.AuthorizationCode.Value
		authCodeNode := flows.AuthorizationCode.ValueNode
		if authCodeNode == nil {
			authCodeNode = flowsNode
		}

		// Validate Speakeasy-specific extensions only (built-in handles URL validation)
		if err := r.validateOverridableScopesExtension(schemeName, authCode, authCodeNode, "authorizationCode"); err != nil {
			validationErrors = append(validationErrors, err)
		}
	}

	// Check that at least one flow is defined (Speakeasy-specific check)
	if !hasFlows {
		validationErrors = append(validationErrors, &validation.Error{
			Rule:            r.ID(),
			Severity:        r.DefaultSeverity(),
			Node:            flowsNode,
			UnderlyingError: fmt.Errorf("oauth2 security scheme `%s` is missing a flow requires `implicit`, `password`, `clientCredentials` or `authorizationCode`", schemeName),
		})
	}

	return validationErrors
}

func (r *ValidateSecurity) validateAPIKeyAuthorizationHeaderHint(schemeName string, scheme *openapi.SecurityScheme, schemeNode *yaml.Node) error {
	schemeCore := scheme.GetCore()
	if schemeCore == nil {
		return nil
	}

	name := scheme.GetName()
	if name == "" {
		return nil
	}

	if scheme.GetIn() != openapi.SecuritySchemeInHeader || !strings.EqualFold(name, "Authorization") {
		return nil
	}

	nameNode := schemeCore.Name.GetKeyNodeOrRoot(schemeNode)
	return &validation.Error{
		Rule:            r.ID(),
		Severity:        validation.SeverityHint,
		Node:            nameNode,
		UnderlyingError: fmt.Errorf("apiKey security scheme `%s` uses header `Authorization`; did you mean `type=http scheme=bearer`?", schemeName),
	}
}

func (r *ValidateSecurity) validateSecurityRefs(_ string, securityReqs []*openapi.SecurityRequirement, schemeMap map[string]*openapi.ReferencedSecurityScheme, securityNode *yaml.Node, overrideToOriginal map[string][]schemeOverrideInfo) []error {
	var validationErrors []error

	for i, secReq := range securityReqs {
		if secReq == nil || secReq.Len() == 0 {
			continue
		}

		// Iterate through each scheme reference in the requirement
		// SecurityRequirement embeds *sequencedmap.Map, so use All()
		for schemeKey := range secReq.All() {
			// Check if the scheme is defined in components (Speakeasy-specific warning)
			found := false
			if _, ok := schemeMap[schemeKey]; ok {
				found = true
			} else if infos, ok := overrideToOriginal[schemeKey]; ok {
				for _, info := range infos {
					if _, ok := schemeMap[info.originalName]; ok {
						found = true
						break
					}
				}
			}

			if !found {
				// Try to get the specific node if possible, otherwise use security node
				node := securityNode
				if securityNode != nil && securityNode.Kind == yaml.SequenceNode && i < len(securityNode.Content) {
					reqNode := securityNode.Content[i]
					// Try to get the key node for this specific scheme from the core
					if secReq.GetCore() != nil {
						keyNode := secReq.GetCore().GetMapKeyNodeOrRoot(schemeKey, reqNode)
						if keyNode != nil {
							node = keyNode
						}
					}
				}

				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        validation.SeverityWarning,
					Node:            node,
					UnderlyingError: fmt.Errorf("security scheme `%s` not found, assuming `type: apiKey`, `in: header`, `name: Authorization`", schemeKey),
				})
			}
		}
	}

	return validationErrors
}

func (r *ValidateSecurity) validateOverridableScopesExtension(schemeName string, flow *core.OAuthFlow, flowNode *yaml.Node, flowType string) error {
	if !flow.Extensions.IsInitialized() || flow.Extensions.Len() == 0 {
		return nil
	}

	// Validate Speakeasy-specific extension
	if _, ok := flow.Extensions.Get("x-speakeasy-overridable-scopes"); ok {
		if flowType != "clientCredentials" {
			return &validation.Error{
				Rule:            r.ID(),
				Severity:        validation.SeverityWarning,
				Node:            flowNode,
				UnderlyingError: fmt.Errorf("security scheme `%s` uses `x-speakeasy-overridable-scopes` extension which is not currently supported for OAuth2 `%s` flow", schemeName, flowType),
			}
		}
	}

	return nil
}
