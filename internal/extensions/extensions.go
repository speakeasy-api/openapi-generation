package extensions

import (
	"regexp"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/extensions"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	"github.com/speakeasy-api/openapi/openapi"
	"gopkg.in/yaml.v3"
)

// modelNamespaceValidChars matches only valid namespace characters: alphanumeric, underscore, hyphen, and dot.
// Nested namespaces (containing /) are not currently supported.
var modelNamespaceValidChars = regexp.MustCompile(`^[a-zA-Z0-9_\-\.]+$`)

const (
	ErrUnmarshal = errors.Error("failed to unmarshal extension")
)

type Extension int

func (e Extension) Name() string {
	return extensionNames[e]
}

const (
	ExtUsageExample Extension = iota
	ExtRetries
	ExtServerID
	ExtNameOverride
	ExtInclude
	ExtIgnore
	ExtGlobals
	ExtGlobalsHidden
	ExtExample
	ExtGroup
	ExtEnums
	ExtEnumDescriptions
	ExtEnumFormat
	ExtDeprecationReplacement
	ExtDeprecationMessage
	ExtPagination
	ExtTypeOverride
	ExtErrors
	ExtErrorMessage
	ExtExtensionRewrite
	ExtDocs
	ExtTest
	ExtTestInternalDirectives
	ExtTestID
	ExtTestIgnore
	ExtTestServer
	ExtDocsRateLimits
	ExtMaxMethodParams
	ExtExampleUnset
	ExtUnknownValues
	ExtFlattenRequest
	ExtTimeout
	ExtSSESentinel
	ExtTransformFromAPI
	ExtTransformToAPI
	ExtCustomSecurityScheme
	ExtParamEncodingOverride
	ExtWebhooks
	ExtReactHook
	ExtMCP
	ExtTokenEndpointAuth
	ExtEntity
	ExtEntityDescription
	ExtEntityOperation
	ExtEntityOperations
	ExtEntityVersion
	ExtMatch
	ExtTerraformAliasTo
	ExtTerraformCustomDefault
	ExtTerraformIgnore
	ExtTerraformWriteOnly
	ExtResponseFilter
	ExtTokenEndpointAdditionalPropertiess
	ExtWrappedAttribute
	ExtSSEOverload
	ExtOverridableOAuth2Scopes
	ExtPolling
	ExtEntityMissingCodes
	ExtAllowEmptyValue
	ExtModelNamespace
	ExtDiscriminator
	ExtBase64InputMode
	ExtPublicExports
	ExtGoOptionalMethodArguments
	ExtCLICommands
	ExtCLIErrors
)

