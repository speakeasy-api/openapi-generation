package generate

import (
	"context"
	"fmt"
	"mime"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/contenttypes"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/openapi"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi-generation/v2/internal/schemas"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/jsonschema/oas3"
	oas "github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/references"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

type handleReqParams struct {
	Scope               ast.Scope
	RequestBody         *oas.ReferencedRequestBody
	Servers             *ast.Servers
	Retries             *extensions.Retries
	Timeout             *int64
	DocsRateLimits      []extensions.RateLimit
	Pagination          *extensions.Pagination
	OpID                string
	Op                  *oas.Operation
	PathItemParameters  []*oas.ReferencedParameter
	GlobalNameOverrides []*extensions.NameOverride
	Globals             *ast.TypeDef
	SecuritySchemes     *sequencedmap.Map[string, *oas.ReferencedSecurityScheme]
	Errors              *extensions.Errors
	DocInfo             *document.DocumentInfo
}

type request struct {
	UsesUserAgentHeader bool
	Request             *ast.Request
	GlobalParams        ast.Fields
}

type requestBody struct {
	field               *ast.FieldDef
	contentType         string
	matchedContentTypes []string
	examples            []*ast.Example
}

func (g *Generator) handleRequests(ctx context.Context, contextStack ast.ContextStack, params handleReqParams) ([]request, error) {
	reqBodyFields, err := g.handleRequestBody(ctx, append(contextStack, ast.ContextFrame{
		Type:       ast.ContextTypeRequestResponse,
		Identifier: ast.InternalTypeRequest,
	}), params)
	if err != nil {
		return nil, err
	}
	// If there are no request body fields, we still need to check if there are any other request options
	if len(reqBodyFields) == 0 {
		reqBodyFields = []requestBody{{}}
	}

	requests := []request{}

	for _, reqBody := range reqBodyFields {
		var id string

		// Create local copies of the stack
		reqBodyContextStack := make(ast.ContextStack, len(contextStack))
		copy(reqBodyContextStack, contextStack)

		for _, frame := range reqBodyContextStack {
			if frame.Type == ast.ContextTypeOperation {
				id = frame.Identifier
				if len(reqBodyFields) > 1 && reqBody.field != nil && reqBody.field.SerializationMethod != nil {
					if !g.subsystem.Config.Generation.NameResolutionAtLeastShortest() || *reqBody.field.SerializationMethod != ast.SerializationMethodJSON {
						// Only prefix if it's not a JSON request body
						id = fmt.Sprintf("%s_%s", id, *reqBody.field.SerializationMethod)
					}
				}

				reqBodyContextStack.UpdateOperation(id)
				break
			}
		}

		if reqBody.field != nil {
			// We need to update the operation context in any types already generated if they need to be scoped by serialization method
			for _, t := range reqBody.field.Type.Walk() {
				if t.ContextStack.HasFrameOfType(ast.ContextTypeOperation) {
					g.subsystem.Register.UnregisterType(t)
					t.ContextStack.UpdateOperation(id)
					g.subsystem.Register.RegisterType(ctx, t, true) // TODO: might have to update the pointer to the type def
				}
			}
		}

		opParams, opGlobalParams, usesUserAgentHeader, err := g.handleParameters(ctx, handleParametersOptions{
			contextStack:        reqBodyContextStack,
			itemParams:          params.PathItemParameters,
			paramsOAS:           params.Op.Parameters,
			globalNameOverrides: params.GlobalNameOverrides,
			scope:               params.Scope,
			globals:             params.Globals,
			opID:                params.OpID,
			docInfo:             params.DocInfo,
		})
		if err != nil {
			return nil, err
		}

		if opParams == nil && opGlobalParams == nil && reqBody.field == nil && params.Servers == nil && params.Retries == nil {
			return nil, nil
		}

		copyReqParams := &ast.RequestParams{}
		fields := ast.Fields{}

		if opParams != nil {
			if opParams.PathParams != nil {
				for _, param := range opParams.PathParams {
					if !param.Hidden {
						fields = append(fields, param.Field)
					}
				}

				copyReqParams.PathParams = make([]*ast.Param, len(opParams.PathParams))
				copy(copyReqParams.PathParams, opParams.PathParams)
			}
			if opParams.QueryParams != nil {
				for _, param := range opParams.QueryParams {
					if !param.Hidden {
						fields = append(fields, param.Field)
					}
				}

				copyReqParams.QueryParams = make([]*ast.Param, len(opParams.QueryParams))
				copy(copyReqParams.QueryParams, opParams.QueryParams)
			}
			if opParams.HeaderParams != nil {
				for _, param := range opParams.HeaderParams {
					if !param.Hidden {
						fields = append(fields, param.Field)
					}
				}

				copyReqParams.HeaderParams = make([]*ast.Param, len(opParams.HeaderParams))
				copy(copyReqParams.HeaderParams, opParams.HeaderParams)
			}
		}

		if reqBody.field != nil {
			// Check if we should preserve field names without sanitization
			preserveFieldNames := utils.ToBool(g.subsystem.Config.GetLanguageConfigValue("preserveModelFieldNames"))
			fields = fields.MustAddFieldWithOptions(reqBody.field, ast.FieldAddOptions{
				MaintainOriginalOrder: g.subsystem.Config.MaintainOpenAPIOrder(),
				Sanitize:              !preserveFieldNames,
			})
		}

		if copyReqParams.IsEmpty() {
			copyReqParams = nil
		}

		requestRequired := false
		for _, fields := range fields {
			if !fields.Optional {
				requestRequired = true
				break
			}
		}

		if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureMethodArguments) {
			val := g.subsystem.Config.GetLanguageConfigValue("methodArguments")
			if val == nil || val == "" || val == "require-security-and-request" {
				requestRequired = true
			}
		}

		req := &ast.Request{
			Params: copyReqParams,
			Field: &ast.FieldDef{
				Name:        "request",
				Annotations: ast.Annotations{&ast.RequestWrapperAnnotation{}, &ast.NeedsCasingAnnotation{}},
				Type: g.subsystem.Register.RegisterType(ctx, ast.NewType(&ast.TypeDef{
					Name:  id + "_" + ast.InternalTypeRequest,
					Type:  ast.DataTypeClass,
					Scope: params.Scope,
				}, ast.ContextStack{
					ast.ContextFrame{
						Type:       ast.ContextTypeRequestResponse,
						Identifier: ast.InternalTypeRequest,
					},
				}), true),
				Optional: !requestRequired,
			},
			RequestBody:         reqBody.field,
			MatchedContentTypes: reqBody.matchedContentTypes,
			Examples:            reqBody.examples,
		}

		requestBody, err := resolution.Resolve(ctx, params.RequestBody, params.DocInfo)
		if err != nil {
			return nil, err
		}
		if req.RequestBody != nil && requestBody.Required != nil {
			req.IsRequestBodyRequired = *requestBody.Required
		}

		req.Field.Type.Fields = g.flattenRequest(fields)

		if len(req.Field.Type.Fields) == 0 && req.RequestBody != nil {
			if g.subsystem.Config.Generation.NameResolutionAtLeastShortest() {
				// Avoid conflict with the request body name
				g.subsystem.Register.UnregisterType(req.Field.Type)
			}
			req.Field = req.RequestBody
			req.IsRequestBody = true
			req.Field.Name = "request" // Reset to a default name this doesn't get used for a fully flattened request at serialization time but will provide less confusion in generated code
			req.Field.OriginalName = ""
			if g.subsystem.Config.Generation.NameResolutionAtLeastShortest() && req.Field.Type.Name == ast.HumanizedRequestBody {
				g.subsystem.Register.UnregisterType(req.Field.Type)
				// Update the request body to be called just "request" since it's flat
				req.Field.Type.Name = ast.InternalTypeRequest
				req.Field.Type.ContextStack.MarkUsed(ast.ContextTypeRequestResponse)
				req.Field.Type.ContextStack.Update(ast.ContextTypeRequestBody, "requestBody", "body")
				req.Field.Type.ContextStack.MarkUnused(ast.ContextTypeRequestBody)
				g.subsystem.Register.RegisterType(ctx, req.Field.Type, true)
			}
		}

		if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureMethodArguments) {
			val := g.subsystem.Config.GetLanguageConfigValue("methodArguments")

			if val == "infer-optional-args" {
				g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureMethodArguments)

				// We must respect the requiredness of the request body, ie the `allFieldsOptional` only applies to the request wrapper
				if req.Field.Type.IsTypeWithFields() && !req.IsRequestBody {
					allFieldsOptional := true

					for _, field := range req.Field.Type.Fields {
						if !field.Optional {
							allFieldsOptional = false
						}
					}

					if allFieldsOptional {
						req.Field.Optional = true
					}
				}
			}
		}

		requests = append(requests, request{
			UsesUserAgentHeader: usesUserAgentHeader,
			Request:             req,
			GlobalParams:        opGlobalParams,
		})
	}

	if len(requests) == 0 {
		return nil, nil
	}

	return requests, nil
}

