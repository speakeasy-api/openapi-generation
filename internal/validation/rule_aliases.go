package validation

// VacuumRuleAliases maps old vacuum rule names to new linter rule names
// This allows users to continue using old rule names in their lint.yaml configurations
var VacuumRuleAliases = map[string]string{
	// Semantic/Validation rules
	"typed-enum":                         "semantic-typed-enum",
	"path-params":                        "semantic-path-params",
	"path-declarations-must-exist":       "semantic-path-declarations",
	"path-not-include-query":             "semantic-path-query",
	"duplicated-entry-in-enum":           "semantic-duplicated-enum",
	"no-eval-in-markdown":                "semantic-no-eval-in-markdown",
	"no-script-tags-in-markdown":         "semantic-no-script-tags-in-markdown",
	"operation-operationId":              "semantic-operation-operation-id",
	"operation-operationId-valid-in-url": "semantic-operation-id-valid-in-url",
	"no-ambiguous-paths":                 "semantic-no-ambiguous-paths",
	"oas3-unused-component":              "semantic-unused-component",
	"oas2-unused-definition":             "semantic-unused-component",
	"oas3-missing-example":               "oas3-example-missing",

	// Style rules
	"contact-properties":          "style-contact-properties",
	"info-contact":                "style-info-contact",
	"info-description":            "style-info-description",
	"info-license":                "style-info-license",
	"license-url":                 "style-license-url",
	"openapi-tags":                "style-openapi-tags",
	"openapi-tags-alphabetical":   "style-tags-alphabetical",
	"operation-tags":              "style-operation-tags",
	"operation-description":       "style-operation-description",
	"component-description":       "style-component-description",
	"tag-description":             "style-tag-description",
	"no-$ref-siblings":            "style-no-ref-siblings",
	"oas3-no-$ref-siblings":       "style-no-ref-siblings",
	"oas3-host-not-example":       "style-oas3-host-not-example",
	"oas3-server-trailing-slash":  "style-oas3-host-trailing-slash",
	"oas3-parameter-description":  "style-oas3-parameter-description",
	"description-duplication":     "style-description-duplication",
	"oas3-api-servers":            "style-oas3-api-servers",
	"no-http-verbs-in-path":       "style-no-verbs-in-path",
	"paths-kebab-case":            "style-paths-kebab-case",
	"operation-4xx-response":      "style-operation-error-response",
	"operation-success-response":  "style-operation-success-response",
	"operation-singular-tag":      "style-operation-singular-tag",
	"operation-tag-defined":       "style-operation-tag-defined",
	"path-keys-no-trailing-slash": "style-path-trailing-slash",

	// OWASP rules (most are identical, these have slight differences)
	"owasp-no-additionalProperties":          "owasp-no-additional-properties",
	"owasp-constrained-additionalProperties": "owasp-additional-properties-constrained",

	// Style rules (alternate old names)
	"oas3-host-trailing-slash": "style-oas3-host-trailing-slash",

	// Custom generator rules (old names without prefix -> new generator- prefix)
	"duplicate-schema-name":      "generator-duplicate-schema-name",
	"duplicate-schemas":          "generator-duplicate-inline-schemas", // old shorthand used in docs
	"duplicate-operation-name":   "generator-duplicate-operation-name",
	"duplicate-errors":           "generator-duplicate-errors",
	"duplicate-inline-schemas":   "generator-duplicate-inline-schemas",
	"duplicate-path-params":      "generator-duplicate-path-params",
	"duplicate-properties":       "generator-duplicate-properties",
	"duplicate-tag":              "generator-duplicate-tag",
	"missing-error-response":     "generator-missing-error-response",
	"missing-examples":           "generator-missing-examples",
	"pagination":                 "generator-pagination",
	"retries":                    "generator-retries",
	"validate-composite-schemas": "generator-validate-composite-schemas",
	"validate-consts-defaults":   "generator-validate-consts-defaults",
	"validate-content-type":      "generator-validate-content-type",
	"validate-deprecation":       "generator-validate-deprecation",
	"validate-document":          "speakeasy-validate-document",
	"validate-enums":             "generator-validate-enums",
	"validate-extensions":        "generator-validate-extensions",
	"validate-not-schemas":       "generator-validate-not-schemas",
	"validate-parameters":        "generator-validate-parameters",
	"validate-paths":             "generator-validate-paths",
	"validate-requests":          "generator-validate-requests",
	"validate-responses":         "generator-validate-responses",
	"validate-security":          "generator-validate-security",
	"validate-servers":           "generator-validate-servers",
	"validate-types":             "generator-validate-types",
}

// TranslateRuleName translates an old vacuum rule name to the new linter rule name
// If no translation exists, it returns the original name
func TranslateRuleName(oldName string) string {
	if newName, ok := VacuumRuleAliases[oldName]; ok {
		return newName
	}
	return oldName // Return as-is if no translation needed
}

// ReverseTranslateRuleName translates a new linter rule name back to old vacuum name
// This is useful for displaying errors with legacy rule names for backward compatibility
func ReverseTranslateRuleName(newName string) string {
	for oldName, mappedNewName := range VacuumRuleAliases {
		if mappedNewName == newName {
			return oldName
		}
	}
	return newName // Return as-is if no reverse translation needed
}