var extensionNames = map[Extension]string{
	ExtUsageExample:                       "x-speakeasy-usage-example",
	ExtRetries:                            "x-speakeasy-retries",
	ExtServerID:                           "x-speakeasy-server-id",
	ExtNameOverride:                       "x-speakeasy-name-override",
	ExtInclude:                            "x-speakeasy-include",
	ExtIgnore:                             "x-speakeasy-ignore",
	ExtGlobals:                            "x-speakeasy-globals",
	ExtGlobalsHidden:                      "x-speakeasy-globals-hidden",
	ExtExample:                            "x-speakeasy-example",
	ExtGroup:                              "x-speakeasy-group",
	ExtEnums:                              "x-speakeasy-enums",
	ExtEnumDescriptions:                   "x-speakeasy-enum-descriptions",
	ExtEnumFormat:                         "x-speakeasy-enum-format",
	ExtDeprecationReplacement:             "x-speakeasy-deprecation-replacement",
	ExtDeprecationMessage:                 "x-speakeasy-deprecation-message",
	ExtPagination:                         "x-speakeasy-pagination",
	ExtTypeOverride:                       "x-speakeasy-type-override",
	ExtErrors:                             "x-speakeasy-errors",
	ExtErrorMessage:                       "x-speakeasy-error-message",
	ExtExtensionRewrite:                   "x-speakeasy-extension-rewrite",
	ExtDocs:                               "x-speakeasy-docs",
	ExtTest:                               "x-speakeasy-test",
	ExtTestInternalDirectives:             "x-speakeasy-test-internal-directives",
	ExtTestID:                             "x-speakeasy-test-internal-id",
	ExtTestIgnore:                         "x-speakeasy-test-ignore",
	ExtTestServer:                         "x-speakeasy-test-server",
	ExtDocsRateLimits:                     "x-speakeasy-docs-rate-limits",
	ExtMaxMethodParams:                    "x-speakeasy-max-method-params",
	ExtExampleUnset:                       "x-speakeasy-example-unset",
	ExtUnknownValues:                      "x-speakeasy-unknown-values",
	ExtFlattenRequest:                     "x-speakeasy-flatten-request",
	ExtTimeout:                            "x-speakeasy-timeout",
	ExtSSESentinel:                        "x-speakeasy-sse-sentinel",
	ExtCustomSecurityScheme:               "x-speakeasy-custom-security-scheme",
	ExtParamEncodingOverride:              "x-speakeasy-param-encoding-override",
	ExtWebhooks:                           "x-speakeasy-webhooks",
	ExtTransformFromAPI:                   "x-speakeasy-transform-from-api",
	ExtTransformToAPI:                     "x-speakeasy-transform-to-api",
	ExtReactHook:                          "x-speakeasy-react-hook",
	ExtMCP:                                "x-speakeasy-mcp",
	ExtTokenEndpointAuth:                  "x-speakeasy-token-endpoint-authentication",
	ExtEntity:                             "x-speakeasy-entity",
	ExtEntityDescription:                  "x-speakeasy-entity-description",
	ExtEntityOperation:                    "x-speakeasy-entity-operation",
	ExtEntityOperations:                   "x-speakeasy-entity-operations",
	ExtEntityVersion:                      "x-speakeasy-entity-version",
	ExtMatch:                              "x-speakeasy-match",
	ExtPolling:                            "x-speakeasy-polling",
	ExtTerraformAliasTo:                   "x-speakeasy-terraform-alias-to",
	ExtTerraformCustomDefault:             "x-speakeasy-terraform-custom-default",
	ExtTerraformIgnore:                    "x-speakeasy-terraform-ignore",
	ExtTerraformWriteOnly:                 "x-speakeasy-terraform-write-only",
	ExtResponseFilter:                     "x-speakeasy-response-filter",
	ExtTokenEndpointAdditionalPropertiess: "x-speakeasy-token-endpoint-additional-properties",
	ExtWrappedAttribute:                   "x-speakeasy-wrapped-attribute",
	ExtSSEOverload:                        "x-speakeasy-sse-overload",
	ExtOverridableOAuth2Scopes:            "x-speakeasy-overridable-scopes",
	ExtEntityMissingCodes:                 "x-speakeasy-entity-missing-codes",
	ExtAllowEmptyValue:                    "x-speakeasy-allow-empty-value",
	ExtModelNamespace:                     "x-speakeasy-model-namespace",
	ExtDiscriminator:                      "x-speakeasy-discriminator",
	ExtBase64InputMode:                    "x-speakeasy-base64-input-mode",
	ExtPublicExports:                      "x-speakeasy-exports",
	ExtCLICommands:                        "x-speakeasy-cli-commands",
	ExtCLIErrors:                          "x-speakeasy-cli-errors",
	ExtGoOptionalMethodArguments:          "x-speakeasy-go-optional-method-arguments",
}

type OAExtensions = *extensions.Extensions

type Extensions struct {
	rewrites map[string]string
	target   types.Target
}

func New(target types.Target) *Extensions {
	return &Extensions{
		rewrites: map[string]string{},
		target:   target,
	}
}

type HandleRewriteExtensionOption func(*HandleRewriteExtensionOptions)

type HandleRewriteExtensionOptions struct {
	Extensions OAExtensions
	Node       *yaml.Node
}

func WithDocumentExtensions(extensions OAExtensions) HandleRewriteExtensionOption {
	return func(o *HandleRewriteExtensionOptions) {
		o.Extensions = extensions
	}
}

func WithRewritesExtensionNode(node *yaml.Node) HandleRewriteExtensionOption {
	return func(o *HandleRewriteExtensionOptions) {
		o.Node = node
	}
}

func (e *Extensions) GetResolvedName(ext Extension) string {
	name := ext.Name()

	if len(e.rewrites) == 0 {
		return name
	}

	if rewrite, ok := e.rewrites[name]; ok {
		return rewrite
	}

	return name
}

