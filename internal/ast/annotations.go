package ast

import (
	"slices"
)

type EncodingAnnotation struct {
	MediaType string `yaml:",omitempty"`
}

var _ Annotation = &EncodingAnnotation{}

// Clone creates a deep copy of the EncodingAnnotation.
func (e *EncodingAnnotation) Clone() Annotation {
	if e == nil {
		return nil
	}

	return &EncodingAnnotation{
		MediaType: e.MediaType,
	}
}

func (e *EncodingAnnotation) IsType(t AnnotationType) bool {
	return e.Type() == t
}

func (e *EncodingAnnotation) IsSameType(a Annotation) bool {
	return e.Type() == a.Type()
}

func (e *EncodingAnnotation) Type() AnnotationType {
	return AnnotationTypeEncoding
}

func (e *EncodingAnnotation) IsEqual(a Annotation) bool {
	if !e.IsType(a.Type()) {
		return false
	}

	anno, ok := a.(*EncodingAnnotation)
	if !ok {
		return false
	}

	return anno.MediaType == e.MediaType
}

type FormAnnotation struct {
	Name      string   `yaml:",omitempty"`
	JSON      bool     `yaml:",omitempty"`
	Style     string   `yaml:",omitempty"`
	Explode   bool     `yaml:",omitempty"`
	FieldType *TypeDef `yaml:",omitempty"`
}

var _ Annotation = &FormAnnotation{}

// Clone creates a deep copy of the FormAnnotation.
func (f *FormAnnotation) Clone() Annotation {
	if f == nil {
		return nil
	}

	var fieldType *TypeDef

	if f.FieldType != nil {
		fieldType = f.FieldType.Clone()
	}

	return &FormAnnotation{
		Explode:   f.Explode,
		FieldType: fieldType,
		JSON:      f.JSON,
		Name:      f.Name,
		Style:     f.Style,
	}
}

func (f *FormAnnotation) IsType(t AnnotationType) bool {
	return f.Type() == t
}

func (f *FormAnnotation) IsSameType(a Annotation) bool {
	return f.Type() == a.Type()
}

func (f *FormAnnotation) Type() AnnotationType {
	return AnnotationTypeForm
}

func (f *FormAnnotation) IsEqual(a Annotation) bool {
	if !f.IsType(a.Type()) {
		return false
	}

	anno := a.(*FormAnnotation)

	if f.Name != anno.Name || f.JSON != anno.JSON || f.Style != anno.Style || f.Explode != anno.Explode {
		return false
	}

	if (f.FieldType == nil && anno.FieldType != nil) || (f.FieldType != nil && anno.FieldType == nil) {
		return false
	}

	if f.FieldType == nil && anno.FieldType == nil {
		return true
	}

	return f.FieldType.IsEqual(anno.FieldType) == nil
}

type JSONAnnotation struct {
	Ignore    bool   `yaml:",omitempty"`
	FieldName string `yaml:",omitempty"`
}

var _ Annotation = &JSONAnnotation{}

// Clone creates a deep copy of the JSONAnnotation.
func (j *JSONAnnotation) Clone() Annotation {
	if j == nil {
		return nil
	}

	return &JSONAnnotation{
		FieldName: j.FieldName,
		Ignore:    j.Ignore,
	}
}

func (j *JSONAnnotation) IsType(t AnnotationType) bool {
	return j.Type() == t
}

func (j *JSONAnnotation) IsSameType(a Annotation) bool {
	return j.Type() == a.Type()
}

func (j *JSONAnnotation) Type() AnnotationType {
	return AnnotationTypeJSON
}

func (j *JSONAnnotation) IsEqual(a Annotation) bool {
	if !j.IsType(a.Type()) {
		return false
	}

	return j.FieldName == a.(*JSONAnnotation).FieldName
}

type MultipartFormAnnotation struct {
	File      bool     `yaml:",omitempty"`
	Content   bool     `yaml:",omitempty"`
	JSON      bool     `yaml:",omitempty"`
	Name      string   `yaml:",omitempty"`
	FieldType *TypeDef `yaml:",omitempty"`
}

var _ Annotation = &MultipartFormAnnotation{}

// Clone creates a deep copy of the MultipartFormAnnotation.
func (m *MultipartFormAnnotation) Clone() Annotation {
	if m == nil {
		return nil
	}

	var fieldType *TypeDef

	if m.FieldType != nil {
		fieldType = m.FieldType.Clone()
	}

	return &MultipartFormAnnotation{
		Content:   m.Content,
		FieldType: fieldType,
		File:      m.File,
		JSON:      m.JSON,
		Name:      m.Name,
	}
}

