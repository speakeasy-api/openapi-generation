package pre_apply_union_discriminators

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

type discriminatorInfo struct {
	propertyName string
	value        string
}

func noDiscriminator() discriminatorInfo {
	return discriminatorInfo{}
}

// Tracker tracks how types are used with discriminators across the AST
type Tracker struct {
	log                               logging.Logger
	shouldPreApplyUnionDiscriminators bool
	usedWithSameDiscriminatorAlways   map[*ast.TypeDef]discriminatorInfo
	usedWithoutDiscriminator          map[*ast.TypeDef]bool
	usedWithDifferentDiscriminators   map[*ast.TypeDef]bool
}

func NewTracker(logger logging.Logger, preApplyUnionDiscriminators bool) *Tracker {
	return &Tracker{
		log:                               logger,
		shouldPreApplyUnionDiscriminators: preApplyUnionDiscriminators,
		usedWithSameDiscriminatorAlways:   make(map[*ast.TypeDef]discriminatorInfo),
		usedWithoutDiscriminator:          make(map[*ast.TypeDef]bool),
		usedWithDifferentDiscriminators:   make(map[*ast.TypeDef]bool),
	}
}

func (t *Tracker) recordUsage(typeDef *ast.TypeDef, viaDiscriminator discriminatorInfo) {
	if viaDiscriminator.value == "" {
		t.usedWithoutDiscriminator[typeDef] = true
		return
	}

	if t.usedWithDifferentDiscriminators[typeDef] {
		return
	}

	existing, ok := t.usedWithSameDiscriminatorAlways[typeDef]
	if !ok {
		t.usedWithSameDiscriminatorAlways[typeDef] = viaDiscriminator
		return
	}

	if viaDiscriminator.value != existing.value || viaDiscriminator.propertyName != existing.propertyName {
		t.usedWithDifferentDiscriminators[typeDef] = true
		t.log.Debug(fmt.Sprintf("type %q has inconsistent discriminator: %q=%q vs %q=%q",
			typeDef.Name, viaDiscriminator.propertyName, viaDiscriminator.value, existing.propertyName, existing.value))
	}
}

func (t *Tracker) WalkType(typeDef *ast.TypeDef) {
	visited := make(map[*ast.TypeDef]bool)
	t.walkType(typeDef, noDiscriminator(), visited)
}

func (t *Tracker) walkType(typeDef *ast.TypeDef, viaDiscriminator discriminatorInfo, visited map[*ast.TypeDef]bool) {
	if typeDef == nil || visited[typeDef] {
		return
	}
	visited[typeDef] = true

	t.recordUsage(typeDef, viaDiscriminator)

	if typeDef.Type == ast.DataTypeUnion && typeDef.Discriminator != nil {
		t.walkDiscriminatedUnionMembers(typeDef, visited)
		return
	}

	for _, child := range typeDef.Children {
		t.walkType(child, noDiscriminator(), visited)
	}
}

func (t *Tracker) walkDiscriminatedUnionMembers(typeDef *ast.TypeDef, visited map[*ast.TypeDef]bool) {
	propertyName := typeDef.Discriminator.TypePropertyName
	mapping := typeDef.Discriminator.Mapping

	for _, member := range typeDef.AssociatedTypes {
		info := lookupDiscriminatorValue(member, propertyName, mapping)

		if info.value == "" {
			t.log.Debug(fmt.Sprintf("type %q in union %q has no discriminator mapping", member.Name, typeDef.Name))
		}

		t.walkType(member, info, visited)
	}
}

func (t *Tracker) ApplyDiscriminators() {
	for member, info := range t.usedWithSameDiscriminatorAlways {
		if t.usedWithDifferentDiscriminators[member] {
			continue
		}

		if IsDiscriminatorAlreadyAppliedAsConst(member, info.propertyName, info.value) {
			member.DiscriminatorPreApplied = info.propertyName
			continue
		}

		if !t.shouldPreApplyUnionDiscriminators {
			continue
		}

		if t.usedWithoutDiscriminator[member] {
			continue
		}

		existingField := member.Fields.GetField(info.propertyName)
		if existingField != nil {
			t.log.Debug(fmt.Sprintf("updating field %q to const value %q on type %q", info.propertyName, info.value, member.Name))
			existingField.Const = &ast.AnyValue{Value: info.value}
			existingField.Optional = false
			existingField.Nullable = false
			existingField.Default = nil
			existingField.Type.Type = ast.DataTypeString
			member.DiscriminatorPreApplied = info.propertyName
		}
	}
}

func IsDiscriminatorAlreadyAppliedAsConst(member *ast.TypeDef, propertyName string, value string) bool {
	existingField := member.Fields.GetField(propertyName)
	if existingField != nil {
		isNotNullish := !existingField.Optional && !existingField.Nullable
		isString := existingField.Type.Type == ast.DataTypeString
		hasExactConstValue := existingField.Const != nil && existingField.Const.Value == value
		isDefaultNil := existingField.Default == nil
		return isNotNullish && isString && hasExactConstValue && isDefaultNil
	}
	return false
}

func lookupDiscriminatorValue(member *ast.TypeDef, propertyName string, mapping ast.DiscriminatorMappings) discriminatorInfo {
	var value string
	var count int
	for _, m := range mapping {
		if m.Type == member {
			value = m.Name
			count++
		}
	}

	// Multiple mappings or no mapping = no discriminator
	if count != 1 {
		return noDiscriminator()
	}

	return discriminatorInfo{propertyName, value}
}
