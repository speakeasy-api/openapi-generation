package ast

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// AnnotationType represents the various types of annotation
type AnnotationType string

const (
	AnnotationTypeJSON              = "json"
	AnnotationTypeSecurity          = "security"
	AnnotationTypeOperationSecurity = "opSecurity"
	AnnotationTypeParam             = "param"
	AnnotationTypeRequestWrapper    = "requestWrapper"
	AnnotationTypeRequest           = "request"
	AnnotationTypeMultipartForm     = "multipartForm"
	AnnotationTypeForm              = "form"
	AnnotationTypeEncoding          = "encoding"
	AnnotationTypeResponse          = "response"
	AnnotationTypeNeedsCasing       = "needsCasing"
)

// Annotation represents metadata that controls serialization and deserialization for a Field associated with a Type
type Annotation interface {
	// Should return a deep copy of the Annotation.
	Clone() Annotation

	// Should return true if the Annotation is equal to the given Annotation.
	IsEqual(a Annotation) bool

	// Should return true if the Annotation is of the same concrete type as the given Annotation.
	IsSameType(a Annotation) bool

	// Should return true if the Annotation is of the given type.
	IsType(t AnnotationType) bool

	// Should return the type of the Annotation.
	Type() AnnotationType
}

// Collection of Annotations.
type Annotations []Annotation

// Clone creates a deep copy of the Annotations
func (a Annotations) Clone() Annotations {
	if a == nil {
		return nil
	}

	cloned := make(Annotations, len(a))

	for i, annotation := range a {
		cloned[i] = annotation.Clone()
	}

	return cloned
}

func (a Annotations) Match(matchers Matchers) error {
	if matchers.Annotations != nil {
		return matchers.Annotations(a)
	}

	return nil
}

func (a Annotations) MarshalYAML() (any, error) {
	annos := []*yaml.Node{}

	for _, a := range a {
		var node yaml.Node
		if err := node.Encode(a); err != nil {
			return nil, err
		}

		node.Content = append(
			node.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "type"},
			&yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: string(a.Type())},
		)

		annos = append(annos, &node)
	}

	node := yaml.Node{
		Kind:    yaml.SequenceNode,
		Tag:     "!!seq",
		Content: annos,
	}

	return node, nil
}

func (a *Annotations) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.SequenceNode {
		return fmt.Errorf("expected sequence node, got %v", n.Kind)
	}

	for _, n := range n.Content {

		type anno struct {
			Type string
		}

		var an anno
		if err := n.Decode(&an); err != nil {
			return err
		}

		switch an.Type {
		case AnnotationTypeJSON:
			var anno JSONAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		case AnnotationTypeSecurity:
			var anno SecurityAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		case AnnotationTypeOperationSecurity:
			var anno OperationSecurityAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		case AnnotationTypeParam:
			var anno ParamAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		case AnnotationTypeRequestWrapper:
			var anno RequestWrapperAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		case AnnotationTypeRequest:
			var anno RequestAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		case AnnotationTypeMultipartForm:
			var anno MultipartFormAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		case AnnotationTypeForm:
			var anno FormAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		case AnnotationTypeEncoding:
			var anno EncodingAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		case AnnotationTypeResponse:
			var anno ResponseAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		case AnnotationTypeNeedsCasing:
			var anno NeedsCasingAnnotation
			if err := n.Decode(&anno); err != nil {
				return err
			}
			*a = append(*a, &anno)
		default:
			return fmt.Errorf("unknown annotation type: %s", an.Type)
		}
	}

	return nil
}

func (aa Annotations) Find(anno Annotation) (int, Annotation) {
	for i, a := range aa {
		if a.IsSameType(anno) {
			return i, a
		}
	}
	return -1, nil
}

func (aa Annotations) Get(typ AnnotationType) Annotation {
	for _, a := range aa {
		if a.IsType(typ) {
			return a
		}
	}
	return nil
}

// Returns the ParamAnnotation from the Annotations if present.
func (aa Annotations) GetParam() *ParamAnnotation {
	annotation := aa.Get(AnnotationTypeParam)

	if paramAnnotation, ok := annotation.(*ParamAnnotation); ok {
		return paramAnnotation
	}

	return nil
}

func (aa Annotations) Has(typ AnnotationType) bool {
	for _, a := range aa {
		if a.IsType(typ) {
			return true
		}
	}
	return false
}

func (aa Annotations) HasEqual(anno Annotation) bool {
	for _, a := range aa {
		if a.IsEqual(anno) {
			return true
		}
	}

	return false
}

func (aa *Annotations) Append(anno Annotation) {
	if !aa.HasEqual(anno) {
		*aa = append(*aa, anno)
	}
}