func (m *MultipartFormAnnotation) IsType(t AnnotationType) bool {
	return m.Type() == t
}

func (m *MultipartFormAnnotation) IsSameType(a Annotation) bool {
	return m.Type() == a.Type()
}

func (m *MultipartFormAnnotation) Type() AnnotationType {
	return AnnotationTypeMultipartForm
}

func (m *MultipartFormAnnotation) IsEqual(a Annotation) bool {
	if !m.IsType(a.Type()) {
		return false
	}

	anno := a.(*MultipartFormAnnotation)

	if m.File != anno.File || m.Content != anno.Content || m.JSON != anno.JSON || m.Name != anno.Name {
		return false
	}

	if (m.FieldType == nil && anno.FieldType != nil) || (m.FieldType != nil && anno.FieldType == nil) {
		return false
	}

	if m.FieldType == nil && anno.FieldType == nil {
		return true
	}

	return m.FieldType.IsEqual(anno.FieldType) == nil
}

const (
	ParamTypeQueryParam = "queryParam"
	ParamTypePathParam  = "pathParam"
	ParamTypeHeader     = "header"
)

// Parameter annotation data.
type ParamAnnotation struct {
	ParamType     string   `yaml:",omitempty"`
	Name          string   `yaml:",omitempty"`
	Serialization string   `yaml:",omitempty"`
	Style         string   `yaml:",omitempty"`
	Explode       bool     `yaml:",omitempty"`
	FieldType     *TypeDef `yaml:",omitempty"`
	AllowReserved bool     `yaml:",omitempty"`

	// Enabled when the parameter is global parameter
	IsGlobal bool `yaml:",omitempty"`
	// Contains the operations this parameter is used in if IsGlobal is true
	OperationsForGlobal []string `yaml:",omitempty"`

	// Enabled when the parameter is local to an operation but has a global representation
	HasGlobal bool `yaml:",omitempty"`

	// Enabled when the parameter should be hidden, where the value should not
	// be settable from an operation and only at the SDK level. Hidden is only set
	// if the parameter is global and the x-speakeasy-globals-hidden extension
	// is enabled.
	Hidden bool `yaml:",omitempty"`

	// Enabled when the parameter is global and the path or operation marks it
	// as required. Global parameters are implicitly marked as optional, so this
	// captures the overriding value requirement.
	RequiredForOperation bool `yaml:",omitempty"`
}

var _ Annotation = &ParamAnnotation{}

// Clone creates a deep copy of the ParamAnnotation.
func (p *ParamAnnotation) Clone() Annotation {
	if p == nil {
		return nil
	}

	var fieldType *TypeDef

	if p.FieldType != nil {
		fieldType = p.FieldType.Clone()
	}

	return &ParamAnnotation{
		AllowReserved:        p.AllowReserved,
		Explode:              p.Explode,
		FieldType:            fieldType,
		HasGlobal:            p.HasGlobal,
		Hidden:               p.Hidden,
		IsGlobal:             p.IsGlobal,
		Name:                 p.Name,
		OperationsForGlobal:  slices.Clone(p.OperationsForGlobal),
		ParamType:            p.ParamType,
		RequiredForOperation: p.RequiredForOperation,
		Serialization:        p.Serialization,
		Style:                p.Style,
	}
}

func (p *ParamAnnotation) IsType(t AnnotationType) bool {
	return p.Type() == t
}

func (p *ParamAnnotation) IsSameType(a Annotation) bool {
	return p.Type() == a.Type() && p.ParamType == a.(*ParamAnnotation).ParamType
}

func (p *ParamAnnotation) Type() AnnotationType {
	return AnnotationTypeParam
}

func (p *ParamAnnotation) IsEqual(a Annotation) bool {
	if !p.IsType(a.Type()) {
		return false
	}

	anno := a.(*ParamAnnotation)

	if p.ParamType != anno.ParamType || p.Name != anno.Name || p.Style != anno.Style || p.Explode != anno.Explode {
		return false
	}

	if (p.FieldType == nil && anno.FieldType != nil) || (p.FieldType != nil && anno.FieldType == nil) {
		return false
	}

	if p.FieldType == nil && anno.FieldType == nil {
		return true
	}

	return p.FieldType.IsEqual(anno.FieldType) == nil
}