func (e *Extensions) HandleRewriteExtension(opts ...HandleRewriteExtensionOption) error {
	var options HandleRewriteExtensionOptions

	for _, opt := range opts {
		opt(&options)
	}

	if options.Extensions != nil && options.Extensions.Len() > 0 {
		rewrites, err := getExtensionValue[map[string]string](ExtExtensionRewrite.Name(), options.Extensions, nil)
		if err != nil {
			return err
		}

		if len(rewrites) > 0 {
			e.rewrites = rewrites
		}
	} else if options.Node != nil {
		var rewrites map[string]string
		if err := options.Node.Decode(&rewrites); err != nil {
			return errors.NewValidationError("failed to unmarshal "+ExtExtensionRewrite.Name(), options.Node, ErrUnmarshal.Wrap(err))
		}

		if len(rewrites) > 0 {
			e.rewrites = rewrites
		}
	}

	return nil
}

func (e *Extensions) IsUsageExample(extensions OAExtensions) (bool, error) {
	ext, err := getExtensionValue[any](e.GetResolvedName(ExtUsageExample), extensions, nil)
	if err != nil {
		return false, err
	}
	if ext == false {
		return false, nil
	}
	return ext != nil, nil
}

func (e *Extensions) Ignore(extensions OAExtensions) (bool, error) {
	result, err := e.HandleIgnoreExtension(extensions)

	if err != nil {
		return false, err
	}

	if result == nil {
		return false, nil
	}

	return *result, nil
}

func (e *Extensions) TypeOverride(extensions OAExtensions) (string, error) {
	return getExtensionValueWithValidation(e.GetResolvedName(ExtTypeOverride), extensions, "", func(t string) error {
		if t != "any" {
			return errors.Error("x-speakeasy-type-override only supports type any")
		}
		return nil
	})
}

func (e *Extensions) GetServerID(server *openapi.Server) (string, error) {
	if server.GetExtensions().Len() == 0 {
		return "", nil
	}

	return getExtensionValue(e.GetResolvedName(ExtServerID), server.GetExtensions(), "")
}

func (e *Extensions) IncludeSchema(schema *oas3.JSONSchema[oas3.Concrete]) (bool, error) {
	if schema.IsBool() {
		return false, nil
	}

	if schema.GetSchema().GetExtensions().Len() == 0 {
		return false, nil
	}

	return getExtensionValue(e.GetResolvedName(ExtInclude), schema.GetSchema().GetExtensions(), false)
}

func (e *Extensions) GetSecuritySchemeExample(scheme *openapi.SecurityScheme) *yaml.Node {
	if scheme.GetExtensions().Len() == 0 {
		return nil
	}

	extension, _ := e.findExtension(scheme.GetExtensions(), ExtExample)
	return extension
}

func (e *Extensions) GetGroups(operation *openapi.Operation) ([]string, error) {
	groups, _, err := e.GetGroupsWithNode(operation)
	return groups, err
}

// GetGroupsWithNode returns the groups for an operation along with the yaml node for error reporting
func (e *Extensions) GetGroupsWithNode(operation *openapi.Operation) ([]string, *yaml.Node, error) {
	if operation.GetExtensions().Len() == 0 {
		return nil, nil, nil
	}

	// Get the extension node
	groupExtNode, ok := e.findExtension(operation.GetExtensions(), ExtGroup)
	if !ok {
		return nil, nil, nil
	}

	group, err := getExtensionValue[*string](e.GetResolvedName(ExtGroup), operation.GetExtensions(), nil)
	if err == nil {
		if group == nil {
			return nil, nil, nil
		}

		return []string{*group}, groupExtNode, nil
	}
	if errors.Is(err, ErrUnmarshal) {
		groups, err := getExtensionValue[[]string](e.GetResolvedName(ExtGroup), operation.GetExtensions(), nil)
		return groups, groupExtNode, err
	}

	return nil, nil, err
}

