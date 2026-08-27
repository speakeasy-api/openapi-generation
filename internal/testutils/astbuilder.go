package testutils

import (
	"fmt"
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi/pointer"
	"gopkg.in/yaml.v3"
)

// TypeDefBuilder provides a fluent interface for building TypeDef structures
type TypeDefBuilder struct {
	typeDef *ast.TypeDef
}

// NewTypeDef creates a new TypeDefBuilder with basic defaults
func NewTypeDef(dataType ast.DataType) *TypeDefBuilder {
	typeDef := &ast.TypeDef{
		Type: dataType,
		Extensions: &ast.TypeDefExtensions{
			All: make(map[string]any),
		},
	}

	// Only set Validations for primitive types that need them
	if dataType == ast.DataTypeString || dataType == ast.DataTypeInteger || dataType == ast.DataTypeNumber {
		typeDef.Validations = &ast.Validations{}
	}

	return &TypeDefBuilder{
		typeDef: typeDef,
	}
}

// WithLocation sets the OpenAPI location
func (b *TypeDefBuilder) WithLocation(node *yaml.Node) *TypeDefBuilder {
	b.typeDef.Location = &ast.OpenAPILocation{
		Node: node,
	}
	return b
}

// WithName sets the type name
func (b *TypeDefBuilder) WithName(name string) *TypeDefBuilder {
	b.typeDef.Name = name
	b.typeDef.OriginalName = name
	return b
}

// WithOriginalName sets the original name separately
func (b *TypeDefBuilder) WithOriginalName(originalName string) *TypeDefBuilder {
	b.typeDef.OriginalName = originalName
	return b
}

// WithScope sets the scope
func (b *TypeDefBuilder) WithScope(scope ast.Scope) *TypeDefBuilder {
	b.typeDef.Scope = scope
	return b
}

// WithRegistered marks the type as registered
func (b *TypeDefBuilder) WithRegistered() *TypeDefBuilder {
	b.typeDef.Registered = true
	return b
}

// WithOriginalNameFrozen marks the original name as frozen
func (b *TypeDefBuilder) WithOriginalNameFrozen() *TypeDefBuilder {
	b.typeDef.OriginalNameFrozen = true
	return b
}

// WithFields adds fields to the type
func (b *TypeDefBuilder) WithFields(fields ...*ast.FieldDef) *TypeDefBuilder {
	b.typeDef.Fields = append(b.typeDef.Fields, fields...)
	return b
}

// WithEmptyFields explicitly sets Fields to an empty slice (not nil)
func (b *TypeDefBuilder) WithEmptyFields() *TypeDefBuilder {
	b.typeDef.Fields = []*ast.FieldDef{}
	return b
}

// WithAssociatedTypes adds associated types for unions
func (b *TypeDefBuilder) WithAssociatedTypes(types ...*ast.TypeDef) *TypeDefBuilder {
	b.typeDef.AssociatedTypes = append(b.typeDef.AssociatedTypes, types...)
	return b
}

// WithEnum sets enum properties
func (b *TypeDefBuilder) WithEnum(enumType *ast.TypeDef, values []string, format string) *TypeDefBuilder {
	b.typeDef.Enum = &ast.Enum{
		Type:         ast.NewType(enumType, nil),
		Values:       values,
		Format:       format,
		Descriptions: make(map[string]string),
	}
	return b
}

// WithContextStack sets the context stack
func (b *TypeDefBuilder) WithContextStack(stack ast.ContextStack) *TypeDefBuilder {
	b.typeDef.ContextStack = stack
	return b
}

// WithUsedInUnion marks the type as used in a union
func (b *TypeDefBuilder) WithUsedInUnion() *TypeDefBuilder {
	b.typeDef.UsedInUnion = true
	return b
}

// WithValidations sets custom validations (overrides default)
func (b *TypeDefBuilder) WithValidations(validations *ast.Validations) *TypeDefBuilder {
	b.typeDef.Validations = validations
	return b
}

// WithComplexAny sets the ComplexAny flag
func (b *TypeDefBuilder) WithComplexAny(complexAny bool) *TypeDefBuilder {
	b.typeDef.ComplexAny = complexAny
	return b
}

// WithIsComponent sets the IsComponent flag
func (b *TypeDefBuilder) WithIsComponent(isComponent bool) *TypeDefBuilder {
	b.typeDef.IsComponent = isComponent
	return b
}

