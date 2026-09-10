package ast

import (
	"fmt"
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/openapi-generation/v2/internal/sanitization"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	config "github.com/speakeasy-api/sdk-gen-config"
)

type ContextType string

const (
	ContextTypeRefType            ContextType = "refType" // normally "schemas" or "responses" but sometimes "properties"
	ContextTypeRefName            ContextType = "refName" // eg normally the component name but sometimes overridden with x-speakeasy-name-override
	ContextTypeProperty           ContextType = "property"
	ContextTypeInputOutput        ContextType = "inputOutput"
	ContextTypeOneOf              ContextType = "oneOf"
	ContextTypeOneOfPosition      ContextType = "oneOfPosition"
	ContextTypeConstProperty      ContextType = "constProperty"
	ContextTypeRequestResponse    ContextType = "requestResponse"
	ContextTypeRequestMediaType   ContextType = "requestMediaType"
	ContextTypeRequestBody        ContextType = "requestBody"
	ContextTypeResponseStatusCode ContextType = "responseStatusCode"
	ContextTypeResponseError      ContextType = "responseError"
	ContextTypeResponseMediaType  ContextType = "responseMediaType"
	ContextTypeResponseBody       ContextType = "responseBody"
	ContextTypeOperation          ContextType = "operation"
	ContextTypeOperationTag       ContextType = "operationTag"
	ContextTypeGroup              ContextType = "group"
	ContextTypeMainSDK            ContextType = "mainSDK"
	ContextTypeParameter          ContextType = "parameter"
	ContextTypeComponent          ContextType = "component"
	ContextTypeRegisterDuplicate  ContextType = "registerDuplicate"
	ContextTypeModelNamespace     ContextType = "modelNamespace" // x-speakeasy-model-namespace extension value
)

type ContextFrame struct {
	Type                ContextType `yaml:",omitempty"` // The Type of the context which represents things like whether it came from a request or response etc
	Identifier          string      `yaml:",omitempty"` // The identifier of the context, e.g. the operation id or parent schema ref
	IdentifierForNaming *string     `yaml:",omitempty"` // The humanized identifier of the context, e.g. the operation id or parent schema ref
	Used                bool        `yaml:",omitempty"` // Whether or not the context has been used for renaming
	MustUse             bool        `yaml:",omitempty"` // Deprecated: Whether or not the context must be used for renaming
}

// Clone creates a deep copy of the ContextFrame
func (f ContextFrame) Clone() ContextFrame {
	return ContextFrame{
		Identifier:          f.Identifier,
		IdentifierForNaming: clonePtr(f.IdentifierForNaming),
		MustUse:             f.MustUse,
		Type:                f.Type,
		Used:                f.Used,
	}
}

func (f ContextFrame) DisplayName() string {
	if f.IdentifierForNaming != nil {
		return *f.IdentifierForNaming
	}
	return f.Identifier
}

func (f *ContextFrame) MarkUsed() {
	f.Used = true
}

type ContextStack []ContextFrame

// Returns a deep copy of the ContextStack.
func (s ContextStack) Clone() ContextStack {
	if s == nil {
		return nil
	}

	cloned := make(ContextStack, len(s))

	for i, frame := range s {
		cloned[i] = frame.Clone()
	}

	return cloned
}

func (s ContextStack) Match(matchers Matchers) error {
	if matchers.ContextStack != nil {
		return matchers.ContextStack(s)
	}

	return nil
}

func (s ContextStack) FindLastFrameOfType(typ ContextType) *ContextFrame {
	for i := len(s) - 1; i >= 0; i-- {
		if (s)[i].Type == typ {
			return &(s)[i]
		}
	}

	return nil
}

func (s ContextStack) LastFrame() *ContextFrame {
	if len(s) == 0 {
		return nil
	}

	return &s[len(s)-1]
}

func (s *ContextStack) PopLastFrame() {
	if len(*s) == 0 {
		return
	}

	*s = (*s)[:len(*s)-1]
}

func (s ContextStack) HasFrameOfType(typ ContextType) bool {
	for i := len(s) - 1; i >= 0; i-- {
		if (s)[i].Type == typ {
			return true
		}
	}

	return false
}

func (s ContextStack) String() string {
	builder := strings.Builder{}
	for i, frame := range s {
		fmt.Fprintf(&builder, "%s:%s", frame.Type, frame.Identifier)
		if i < len(s)-1 {
			builder.WriteString(" ")
		}
	}
	return builder.String()
}

func (s ContextStack) GetGroups() []ContextFrame {
	var tags []ContextFrame
	for _, frame := range s {
		if frame.Type == ContextTypeGroup {
			tags = append(tags, frame)
		}
	}
	return tags
}

func (s *ContextStack) IsUsed(typ ContextType) bool {
	for i := 0; i < len(*s); i++ {
		if (*s)[i].Type == typ {
			return (*s)[i].Used
		}
	}

	return false
}

func (s *ContextStack) MarkUsed(typ ContextType) {
	for i := 0; i < len(*s); i++ {
		if (*s)[i].Type == typ {
			(*s)[i].Used = true
		}
	}
}

func (s *ContextStack) MarkUnused(typ ContextType) {
	for i := 0; i < len(*s); i++ {
		if (*s)[i].Type == typ {
			(*s)[i].Used = false
		}
	}
}

