package validation

// RulesetConfig defines configuration for a ruleset mapping
type RulesetConfig struct {
	NewName     string
	Description string
	Extends     []string
}

// RulesetMappings maps old ruleset names to new configurations
// This allows users to continue using old ruleset names while migrating to new names
var RulesetMappings = map[string]RulesetConfig{
	"vacuum": {
		NewName:     "openapi-standard",
		Description: "Standard OpenAPI validation rules (formerly vacuum ruleset)",
		Extends:     []string{"recommended"},
	},
	"owasp": {
		NewName:     "owasp",
		Description: "OWASP security rules",
		Extends:     []string{},
	},
	"speakeasy-recommended": {
		NewName:     "speakeasy-recommended",
		Description: "Recommended rules for Speakeasy SDK generation",
		Extends:     []string{"recommended"},
	},
	"speakeasy-generation": {
		NewName:     "speakeasy-generation",
		Description: "Critical rules for SDK generation",
		Extends:     []string{},
	},
	"speakeasy-unsupported-constructs": {
		NewName:     "speakeasy-unsupported-constructs",
		Description: "Generation rules plus checks for constructs the generator cannot represent",
		Extends:     []string{"speakeasy-generation"},
	},
	"speakeasy-openapi": {
		NewName:     "speakeasy-openapi",
		Description: "Broader OpenAPI validation",
		Extends:     []string{"recommended"},
	},
}

// TranslateRulesetName translates an old ruleset name to the new name
// If no translation exists, it returns the original name
func TranslateRulesetName(oldName string) string {
	if config, ok := RulesetMappings[oldName]; ok {
		return config.NewName
	}
	return oldName
}

// GetRulesetMetadata returns the ruleset metadata for a given ruleset name
func GetRulesetMetadata(name string) (RulesetConfig, bool) {
	config, ok := RulesetMappings[name]
	return config, ok
}