// WithComments sets the comments
func (b *TypeDefBuilder) WithComments(comments *ast.Comment) *TypeDefBuilder {
	b.typeDef.Comments = comments
	return b
}

// WithDiscriminator sets the discriminator
func (b *TypeDefBuilder) WithDiscriminator(discriminator *ast.Discriminator) *TypeDefBuilder {
	b.typeDef.Discriminator = discriminator
	return b
}

// WithInput sets the Input flag
func (b *TypeDefBuilder) WithInput(input bool) *TypeDefBuilder {
	b.typeDef.Input = input
	return b
}

// WithOutput sets the Output flag
func (b *TypeDefBuilder) WithOutput(output bool) *TypeDefBuilder {
	b.typeDef.Output = output
	return b
}

// WithTruncated sets the Truncated flag
func (b *TypeDefBuilder) WithTruncated(truncated bool) *TypeDefBuilder {
	b.typeDef.Truncated = truncated
	return b
}

// WithDeduplicatedContextStacks sets the deduplicated context stacks
func (b *TypeDefBuilder) WithDeduplicatedContextStacks(stacks ...ast.ContextStack) *TypeDefBuilder {
	b.typeDef.DeduplicatedContextStacks = stacks
	return b
}

// WithExamples sets the examples
func (b *TypeDefBuilder) WithExamples(examples ...*ast.Example) *TypeDefBuilder {
	if len(examples) == 0 {
		b.typeDef.Examples = ast.Examples{}
	} else {
		b.typeDef.Examples = append(b.typeDef.Examples, examples...)
	}
	return b
}

// WithItemType sets the item type for arrays and maps
func (b *TypeDefBuilder) WithItemType(itemType *ast.TypeDef) *TypeDefBuilder {
	b.typeDef.ItemType = itemType
	return b
}

// WithContainsNull sets the ContainsNull flag for arrays
func (b *TypeDefBuilder) WithContainsNull(containsNull bool) *TypeDefBuilder {
	b.typeDef.ContainsNull = containsNull
	return b
}

// WithFormat sets the format
func (b *TypeDefBuilder) WithFormat(format string) *TypeDefBuilder {
	b.typeDef.Format = format
	return b
}

// WithExtensions sets the extensions
func (b *TypeDefBuilder) WithExtensions(extensions map[string]interface{}) *TypeDefBuilder {
	b.typeDef.Extensions.All = extensions
	return b
}

// WithModelNamespace sets the model namespace
func (b *TypeDefBuilder) WithModelNamespace(namespace string) *TypeDefBuilder {
	b.typeDef.Extensions.ModelNamespace = &namespace
	return b
}

// WithIsMultipartFile sets the IsMultipartFile flag
func (b *TypeDefBuilder) WithIsMultipartFile(isMultipartFile bool) *TypeDefBuilder {
	b.typeDef.IsMultipartFile = isMultipartFile
	return b
}

// Build returns the constructed TypeDef
func (b *TypeDefBuilder) Build() *ast.TypeDef {
	return b.typeDef
}

// FieldDefBuilder provides a fluent interface for building FieldDef structures
type FieldDefBuilder struct {
	fieldDef *ast.FieldDef
}

// NewFieldDef creates a new FieldDefBuilder
func NewFieldDef(name string, typeDef *ast.TypeDef) *FieldDefBuilder {
	return &FieldDefBuilder{
		fieldDef: &ast.FieldDef{
			Name: name,
			Type: ast.NewType(typeDef, nil),
		},
	}
}

// WithOriginalName sets the original name
func (b *FieldDefBuilder) WithOriginalName(name string) *FieldDefBuilder {
	b.fieldDef.OriginalName = name
	return b
}

// WithNullable sets the nullable flag
func (b *FieldDefBuilder) WithNullable(nullable bool) *FieldDefBuilder {
	b.fieldDef.Nullable = nullable
	return b
}

// WithOptional sets the optional flag
func (b *FieldDefBuilder) WithOptional(optional bool) *FieldDefBuilder {
	b.fieldDef.Optional = optional
	return b
}

// WithAnnotations adds annotations
func (b *FieldDefBuilder) WithAnnotations(annotations ...ast.Annotation) *FieldDefBuilder {
	b.fieldDef.Annotations = append(b.fieldDef.Annotations, annotations...)
	return b
}