func (s *ContextStack) Append(typ ContextType, identifier string) {
	*s = append(*s, ContextFrame{Type: typ, Identifier: identifier})
}

// Appends a group frame to the context stack.
func (s *ContextStack) AppendGroup(name string) {
	s.Append(ContextTypeGroup, name)
}

// Appends a main SDK frame to the context stack.
func (s *ContextStack) AppendMainSDK(name string) {
	s.Append(ContextTypeMainSDK, name)
}

func (s *ContextStack) AppendWithHumanized(typ ContextType, identifier string, identifierForNaming string) {
	*s = append(*s, ContextFrame{Type: typ, Identifier: identifier, IdentifierForNaming: &identifierForNaming})
}

const HumanizedRequestBody = "RequestBody"

func (s *ContextStack) AppendRequestBody(fixes *config.Fixes) {
	s.AppendWithHumanized(ContextTypeRequestBody, "requestBody", HumanizedRequestBody)
}

const HumanizedResponseBody = "ResponseBody"

func (s *ContextStack) AppendResponseBody(fixes *config.Fixes) {
	s.AppendWithHumanized(ContextTypeResponseBody, "responseBody", HumanizedResponseBody)
}

func (s *ContextStack) AppendResponseStatusCode(statusCode string, fixes *config.Fixes) {
	s.AppendWithHumanized(ContextTypeResponseStatusCode, statusCode, sanitization.HumanizeStatusCode(statusCode))
}

func (s *ContextStack) AppendResponseMediaType(identifier string, fixes *config.Fixes) {
	if !fixes.NameResolutionFeb2025 {
		s.Append(ContextTypeResponseMediaType, oldSanitizeMediaType(identifier))
		return
	}

	s.AppendWithHumanized(ContextTypeResponseMediaType, strings.ToLower(identifier), sanitization.HumanizeMediaType(identifier))
}

func (s *ContextStack) AppendOperation(operation string) {
	s.AppendWithHumanized(ContextTypeOperation, operation, sanitization.SanitizeName(operation))
}

func (s *ContextStack) UpdateOperation(operation string) {
	s.Update(ContextTypeOperation, operation, sanitization.SanitizeName(operation))
}

func (s *ContextStack) Update(typ ContextType, identifier string, humanizedIdentifier string) {
	frame := s.FindLastFrameOfType(typ)
	if frame != nil {
		frame.Identifier = identifier
		frame.IdentifierForNaming = &humanizedIdentifier
	}
}

func (s *ContextStack) AppendRequestMediaType(mediaType string, fixes *config.Fixes) {
	if !fixes.NameResolutionFeb2025 {
		s.Append(ContextTypeRequestMediaType, oldSanitizeMediaType(mediaType))
		return
	}

	s.AppendWithHumanized(ContextTypeRequestMediaType, strings.ToLower(mediaType), sanitization.HumanizeMediaType(mediaType))
}

// These may be used to suffix types
var humanizedRefTypes = map[string]string{
	"schemas":         "",
	"responses":       "response",
	"parameters":      "parameter",
	"examples":        "example",
	"requestbodies":   "requestBody",
	"headers":         "header",
	"securityschemes": "securityScheme",
	"links":           "link",
	"callbacks":       "callback",
	"pathitems":       "pathItem",
	"paths":           "path",
}

func (s *ContextStack) AppendRefType(refType string) {
	humanizedRefType := refType
	if x, ok := humanizedRefTypes[strings.ToLower(refType)]; ok {
		humanizedRefType = x
	}

	s.AppendWithHumanized(ContextTypeRefType, refType, humanizedRefType)
}

// AppendModelNamespace appends a model namespace frame to the context stack.
// This is used to differentiate types that have the x-speakeasy-model-namespace extension.
func (s *ContextStack) AppendModelNamespace(namespace string) {
	if namespace == "" {
		return
	}
	s.Append(ContextTypeModelNamespace, namespace)
}

func (s *ContextStack) Filter(fn func(ContextFrame) bool) {
	newStack := ContextStack{}
	for _, frame := range *s {
		if fn(frame) {
			newStack = append(newStack, frame)
		}
	}
	*s = newStack
}

// Collection of ContextStack.
type ContextStacks []ContextStack

// Clone creates a deep copy of the ContextStacks.
func (s ContextStacks) Clone() ContextStacks {
	if s == nil {
		return nil
	}

	cloned := make(ContextStacks, len(s))

	for i, stack := range s {
		cloned[i] = stack.Clone()
	}

	return cloned
}

// oldSanitizeMediaType is the old sanitize method that was used before the new name resolution changes
func oldSanitizeMediaType(mediaType string) string {
	if mediaType == "*/*" {
		return "Wildcard"
	}
	parts := strings.Split(mediaType, "/")
	sanitizeParts := utils.MapArray(parts, func(s string) string {
		if s == "*" {
			return "Wildcard"
		}
		return s
	})
	sanitizeParts = utils.MapArray(sanitizeParts, strcase.ToGoPascal)
	sanitizeParts = utils.MapArray(sanitizeParts, func(s string) string {
		parts := strings.Split(s, "-")
		sanitizeParts := utils.MapArray(parts, strcase.ToGoPascal)
		return strings.Join(sanitizeParts, "")
	})
	return strings.Join(sanitizeParts, "")
}