// Helpers for determining if a set of fields contains parameters

func HasParams(fields Fields) bool {
	for _, field := range fields {
		if field.Annotations.Has(AnnotationTypeParam) {
			return true
		}
	}

	return false
}

func HasPathParams(fields Fields) bool {
	for _, field := range fields {
		if field.Annotations.Has(AnnotationTypeParam) {
			if field.Annotations.Get(AnnotationTypeParam).(*ParamAnnotation).ParamType == ParamTypePathParam {
				return true
			}
		}
	}

	return false
}

type RequestAnnotation struct {
	MediaType string `yaml:",omitempty"`
}

var _ Annotation = &RequestAnnotation{}

// Clone creates a deep copy of the RequestAnnotation.
func (r *RequestAnnotation) Clone() Annotation {
	if r == nil {
		return nil
	}

	return &RequestAnnotation{
		MediaType: r.MediaType,
	}
}

func (r *RequestAnnotation) IsType(t AnnotationType) bool {
	return r.Type() == t
}

func (r *RequestAnnotation) IsSameType(a Annotation) bool {
	return r.Type() == a.Type()
}

func (r *RequestAnnotation) Type() AnnotationType {
	return AnnotationTypeRequest
}

func (r *RequestAnnotation) IsEqual(a Annotation) bool {
	if !r.IsType(a.Type()) {
		return false
	}

	return r.MediaType == a.(*RequestAnnotation).MediaType
}

type RequestWrapperAnnotation struct{}

var _ Annotation = &RequestWrapperAnnotation{}

// Clone creates a deep copy of the RequestWrapperAnnotation.
func (r *RequestWrapperAnnotation) Clone() Annotation {
	if r == nil {
		return nil
	}

	return &RequestWrapperAnnotation{}
}

func (r *RequestWrapperAnnotation) IsType(t AnnotationType) bool {
	return r.Type() == t
}

func (r *RequestWrapperAnnotation) IsSameType(a Annotation) bool {
	return r.Type() == a.Type()
}

func (r *RequestWrapperAnnotation) Type() AnnotationType {
	return AnnotationTypeRequestWrapper
}

func (r *RequestWrapperAnnotation) IsEqual(a Annotation) bool {
	return r.IsType(a.Type())
}

type ResponseAnnotation struct {
	ResultField bool `yaml:",omitempty"`
}

var _ Annotation = &ResponseAnnotation{}

// Clone creates a deep copy of the ResponseAnnotation.
func (r *ResponseAnnotation) Clone() Annotation {
	if r == nil {
		return nil
	}

	return &ResponseAnnotation{
		ResultField: r.ResultField,
	}
}

func (r *ResponseAnnotation) IsType(t AnnotationType) bool {
	return r.Type() == t
}

func (r *ResponseAnnotation) IsSameType(a Annotation) bool {
	return r.Type() == a.Type()
}

func (r *ResponseAnnotation) Type() AnnotationType {
	return AnnotationTypeResponse
}

func (r *ResponseAnnotation) IsEqual(a Annotation) bool {
	if !r.IsType(a.Type()) {
		return false
	}

	return r.ResultField == a.(*ResponseAnnotation).ResultField
}

type SecurityAnnotation struct {
	FieldName string `yaml:",omitempty"`

	// Security scheme type from the OAS Security Scheme object "type" field.
	// Values include "apiKey", "http", "oauth2", and "openIdConnect".
	SecType string `yaml:",omitempty"`

	// Underlying security type. Usage is dependent on SecType:
	// - apiKey: Location of API key from the OAS Security Scheme object "in"
	//           field. Values include "cookie", "header", and "query".
	// - http: HTTP Authentication scheme from the OAS Security Scheme object
	//         "scheme" field, normalized to lowercase. Values include "basic",
	//         "bearer", and "custom".
	// - oauth2: OAuth2 Flow type from the OAuth Flows object field name,
	//           normalized to lowercase snakecase. Values include
	//           "client_credentials" and "password".
	// - openIdConnect: N/A
	SubType string `yaml:",omitempty"`

	// Option (a.k.a. "OptionWrapper") marks this field as a wrapper struct
	// that groups one or more scheme fields together.
	// Option is false for flat configurations, i.e. when multiple simple
	// schemes are in either a OR or AND group (but not both).
	// Option is set for all fields when the security definition is a
	// combination of OR and AND relationships.
	Option bool `yaml:",omitempty"`

	// Scheme marks this field as a security scheme entry point. True for
	// flattened credential fields, unflattened security classes, and fields
	// inside OptionWrapper structs.
	Scheme bool `yaml:",omitempty"`

	// SecurityOption (a.k.a. "Alternative") marks this field as one of multiple
	// independent OR alternatives for authentication. Set on both OptionWrappers
	// and flattened scheme fields when numSchemes > 1. When false, the field
	// is either the only scheme available or part of an AND group.
	SecurityOption bool `yaml:",omitempty"`

	// Composite marks this field as part of an AND group of security schemes.
	Composite bool `yaml:",omitempty"`

	// SchemeKey represents the key of the security scheme in the security field
	SchemeKey string `yaml:",omitempty"`
}

