package infer_union_discriminator

import (
	"errors"
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

// ProcessUnionType processes a single union type and infers its discriminator if possible
func ProcessUnionType(logger logging.Logger, typeDef *ast.TypeDef, schemaPath string) {
	if typeDef.Discriminator != nil {
		logDiscriminatorResult(logger, typeDef, nil, schemaPath)
		return
	}

	err := findBestDiscriminator(typeDef)
	logDiscriminatorResult(logger, typeDef, err, schemaPath)
}

// findBestDiscriminator finds the best discriminator property for a union
func findBestDiscriminator(typeDef *ast.TypeDef) error {
	// Must have at least two members
	if len(typeDef.AssociatedTypes) < 2 {
		return errors.New("union must have at least two members")
	}

	// All members must be objects (classes or errors)
	for _, member := range typeDef.AssociatedTypes {
		if !member.IsTypeWithFields() {
			return errors.New("all union members must be objects (classes or errors)")
		}
	}

	// Collect all property names from the first member
	firstMember := typeDef.AssociatedTypes[0]

	// For each property, check if it's a valid discriminator across all members
	var best *discriminator

	for _, f := range firstMember.Fields {
		propName := getOriginalFieldName(f)

		// Parse discriminator for this property
		parsed := parseDiscriminator(propName, typeDef.AssociatedTypes)
		if parsed == nil {
			continue
		}

		// Compare with current best using specificity
		best = better(best, parsed)
	}

	if best == nil {
		return errors.New("no valid discriminator property found")
	}

	// Build final discriminator from best parsed
	discriminator, err := buildDiscriminatorFromParsed(best)
	if err != nil {
		return err
	}

	typeDef.Discriminator = discriminator
	return nil
}

// buildDiscriminatorFromParsed creates an ast.Discriminator from a parsed discriminator
// Only supports CONST and SINGLE specificity with string type
func buildDiscriminatorFromParsed(d *discriminator) (*ast.Discriminator, error) {
	if d.Specificity > DiscriminatorSpecificityMulti {
		return nil, fmt.Errorf("unsupported specificity: %d", d.Specificity)
	}

	// Only support string type
	if d.TypeDefType != ast.DataTypeString {
		return nil, fmt.Errorf("unsupported type: %s (only string supported)", d.TypeDefType)
	}

	mappings := make(ast.DiscriminatorMappings, 0, len(d.Mapping))

	for _, mapping := range d.Mapping {
		if mapping == nil {
			continue
		}

		// Use the first value as the mapping name
		if len(mapping.Consts) == 0 {
			return nil, errors.New("mapping has no values")
		}

		for _, c := range mapping.Consts {
			mappings = append(mappings, &ast.DiscriminatorMapping{
				Name: fmt.Sprint(c.Value),
				Type: mapping.Type,
			})
		}
	}

	return &ast.Discriminator{
		TypePropertyName: d.PropertyName,
		Mapping:          mappings,
		Inferred:         true,
	}, nil
}

// parseDiscriminator determines if the fields can be used as a discriminator
// Returns a discriminator with property name, type, specificity and mappings, or nil if not valid
func parseDiscriminator(propertyName string, members []*ast.TypeDef) *discriminator {
	if len(members) < 2 {
		return nil
	}

	// Extract fields from members
	fields := make([]*ast.FieldDef, len(members))
	for i, member := range members {
		field := findFieldByName(member, propertyName)
		if field == nil {
			return nil
		}
		// Property must be required on all members
		if field.Optional {
			return nil
		}
		fields[i] = field
	}

	d := &discriminator{
		PropertyName: propertyName,
		Specificity:  DiscriminatorSpecificityConst,
		Mapping:      make([]*discriminatorMapping, len(members)),
	}

	// Check each pair of fields to ensure they're distinct
	for i := range len(fields) {
		for j := i + 1; j < len(fields); j++ {
			a, b := fields[i], fields[j]

			typeA := parsePrimitiveDataType(a)
			if typeA == "" {
				return nil
			}
			typeB := parsePrimitiveDataType(b)
			if typeB == "" {
				return nil
			}

			// Set the TypeDefType on first encounter
			if d.TypeDefType == "" {
				d.TypeDefType = typeA
			}

			// Different types means discrimination by type
			if typeA != typeB {
				d.TypeDefType = ast.DataTypeAny
				d.Specificity = DiscriminatorSpecificityType
				continue
			}

			// The discriminator has mixed types *and* shared type
			// eg A={kind:1} B={kind:"dog"} C={kind:"cat"}
			// it's unclear how to build a discriminator from this
			if d.TypeDefType == ast.DataTypeAny {
				return nil
			}

			// Same type - check for distinct const values
			constsA := getPossibleConsts(a)
			constsB := getPossibleConsts(b)

			// Same type but no const values found
			if constsA == nil || constsB == nil {
				return nil
			}

			// If they overlap, the discriminator is invalid
			if hasOverlap(constsA, constsB) {
				return nil
			}

			// Populate mappings
			d.Mapping[i] = &discriminatorMapping{constsA, members[i]}
			d.Mapping[j] = &discriminatorMapping{constsB, members[j]}

			// Update specificity based on enum types (only widen, never narrow)
			specificity := max(d.Specificity, getSpecificity(a), getSpecificity(b))
			d.Specificity = specificity
		}
	}

	return d
}
