package ast

import (
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	generatorErrors "github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"gopkg.in/yaml.v3"
)

const (
	ASTVersion       = "0.3.0"
	ErrTerminateWalk = generatorErrors.Error("walk terminated")
)

type BucketedTypes = *sequencedmap.Map[string, *sequencedmap.Map[string, TypeDefs]]

func NewBucketedTypes() BucketedTypes {
	return sequencedmap.New[string, *sequencedmap.Map[string, TypeDefs]]()
}

type AST struct {
	ASTVersion string

	// Given that this AST is primarily concerned with producing an AST
	// for the *generated* SDK, there are often cases where we don't want
	// to pollute the resulting AST with lower level OpenAPI document structures.
	// For the docs product where we produce a "view model" AST, we still need to
	// access these underlying constructs so we embed the OpenAPI document here.
	OpenAPIDocument *openapi.OpenAPI
	MainSDK         *SDK
	Webhooks        *sequencedmap.Map[string, []Operation]
	// Arazzo contains test-generation workflow graph data derived from
	// x-speakeasy-test Arazzo documents.
	Arazzo           *Arazzo
	Tests            *Tests
	Components       *sequencedmap.Map[string, *TypeDef]
	UsedFeatures     map[string]bool
	UsedTypes        map[string]bool
	BucketedTypes    BucketedTypes
	OperationServers *sequencedmap.Map[string, *Servers]
	PublicExports    *PublicExports

	// CLICommands is the decoded x-speakeasy-cli-commands document extension:
	// declaratively defined CLI intent commands rendered by the cli target.
	CLICommands *extensions.CLICommandManifest

	// CLIErrors is the decoded x-speakeasy-cli-errors document extension:
	// global reason and type hint rules rendered by the cli target.
	CLIErrors *extensions.CLIErrorManifest

	// CLIBodySchemas maps operationId → self-contained JSON Schema for the
	// operation's application/json request body (components bundled under
	// $defs). Rendered by the cli target as an exact --schema surface.
	CLIBodySchemas map[string]string

	// Terraform Provider as gathered from various x-speakeasy-entity*
	// extensions across the source document. Only set when generation target
	// is "terraform".
	TerraformProvider *TerraformProvider
}

type AnyValue struct {
	Value any
}

// Clone creates a shallow copy of the AnyValue
func (v *AnyValue) Clone() *AnyValue {
	if v == nil {
		return nil
	}

	cloned := &AnyValue{}

	if v.Value != nil {
		cloned.Value = cloneAnyValue(v.Value)
	}

	return cloned
}

func cloneAnyValue(v any) any {
	if v == nil {
		return nil
	}

	switch val := v.(type) {
	case []any:
		return slices.Clone(val)
	case map[string]any:
		return maps.Clone(val)
	default:
		cloned := v

		return cloned
	}
}

func NewAST() *AST {
	return &AST{
		Components: sequencedmap.New[string, *TypeDef](),
		Webhooks:   sequencedmap.New[string, []Operation](),
	}
}

// Returns all operations which are entity operations.
func (a AST) EntityOperations() (*sequencedmap.Map[string, *Operation], error) {
	result := sequencedmap.New[string, *Operation]()

	err := a.Walk(func(n Node, parents []Node, a *AST) error {
		return n.Match(Matchers{
			Operation: func(o *Operation) error {
				if o.Extensions == nil || o.Extensions.EntityOperation == nil {
					return nil
				}

				result.Set(o.ID, o)

				return nil
			},
		})
	})

	return result, err
}