// If the request body is the only parameter (no query, header, or path params), it returns nil,
func (g *Generator) flattenRequest(fields ast.Fields) ast.Fields {
	flattenedFields := ast.Fields{}

	for _, field := range fields {
		if field.Annotations.Has(ast.AnnotationTypeRequest) {
			// request body is the only field
			if len(fields) == 1 {
				return nil
			}

			bodyFieldDefaultName := ast.HumanizedRequestBody
			bodyFieldNameIfComponent := field.Type.Name

			if g.subsystem.Config.Generation.RequestBodyFieldName != "" {
				bodyFieldDefaultName = g.subsystem.Config.Generation.RequestBodyFieldName
				bodyFieldNameIfComponent = g.subsystem.Config.Generation.RequestBodyFieldName
			}
			field.Name = bodyFieldDefaultName
			if field.Type.IsComponent && field.Type.Name != "" {
				field.Name = bodyFieldNameIfComponent
			}

			flattenedFields = g.addField(flattenedFields, field)
		} else {
			flattenedFields = g.addField(flattenedFields, field)
		}
	}

	if g.subsystem.Config.Generation.Fixes.ParameterOrderingFeb2024 {
		slices.SortStableFunc(flattenedFields, func(i, j *ast.FieldDef) int {
			// sort by parameter index ascending but if parameter index is nil leave it in the same order at the end

			if i.ParameterIndex == nil && j.ParameterIndex == nil {
				return 0
			}

			if i.ParameterIndex == nil {
				return 1
			}

			if j.ParameterIndex == nil {
				return -1
			}

			if *i.ParameterIndex < *j.ParameterIndex {
				return -1
			}

			if *i.ParameterIndex > *j.ParameterIndex {
				return 1
			}

			return 0
		})
	}

	return flattenedFields
}