// WithJSONAnnotation adds a JSON annotation
func (b *FieldDefBuilder) WithJSONAnnotation(fieldName string) *FieldDefBuilder {
	b.fieldDef.Annotations = append(b.fieldDef.Annotations, &ast.JSONAnnotation{
		FieldName: fieldName,
	})
	return b
}

// WithDefault sets the default value
func (b *FieldDefBuilder) WithDefault(defaultValue *ast.AnyValue) *FieldDefBuilder {
	b.fieldDef.Default = defaultValue
	return b
}

// WithConst sets the const value
func (b *FieldDefBuilder) WithConst(constValue *ast.AnyValue) *FieldDefBuilder {
	b.fieldDef.Const = constValue
	return b
}

// WithIsAdditionalProperties sets the IsAdditionalProperties flag
func (b *FieldDefBuilder) WithIsAdditionalProperties(isAdditionalProperties bool) *FieldDefBuilder {
	b.fieldDef.IsAdditionalProperties = isAdditionalProperties
	return b
}

// WithComments sets the comments for the field
func (b *FieldDefBuilder) WithComments(comments *ast.Comment) *FieldDefBuilder {
	b.fieldDef.Comments = comments
	return b
}

// WithParameterIndex sets the parameter index
func (b *FieldDefBuilder) WithParameterIndex(index *int) *FieldDefBuilder {
	b.fieldDef.ParameterIndex = index
	return b
}

// Build returns the constructed FieldDef
func (b *FieldDefBuilder) Build() *ast.FieldDef {
	return b.fieldDef
}

// ContextStackBuilder provides a fluent interface for building context stacks
type ContextStackBuilder struct {
	stack ast.ContextStack
}

// NewContextStack creates a new ContextStackBuilder
func NewContextStack() *ContextStackBuilder {
	return &ContextStackBuilder{
		stack: ast.ContextStack{},
	}
}

// WithOneOf adds a oneOf context frame
func (b *ContextStackBuilder) WithOneOf(identifier string) *ContextStackBuilder {
	b.stack = append(b.stack, ast.ContextFrame{
		Type:       ast.ContextTypeOneOf,
		Identifier: identifier,
	})
	return b
}

// WithOneOfPosition adds a oneOfPosition context frame
func (b *ContextStackBuilder) WithOneOfPosition(position string, identifierForNaming *string) *ContextStackBuilder {
	b.stack = append(b.stack, ast.ContextFrame{
		Type:                ast.ContextTypeOneOfPosition,
		Identifier:          position,
		IdentifierForNaming: identifierForNaming,
	})
	return b
}

// WithStandardOneOfContext adds the standard oneOf context pattern (oneOf + oneOfPosition)
func (b *ContextStackBuilder) WithStandardOneOfContext(position string) *ContextStackBuilder {
	b.stack = append(b.stack,
		ast.ContextFrame{
			Type:       "oneOf",
			Identifier: "",
		},
		ast.ContextFrame{
			Type:                "oneOfPosition",
			Identifier:          position,
			IdentifierForNaming: pointer.From(""),
		},
	)
	return b
}

// WithRefType adds a reference type context frame
func (b *ContextStackBuilder) WithRefType(refType string) *ContextStackBuilder {
	b.stack = append(b.stack, ast.ContextFrame{
		Type:                ast.ContextTypeRefType,
		Identifier:          refType,
		IdentifierForNaming: pointer.From(""),
	})
	return b
}

// WithRefName adds a reference name context frame
func (b *ContextStackBuilder) WithRefName(refName string) *ContextStackBuilder {
	b.stack = append(b.stack, ast.ContextFrame{
		Type:                ast.ContextTypeRefName,
		Identifier:          refName,
		IdentifierForNaming: pointer.From(refName),
	})
	return b
}

// WithComponent adds a component context frame. The frame is marked used to
// match the walker, which consumes it under any non-legacy name resolution mode.
func (b *ContextStackBuilder) WithComponent() *ContextStackBuilder {
	b.stack = append(b.stack, ast.ContextFrame{
		Type:       ast.ContextTypeComponent,
		Identifier: "true",
		Used:       true,
	})
	return b
}

// WithModelNamespace adds a modelNamespace context frame
func (b *ContextStackBuilder) WithModelNamespace(namespace string) *ContextStackBuilder {
	b.stack = append(b.stack, ast.ContextFrame{
		Type:       ast.ContextTypeModelNamespace,
		Identifier: namespace,
	})
	return b
}

