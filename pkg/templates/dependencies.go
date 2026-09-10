package templates

import (
	"sort"

	"github.com/mitchellh/mapstructure"
	"github.com/speakeasy-api/openapi-generation/v2/internal/executor"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

// TemplateDependency represents a structured dependency with CPE identifier for security scanning
type TemplateDependency struct {
	Name      string `mapstructure:"name"`
	Version   string `mapstructure:"version"`
	CPE       string `mapstructure:"cpe"`
	Ecosystem string `mapstructure:"ecosystem"`
	Category  string `mapstructure:"category"`
	Condition string `mapstructure:"condition,omitempty"`
}

// DependencyWithTemplate associates a dependency with its source template
type DependencyWithTemplate struct {
	TemplateName string
	Dependency   TemplateDependency
}

// GetTemplateDependencies returns the structured dependency list for a given target
func GetTemplateDependencies(target types.Target) ([]TemplateDependency, error) {
	e, err := executor.New(target, "config.ts")
	if err != nil {
		return nil, err
	}

	val, err := e.Run("getTemplateDependencies")
	if err != nil {
		return nil, err
	}

	// The JS function returns a Record<string, TemplateDependency> object
	var depsMap map[string]TemplateDependency
	if err := mapstructure.Decode(val, &depsMap); err != nil {
		return nil, err
	}

	// Convert map to slice and sort for deterministic output
	deps := make([]TemplateDependency, 0, len(depsMap))
	for _, dep := range depsMap {
		deps = append(deps, dep)
	}
	sort.Slice(deps, func(i, j int) bool {
		if deps[i].Name != deps[j].Name {
			return deps[i].Name < deps[j].Name
		}
		if deps[i].Version != deps[j].Version {
			return deps[i].Version < deps[j].Version
		}
		return deps[i].Condition < deps[j].Condition
	})

	return deps, nil
}

// GetAllTemplateDependencies returns the structured dependency list for all targets
func GetAllTemplateDependencies() (map[string][]TemplateDependency, error) {
	result := make(map[string][]TemplateDependency)

	for _, templateName := range GetAvailableTemplates() {
		target := types.NewTargetFromTemplate(templateName)
		deps, err := GetTemplateDependencies(target)
		if err != nil {
			// Skip targets that don't have getTemplateDependencies yet
			continue
		}
		result[templateName] = deps
	}

	return result, nil
}

// GetUniqueCPEs returns a deduplicated list of all CPE identifiers across all targets
func GetUniqueCPEs() ([]string, error) {
	allDeps, err := GetAllTemplateDependencies()
	if err != nil {
		return nil, err
	}

	cpeSet := make(map[string]struct{})
	for _, deps := range allDeps {
		for _, dep := range deps {
			if dep.CPE != "" {
				cpeSet[dep.CPE] = struct{}{}
			}
		}
	}

	cpes := make([]string, 0, len(cpeSet))
	for cpe := range cpeSet {
		cpes = append(cpes, cpe)
	}
	sort.Strings(cpes)

	return cpes, nil
}

// GetCPEToTemplatesMap returns a map of CPE to all templates that use it
// This properly handles cases where multiple templates share the same dependency
func GetCPEToTemplatesMap() (map[string][]DependencyWithTemplate, error) {
	allDeps, err := GetAllTemplateDependencies()
	if err != nil {
		return nil, err
	}

	result := make(map[string][]DependencyWithTemplate)
	for templateName, deps := range allDeps {
		for _, dep := range deps {
			if dep.CPE != "" {
				result[dep.CPE] = append(result[dep.CPE], DependencyWithTemplate{
					TemplateName: templateName,
					Dependency:   dep,
				})
			}
		}
	}

	return result, nil
}
