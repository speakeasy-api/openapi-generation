package ast

// Collection of TypeDef.
type TypeDefs []*TypeDef

// Returns true if all TypeDefs have the given FieldDef that is equal for
// Terraform usage.
func (t TypeDefs) AllHaveTerraformEqualField(field *FieldDef) bool {
	for _, typeDef := range t {
		otherField := typeDef.Fields.GetField(field.Name)

		if otherField == nil {
			return false
		}

		if !field.Type.IsTerraformEqual(otherField.Type) {
			return false
		}
	}

	return true
}

// Clone creates a deep copy of the TypeDefs.
func (t TypeDefs) Clone() []*TypeDef {
	if t == nil {
		return nil
	}

	return t.clone(nil, nil)
}

// clone creates a deep copy of the TypeDefs, tracking visited FieldDefs and
// TypeDefs to avoid infinite recursion on circular references.
func (t TypeDefs) clone(fieldDefVisited map[*FieldDef]*FieldDef, typeDefVisited map[*TypeDef]*TypeDef) TypeDefs {
	if t == nil {
		return nil
	}

	if fieldDefVisited == nil {
		fieldDefVisited = make(map[*FieldDef]*FieldDef)
	}

	if typeDefVisited == nil {
		typeDefVisited = make(map[*TypeDef]*TypeDef)
	}

	allVisited := true

	for _, typeDef := range t {
		if _, ok := typeDefVisited[typeDef]; !ok {
			allVisited = false
			break
		}
	}

	if allVisited {
		return t
	}

	cloned := make(TypeDefs, len(t))

	for i, td := range t {
		cloned[i] = td.clone(fieldDefVisited, typeDefVisited)
	}

	return cloned
}

// deepClone creates fully independent copies using stack-based cycle detection.
// See TypeDef.DeepClone for details on why this exists.
func (t TypeDefs) deepClone(fieldDefStack map[*FieldDef]*FieldDef, typeDefStack map[*TypeDef]*TypeDef) TypeDefs {
	if t == nil {
		return nil
	}

	if fieldDefStack == nil {
		fieldDefStack = make(map[*FieldDef]*FieldDef)
	}

	if typeDefStack == nil {
		typeDefStack = make(map[*TypeDef]*TypeDef)
	}

	cloned := make(TypeDefs, len(t))

	for i, td := range t {
		cloned[i] = td.deepClone(fieldDefStack, typeDefStack)
	}

	return cloned
}

// Returns Fields that are common across TypeDefs for Terraform usage.
func (t TypeDefs) TerraformCommonFields() Fields {
	if len(t) == 0 {
		return nil
	}

	firstTypeDef := t[0]

	if firstTypeDef == nil {
		return nil
	}

	if len(t) == 1 {
		return firstTypeDef.Fields
	}

	// Check type compatibility across all AssociatedTypes.
	// If types are incompatible (e.g., string vs object), return empty fields.
	for i := 1; i < len(t); i++ {
		if t[i] == nil {
			continue
		}

		// Try to merge the types to check compatibility
		_, err := firstTypeDef.TerraformMergeDataType(t[i])
		if err != nil {
			// Types are incompatible, return no common fields
			return nil
		}
	}

	commonFields := make(Fields, 0)
	remainingTypeDefs := t[1:]

	for _, field := range firstTypeDef.Fields {
		// Only primitive types are hoisted.
		if !field.Type.IsTerraformPrimitiveType() {
			continue
		}

		if !remainingTypeDefs.AllHaveTerraformEqualField(field) {
			continue
		}

		commonFields = append(commonFields, field)
	}

	return commonFields
}