// WithConstProperty adds a constProperty context frame
func (b *ContextStackBuilder) WithConstProperty(identifier string) *ContextStackBuilder {
	b.stack = append(b.stack, ast.ContextFrame{
		Type:                ast.ContextTypeConstProperty,
		Identifier:          identifier,
		IdentifierForNaming: pointer.From(identifier),
	})
	return b
}

// WithIdentifierForNaming sets the IdentifierForNaming for the last frame in the stack
func (b *ContextStackBuilder) WithIdentifierForNaming(identifierForNaming *string) *ContextStackBuilder {
	if len(b.stack) > 0 {
		b.stack[len(b.stack)-1].IdentifierForNaming = identifierForNaming
	}
	return b
}

// WithInputOutput adds an inputOutput context frame
func (b *ContextStackBuilder) WithInputOutput(suffix string) *ContextStackBuilder {
	b.stack = append(b.stack, ast.ContextFrame{
		Type:       ast.ContextTypeInputOutput,
		Identifier: suffix,
	})
	return b
}

// Build returns the constructed context stack
func (b *ContextStackBuilder) Build() ast.ContextStack {
	return b.stack
}

// NewEnumTypeDef creates a complete enum TypeDef with the given values and format
func NewEnumTypeDef(enumType ast.DataType, values []string, format string) *TypeDefBuilder {
	// Create the underlying type without validations for enum types
	underlyingType := &ast.TypeDef{
		Type: enumType,
		Extensions: &ast.TypeDefExtensions{
			All: make(map[string]any),
		},
		// Explicitly do not set Validations for enum underlying types
	}

	typeDef := &ast.TypeDef{
		Type: ast.DataTypeEnum,
		Extensions: &ast.TypeDefExtensions{
			All: make(map[string]any),
		},
		Enum: &ast.Enum{
			Type:         ast.NewType(underlyingType, nil),
			Values:       values,
			Format:       format,
			Descriptions: make(map[string]string),
		},
	}

	return &TypeDefBuilder{
		typeDef: typeDef,
	}
}

// WithEnumUnderlyingLocation sets the location on the underlying type for enum TypeDefs
func (b *TypeDefBuilder) WithEnumUnderlyingLocation(node *yaml.Node) *TypeDefBuilder {
	if b.typeDef.Enum != nil && b.typeDef.Enum.Type != nil {
		// The Type field contains a TypeDef pointer, access it directly
		b.typeDef.Enum.Type.Location = &ast.OpenAPILocation{
			Node: node,
		}
	}
	return b
}

// WithEnumNames sets the names on the enum for custom enum names
func (b *TypeDefBuilder) WithEnumNames(names []string) *TypeDefBuilder {
	if b.typeDef.Enum != nil {
		b.typeDef.Enum.Names = names
	}
	return b
}

// NewDiscriminator creates a new discriminator with the given property name
func NewDiscriminator(propertyName string, mappings ...*ast.DiscriminatorMapping) *ast.Discriminator {
	return &ast.Discriminator{
		TypePropertyName: propertyName,
		Mapping:          mappings,
	}
}

// NewDiscriminatorMapping creates a new discriminator mapping
func NewDiscriminatorMapping(name string, typeDef *ast.TypeDef) *ast.DiscriminatorMapping {
	return &ast.DiscriminatorMapping{
		Name: name,
		Type: typeDef,
	}
}

// NewIntExample creates a new example with the given integer value
func NewIntExample(value string) *ast.Example {
	return &ast.Example{
		Value: &yaml.Node{
			Kind:   yaml.ScalarNode,
			Style:  0, // Plain style
			Tag:    "!!int",
			Value:  value,
			Line:   10,
			Column: 16,
		},
	}
}

// NewFloatExample creates a new example with the given float value
func NewFloatExample(value string) *ast.Example {
	return &ast.Example{
		Value: &yaml.Node{
			Kind:   yaml.ScalarNode,
			Style:  0, // Plain style
			Tag:    "!!float",
			Value:  value,
			Line:   10,
			Column: 16,
		},
	}
}

// NewSecurityAnnotation creates a security annotation with the given parameters
func NewSecurityAnnotation(fieldName, secType, subType, schemeKey string, scheme, securityOption bool) *ast.SecurityAnnotation {
	return &ast.SecurityAnnotation{
		FieldName:      fieldName,
		SecType:        secType,
		SubType:        subType,
		Scheme:         scheme,
		SecurityOption: securityOption,
		SchemeKey:      schemeKey,
	}
}

