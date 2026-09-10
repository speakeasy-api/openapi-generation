package terraform

import (
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/casing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
)

// AttributeName converts a name to Terraform attribute format (snake_case,
// lowercase, no trailing underscore). This matches the TypeScript
// sanitizeTFStateName function behavior.
func AttributeName(name string) string {
	return strings.TrimSuffix(casing.New().ToSnake(name), "_")
}

// ProviderTypeName returns the resolved provider type name used as the prefix
// for all resource, data source, ephemeral resource, and action type names. It
// returns the providerTypeNameOverride generation configuration value if set,
// otherwise falls back to packageName. The returned value is normalized to
// lowercase kebab-case so it is safe to embed in generated code and Terraform
// configuration. This matches the TypeScript getProviderTypeName function
// behavior.
func ProviderTypeName(generationConfig map[string]any) string {
	name, _ := generationConfig["providerTypeNameOverride"].(string)

	if name == "" {
		name, _ = generationConfig["packageName"].(string)
	}

	if name == "" {
		return ""
	}

	return casing.New().ToKebab(sanitization.SanitizeName(name))
}

// TypeName returns the full Terraform type name for a resource, data source,
// ephemeral resource, or action by joining the provider type name and the
// snake_case form of the entity name with an underscore. This matches the
// TypeScript sanitizeResourceName function behavior.
func TypeName(providerTypeName, entityName string) string {
	return providerTypeName + "_" + casing.New().ToSnake(sanitization.SanitizeName(entityName))
}
