package ast

import (
	"fmt"
	"slices"
	"strings"
)

type ArgumentsOptions struct {
	// ParamFlatteningBodyFirst affects whether param fields come before the
	// request body field or the other way around.
	ParamFlatteningParamsFirst bool

	// MaxMethodParams is the maximum number of parameters that a method can
	// have and still be eligible for parameter flattening.
	MaxMethodParams int

	// FlattenRequest indicates if request body fields should be flattened.
	FlattenRequest bool

	// In many languages, for convenience we don't render const fields in
	// models / arguments because we just go ahead and fill it in
	// In Typescript, due to structural typing this can cause issues in unions
	ConstFieldsAlwaysOptional *bool
}

// Arguments provides a list of fields that is used to generate arguments for
// SDK methods including informing on whether or not the fields were generated
// as a result of parameter and/or body flattening. The details provided can be
// used to rehydrate request models for use in method bodies.
type Arguments struct {
	// Flattening is the strategy used to flatten the request into a list of
	// SDK method arguments.
	// - "none": no flattening is applied. Sorted _may_ contain up to two fields
	//   representing the request and security inputs.
	//
	// - "all": parameters and request body fields have been exploded into the
	//   Sorted list. This implies that an operation contains both parameters
	//   _and_ a body. Security field may also be included depending on
	//   per-operation security requirements.
	//
	// - "params": parameters have been exploded into the Sorted and there may
	//   be a single field representing the request body if an operation has a
	//   body. Security field may also be included depending on per-operation
	//   security requirements.
	//
	// - "body": request body has been exploded into the Sorted list. This also
	//   means that the operation does not have parameters. Security field may
	//   also be included depending on per-operation security requirements.
	Flattening string `yaml:",omitempty"` // "none", "all", "body", "params"

	// Sorted is a list of the arguments to template in SDK methods. It includes
	// per-operation security, request parameters, and request body fields
	// depending on the flattening strategy. The ordering of fields is based
	// on a combination of criteria such as optionality and whether a field
	// represents additional properties.
	Sorted Fields `yaml:",omitempty"`

	// ParamFields is the subset of fields in Sorted that contains all request
	// parameters if the request is flattened. This will be an empty slice if
	// there are no parameters or the request is not flattened.
	ParamFields Fields `yaml:",omitempty"`

	// BodyFields is the subset of fields in Sorted that contains all request
	// body fields if the body is flattened. This will be an empty slice if the
	// body is not flattened.
	//
	// NOTE: This field is mutually exclusive with `BodyField`. Only one of them
	// is populated based on the flattening strategy.
	BodyFields Fields `yaml:",omitempty"`

	// BodyField is set when the request body is flattened. It is used to
	// determine how to reconstruct the request model.
	//
	// NOTE: This field is mutually exclusive with `BodyFields`. Only one of
	// them is populated based on the flattening strategy.
	BodyField *FieldDef `yaml:",omitempty"`

	Warning error `yaml:",omitempty"`
}

func (a *Arguments) Match(matchers Matchers) error {
	if matchers.Arguments != nil {
		return matchers.Arguments(a)
	}

	return nil
}

func (a Arguments) IsBodyField(f *FieldDef) bool {
	return a.BodyField == f
}

func (a Arguments) IsFieldInBody(f *FieldDef) bool {
	return slices.Contains(a.BodyFields, f)
}

func (a Arguments) IsFieldInParams(f *FieldDef) bool {
	return slices.Contains(a.ParamFields, f)
}

func NewArguments(op *Operation, security *FieldDef, options ArgumentsOptions) (*Arguments, error) {
	r := op.Request

	if r == nil {
		f := Fields{}
		if security != nil {
			f = append(f, security)
		}
		return &Arguments{Flattening: "none", Sorted: f}, nil
	}

	defaultArgs := &Arguments{
		Flattening: "none",
		Sorted:     injectSecurityArg(Fields{r.Field}, security),
		BodyField:  r.RequestBody,
	}

	hasClassBody := r.RequestBody != nil && r.RequestBody.Type.Type == DataTypeClass && len(r.RequestBody.Type.Fields) > 0
	canFlattenBody := options.FlattenRequest && hasClassBody
	hasVisibleParams := r.Params.HasVisibleParams()

	flattening := "none"

	switch {
	case hasVisibleParams && canFlattenBody && !r.IsRequestBody:
		flattening = "all"
	case canFlattenBody:
		flattening = "body"
	case hasVisibleParams:
		flattening = "params"
	}

	if flattening == "none" {
		return defaultArgs, nil
	}

	args := Fields{}
	paramFields := Fields{}
	bodyFields := Fields{}
	var bodyField *FieldDef
	switch flattening {
	case "all":
		paramFields = append(paramFields, unpackParamFields(r.Field.Type, options)...)
		bodyFields = unpackFieldDef(r.RequestBody, options)
		bodyField = r.RequestBody

		if options.ParamFlatteningParamsFirst {
			args = append(args, paramFields...)
			args = append(args, bodyFields...)
		} else {
			args = append(args, bodyFields...)
			args = append(args, paramFields...)
		}
	case "params":
		bodyField = r.RequestBody
		paramFields = append(paramFields, unpackParamFields(r.Field.Type, options)...)
		reqBody := Fields{}
		if r.RequestBody != nil {
			reqBody = Fields{r.RequestBody}
		}

		if options.ParamFlatteningParamsFirst {
			args = append(args, paramFields...)
			args = append(args, reqBody...)
		} else {
			args = append(args, reqBody...)
			args = append(args, paramFields...)
		}
	case "body":
		bodyFields = unpackFieldDef(r.RequestBody, options)
		args = append(args, bodyFields...)
	default:
		return nil, fmt.Errorf("assertion error: unexpected flattening strategy: %s", flattening)
	}

	// If we're only flattening params and not forcing full flattening then we
	// test against maxMethodParams.
	if flattening == "params" && !options.FlattenRequest && len(args) > options.MaxMethodParams {
		return defaultArgs, nil
	}

	var warning error
	renamedFields := resolveFieldNameConflicts(op, paramFields, bodyFields)
	if len(renamedFields) > 0 {
		warning = fmt.Errorf("operation '%s' has conflicting names in the parameters and requestBody: %v: use x-speakeasy-name-override to re-alias one of them", op.ID, strings.Join(quoteNames(renamedFields), ", "))
	}

	conflicts := checkForConflicts(paramFields, bodyFields)
	if len(conflicts) > 0 {
		defaultArgs.Warning = fmt.Errorf("operation '%s' with flattened parameters and request body contains conflicting field names: %v: use x-speakeasy-name-override on parameter or request fields to re-alias one of them", op.ID, strings.Join(quoteNames(conflicts), ", "))
		return defaultArgs, nil
	}

	return &Arguments{
		Flattening:  flattening,
		Sorted:      injectSecurityArg(sortArguments(args), security),
		ParamFields: paramFields,
		BodyFields:  bodyFields,
		BodyField:   bodyField,
		Warning:     warning,
	}, nil
}