var _ Annotation = &SecurityAnnotation{}

// Clone creates a deep copy of the SecurityAnnotation.
func (s *SecurityAnnotation) Clone() Annotation {
	if s == nil {
		return nil
	}

	return &SecurityAnnotation{
		FieldName:      s.FieldName,
		Option:         s.Option,
		Scheme:         s.Scheme,
		SchemeKey:      s.SchemeKey,
		SecType:        s.SecType,
		SecurityOption: s.SecurityOption,
		Composite:      s.Composite,
		SubType:        s.SubType,
	}
}

func (s *SecurityAnnotation) IsType(t AnnotationType) bool {
	return s.Type() == t
}

func (s *SecurityAnnotation) IsSameType(a Annotation) bool {
	return s.Type() == a.Type()
}

func (s *SecurityAnnotation) Type() AnnotationType {
	return AnnotationTypeSecurity
}

func (s *SecurityAnnotation) IsEqual(a Annotation) bool {
	if !s.IsType(a.Type()) {
		return false
	}

	other := a.(*SecurityAnnotation)

	return s.FieldName == other.FieldName && s.SecType == other.SecType && s.SubType == other.SubType && s.Option == other.Option && s.Scheme == other.Scheme
}

func (s *SecurityAnnotation) Merge(a *SecurityAnnotation) {
	if a.FieldName != "" {
		s.FieldName = a.FieldName
	}

	if a.SecType != "" {
		s.SecType = a.SecType
	}

	if a.SubType != "" {
		s.SubType = a.SubType
	}

	if a.Option {
		s.Option = a.Option
	}

	if a.Scheme {
		s.Scheme = a.Scheme
	}

	if a.SecurityOption {
		s.SecurityOption = a.SecurityOption
	}

	if a.Composite {
		s.Composite = a.Composite
	}

	if a.SchemeKey != "" {
		s.SchemeKey = a.SchemeKey
	}
}

type OperationSecurityAnnotation struct{}

var _ Annotation = &OperationSecurityAnnotation{}

// Clone creates a deep copy of the OperationSecurityAnnotation.
func (o *OperationSecurityAnnotation) Clone() Annotation {
	if o == nil {
		return nil
	}

	return &OperationSecurityAnnotation{}
}

func (r *OperationSecurityAnnotation) IsType(t AnnotationType) bool {
	return r.Type() == t
}

func (r *OperationSecurityAnnotation) IsSameType(a Annotation) bool {
	return r.Type() == a.Type()
}

func (r *OperationSecurityAnnotation) Type() AnnotationType {
	return AnnotationTypeOperationSecurity
}

func (r *OperationSecurityAnnotation) IsEqual(a Annotation) bool {
	return r.IsType(a.Type())
}

type NeedsCasingAnnotation struct{}

var _ Annotation = &NeedsCasingAnnotation{}

// Clone creates a deep copy of the NeedsCasingAnnotation.
func (n *NeedsCasingAnnotation) Clone() Annotation {
	if n == nil {
		return nil
	}

	return &NeedsCasingAnnotation{}
}

func (n *NeedsCasingAnnotation) IsType(t AnnotationType) bool {
	return n.Type() == t
}

func (n *NeedsCasingAnnotation) IsSameType(a Annotation) bool {
	return n.Type() == a.Type()
}

func (n *NeedsCasingAnnotation) Type() AnnotationType {
	return AnnotationTypeNeedsCasing
}

func (n *NeedsCasingAnnotation) IsEqual(a Annotation) bool {
	return n.IsType(a.Type())
}