func (a AST) MarshalYAML() (any, error) {
	type ast struct {
		ASTVersion        string
		MainSDK           *SDK
		Webhooks          *sequencedmap.Map[string, []Operation] `yaml:",omitempty"`
		Components        *sequencedmap.Map[string, yaml.Node]
		UsedFeatures      map[string]bool
		TerraformProvider *TerraformProvider `yaml:",omitempty"`
	}

	am := ast{
		ASTVersion:        ASTVersion,
		MainSDK:           a.MainSDK,
		Webhooks:          a.Webhooks,
		Components:        sequencedmap.New[string, yaml.Node](),
		UsedFeatures:      a.UsedFeatures,
		TerraformProvider: a.TerraformProvider,
	}

	for registeredID, v := range a.Components.All() {
		currID := v.GetRegistrationID()

		if currID != registeredID {
			return nil, fmt.Errorf("type %q has a different registration id than the one in the AST: %q", registeredID, currID)
		}

		v.Reference = currID
	}

	// Sort keys for stable lockfile generation. Use a different methodology,
	// such as additional logging, if debugging functionality affected by
	// ordering.
	sortedKeys := slices.Collect(a.Components.Keys())
	slices.Sort(sortedKeys)

	for _, k := range sortedKeys {
		v, _ := a.Components.Get(k)
		t := *v
		t.Reference = ""

		var node yaml.Node
		if err := node.Encode(t); err != nil {
			return nil, err
		}

		am.Components.Set(k, node)
	}

	var amNode yaml.Node
	if err := amNode.Encode(am); err != nil {
		return nil, err
	}

	for _, v := range a.Components.All() {
		v.Reference = ""
	}

	return amNode, nil
}

func (a *AST) Resolve() {
	_ = a.Walk(func(n Node, parents []Node, a *AST) error {
		return n.Match(Matchers{
			Annotations: func(annos Annotations) error {
				for _, ann := range annos {
					switch v := ann.(type) {
					case *EncodingAnnotation:
					case *FormAnnotation:
						v.FieldType = a.ResolveTypeDef(v.FieldType)
					case *JSONAnnotation:
					case *MultipartFormAnnotation:
						v.FieldType = a.ResolveTypeDef(v.FieldType)
					case *ParamAnnotation:
						v.FieldType = a.ResolveTypeDef(v.FieldType)
					case *RequestAnnotation:
					case *RequestWrapperAnnotation:
					case *ResponseAnnotation:
					case *SecurityAnnotation:
					case *OperationSecurityAnnotation:
					default:
						panic(fmt.Errorf("unknown annotation type: %T", ann))
					}
				}
				return nil
			},
			FieldDef: func(f *FieldDef) error {
				f.Type = a.ResolveTypeDef(f.Type)
				return nil
			},
			Operation: func(op *Operation) error {
				for i, c := range op.Callbacks {
					op.Callbacks[i] = a.ResolveTypeDef(c)
				}

				op.Globals = a.ResolveTypeDef(op.Globals)

				// Match against the immediate parent (if available)
				if len(parents) > 0 {
					last := parents[len(parents)-1]
					_ = last.Match(Matchers{
						SDK: func(s *SDK) error {
							op.OwningSDK = s
							return nil
						},
					})
				}

				return nil
			},
			Response: func(r *Response) error {
				r.Type = a.ResolveTypeDef(r.Type)
				return nil
			},
			SDK: func(s *SDK) error {
				s.Type = a.ResolveTypeDef(s.Type)
				for i, at := range s.AdditionalTypes {
					s.AdditionalTypes[i] = a.ResolveTypeDef(at)
				}
				s.Globals = a.ResolveTypeDef(s.Globals)
				return nil
			},
			ServerVariable: func(s *ServerVariable) error {
				s.Type = a.ResolveTypeDef(s.Type)
				return nil
			},
			TypeDef: func(td *TypeDef) error {
				td.ItemType = a.ResolveTypeDef(td.ItemType)
				for i, at := range td.AssociatedTypes {
					td.AssociatedTypes[i] = a.ResolveTypeDef(at)
				}
				return nil
			},
			DiscriminatorMapping: func(dm *DiscriminatorMapping) error {
				dm.Type = a.ResolveTypeDef(dm.Type)
				return nil
			},
			Enum: func(e *Enum) error {
				e.Type = a.ResolveTypeDef(e.Type)
				return nil
			},
			OperationExtensions: func(oe *OperationExtensions) error {
				if oe.Pagination != nil {
					extensions.ComputePaginationDotNotation(oe.Pagination)
				}
				return nil
			},
		})
	})
}