func quoteNames(names []string) []string {
	quoted := make([]string, len(names))
	for i, name := range names {
		quoted[i] = fmt.Sprintf("%q", name)
	}

	return quoted
}

// Returns a list of original field names before they were renamed to resolve conflicts.
func resolveFieldNameConflicts(op *Operation, paramFields, bodyFields Fields) []string {
	conflicts := []string{}
	existingNames := make(map[string]bool)
	allFields := append(bodyFields, paramFields...)

	opSecurityName := ""
	if op.Security != nil {
		opSecurityName = SanitizeFieldName(op.Security.Name)
		existingNames[opSecurityName] = true
	}

	for i, field := range allFields {
		name := SanitizeFieldName(field.Name)
		isParam := i >= len(bodyFields)

		// Found a conflict
		if existingNames[name] {
			// Keep track of renamed fields for the warning
			if !slices.Contains(conflicts, field.OriginalName) {
				conflicts = append(conflicts, field.OriginalName)
			}

			// Only params can be renamed
			if isParam || name == opSecurityName {
				field.Name = renameFieldNameDueToConflict(field, existingNames)
				existingNames[field.Name] = true
			}
		}

		existingNames[name] = true
	}

	return conflicts
}

func renameFieldNameDueToConflict(field *FieldDef, existingNames map[string]bool) string {
	for i := 0; i < 6; i++ {
		suffix := "_param"
		if i > 0 {
			suffix = fmt.Sprintf("_%d", i)
		}

		// If the field name doesn't already end with the suffix and the new
		// name doesn't already exist then we can use it.
		if !strings.HasSuffix(field.Name, suffix) && !existingNames[field.Name+suffix] {
			return field.Name + suffix
		}
	}

	return field.Name
}

func checkForConflicts(paramFields Fields, bodyFields Fields) []string {
	if len(paramFields) == 0 || len(bodyFields) == 0 {
		return nil
	}

	names := make(map[string]struct{}, len(paramFields))
	for _, field := range paramFields {
		names[SanitizeFieldName(field.Name)] = struct{}{}
	}

	conflicts := []string{}
	for _, field := range bodyFields {
		if _, ok := names[SanitizeFieldName(field.Name)]; ok {
			conflicts = append(conflicts, field.Name)
		}
	}

	return conflicts
}

func injectSecurityArg(args Fields, security *FieldDef) Fields {
	if security == nil {
		return args
	}

	if security.Optional {
		return append(args, security)
	} else {
		return append(Fields{security}, args...)
	}
}

func sortArguments(args Fields) Fields {
	args = slices.Clone(args)

	slices.SortStableFunc(args, func(a, b *FieldDef) int {
		switch {
		case a.IsAdditionalProperties:
			return 1
		case b.IsAdditionalProperties:
			return -1
		case a.Optional && !b.Optional:
			return 1
		case !a.Optional && b.Optional:
			return -1
		default:
			return 0
		}
	})

	return args
}

func unpackFieldDef(field *FieldDef, options ArgumentsOptions) Fields {
	if field == nil {
		return []*FieldDef{}
	}

	if field.Type.Type != DataTypeClass {
		panic(fmt.Errorf("assertion error: %s: expected field type to be class: %s", field.Name, field.Type.Type))
	}

	if field.Type.Fields == nil {
		return []*FieldDef{}
	}

	fields := Fields{}

	for _, f := range field.Type.Fields {
		// If it's a const and we always treat them as optional then we skip them
		if f.Const != nil && (options.ConstFieldsAlwaysOptional == nil || *options.ConstFieldsAlwaysOptional) {
			continue
		}

		fields = append(fields, f)
	}

	return fields
}

func unpackParamFields(def *TypeDef, options ArgumentsOptions) Fields {
	if def == nil || def.Type != DataTypeClass {
		return Fields{}
	}

	result := make(Fields, 0, len(def.Fields))
	for _, f := range def.Fields {
		// If it's a const and we always treat them as optional then we skip them
		if f.Const != nil && (options.ConstFieldsAlwaysOptional == nil || *options.ConstFieldsAlwaysOptional) {
			continue
		}

		if !f.Annotations.Has(AnnotationTypeParam) {
			continue
		}

		result = append(result, f)
	}
	return result
}