func (g *Generator) addField(fields ast.Fields, field *ast.FieldDef) ast.Fields {
	if field == nil {
		panic("field is nil")
	}
	if field.Name == "" {
		panic("field name is empty")
	}

	// Check if we should preserve field names without sanitization
	preserveFieldNames := utils.ToBool(g.subsystem.Config.GetLanguageConfigValue("preserveModelFieldNames"))
	sanitize := !preserveFieldNames

	// Picked 1000 as a random number to avoid infinite looping
	totalIterationsAllowed := 1000
	for i := 0; i < totalIterationsAllowed; i++ {
		exists, idx := ast.FieldExists(fields, field, sanitize)

		if exists {
			existingField := fields[idx]
			existingParamAnn := existingField.Annotations.Get(ast.AnnotationTypeParam)
			if existingParamAnn != nil {
				existingField.Name = utils.SuffixParamType(existingField.Name, existingParamAnn.(*ast.ParamAnnotation).ParamType)
				fields[idx] = existingField
			}

			paramAnn := field.Annotations.Get(ast.AnnotationTypeParam)
			if paramAnn != nil {
				// Gotcha: For path and query params we add path and query suffix respectively
				// For header params we dont add any suffix. The reason we are not fixing it is because
				// it will be a breaking change for our clients with no additional value for them.
				field.Name = utils.SuffixParamType(field.Name, paramAnn.(*ast.ParamAnnotation).ParamType)
			} else {
				field.Name = utils.IncrementName(field.Name)
			}

			// We compare field names to check if disambiguation worked
			// We increment the name if the existingField.name & field.name are the same because
			// we try to disambiguate the name by adding a suffix to the field name. The suffix is the value from the `in` field.
			// If adding the `in` field does not disambiguate the name then the only way to make the name unique is to increment the name.
			var namesMatch bool
			if sanitize {
				namesMatch = ast.SanitizeFieldName(existingField.Name) == ast.SanitizeFieldName(field.Name)
			} else {
				namesMatch = existingField.Name == field.Name
			}

			if namesMatch {
				field.Name = utils.IncrementName(field.Name)
			}
			if i == totalIterationsAllowed-1 {
				panic(fmt.Sprintf("\n 2 request fields have the same name.\n The collision is happening between %+v and %+v", field.Type.Location, existingField.Type.Location))
			}
		} else {
			break
		}
	}

	ff, err := fields.AddFieldWithOptions(field, ast.FieldAddOptions{
		MaintainOriginalOrder: g.subsystem.Config.MaintainOpenAPIOrder(),
		Sanitize:              sanitize,
	})
	if err != nil {
		panic(err)
	}

	return ff
}