// NewTypeDefWithoutValidations creates a TypeDefBuilder with the given parameters
func NewTypeDefWithoutValidations(dataType ast.DataType) *TypeDefBuilder {
	typeDef := &ast.TypeDef{
		Type: dataType,
		Extensions: &ast.TypeDefExtensions{
			All: make(map[string]any),
		},
	}

	return &TypeDefBuilder{
		typeDef: typeDef,
	}
}

// SecurityBuilder provides a fluent interface for building Security structures
type SecurityBuilder struct {
	security *ast.Security
}

// NewSecurity creates a new SecurityBuilder
func NewSecurity() *SecurityBuilder {
	return &SecurityBuilder{
		security: &ast.Security{
			SecurityConfig: ast.SecurityConfig{
				OAuth2Config: ast.OAuth2Config{},
			},
		},
	}
}

// WithSecurity sets the security field
func (b *SecurityBuilder) WithSecurity(security *ast.FieldDef) *SecurityBuilder {
	b.security.Security = security
	return b
}

// WithRequirements adds security requirements. Each requirement is a sorted slice of scheme keys.
func (b *SecurityBuilder) WithRequirements(reqs ...ast.SecurityRequirement) *SecurityBuilder {
	if reqs == nil {
		b.security.Requirements = []ast.SecurityRequirement{}
	} else {
		b.security.Requirements = reqs
	}

	return b
}

// WithSchemes is a convenience wrapper around WithRequirements that accepts
// dash-separated scheme keys (e.g. "ApiKeyAuth", "ApiKeyAuth-BasicAuth").
// Each argument becomes one SecurityRequirement; dashes split into AND schemes.
func (b *SecurityBuilder) WithSchemes(schemes ...string) *SecurityBuilder {
	if len(schemes) == 0 {
		return b.WithRequirements()
	}
	reqs := make([]ast.SecurityRequirement, 0, len(schemes))
	for _, s := range schemes {
		var req ast.SecurityRequirement
		for _, part := range strings.Split(s, "-") {
			req = append(req, ast.SecurityScheme(part))
		}
		reqs = append(reqs, req)
	}
	return b.WithRequirements(reqs...)
}

// WithOAuth2Config sets the OAuth2 config with Comments and AvailableScopes
//
//	`schemeName` is the security scheme name (e.g., "OAuth2ClientCredentials")
//	`flow` is the OAuth2 flow type in snake_case (e.g., "client_credentials")
//	`description` is the security scheme description
//	`requiredScopes` are the *required* OAuth2 scopes
//	`availableScopes` holds *available* OAuth2 scopes and their description
func (b *SecurityBuilder) WithOAuth2Config(schemeName string, flow string, description string, requiredScopes []string, availableScopes []ast.OAuth2Scope) *SecurityBuilder {
	if b.security.SecurityConfig.OAuth2Config == nil {
		b.security.SecurityConfig.OAuth2Config = make(ast.OAuth2Config)
	}

	// Generate the same summary as the one set by `GetOAuth2FlowConfig` in `security/security.go`
	summary := fmt.Sprintf("Available scopes for the %s OAuth 2.0 scheme (%s flow).", schemeName, strcase.ToCamel(flow))

	b.security.SecurityConfig.OAuth2Config[schemeName] = ast.OAuth2FlowConfig{
		Flow:            ast.OAuth2Flow(flow),
		Enabled:         true,
		Comments:        ast.Comment{Summary: summary, Description: description},
		RequiredScopes:  requiredScopes,
		AvailableScopes: availableScopes,
	}
	return b
}

func (b *SecurityBuilder) WithHoistedSecurityConfig(equivalent bool, fields []ast.HoistedSecurityField) *SecurityBuilder {
	b.security.SecurityConfig.HoistedSecurityConfig = &ast.HoistedSecurityConfig{
		Equivalent: equivalent,
		Fields:     fields,
	}
	return b
}

// WithDisabled sets the disabled flag
func (b *SecurityBuilder) WithDisabled(disabled bool) *SecurityBuilder {
	b.security.SecurityConfig.Disabled = disabled
	return b
}

// WithOptionalityReason sets the optionality reason
func (b *SecurityBuilder) WithOptionalityReason(reason ast.SecurityOptionalityReason) *SecurityBuilder {
	b.security.SecurityConfig.OptionalityReason = reason
	return b
}

// Build returns the constructed Security
func (b *SecurityBuilder) Build() *ast.Security {
	return b.security
}