// GetModelNamespace returns the namespace for a schema, used to organize component schemas into
// namespace folders. This enables multiple components with the same name to exist in different
// namespaces without conflict.
//
// The namespace value must contain only alphanumeric characters, underscores, hyphens, and dots.
// Nested namespaces (containing forward slashes) are not currently supported.
func (e *Extensions) GetModelNamespace(extensions OAExtensions) (string, error) {
	if extensions.Len() == 0 {
		return "", nil
	}

	namespace, err := getExtensionValueWithValidation(
		e.GetResolvedName(ExtModelNamespace),
		extensions,
		(*string)(nil),
		validateModelNamespace,
	)
	if err != nil {
		return "", err
	}

	if namespace == nil {
		return "", nil
	}

	return *namespace, nil
}

// validateModelNamespace ensures the namespace value contains only valid characters.
// Valid characters are: alphanumeric (a-z, A-Z, 0-9), underscore (_), hyphen (-), and dot (.).
// Nested namespaces (containing /) are not currently supported.
func validateModelNamespace(namespace *string) error {
	if namespace == nil {
		return nil
	}

	ns := *namespace

	// Check for empty namespace
	if ns == "" {
		return errors.Error("x-speakeasy-model-namespace cannot be empty")
	}

	// Check if namespace matches the valid pattern
	if !modelNamespaceValidChars.MatchString(ns) {
		return errors.Error(
			"x-speakeasy-model-namespace must contain only alphanumeric characters (a-z, A-Z, 0-9), " +
				"underscores (_), hyphens (-), and dots (.). " +
				"Nested namespaces using forward slashes (/) are not currently supported",
		)
	}

	// Check if namespace starts with a number (would cause compilation errors)
	if ns[0] >= '0' && ns[0] <= '9' {
		return errors.Error("x-speakeasy-model-namespace cannot start with a number")
	}

	return nil
}

func (e *Extensions) GetDeprecationReplacement(extensions OAExtensions) (string, error) {
	if extensions.Len() == 0 {
		return "", nil
	}

	return getExtensionValue(e.GetResolvedName(ExtDeprecationReplacement), extensions, "")
}

func (e *Extensions) GetDeprecationMessage(extensions OAExtensions) (string, error) {
	const defaultMessage string = "This will be removed in a future release, please migrate away from it as soon as possible"

	if extensions.Len() == 0 {
		return defaultMessage, nil
	}

	return getExtensionValue(e.GetResolvedName(ExtDeprecationMessage), extensions, defaultMessage)
}

func (e *Extensions) GetUsageConfig(extensions OAExtensions) (*UsageExampleConfig, error) {
	var usageExtension *yaml.Node
	for ext, node := range extensions.All() {
		if ext == e.GetResolvedName(ExtUsageExample) {
			var usageExample bool
			if err := node.Decode(&usageExample); err == nil && usageExample {
				return &UsageExampleConfig{}, nil
			}

			usageExtension = node
		}
	}

	var config UsageExampleConfig
	if err := usageExtension.Decode(&config); err != nil {
		return nil, errors.NewValidationError("failed to unmarshal "+ExtUsageExample.Name(), usageExtension, ErrUnmarshal.Wrap(err))
	}
	return &config, nil
}

func (e *Extensions) GetCustomDocs(extensions OAExtensions) (Comments, error) {
	if extensions.Len() == 0 {
		return Comments{}, nil
	}

	return getExtensionValue(e.GetResolvedName(ExtDocs), extensions, Comments{})
}

func (e *Extensions) GetResolvedSchemaName(extensions OAExtensions, originalName string) (string, error) {
	if extensions.Len() == 0 {
		return originalName, nil
	}

	return getExtensionValue(e.GetResolvedName(ExtNameOverride), extensions, originalName)
}

func (e *Extensions) GetPropertyName(s *oas3.JSONSchema[oas3.Referenceable], resolvedSchema *oas3.JSONSchema[oas3.Concrete], originalName string, ignoreResolvedNameOverride bool) (string, error) {
	// When `x-speakeasy-name-override` is defined as a sibling to a $ref, it must take
	// precedence over any name override that may have been defined at the component level
	if s.IsReference() && s.GetSchema().GetExtensions().Len() > 0 {
		newName, err := getExtensionValue(e.GetResolvedName(ExtNameOverride), s.GetSchema().GetExtensions(), originalName)
		if err != nil {
			return originalName, err
		}
		if newName != originalName {
			return newName, nil
		}
	}

	// When a property references a component via $ref, the component's x-speakeasy-name-override
	// is intended to rename the type/class, not the property field. Without this guard the
	// component-level override bleeds into the property name. See `NameOverrideFeb2026` flag.
	if ignoreResolvedNameOverride && s.IsReference() {
		return originalName, nil
	}

	return e.GetResolvedSchemaName(resolvedSchema.GetExtensions(), originalName)
}