func (a *AST) ResolveTypeDef(t *TypeDef) *TypeDef {
	if t == nil {
		return nil
	}

	if t.Reference != "" {
		if a.Components == nil {
			panic(fmt.Errorf("failed to resolve type %q: components not set", t.Reference))
		}

		rt, ok := a.Components.Get(t.Reference)
		if !ok {
			panic(fmt.Errorf("failed to resolve type %q", t.Reference))
		}

		t = rt
	}

	return t
}

type Matchers struct {
	SDK                  func(*SDK) error
	Servers              func(*Servers) error
	Server               func(*Server) error
	ServerVariable       func(*ServerVariable) error
	FieldDef             func(*FieldDef) error
	Annotations          func(Annotations) error
	TypeDef              func(*TypeDef) error
	ContextStack         func(ContextStack) error
	Validations          func(*Validations) error
	Enum                 func(*Enum) error
	Discriminator        func(*Discriminator) error
	DiscriminatorMapping func(*DiscriminatorMapping) error
	Examples             func(Examples) error
	Operation            func(*Operation) error
	Request              func(*Request) error
	RequestParams        func(*RequestParams) error
	Param                func(*Param) error
	Response             func(*Response) error
	SubResponse          func(*SubResponse) error
	ResponseBodyContent  func(*ResponseBodyContent) error
	OperationExtensions  func(*OperationExtensions) error
	TypeDefExtensions    func(*TypeDefExtensions) error
	Arguments            func(*Arguments) error
	Comment              func(*Comment) error
	ExtendedComment      func(*ExtendedComment) error
	ExternalDocs         func(*ExternalDocs) error
}

type Node interface {
	Match(Matchers) error
}

type NodeType string

const (
	NodeTypeSDK                  NodeType = "SDK"
	NodeTypeServers              NodeType = "Servers"
	NodeTypeServer               NodeType = "Server"
	NodeTypeServerVariable       NodeType = "ServerVariable"
	NodeTypeFieldDef             NodeType = "FieldDef"
	NodeTypeAnnotations          NodeType = "Annotations"
	NodeTypeTypeDef              NodeType = "TypeDef"
	NodeTypeContextStack         NodeType = "ContextStack"
	NodeTypeValidations          NodeType = "Validations"
	NodeTypeEnum                 NodeType = "Enum"
	NodeTypeDiscriminator        NodeType = "Discriminator"
	NodeTypeDiscriminatorMapping NodeType = "DiscriminatorMapping"
	NodeTypeExamples             NodeType = "Examples"
	NodeTypeOperation            NodeType = "Operation"
	NodeTypeRequest              NodeType = "Request"
	NodeTypeRequestParams        NodeType = "RequestParams"
	NodeTypeParam                NodeType = "Param"
	NodeTypeResponse             NodeType = "Response"
	NodeTypeSubResponse          NodeType = "SubResponse"
	NodeTypeResponseBodyContent  NodeType = "ResponseBodyContent"
	NodeTypeOperationExtensions  NodeType = "OperationExtensions"
	NodeTypeTypeDefExtensions    NodeType = "TypeDefExtensions"
	NodeTypeArguments            NodeType = "Arguments"
	NodeTypeComment              NodeType = "Comment"
	NodeTypeExtendedComment      NodeType = "ExtendedComment"
	NodeTypeExternalDocs         NodeType = "ExternalDocs"
)

