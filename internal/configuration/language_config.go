package configuration

import (
	config "github.com/speakeasy-api/sdk-gen-config"
)

// Returns true if the given language configuration has the
// "maintainTagBasedOrdering" setting enabled.
func LanguageConfigHasMaintainTagBasedOrdering(langConfig config.LanguageConfig) bool {
	valueRaw, ok := langConfig.Cfg["maintainTagBasedOrdering"]

	if !ok {
		return false
	}

	value, ok := valueRaw.(bool)

	if !ok {
		return false
	}

	return value
}