func (g *Generator) handleRequestBody(ctx context.Context, contextStack ast.ContextStack, params handleReqParams) ([]requestBody, error) {
	reqBody, err := resolution.Resolve(ctx, params.RequestBody, params.DocInfo)
	if err != nil {
		return nil, err
	}
	if reqBody == nil || reqBody.Content.Len() == 0 {
		return nil, nil
	}

	requestBodies := sequencedmap.New[ast.SerializationMethod, requestBody]()

	for contentType, typeObj := range reqBody.Content.AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
		ignore, err := g.subsystem.Extensions.Ignore(typeObj.GetExtensions())
		if err != nil {
			return nil, err
		}
		if ignore {
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureIgnores)
			continue
		}

		handleEncoding := false

		mediaType, _, _ := mime.ParseMediaType(contentType)

		childContextStack := contextStack

		if reqBody.Content.Len() > 1 {
			childContextStack.AppendRequestMediaType(mediaType, g.subsystem.Config.Generation.GetNameResolution())
		}

		childContextStack.AppendRequestBody()

		serializationMethod := ast.SerializationMethodRAW

		var strType *ast.TypeDef

		schema := typeObj.Schema

		var resolvedSchema *oas3.JSONSchema[oas3.Concrete]
		if schema != nil {
			resolvedSchema, err = resolution.Resolve(ctx, schema, params.DocInfo)
			if err != nil {
				return nil, err
			}
		}

		var parentComponentRef references.Reference
		parentComponentDescription := ""
		// If this is a component request and only has one potential inline schema lets use this component name if possible
		if params.RequestBody.IsReference() && reqBody.Content.Len() == 1 {
			parentComponentRef = params.RequestBody.GetReference()
			parentComponentDescription = params.RequestBody.GetDescription() // Reference level description takes precedence
			if parentComponentDescription == "" {
				parentComponentDescription = reqBody.GetDescription()
			}
		}

		if schema != nil {
			typ, _, _ := openapi.GetResolvedType(ctx, resolvedSchema)
			if typ == "string" {
				f, err := g.schemas.HandleSchema(ctx, schemas.Params{
					ContextStack:               childContextStack,
					Schema:                     schema,
					Scope:                      params.Scope,
					ParentComponentRef:         parentComponentRef,
					ParentComponentDescription: parentComponentDescription,
					DocInfo:                    params.DocInfo,
				})
				if err != nil {
					return nil, err
				}
				strType = f.Type
			}

			// If this schema is an inline schema within a requestBody reference we can use the requestBody reference to
			// create a shared type instead of an inline type
			if !schema.IsReference() && params.RequestBody.IsReference() && reqBody.Content.Len() == 1 {
				schema = oas3.NewReferencedScheme(ctx, *params.RequestBody.Reference, resolvedSchema)
			}
		}

		node := reqBody.GetCore().Content.GetMapKeyNodeOrRoot(contentType, reqBody.GetRootNode())

		isString := strType != nil && strType.Type == ast.DataTypeString

		switch {
		case contenttypes.IsJSON(contentType):
			serializationMethod = ast.SerializationMethodJSON
		case contenttypes.IsMultipart(contentType) && !isString:
			serializationMethod = ast.SerializationMethodMultipart
			handleEncoding = true
		case contenttypes.IsURLEncoded(contentType) && !isString:
			serializationMethod = ast.SerializationMethodForm
			handleEncoding = true
		case contenttypes.IsXML(contentType) && !isString:
			logging.LogWarning(ctx, "unsupported", errors.NewUnsupportedError("xml request bodies are not currently supported. Handling as bytes", node))
		case contenttypes.IsCSV(contentType) && !isString:
			logging.LogWarning(ctx, "unsupported", errors.NewUnsupportedError("csv request bodies are not currently supported. Handling as bytes", node))
		case contenttypes.IsYAML(contentType) && !isString:
			logging.LogWarning(ctx, "unsupported", errors.NewUnsupportedError("yaml request bodies are not currently supported. Handling as bytes", node))
		default:
			if isString {
				serializationMethod = ast.SerializationMethodString
			}
			// TODO: maybe handle text base types such as text/html, text/plain, text/javascript, etc as strings
			// TODO: capture metrics for unsupported media types
		}

		if !handleEncoding && typeObj.GetEncoding().Len() > 0 {
			// TODO this may actually be a validation issue
			return nil, errors.NewUnsupportedError("request body with encoding not currently supported for non application/x-www-form-urlencoded content types", typeObj.GetCore().Encoding.GetKeyNodeOrRoot(typeObj.GetRootNode()))
		}

		var reqField *ast.FieldDef

		if schema != nil {
			var err error
			reqField, err = g.schemas.HandleSchema(ctx, schemas.Params{
				ContextStack:               childContextStack,
				SerializationMethod:        serializationMethod,
				Schema:                     schema,
				Encoding:                   typeObj.Encoding,
				Scope:                      params.Scope,
				IsRequest:                  true,
				Depth:                      1,
				ParentComponentRef:         parentComponentRef,
				ParentComponentDescription: parentComponentDescription,
				DocInfo:                    params.DocInfo,
			})
			if err != nil {
				return nil, err
			}

			reqField.Type.IsInlineRequestBody = reqField.Type.Scope != ast.ScopeShared

			// If the request body was an inline type make sure it always gets prefixed with the operation name
			if reqField.Type.Scope != ast.ScopeShared && reqField.Type.ContextStack.HasFrameOfType(ast.ContextTypeOperation) {
				opFrame := reqField.Type.ContextStack.FindLastFrameOfType(ast.ContextTypeOperation)
				opFrame.MustUse = true
			}
		}

		if reqField != nil && serializationMethod != ast.SerializationMethodRAW {
			if schema != nil {
				if serializationMethod == ast.SerializationMethodForm {
					node := resolvedSchema.GetRootNode()
					if resolvedSchema.IsSchema() {
						node = resolvedSchema.GetSchema().GetCore().Type.GetKeyNodeOrRoot(resolvedSchema.GetSchema().GetRootNode())
					}

					if !reqField.Type.IsObjectType() {
						return nil, errors.NewUnsupportedError("application/x-www-form-urlencoded request body not currently supported for non object types", node)
					}

					if reqField.Type.IsContainer() && !reqField.Type.ItemType.IsPrimitive() {
						return nil, errors.NewUnsupportedError("application/x-www-form-urlencoded request body not currently supported for complex map types", node)
					}
				}
			}

			reqField.Name = "Request"
			reqField.Optional = reqBody.Required == nil || !*reqBody.Required
			reqField.SerializationMethod = &serializationMethod
			reqField.Annotations = append(reqField.Annotations, &ast.RequestAnnotation{MediaType: mediaType}, &ast.NeedsCasingAnnotation{})

			replace := true

			if existingRequestBody, ok := requestBodies.Get(serializationMethod); ok {
				existingMediaType := ""

				anno := existingRequestBody.field.Annotations.Get(ast.AnnotationTypeRequest)
				if anno != nil {
					reqAnno, ok := anno.(*ast.RequestAnnotation)
					if ok {
						existingMediaType = reqAnno.MediaType
					}
				}

				// Trying to retain the more common media type
				if existingMediaType != "" {
					switch *reqField.SerializationMethod {
					case ast.SerializationMethodJSON:
						replace = existingMediaType != "application/json" && !strings.HasPrefix(existingMediaType, "application/json")
					case ast.SerializationMethodForm:
						replace = existingMediaType != "application/x-www-form-urlencoded" && !strings.HasPrefix(existingMediaType, "application/x-www-form-urlencoded")
					case ast.SerializationMethodMultipart:
						replace = existingMediaType != "multipart/form-data" && !strings.HasPrefix(existingMediaType, "multipart/form-data")
					case ast.SerializationMethodString:
						replace = existingMediaType != "text/plain" && !strings.HasPrefix(existingMediaType, "text/plain")
					}
				}
			}

			examples := []*ast.Example{}
			if typeObj.GetExamples().Len() > 0 {
				for name, ex := range typeObj.GetExamples().All() {
					example, err := resolution.Resolve(ctx, ex, params.DocInfo)
					if err != nil {
						return nil, err
					}

					examples = append(examples, ast.NewExample(name, example.GetDescription(), resolveExampleValue(example)))
				}
			} else if typeObj.Example != nil {
				examples = append(examples, ast.NewExample("", "", typeObj.Example))
			}

			matchedContentTypes := []string{}
			current, ok := requestBodies.Get(serializationMethod)
			if ok {
				matchedContentTypes = current.matchedContentTypes
			}

			matchedContentTypes = append(matchedContentTypes, contentType)

			if replace {
				requestBodies.Set(serializationMethod, requestBody{
					field:               reqField,
					contentType:         contentType,
					matchedContentTypes: matchedContentTypes,
					examples:            examples,
				})
			} else {
				current.matchedContentTypes = matchedContentTypes
				requestBodies.Set(serializationMethod, current)
			}
		} else {
			dataType := ast.DataTypeBytes
			if g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureUploadStreams) {
				g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureUploadStreams)
				dataType = ast.DataTypeRequestStream
			}

			var typ ast.TypeDef
			if reqField != nil {
				typ = *reqField.Type
				typ.Type = dataType
			} else {
				typ = *ast.NewType(&ast.TypeDef{Type: dataType}, nil)
			}

			requestBodies.Set(serializationMethod, requestBody{
				field: &ast.FieldDef{
					Name: "Request",
					Type: &typ,
					Annotations: []ast.Annotation{
						&ast.RequestAnnotation{MediaType: mediaType},
						&ast.NeedsCasingAnnotation{},
					},
					Optional:            reqBody.Required == nil || !*reqBody.Required,
					SerializationMethod: &serializationMethod,
				},
				contentType:         contentType,
				matchedContentTypes: []string{contentType},
			})
		}

		// Add request description to field and type if missing from schema
		if reqB, ok := requestBodies.Get(serializationMethod); ok {
			if reqBody.GetDescription() != "" {
				if reqB.field.Comments == nil || reqB.field.Comments.Description == "" {
					if reqB.field.Comments == nil {
						reqB.field.Comments = &ast.Comment{}
					}

					reqB.field.Comments.Description = reqBody.GetDescription()
				}

				if (reqB.field.Type.Comments == nil || reqB.field.Type.Comments.Description == "") && reqB.field.Type.Scope != ast.ScopeShared {
					if reqB.field.Type.Comments == nil {
						reqB.field.Type.Comments = &ast.Comment{}
					}

					reqB.field.Type.Comments.Description = reqBody.GetDescription()
				}
			}
		}
	}

	if requestBodies.Len() == 0 {
		return nil, nil
	}

	sameTypes := true
	var currentType *ast.TypeDef

	for _, reqB := range requestBodies.All() {
		if currentType == nil {
			currentType = reqB.field.Type
			continue
		}

		err := currentType.IsEqual(reqB.field.Type, ast.SkipAnnotations())
		sameTypes = sameTypes && err == nil
	}

	if !sameTypes {
		var ret []requestBody
		for _, reqB := range requestBodies.All() {
			ret = append(ret, reqB)
		}
		return ret, nil
	}

	matchedContentTypes := []string{}
	examples := []*ast.Example{}

	for _, reqBody := range requestBodies.All() {
		matchedContentTypes = append(matchedContentTypes, reqBody.matchedContentTypes...)
		examples = append(examples, reqBody.examples...)
	}

	if jsonBody, ok := requestBodies.Get(ast.SerializationMethodJSON); ok {
		jsonBody.matchedContentTypes = matchedContentTypes
		jsonBody.examples = examples
		return []requestBody{jsonBody}, nil
	}
	if multipartBody, ok := requestBodies.Get(ast.SerializationMethodMultipart); ok {
		multipartBody.matchedContentTypes = matchedContentTypes
		multipartBody.examples = examples
		return []requestBody{multipartBody}, nil
	}
	if formBody, ok := requestBodies.Get(ast.SerializationMethodForm); ok {
		formBody.matchedContentTypes = matchedContentTypes
		formBody.examples = examples
		return []requestBody{formBody}, nil
	}
	if stringBody, ok := requestBodies.Get(ast.SerializationMethodString); ok {
		stringBody.matchedContentTypes = matchedContentTypes
		stringBody.examples = examples
		return []requestBody{stringBody}, nil
	}
	if rawBody, ok := requestBodies.Get(ast.SerializationMethodRAW); ok {
		rawBody.matchedContentTypes = matchedContentTypes
		rawBody.examples = examples
		return []requestBody{rawBody}, nil
	}

	return nil, nil
}