func GetNodeType(n Node) NodeType {
	switch n.(type) {
	case *SDK:
		return NodeTypeSDK
	case *Servers:
		return NodeTypeServers
	case *Server:
		return NodeTypeServer
	case *ServerVariable:
		return NodeTypeServerVariable
	case *FieldDef:
		return NodeTypeFieldDef
	case Annotations:
		return NodeTypeAnnotations
	case *TypeDef:
		return NodeTypeTypeDef
	case ContextStack:
		return NodeTypeContextStack
	case *Validations:
		return NodeTypeValidations
	case *Enum:
		return NodeTypeEnum
	case *Discriminator:
		return NodeTypeDiscriminator
	case *DiscriminatorMapping:
		return NodeTypeDiscriminatorMapping
	case Examples:
		return NodeTypeExamples
	case *Operation:
		return NodeTypeOperation
	case *Request:
		return NodeTypeRequest
	case *RequestParams:
		return NodeTypeRequestParams
	case *Param:
		return NodeTypeParam
	case *Response:
		return NodeTypeResponse
	case *SubResponse:
		return NodeTypeSubResponse
	case *ResponseBodyContent:
		return NodeTypeResponseBodyContent
	case *OperationExtensions:
		return NodeTypeOperationExtensions
	case *TypeDefExtensions:
		return NodeTypeTypeDefExtensions
	case *Arguments:
		return NodeTypeArguments
	case *Comment:
		return NodeTypeComment
	case *ExtendedComment:
		return NodeTypeExtendedComment
	case *ExternalDocs:
		return NodeTypeExternalDocs
	default:
		panic(fmt.Errorf("unknown node type: %T", n))
	}
}

// VisitFn now takes a slice of parents.
type VisitFn func(Node, []Node, *AST) error

func (a *AST) Walk(visit VisitFn) error {
	visited := make(map[string]bool)

	if err := a.WalkSDK(a.MainSDK, []Node{}, visit, visited); err != nil && !errors.Is(err, ErrTerminateWalk) {
		return err
	}

	for _, ops := range a.Webhooks.All() {
		for i := range ops {
			if err := a.WalkOperation(&ops[i], []Node{}, visit, visited); err != nil {
				if errors.Is(err, ErrTerminateWalk) {
					return nil
				}
				return err
			}
		}
	}

	for _, v := range a.Components.All() {
		if err := a.WalkTypeDef(v, []Node{}, visit, visited); err != nil {
			if errors.Is(err, ErrTerminateWalk) {
				return nil
			}
			return err
		}
	}

	return nil
}