func (e *Extensions) GetMaxMethodParams(extensions OAExtensions) (*int, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	return getExtensionValue[*int](e.GetResolvedName(ExtMaxMethodParams), extensions, nil)
}

func (e *Extensions) GetFlattenRequest(extensions OAExtensions) (*bool, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	return getExtensionValue[*bool](e.GetResolvedName(ExtFlattenRequest), extensions, nil)
}

func (e *Extensions) GetSSEOverload(extensions OAExtensions) (bool, error) {
	if extensions.Len() == 0 {
		return false, nil
	}

	return getExtensionValue(e.GetResolvedName(ExtSSEOverload), extensions, false)
}

func getExtensionValue[T any](ext string, extensions OAExtensions, defaultVal T) (T, error) {
	return getExtensionValueWithValidation(ext, extensions, defaultVal, func(T) error { return nil })
}

func getExtensionValueWithValidation[T any](ext string, extensions OAExtensions, defaultVal T, validateFunc func(T) error) (T, error) {
	var zero T

	if extensions.Len() == 0 {
		return zero, nil
	}

	var extension *yaml.Node

	for name, node := range extensions.All() {
		if name == ext {
			extension = node
		}
	}

	if extension == nil {
		return defaultVal, nil
	}

	var value T
	if err := extension.Decode(&value); err != nil {
		return zero, errors.NewValidationError("failed to unmarshal "+ext, extension, ErrUnmarshal.Wrap(err))
	}

	if err := validateFunc(value); err != nil {
		return zero, errors.NewValidationError("invalid value for "+ext, extension, err)
	}

	return value, nil
}

func (e *Extensions) IsExtensionMergable(extName string) bool {
	for orig, rename := range e.rewrites {
		if rename == extName {
			extName = orig
		}
	}

	blacklist := []string{
		ExtInclude.Name(),
		// Name override should not be merged because each schema should keep its own identity
		// When merging oneOf schemas, we don't want members to inherit the parent union's name
		// However, it IS identifying (affects type identity) - see IsExtensionIdentifying
		ExtNameOverride.Name(),
		// Model namespace should not be merged because each schema should keep its own namespace
		// However, it IS identifying (affects type identity) - see IsExtensionIdentifying
		ExtModelNamespace.Name(),
		// Public export declarations are source-site specific: a schema used
		// inside another's allOf/oneOf must not donate its exports to the
		// composed type.
		ExtPublicExports.Name(),
	}

	if slices.Contains(blacklist, extName) {
		return false
	}

	// Ignore any non-speakeasy extensions
	return strings.HasPrefix(extName, "x-speakeasy-")
}

// IsExtensionIdentifying returns true if the extension affects schema identity
// and should prevent a schema from being considered "empty" during allOf processing.
// This is different from IsExtensionMergable - some extensions like x-speakeasy-name-override
// should not be merged into child schemas during oneOf processing, but DO affect identity
// when building unique references for allOf schemas.
func (e *Extensions) IsExtensionIdentifying(extName string) bool {
	for orig, rename := range e.rewrites {
		if rename == extName {
			extName = orig
		}
	}

	blacklist := []string{
		ExtInclude.Name(),
		// Public exports alias whatever type the annotated site resolves to;
		// they must not make a nullable $ref wrapper look non-empty, which
		// would clone the referenced component instead of preserving the ref.
		ExtPublicExports.Name(),
	}

	if slices.Contains(blacklist, extName) {
		return false
	}

	// Ignore any non-speakeasy extensions
	return strings.HasPrefix(extName, "x-speakeasy-")
}

func (e *Extensions) findExtension(exts OAExtensions, ext Extension) (*yaml.Node, bool) {
	for name, node := range exts.All() {
		if name == e.GetResolvedName(ext) {
			return node, true
		}
	}

	return nil, false
}
