package ast

import "strings"

// FindTerraformEquivalentField finds the equivalent field in the given fields
// by matching using either the x-speakeasy-match path configuration or
// sanitized field name comparison.
//
// When useMatchConfig is true and the receiver field has an x-speakeasy-match
// path configured, it resolves the dot-separated path through the target
// fields' nested types. Otherwise, it falls back to matching by sanitized
// field name.
//
// Returns the matched field and the resolved path segments, or nil if no match
// is found.
func (f *FieldDef) FindTerraformEquivalentField(fields Fields, useMatchConfig bool) (*FieldDef, []string) {
	if f == nil || len(fields) == 0 {
		return nil, nil
	}

	// Check for x-speakeasy-match path configuration that indicates where
	// this field's data is sourced from, allowing cross-name equivalence.
	if useMatchConfig && f.Type != nil && f.Type.Extensions != nil && f.Type.Extensions.MatchConfig != nil && f.Type.Extensions.MatchConfig.Path != nil {
		matchedField, accessor := findTerraformMatchPath(fields, *f.Type.Extensions.MatchConfig.Path)
		if matchedField != nil {
			return matchedField, accessor
		}
	}

	// Fall back to sanitized name matching.
	target := SanitizeFieldName(f.Name)

	for _, field := range fields {
		if SanitizeFieldName(field.Name) == target {
			return field, []string{SanitizeFieldName(field.Name)}
		}
	}

	return nil, nil
}

// findTerraformMatchPath walks a dot-separated path through the given fields,
// using sanitized name comparison at each step to resolve through nested types.
func findTerraformMatchPath(fields Fields, matchPath string) (*FieldDef, []string) {
	steps := strings.Split(matchPath, ".")
	accessor := make([]string, 0, len(steps))

	var result *FieldDef

	currentFields := fields

	for _, step := range steps {
		sanitizedStep := SanitizeFieldName(step)

		var matchField *FieldDef

		for _, field := range currentFields {
			if SanitizeFieldName(field.Name) == sanitizedStep {
				matchField = field

				break
			}
		}

		if matchField == nil {
			return nil, nil
		}

		accessor = append(accessor, SanitizeFieldName(matchField.Name))
		result = matchField

		if matchField.Type != nil {
			currentFields = matchField.Type.Fields
		} else {
			currentFields = nil
		}
	}

	if result == nil {
		return nil, nil
	}

	return result, accessor
}

// IsTerraformImportRequired returns true if the field is required for
// Terraform import state operations. A field is required when its param
// annotation has RequiredForOperation set, or when the field is neither
// Optional nor Nullable.
func (f *FieldDef) IsTerraformImportRequired() bool {
	if f == nil {
		return false
	}

	if paramAnnotation := f.Annotations.GetParam(); paramAnnotation != nil && paramAnnotation.RequiredForOperation {
		return true
	}

	return !f.Optional && !f.Nullable
}

// HasMatchConfigPath returns true if the field has a path-only match config
// alias, indicating it references data sourced from another field path.
func (f *FieldDef) HasMatchConfigPath() bool {
	return f.Type != nil &&
		f.Type.Extensions != nil &&
		f.Type.Extensions.MatchConfig != nil &&
		f.Type.Extensions.MatchConfig.Path != nil
}