func (a *AST) WalkSDK(sdk *SDK, parents []Node, visit VisitFn, visited map[string]bool) error {
	if sdk == nil {
		return nil
	}

	pointerID := getPointerID(sdk, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(sdk, parents, a); err != nil {
		return err
	}

	newParents := append(parents, sdk)
	if err := a.WalkTypeDef(sdk.Type, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkServers(sdk.Servers, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkComment(sdk.Comments, newParents, visit, visited); err != nil {
		return err
	}

	for _, s := range sdk.SubSDKs {
		if err := a.WalkSDK(s, newParents, visit, visited); err != nil {
			return err
		}
	}

	if err := a.WalkFieldDef(sdk.Security, newParents, visit, visited); err != nil {
		return err
	}

	for _, o := range sdk.Operations {
		if err := a.WalkOperation(o, newParents, visit, visited); err != nil {
			return err
		}
	}

	for _, t := range sdk.AdditionalTypes {
		if err := a.WalkTypeDef(t, newParents, visit, visited); err != nil {
			return err
		}
	}

	return a.WalkTypeDef(sdk.Globals, newParents, visit, visited)
}

func (a *AST) WalkServers(servers *Servers, parents []Node, visit VisitFn, visited map[string]bool) error {
	if servers == nil {
		return nil
	}

	pointerID := getPointerID(servers, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(servers, parents, a); err != nil {
		return err
	}

	newParents := append(parents, servers)
	for _, server := range servers.Servers {
		if err := a.WalkServer(server, newParents, visit, visited); err != nil {
			return err
		}
	}

	return nil
}

func (a *AST) WalkServer(server *Server, parents []Node, visit VisitFn, visited map[string]bool) error {
	if server == nil {
		return nil
	}

	pointerID := getPointerID(server, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(server, parents, a); err != nil {
		return err
	}

	newParents := append(parents, server)
	if err := a.WalkComment(server.Comments, newParents, visit, visited); err != nil {
		return err
	}

	for _, v := range server.Variables {
		if err := a.WalkServerVariable(v, newParents, visit, visited); err != nil {
			return err
		}
	}

	return nil
}

func (a *AST) WalkServerVariable(variable *ServerVariable, parents []Node, visit VisitFn, visited map[string]bool) error {
	if variable == nil {
		return nil
	}

	pointerID := getPointerID(variable, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(variable, parents, a); err != nil {
		return err
	}

	newParents := append(parents, variable)
	return a.WalkTypeDef(variable.Type, newParents, visit, visited)
}

func (a *AST) WalkOperation(op *Operation, parents []Node, visit VisitFn, visited map[string]bool) error {
	if op == nil {
		return nil
	}

	pointerID := getPointerID(op, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(op, parents, a); err != nil {
		return err
	}

	newParents := append(parents, op)
	if err := a.WalkRequest(op.Request, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkResponse(op.Response, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkFieldDef(op.Security, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkFieldDef(op.GlobalSecurity, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkServers(op.Servers, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkComment(op.Comments, newParents, visit, visited); err != nil {
		return err
	}

	for _, c := range op.Callbacks {
		if err := a.WalkTypeDef(c, newParents, visit, visited); err != nil {
			return err
		}
	}

	if err := a.WalkSDK(op.OwningSDK, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkOperationExtensions(op.Extensions, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkTypeDef(op.Globals, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkArguments(op.Arguments, newParents, visit, visited); err != nil {
		return err
	}

	return nil
}

func (a *AST) WalkRequest(req *Request, parents []Node, visit VisitFn, visited map[string]bool) error {
	if req == nil {
		return nil
	}

	pointerID := getPointerID(req, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(req, parents, a); err != nil {
		return err
	}

	newParents := append(parents, req)
	if err := a.WalkFieldDef(req.Field, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkFieldDef(req.RequestBody, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkRequestParams(req.Params, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkExamples(req.Examples, newParents, visit, visited); err != nil {
		return err
	}

	return nil
}

func (a *AST) WalkRequestParams(params *RequestParams, parents []Node, visit VisitFn, visited map[string]bool) error {
	if params == nil {
		return nil
	}

	pointerID := getPointerID(params, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(params, parents, a); err != nil {
		return err
	}

	newParents := append(parents, params)
	for _, param := range params.QueryParams {
		if err := a.WalkParam(param, newParents, visit, visited); err != nil {
			return err
		}
	}

	for _, param := range params.PathParams {
		if err := a.WalkParam(param, newParents, visit, visited); err != nil {
			return err
		}
	}

	for _, param := range params.HeaderParams {
		if err := a.WalkParam(param, newParents, visit, visited); err != nil {
			return err
		}
	}

	return nil
}

func (a *AST) WalkParam(param *Param, parents []Node, visit VisitFn, visited map[string]bool) error {
	if param == nil {
		return nil
	}

	pointerID := getPointerID(param, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(param, parents, a); err != nil {
		return err
	}

	newParents := append(parents, param)
	if err := a.WalkFieldDef(param.Field, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkExamples(param.Examples, newParents, visit, visited); err != nil {
		return err
	}

	return nil
}

func (a *AST) WalkResponse(response *Response, parents []Node, visit VisitFn, visited map[string]bool) error {
	if response == nil {
		return nil
	}

	pointerID := getPointerID(response, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(response, parents, a); err != nil {
		return err
	}

	newParents := append(parents, response)
	if err := a.WalkTypeDef(response.Type, newParents, visit, visited); err != nil {
		return err
	}

	for _, resp := range response.Responses {
		if err := a.WalkSubResponse(resp, newParents, visit, visited); err != nil {
			return err
		}
	}

	return nil
}

func (a *AST) WalkSubResponse(resp *SubResponse, parents []Node, visit VisitFn, visited map[string]bool) error {
	if resp == nil {
		return nil
	}

	pointerID := getPointerID(resp, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(resp, parents, a); err != nil {
		return err
	}

	newParents := append(parents, resp)
	for _, content := range resp.Content {
		if err := a.WalkResponseBodyContent(content, newParents, visit, visited); err != nil {
			return err
		}
	}

	return nil
}

func (a *AST) WalkResponseBodyContent(content *ResponseBodyContent, parents []Node, visit VisitFn, visited map[string]bool) error {
	if content == nil {
		return nil
	}

	pointerID := getPointerID(content, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(content, parents, a); err != nil {
		return err
	}

	newParents := append(parents, content)
	if err := a.WalkFieldDef(content.Content, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkExamples(content.Examples, newParents, visit, visited); err != nil {
		return err
	}

	return nil
}

func (a *AST) WalkOperationExtensions(extensions *OperationExtensions, parents []Node, visit VisitFn, visited map[string]bool) error {
	if extensions == nil {
		return nil
	}

	pointerID := getPointerID(extensions, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	return visit(extensions, parents, a)
}

func (a *AST) WalkArguments(arguments *Arguments, parents []Node, visit VisitFn, visited map[string]bool) error {
	if arguments == nil {
		return nil
	}

	pointerID := getPointerID(arguments, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(arguments, parents, a); err != nil {
		return err
	}

	newParents := append(parents, arguments)
	for _, field := range arguments.Sorted {
		if err := a.WalkFieldDef(field, newParents, visit, visited); err != nil {
			return err
		}
	}

	for _, field := range arguments.ParamFields {
		if err := a.WalkFieldDef(field, newParents, visit, visited); err != nil {
			return err
		}
	}

	for _, field := range arguments.BodyFields {
		if err := a.WalkFieldDef(field, newParents, visit, visited); err != nil {
			return err
		}
	}

	if err := a.WalkFieldDef(arguments.BodyField, newParents, visit, visited); err != nil {
		return err
	}

	return nil
}

func (a *AST) WalkFieldDef(field *FieldDef, parents []Node, visit VisitFn, visited map[string]bool) error {
	if field == nil {
		return nil
	}

	pointerID := getPointerID(field, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(field, parents, a); err != nil {
		return err
	}

	newParents := append(parents, field)
	if err := a.WalkTypeDef(field.Type, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkComment(field.Comments, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkAnnotations(field.Annotations, newParents, visit, visited); err != nil {
		return err
	}

	return nil
}

func (a *AST) WalkAnnotations(annotations Annotations, parents []Node, visit VisitFn, visited map[string]bool) error {
	if annotations == nil {
		return nil
	}

	pointerID := getPointerID(annotations, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	return visit(annotations, parents, a)
}

func (a *AST) WalkTypeDef(t *TypeDef, parents []Node, visit VisitFn, visited map[string]bool) error {
	if t == nil {
		return nil
	}

	pointerID := getPointerID(t, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(t, parents, a); err != nil {
		return err
	}

	newParents := append(parents, t)
	if err := a.WalkContextStack(t.ContextStack, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkTypeDef(t.ItemType, newParents, visit, visited); err != nil {
		return err
	}

	for _, f := range t.Fields {
		if err := a.WalkFieldDef(f, newParents, visit, visited); err != nil {
			return err
		}
	}

	if err := a.WalkValidations(t.Validations, newParents, visit, visited); err != nil {
		return err
	}

	for _, at := range t.AssociatedTypes {
		if err := a.WalkTypeDef(at, newParents, visit, visited); err != nil {
			return err
		}
	}

	if err := a.WalkEnum(t.Enum, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkComment(t.Comments, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkExamples(t.Examples, newParents, visit, visited); err != nil {
		return err
	}

	if err := a.WalkDiscriminator(t.Discriminator, newParents, visit, visited); err != nil {
		return err
	}

	return nil
}

func (a *AST) WalkContextStack(stack ContextStack, parents []Node, visit VisitFn, visited map[string]bool) error {
	if stack == nil {
		return nil
	}

	pointerID := getPointerID(stack, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	return visit(stack, parents, a)
}

func (a *AST) WalkValidations(validations *Validations, parents []Node, visit VisitFn, visited map[string]bool) error {
	if validations == nil {
		return nil
	}

	pointerID := getPointerID(validations, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	return visit(validations, parents, a)
}

func (a *AST) WalkEnum(enum *Enum, parents []Node, visit VisitFn, visited map[string]bool) error {
	if enum == nil {
		return nil
	}

	pointerID := getPointerID(enum, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(enum, parents, a); err != nil {
		return err
	}

	newParents := append(parents, enum)
	if err := a.WalkTypeDef(enum.Type, newParents, visit, visited); err != nil {
		return err
	}

	return nil
}

func (a *AST) WalkExamples(examples Examples, parents []Node, visit VisitFn, visited map[string]bool) error {
	if examples == nil {
		return nil
	}

	pointerID := getPointerID(examples, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	return visit(examples, parents, a)
}

func (a *AST) WalkDiscriminator(discriminator *Discriminator, parents []Node, visit VisitFn, visited map[string]bool) error {
	if discriminator == nil {
		return nil
	}

	pointerID := getPointerID(discriminator, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(discriminator, parents, a); err != nil {
		return err
	}

	newParents := append(parents, discriminator)
	for _, dm := range discriminator.Mapping {
		if err := a.WalkDiscriminatorMapping(dm, newParents, visit, visited); err != nil {
			return err
		}
	}

	return nil
}

func (a *AST) WalkDiscriminatorMapping(dm *DiscriminatorMapping, parents []Node, visit VisitFn, visited map[string]bool) error {
	if dm == nil {
		return nil
	}

	pointerID := getPointerID(dm, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(dm, parents, a); err != nil {
		return err
	}

	newParents := append(parents, dm)
	if err := a.WalkTypeDef(dm.Type, newParents, visit, visited); err != nil {
		return err
	}

	return nil
}

func (a *AST) WalkComment(comment *Comment, parents []Node, visit VisitFn, visited map[string]bool) error {
	if comment == nil {
		return nil
	}

	pointerID := getPointerID(comment, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	if err := visit(comment, parents, a); err != nil {
		return err
	}

	newParents := append(parents, comment)
	for _, ec := range comment.ExtendedComments {
		if err := a.WalkExtendedComment(ec, newParents, visit, visited); err != nil {
			return err
		}
	}

	if err := a.WalkExternalDocs(comment.ExternalDocs, newParents, visit, visited); err != nil {
		return err
	}

	return nil
}

func (a *AST) WalkExtendedComment(ec *ExtendedComment, parents []Node, visit VisitFn, visited map[string]bool) error {
	if ec == nil {
		return nil
	}

	pointerID := getPointerID(ec, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	return visit(ec, parents, a)
}

func (a *AST) WalkExternalDocs(ed *ExternalDocs, parents []Node, visit VisitFn, visited map[string]bool) error {
	if ed == nil {
		return nil
	}

	pointerID := getPointerID(ed, parents)
	if visited[pointerID] {
		return nil
	}
	visited[pointerID] = true

	return visit(ed, parents, a)
}

func getPointerID(visiting Node, parents []Node) string {
	var parent Node = nil
	if len(parents) > 0 {
		parent = parents[len(parents)-1]
	}

	if parent == nil {
		return fmt.Sprintf("%d", getPointerAddress(visiting))
	}

	return fmt.Sprintf("%d-%d", getPointerAddress(visiting), getPointerAddress(parent))
}

func getPointerAddress(i interface{}) uintptr {
	v := reflect.ValueOf(i)
	if v.Kind() != reflect.Pointer && v.Kind() != reflect.Map && v.Kind() != reflect.Slice {
		panic("The interface does not contain a pointer")
	}
	return v.Pointer()
}
