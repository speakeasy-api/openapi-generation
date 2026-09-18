package generate

import (
	"cmp"
	"context"
	"fmt"
	"mime"
	"slices"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/contenttypes"
	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/licensing"
	"github.com/speakeasy-api/openapi-generation/v2/internal/namer"
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

const (
	contentTypeDesc = "HTTP response content type for this operation"
	statusCodeDesc  = "HTTP response status code for this operation"
	rawResponseDesc = "Raw HTTP response; suitable for custom response parsing"
	rawRequestDesc  = "Raw HTTP request; suitable for debugging"
)

const (
	responseFormatEnvelope     = "envelope"
	responseFormatEnvelopeHTTP = "envelope-http"
	responseFormatFlat         = "flat"
)

type responseContent struct {
	statusToContentType map[string]map[string]ast.SerializationMethod
	responseField       *ast.FieldDef
	usageExample        map[string]map[string]bool
	examples            []*ast.Example
	sseSentinel         string
	mediaType           *oas.MediaType
}

func (g *Generator) getResponseFormat(ctx context.Context) (string, error) {
	if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureResponseFormat) {
		return responseFormatEnvelope, nil
	}
	g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureResponseFormat)

	// Use GetLanguageConfigValue to respect config overlays (e.g., CLI's getConfigOverlay()).
	raw := g.subsystem.Config.GetLanguageConfigValue("responseFormat")
	if raw == nil {
		return responseFormatEnvelope, nil
	}

	val, ok := raw.(string)
	if !ok {
		return responseFormatEnvelope, nil
	}

	switch val {
	case responseFormatEnvelope:
		fallthrough
	case responseFormatEnvelopeHTTP:
		fallthrough
	case responseFormatFlat:
		return val, nil
	default:
		return "", fmt.Errorf("invalid responseFormat: %s", val)
	}
}

type handleResponsesParams struct {
	Scope      ast.Scope
	Pagination *extensions.Pagination
	ErrsExt    *extensions.Errors
	DocInfo    *document.DocumentInfo
}

func (g *Generator) handleResponses(
	ctx context.Context,
	responses oas.Responses,
	contextStack ast.ContextStack,
	params handleResponsesParams,
) (*ast.Response, error) {
	id := contextStack.FindLastFrameOfType(ast.ContextTypeOperation).Identifier
	if g.subsystem.Config.Generation.NameResolutionAtLeastShortest() {
		id = contextStack.FindLastFrameOfType(ast.ContextTypeOperation).DisplayName()
	}

	childContextStack := append(contextStack,
		ast.ContextFrame{
			Type:       ast.ContextTypeRequestResponse,
			Identifier: ast.InternalTypeResponse,
		})

	resFormat, err := g.getResponseFormat(ctx)
	if err != nil {
		return nil, err
	}

	_ = g.subsystem.Register.CloseTypes(nil)

	httpMetadataType := g.generateHTTPMetadataType(ctx)

	responseObj := ast.NewType(&ast.TypeDef{
		Name:             id + "_" + ast.InternalTypeResponse,
		Type:             ast.DataTypeClass,
		Fields:           ast.Fields{},
		Scope:            params.Scope,
		ResponseEnvelope: true,
	}, ast.ContextStack{
		ast.ContextFrame{
			Type:       ast.ContextTypeRequestResponse,
			Identifier: ast.InternalTypeResponse,
		},
	})

	responseFieldMapping := sequencedmap.New[string, responseContent]()
	responseFieldOriginalNames := map[string]string{}

	headersMap := map[string]bool{}
	emptyCodes := map[string]bool{}
	foundCodes := map[string]bool{}

	if responses.Default != nil {
		responses.Init()
		responses.Set("default", responses.Default)
	}

	errsExt := params.ErrsExt
	if errsExt == nil && utils.ToBool(g.subsystem.Config.GetLanguageConfigValue("clientServerStatusCodesAsErrors")) {
		errsExt = &extensions.Errors{
			StatusCodes: []string{
				"4XX",
				"5XX",
			},
		}
	}
	if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureErrors) {
		errsExt = nil
	}

	isErrorStatusCode := func(statusCode string) bool {
		return errsExt != nil && errsExt.IsErrorStatusCode(statusCode)
	}

	for statusCode, r := range responses.AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
		resp, err := resolution.Resolve(ctx, r, params.DocInfo)
		if err != nil {
			return nil, err
		}

		ignore, err := g.subsystem.Extensions.Ignore(resp.GetExtensions())
		if err != nil {
			return nil, err
		}
		if ignore {
			g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureIgnores)
			continue
		}

		// Handle empty status code (not technically valid in OpenAPI but we'll default it)
		if statusCode == "" {
			statusCode = "default"
		}

		if resp.GetHeaders().Len() > 0 {
			headersMap[statusCode] = true
		}

		foundCodes[statusCode] = true

		contentAdded := false

		for contentType, typeObj := range resp.GetContent().AllOrdered(g.subsystem.Config.GetSequencedMapIterationOrder()) {
			if typeObj.Schema == nil && typeObj.ItemSchema == nil {
				continue
			}

			ignore, err := g.subsystem.Extensions.Ignore(typeObj.GetExtensions())
			if err != nil {
				return nil, err
			}
			if ignore {
				g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureIgnores)
				continue
			}

			usageExample, err := g.subsystem.Extensions.IsUsageExample(typeObj.GetExtensions())
			if err != nil {
				return nil, err
			}
			if usageExample {
				g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureExamples)
			}

			mediaType, _, _ := mime.ParseMediaType(contentType)

			contentAdded = true

			var strType *ast.TypeDef

			schema := typeObj.GetSchema()

			// OpenAPI 3.2: For sequential media types, itemSchema describes the per-item type.
			if itemSchema := typeObj.GetItemSchema(); itemSchema != nil {
				if contenttypes.IsJsonL(contentType) || contenttypes.IsJsonSeq(contentType) {
					// For JSONL/json-seq, itemSchema directly replaces schema (no envelope).
					schema = itemSchema
				} else if contenttypes.IsEventStream(contentType) && schema == nil {
					// For SSE, itemSchema describes the data payload. Construct a synthetic
					// SSE envelope schema with data → itemSchema plus standard SSE fields.
					schema = buildSSEEnvelopeSchema(itemSchema)
				}
			}

			if schema == nil {
				continue
			}

			resolvedSchema, err := resolution.Resolve(ctx, schema, params.DocInfo)
			if err != nil {
				return nil, err
			}

			contentContextStack := childContextStack
			if isErrorStatusCode(statusCode) {
				contentContextStack.AppendWithHumanized(ast.ContextTypeResponseError, statusCode, "")
			}

			typ, _, _ := openapi.GetResolvedType(ctx, resolvedSchema)

			contentContextStack.AppendResponseStatusCode(statusCode)
			contentContextStack.AppendResponseMediaType(mediaType, g.subsystem.Config.Generation.GetNameResolution())
			contentContextStack.AppendResponseBody()

			var parentComponentRef references.Reference
			parentComponentDescription := ""
			// If this is a component response and only has one potential inline schema lets use this component name if possible
			if r.IsReference() && resp.GetContent().Len() == 1 {
				parentComponentRef = r.GetReference()
				parentComponentDescription = resp.GetDescription()
			}

			if typ == "string" {
				t, err := g.schemas.HandleSchema(ctx, schemas.Params{
					ContextStack:               contentContextStack,
					Schema:                     schema,
					Scope:                      params.Scope,
					ParentComponentRef:         parentComponentRef,
					ParentComponentDescription: parentComponentDescription,
					DocInfo:                    params.DocInfo,
				})
				if err != nil {
					return nil, err
				}
				strType = t.Type
			}

			node := resp.GetCore().Content.GetMapKeyNodeOrRoot(contentType, resp.GetRootNode())

			serializationMethod := ast.SerializationMethodRAW
			typeBasedSerialization := false

			if strType != nil {
				switch strType.Type {
				case ast.DataTypeString:
					if !contenttypes.IsJSON(contentType) {
						serializationMethod = ast.SerializationMethodString
						typeBasedSerialization = true
					}
				case ast.DataTypeBytes:
					serializationMethod = ast.SerializationMethodRAW
					typeBasedSerialization = true
				case ast.DataTypeResponseStream:
					serializationMethod = ast.SerializationMethodRAW
					typeBasedSerialization = true
				}
			}

			var warnErr error
			// Might want to handle other base types in the future
			if !typeBasedSerialization {
				switch {
				case contenttypes.IsTextPlain(contentType):
					serializationMethod = ast.SerializationMethodString
				case contenttypes.IsJSON(contentType):
					serializationMethod = ast.SerializationMethodJSON
				case contenttypes.IsEventStream(contentType):
					if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureServerEvents) {
						return nil, errors.NewUnsupportedError("server sent events are not currently supported", node)
					}
					accessErr := licensing.ValidateAccountHasFeatureAccess(ctx, features.FeatureServerEvents)
					if accessErr != nil {
						return nil, errors.NewUnsupportedError(accessErr.Error(), node)
					}
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureServerEvents)
					serializationMethod = ast.SerializationMethodEventStream
				case contenttypes.IsJsonL(contentType) || contenttypes.IsJsonSeq(contentType):
					if !g.subsystem.Features.IsFeatureSupported(ctx, features.FeatureJsonlResponses) {
						return nil, errors.NewUnsupportedError("JSONL/JSON Sequence responses are not currently supported", node)
					}
					accessErr := licensing.ValidateAccountHasFeatureAccess(ctx, features.FeatureJsonlResponses)
					if accessErr != nil {
						return nil, errors.NewUnsupportedError(accessErr.Error(), node)
					}
					g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureJsonlResponses)
					serializationMethod = ast.SerializationMethodJsonL
				case contenttypes.IsXML(contentType):
					warnErr = errors.NewUnsupportedError("xml responses are not currently supported. Handling as bytes", node)
				case contenttypes.IsCSV(contentType):
					warnErr = errors.NewUnsupportedError("csv responses are not currently supported. Handling as bytes", node)
				case contenttypes.IsYAML(contentType):
					warnErr = errors.NewUnsupportedError("yaml responses are not currently supported. Handling as bytes", node)
				default:
					// TODO: maybe handle text base types such as text/html, text/plain, text/javascript, etc as strings
					// TODO: capture metrics for unsupported media types
				}
			}

			var resField *ast.FieldDef

			if serializationMethod != ast.SerializationMethodString {
				// Handle complex types with schema processing
				resField, err = g.schemas.HandleSchema(ctx, schemas.Params{
					ContextStack:               contentContextStack,
					SerializationMethod:        serializationMethod,
					Schema:                     schema,
					Scope:                      params.Scope,
					IsRequest:                  false,
					Depth:                      1,
					ParentComponentRef:         parentComponentRef,
					ParentComponentDescription: parentComponentDescription,
					DocInfo:                    params.DocInfo,
				})
				if err != nil {
					return nil, err
				}

				resField.Type.IsInlineResponseBody = resField.Type.Scope != ast.ScopeShared
				if g.subsystem.Config.Generation.NameResolutionAtLeastShortest() {
					if resField.Type.IsInlineResponseBody && isErrorStatusCode(statusCode) && resField.Type.Name == ast.HumanizedResponseBody {
						// Errors look strange if they're called "ResponseBodyError"
						g.subsystem.Register.UnregisterType(resField.Type)
						resField.Type.Name = ""
						resField.Type.ContextStack.MarkUsed(ast.ContextTypeRequestResponse)
						resField.Type.ContextStack.MarkUsed(ast.ContextTypeResponseBody)
						g.subsystem.Register.RegisterType(ctx, resField.Type, false)
					}
				}

				// If the response body was an inline type make sure it always gets prefixed with the operation name
				if resField.Type.Scope != ast.ScopeShared && resField.Type.ContextStack.HasFrameOfType(ast.ContextTypeOperation) {
					opFrame := resField.Type.ContextStack.FindLastFrameOfType(ast.ContextTypeOperation)
					opFrame.MustUse = true
				}

				if !resField.Type.IsComponent && parentComponentRef == "" {
					scopedName := fmt.Sprintf("%s_%s_%s", statusCode, mediaType, resField.Name)
					responseFieldOriginalNames[scopedName] = resField.Name

					resField.Name = scopedName
				}
			} else {
				fieldName, err := g.namer.GetFieldName(ctx, schema, "res", parentComponentRef)
				if err != nil {
					return nil, err
				}

				scopedName := fmt.Sprintf("%s_%s_%s", statusCode, mediaType, fieldName)
				responseFieldOriginalNames[scopedName] = fieldName

				resField = &ast.FieldDef{
					Name: scopedName,
					Type: ast.NewType(&ast.TypeDef{Type: ast.DataTypeString}, nil),
				}
			}
			resField.Optional = true // Response types are always optional

			responseDescription := resp.GetDescription()
			if r.GetDescription() != "" {
				responseDescription = r.GetDescription() // Reference level description overrides response level description
			}

			if responseDescription != "" {
				if resField.Comments == nil || resField.Comments.Description == "" {
					if resField.Comments == nil {
						resField.Comments = &ast.Comment{}
					}

					resField.Comments.Description = responseDescription
				}

				if (resField.Type.Comments == nil || resField.Type.Comments.Description == "") && resField.Type.Scope != ast.ScopeShared {
					if resField.Type.Comments == nil {
						resField.Type.Comments = &ast.Comment{}
					}

					resField.Type.Comments.Description = responseDescription
				}
			}

			switch {
			case serializationMethod == ast.SerializationMethodRAW && (resField.Type.Type == ast.DataTypeBytes || resField.Type.Type == ast.DataTypeResponseStream):
				fallthrough
			case serializationMethod == ast.SerializationMethodString:
				fallthrough
			case serializationMethod == ast.SerializationMethodJSON:
				err := g.registerContentType(ctx, responseFieldMapping, resField, statusCode, contentType, serializationMethod, usageExample, typeObj, params.DocInfo)
				var dupErr *duplicateResponseFieldError
				if errors.As(err, &dupErr) {
					return nil, dupErr.IntoValidationError(typeObj.GetCore().Schema.GetKeyNodeOrRoot(typeObj.GetRootNode()))
				} else if err != nil {
					return nil, err
				}
			case serializationMethod == ast.SerializationMethodEventStream:
				resField.Type.EventStreamEnvelope = true

				eventStreamSentinel := g.subsystem.Extensions.HandleSSESentinelExtensionString(typeObj.GetExtensions())
				wrapper := ast.NewType(&ast.TypeDef{
					Type:                ast.DataTypeEventStream,
					ItemType:            resField.Type,
					EventStreamSentinel: eventStreamSentinel,
				}, nil)
				resField.Type = wrapper
				err := g.registerContentType(ctx, responseFieldMapping, resField, statusCode, contentType, serializationMethod, usageExample, typeObj, params.DocInfo)
				var dupErr *duplicateResponseFieldError
				if errors.As(err, &dupErr) {
					return nil, dupErr.IntoValidationError(typeObj.GetCore().Schema.GetKeyNodeOrRoot(typeObj.GetRootNode()))
				} else if err != nil {
					return nil, err
				}
			case serializationMethod == ast.SerializationMethodJsonL:
				wrapper := ast.NewType(&ast.TypeDef{
					Type:     ast.DataTypeJsonL,
					ItemType: resField.Type,
				}, nil)
				resField.Type = wrapper
				err := g.registerContentType(ctx, responseFieldMapping, resField, statusCode, contentType, serializationMethod, usageExample, typeObj, params.DocInfo)
				var dupErr *duplicateResponseFieldError
				if errors.As(err, &dupErr) {
					return nil, dupErr.IntoValidationError(typeObj.GetCore().Schema.GetKeyNodeOrRoot(typeObj.GetRootNode()))
				} else if err != nil {
					return nil, err
				}
			default:
				if warnErr != nil {
					logging.LogWarning(ctx, "unsupported", warnErr)
				}

				resCont, ok := responseFieldMapping.Get("Body")
				if !ok {
					typ := ast.NewType(&ast.TypeDef{Type: ast.DataTypeBytes}, nil)
					typ.AssociatedTypes = ast.TypeDefs{resField.Type}

					examples := []*ast.Example{}

					if typeObj.GetExamples().Len() > 0 {
						for name, e := range typeObj.GetExamples().All() {
							example, err := resolution.Resolve(ctx, e, params.DocInfo)
							if err != nil {
								return nil, err
							}

							examples = append(examples, ast.NewExample(name, example.GetDescription(), resolveExampleValue(example)))
						}
					} else if typeObj.Example != nil {
						examples = append(examples, ast.NewExample("", "", typeObj.Example))
					}

					resCont = responseContent{
						statusToContentType: map[string]map[string]ast.SerializationMethod{},
						responseField: &ast.FieldDef{
							Name:     "Body",
							Type:     typ,
							Optional: true,
						},
						usageExample: map[string]map[string]bool{},
						examples:     examples,
						mediaType:    typeObj,
					}
				}
				if _, ok := resCont.statusToContentType[statusCode]; !ok {
					resCont.statusToContentType[statusCode] = map[string]ast.SerializationMethod{}
				}
				if _, ok := resCont.usageExample[statusCode]; !ok {
					resCont.usageExample[statusCode] = map[string]bool{}
				}

				resCont.statusToContentType[statusCode][contentType] = ast.SerializationMethodRAW
				resCont.usageExample[statusCode][contentType] = usageExample
				responseFieldMapping.Set("Body", resCont)
			}
		}

		if !contentAdded {
			emptyCodes[statusCode] = true
		}
	}

	res := &ast.Response{}

	subResponses := map[string]*ast.SubResponse{}

	usedErrorStatusCodes := []string{}

	if errsExt != nil {
		for rf := range responseFieldMapping.Values() {
			resField := rf.responseField

			allStatusCodesAreErrors := len(rf.statusToContentType) > 0

			for statusCode := range rf.statusToContentType {
				isErrorCode := errsExt.IsErrorStatusCode(statusCode)

				if !isErrorCode {
					allStatusCodesAreErrors = false
				}
			}

			if allStatusCodesAreErrors {
				// Remove the field from the original names map so we don't consider it for conflict detection
				delete(responseFieldOriginalNames, resField.Name)
			}
		}
	}

	for rf := range responseFieldMapping.Values() {
		resField := rf.responseField
		isCustomClass := resField.Type.IsCustomClass()

		allStatusCodesAreErrors := len(rf.statusToContentType) > 0 && errsExt != nil
		hasErrorStatusCodes := false

		if errsExt != nil {
			for statusCode := range rf.statusToContentType {
				isErrorCode := errsExt.IsErrorStatusCode(statusCode)

				if isErrorCode && !isCustomClass {
					// Treat this as an empty error code
					emptyCodes[statusCode] = true
					logging.LogWarning(ctx, "unsupported", errors.NewUnsupportedError(
						"Error responses with non-object content are not currently supported, treating as empty content",
						rf.mediaType.GetCore().Schema.GetKeyNodeOrRoot(rf.mediaType.GetRootNode())),
					)
				}

				if isErrorCode {
					hasErrorStatusCodes = true
				} else {
					allStatusCodesAreErrors = false
				}
			}
		}

		var errField *ast.FieldDef

		var addResponseValue bool
		if g.subsystem.Config.Generation.Fixes.SharedErrorComponentsApr2025 {
			addResponseValue = true
		} else {
			addResponseValue = resField.Type.Scope == ast.ScopeOperations
		}

		if allStatusCodesAreErrors {
			if isCustomClass {
				t, err := g.cloneTypeToErrorsScope(ctx, resField.Type, errorTypeUpdateOptions{
					isRootType:     true,
					addRawResponse: addResponseValue,
					httpMetaType:   httpMetadataType,
					resFormat:      resFormat,
				})
				if err != nil {
					return nil, err
				}
				resField.Type = t
			}
		} else {
			// This "else" code block is a little bonkers of an edge case (Thomas note).
			//   I've tried to add what I think are useful notes. We might want to revisit this behaviour.
			// Effectively what this does is
			//  1. We're in this code path if we have both a "success" and an "error" response code pointing at exactly the same type
			//  2. So we need to "duplicate" the type so both of them work -- we can both throw an error type, and respond with
			//     the shared type depending on the status code.
			//    (TC note: I checked customers, I think this is conceptually okay but probably just represents mistakes in OAS)
			// updateErrorTypeScope will do the right thing (move it into the error scope if possible)
			// But we still want to make sure the original one is registered. So we do an extra register call there (it's idempotent)

			// This bit of the code path is making sure that we've done conflict detection properly : each
			// content type have the exact same type.
			areOriginalFieldNamesUnique := true
			otherNames := map[string]bool{}
			for _, origName := range responseFieldOriginalNames {
				if _, ok := otherNames[origName]; ok {
					areOriginalFieldNamesUnique = false
					break
				}

				otherNames[origName] = true
			}
			if areOriginalFieldNamesUnique {
				originalName, ok := responseFieldOriginalNames[resField.Name]
				if ok {
					resField.Name = originalName
				}
			}

			// Usually errors are removed from responseObj.Fields -- but in this scenario we need to keep it there.
			responseObj.Fields = responseObj.Fields.MustAddField(resField, g.subsystem.Config.MaintainOpenAPIOrder())

			if isCustomClass {
				// First we close the resField to prevent it being modified / removed from the register
				g.subsystem.Register.CloseType(resField.Type)

				// Okay; now the actual splitting.
				if hasErrorStatusCodes {
					// Now let's make a copy, we're going to put the error here.
					tempErrField := *resField
					// Then we make sure resField.Type is updated.
					t, err := g.cloneTypeToErrorsScope(ctx, resField.Type, errorTypeUpdateOptions{
						isRootType:     true,
						addRawResponse: addResponseValue,
						httpMetaType:   httpMetadataType,
						resFormat:      resFormat,
					})
					if err != nil {
						return nil, err
					}
					// Right now, this "t" is the error type
					tempErrField.Type = t
					errField = &tempErrField
				}

				// Now we need to double check that the type is registered
				resField.Type = g.subsystem.Register.RegisterType(ctx, resField.Type, false)
			}
		}

		statusToContentTypeKeys := utils.GetSortedKeys(rf.statusToContentType)

		for _, statusCode := range statusToContentTypeKeys {
			contentTypes := rf.statusToContentType[statusCode]

			subResp, ok := subResponses[statusCode]
			if !ok {
				subResp = &ast.SubResponse{
					Content: []*ast.ResponseBodyContent{},
				}
			}

			subResp.Code = []string{statusCode}

			isErrorCode := isErrorStatusCode(statusCode)

			contentTypesKeys := utils.GetSortedKeys(contentTypes)

			for _, contentType := range contentTypesKeys {
				method := contentTypes[contentType]

				respBodyContent := &ast.ResponseBodyContent{
					SerializationMethod: string(method),
					ContentType:         contentType,
					Content:             resField,
					UsageExample:        rf.usageExample[statusCode][contentType],
					Examples:            rf.examples,
					SSESentinel:         rf.sseSentinel,
				}

				if isErrorCode && !allStatusCodesAreErrors && errField != nil {
					respBodyContent.Content = errField
				}

				subResp.Content = append(subResp.Content, respBodyContent)
			}

			slices.SortStableFunc(subResp.Content, func(i, j *ast.ResponseBodyContent) int {
				return CompareMediaRanges(i.ContentType, j.ContentType)
			})

			if isErrorCode && (hasErrorStatusCodes || allStatusCodesAreErrors) {
				subResp.Error = true
				usedErrorStatusCodes = append(usedErrorStatusCodes, statusCode)
			}

			subResponses[statusCode] = subResp
		}
	}

	// Check if any 2xx or 3xx status codes were found
	found2xxOr3xx := false
	for statusCode := range foundCodes {
		switch {
		case statusCode[0] == '2', statusCode[0] == '3':
			found2xxOr3xx = true
		case statusCode == "default":
			found2xxOr3xx = !isErrorStatusCode(statusCode)
		}

		if found2xxOr3xx {
			break
		}
	}
	if !found2xxOr3xx {
		// If no 2xx or 3xx status codes were found, add a default 2XX empty response
		emptyCodes["2XX"] = true
	}

	for statusCode := range emptyCodes {
		isErrorCode := isErrorStatusCode(statusCode)

		subResponses[statusCode] = &ast.SubResponse{
			Code:  []string{statusCode},
			Error: isErrorCode,
		}

		if isErrorCode {
			usedErrorStatusCodes = append(usedErrorStatusCodes, statusCode)
		}
	}

	if errsExt != nil {
		for _, statusCode := range errsExt.StatusCodes {
			if !slices.Contains(usedErrorStatusCodes, statusCode) {
				subResponses[statusCode] = &ast.SubResponse{
					Code:  []string{statusCode},
					Error: true,
				}
			}
		}
	}

	resps := []*ast.SubResponse{}

	// The response headers map is redundant when `responseFormat: envelop-http` is used since `HttpMeta` already
	// provides access to response headers.
	// This field was preserved in most languages to prevent unnecessary churn / breaking changes.
	// In Python and Typescript SDKs however, the `Headers` field is only templated for `flat` or `envelope` formats.
	headersAdded := resFormat == responseFormatEnvelopeHTTP && slices.Contains([]string{"typescript", "python"}, g.target.Target)

	statusCodes := utils.GetSortedKeys(subResponses)

	for _, statusCode := range statusCodes {
		_, ok := headersMap[statusCode]
		if ok && !headersAdded {
			headersAdded = true
			responseObj.Fields = responseObj.Fields.MustAddField(&ast.FieldDef{
				Name:              "Headers",
				IsResponseHeaders: true,
				Type: ast.NewType(&ast.TypeDef{
					Type: ast.DataTypeMap,
					ItemType: ast.NewType(&ast.TypeDef{
						Type:     ast.DataTypeArray,
						ItemType: ast.NewType(&ast.TypeDef{Type: ast.DataTypeString}, nil),
					}, nil),
				}, nil),
			}, g.subsystem.Config.MaintainOpenAPIOrder())
		}

		subResp := subResponses[statusCode]
		subResp.Headers = ok

		resps = addSubResponse(resps, subResp)
	}

	resps = sortSubResponses(resps)

	res.Responses = resps

	if params.Pagination != nil {
		responseObj.Extensions.Pagination = params.Pagination
	}

	if resFormat == responseFormatFlat {
		g.flattenResponseObject(ctx, responseObj, res, params, headersAdded, id)
	} else {
		g.addHTTPMetaFields(ctx, resFormat, httpMetadataType, responseObj, res)
	}

	err = g.subsystem.Register.CloseTypes(res)

	return res, err
}

func (g *Generator) flattenResponseObject(ctx context.Context, responseObj *ast.TypeDef, res *ast.Response, params handleResponsesParams, headersAdded bool, id string) {
	switch {
	case len(responseObj.Fields) == 0:
		// Empty response case: no response body, no headers, no nada
		res.Type = nil

	case headersAdded && len(responseObj.Fields) == 1:
		// Only headers, no response body
		res.Type = g.subsystem.Register.RegisterType(ctx, responseObj, false)
	default:
		envelopeNeededStill := params.Pagination != nil || headersAdded

		if !headersAdded || len(responseObj.Fields) > 1 {
			var flattenedResponseType *ast.TypeDef

			if len(responseObj.Fields) == 1 {
				// Either single response body, or just headers
				flattenedResponseType = responseObj.Fields[0].Type
				responseObj.Fields = ast.Fields{} // Clear original fields
			} else {
				// Multiple possible response bodies
				if (headersAdded && len(responseObj.Fields) > 2) || (!headersAdded && len(responseObj.Fields) > 1) {
					// Prepare union of possible response body types
					associatedTypes := ast.TypeDefs{}
					fields := ast.Fields{}

					// Separate headers from body fields
					for _, field := range responseObj.Fields {
						if field.Name != "Headers" {
							associatedTypes = append(associatedTypes, field.Type)
						} else {
							fields = append(fields, field)
						}
					}
					responseObj.Fields = fields // Keep only header fields

					// Create naming suffix based on envelope need
					suffix := ""
					if envelopeNeededStill {
						suffix = "_Result"
					}

					// Register union type for multiple possible response body types
					flattenedResponseType = g.subsystem.Register.RegisterType(ctx, ast.NewType(&ast.TypeDef{
						Type:            ast.DataTypeUnion,
						Name:            id + "_" + ast.InternalTypeResponse + suffix,
						Scope:           params.Scope,
						AssociatedTypes: associatedTypes,
					}, ast.ContextStack{
						ast.ContextFrame{
							Type:       ast.ContextTypeRequestResponse,
							Identifier: ast.InternalTypeResponse,
						},
					}), false)
				} else {
					// One response body and headers
					fields := ast.Fields{}
					for _, field := range responseObj.Fields {
						if field.Name != "Headers" {
							flattenedResponseType = field.Type
						} else {
							fields = append(fields, field)
						}
					}
					responseObj.Fields = fields
				}
			}

			if envelopeNeededStill {
				// Check if any non-error response has no body content.
				// If so, the result field must be Optional since those
				// response paths will set result=None.
				resultOptional := false
				for _, subResp := range res.Responses {
					if !subResp.Error && len(subResp.Content) == 0 {
						resultOptional = true
						break
					}
				}

				responseObj.Fields = responseObj.Fields.MustAddField(&ast.FieldDef{
					Name:     "Result",
					Type:     flattenedResponseType,
					Optional: resultOptional,
					Annotations: ast.Annotations{
						&ast.ResponseAnnotation{ResultField: true},
						&ast.NeedsCasingAnnotation{},
					},
				}, g.subsystem.Config.MaintainOpenAPIOrder())

				res.Type = g.subsystem.Register.RegisterType(ctx, responseObj, false)
			} else {
				// Use flattened type directly when no envelope needed
				res.Type = flattenedResponseType
				if flattenedResponseType.IsInlineResponseBody && g.subsystem.Config.Generation.NameResolutionAtLeastShortest() {
					// We can rename the inline schema to just "response"
					g.renameFlattenedInlineResponseBody(ctx, flattenedResponseType, id)
				}
			}
		}
	}
}

func (g *Generator) renameFlattenedInlineResponseBody(ctx context.Context, t *ast.TypeDef, id string) {
	// We can rename the inline schema to just "response"
	if t.Name == ast.HumanizedResponseBody {
		g.subsystem.Register.UnregisterType(t)
		t.Name = id + "_" + ast.InternalTypeResponse
		t.ContextStack.MarkUsed(ast.ContextTypeOperation)
		t.ContextStack.MarkUsed(ast.ContextTypeRequestResponse)
		t.ContextStack.Update(ast.ContextTypeResponseBody, "responseBody", "body")
		t.ContextStack.MarkUnused(ast.ContextTypeResponseBody)
		g.subsystem.Register.RegisterType(ctx, t, false)
	}
	// The inline schema is an array so we can rename the ItemType to "response"
	if t.ItemType != nil && t.ItemType.Name == ast.HumanizedResponseBody {
		g.subsystem.Register.UnregisterType(t.ItemType)
		t.ItemType.Name = id + "_" + ast.InternalTypeResponse
		t.ItemType.ContextStack.MarkUsed(ast.ContextTypeOperation)
		t.ItemType.ContextStack.MarkUsed(ast.ContextTypeRequestResponse)
		t.ContextStack.Update(ast.ContextTypeResponseBody, "responseBody", "body")
		t.ContextStack.MarkUnused(ast.ContextTypeResponseBody)
		g.subsystem.Register.RegisterType(ctx, t.ItemType, false)
	}
}

func (g *Generator) addHTTPMetaFields(ctx context.Context, resFormat string, httpMetadataType *ast.TypeDef, responseObj *ast.TypeDef, res *ast.Response) {
	var responseFields ast.Fields
	if resFormat == responseFormatEnvelopeHTTP {
		responseFields = ast.Fields{
			{
				Name: "HttpMeta",
				Type: g.subsystem.Register.RegisterType(ctx, httpMetadataType, false),
				Annotations: ast.Annotations{
					&ast.JSONAnnotation{Ignore: true, FieldName: "-"},
					&ast.NeedsCasingAnnotation{},
				},
			},
		}
	} else {
		responseFields = ast.Fields{
			{
				Name:               "ContentType",
				Type:               ast.NewType(&ast.TypeDef{Type: ast.DataTypeString}, nil),
				IsResponseMetadata: true,
				Comments: &ast.Comment{
					Description: contentTypeDesc,
				},
			},
			{
				Name:               "StatusCode",
				Type:               ast.NewType(&ast.TypeDef{Type: ast.DataTypeInt32}, nil),
				IsResponseMetadata: true,
				Comments: &ast.Comment{
					Description: statusCodeDesc,
				},
			},
			{
				Name:               "RawResponse",
				Type:               ast.NewType(&ast.TypeDef{Type: ast.DataTypeResponse}, nil),
				IsResponseMetadata: true,
				Comments: &ast.Comment{
					Description: rawResponseDesc,
				},
			},
		}
	}

	for _, field := range responseObj.Fields {
		responseFields = responseFields.MustAddField(field, g.subsystem.Config.MaintainOpenAPIOrder())
	}
	responseObj.Fields = responseFields

	res.Type = g.subsystem.Register.RegisterType(ctx, responseObj, false)
}

func addSubResponse(subResps []*ast.SubResponse, subResp *ast.SubResponse) []*ast.SubResponse {
	found := false
	for i, existingResp := range subResps {
		if existingResp.IsEqualExceptExactCode(subResp) {
			subResps[i].Code = append(existingResp.Code, subResp.Code...)
			found = true
			break
		}
	}
	if !found {
		subResps = append(subResps, subResp)
	}

	return subResps
}

func sortSubResponses(subResps []*ast.SubResponse) []*ast.SubResponse {
	slices.SortStableFunc(subResps, func(i, j *ast.SubResponse) int {
		if slices.Contains(i.Code, "default") {
			return 1
		}
		if slices.Contains(j.Code, "default") {
			return -1
		}

		return CompareCodePredicates(i.Code, j.Code)
	})

	return subResps
}

func CompareCodePredicates(i, j []string) int {
	iscore, icode := scoreCodePredicate(i)
	jscore, jcode := scoreCodePredicate(j)

	return cmp.Or(
		cmp.Compare(iscore, jscore),
		cmp.Compare(icode, jcode),
		cmp.Compare(len(i), len(j)),
	)
}

// CompareMediaRanges is used to order a list of media ranges such that
// wildcards (`*/*`) are pushed to th end, preceded by ranges (e.g. `text/*`)
// and finally preceded by exact media types (e.g. `application/json`) which
// maintain their document order. This comparator is useful to apply on a list
// of subresponses so that response matching logic attempts the most specific
// media types first.
func CompareMediaRanges(i, j string) int {
	itype, _, err := mime.ParseMediaType(i)
	if err != nil {
		panic(fmt.Errorf("%s: invalid content type: %w", i, err))
	}
	jtype, _, err := mime.ParseMediaType(j)
	if err != nil {
		panic(fmt.Errorf("%s: invalid content type: %w", j, err))
	}

	isub := strings.Split(i, "/")[1]
	jsub := strings.Split(j, "/")[1]

	switch {
	case itype == "*/*" && jtype != "*/*":
		return 1
	case itype != "*/*" && jtype == "*/*":
		return -1
	case isub == "*" && jsub != "*":
		return 1
	case isub != "*" && jsub == "*":
		return -1
	default:
		return 0
	}
}

const MAX_CODE_SCORE = 999

func scoreCodePredicate(codes []string) (int, string) {
	maxscore := 0
	maxcode := ""
	for _, c := range codes {
		code := strings.ToUpper(c)
		score := scoreCode(code)
		if score > maxscore {
			maxscore = score
		}

		if maxcode == "" || cmp.Compare(maxcode, code) < 0 {
			maxcode = code
		}
	}
	return maxscore, maxcode
}

func scoreCode(code string) int {
	switch {
	case code == "DEFAULT":
		return MAX_CODE_SCORE + 1
	case code[1:] == "XX":
		return MAX_CODE_SCORE
	default:
		value, err := strconv.Atoi(code)
		if err != nil {
			return MAX_CODE_SCORE
		}

		return value
	}
}

func (g *Generator) registerContentType(ctx context.Context, mapping *sequencedmap.Map[string, responseContent], resField *ast.FieldDef, statusCode, contentType string, serializationMethod ast.SerializationMethod, usageExample bool, typeObj *oas.MediaType, docInfo *document.DocumentInfo) error {
	content, ok := mapping.Get(resField.Name)
	if !ok {
		examples := []*ast.Example{}

		if typeObj.GetExamples().Len() > 0 {
			for name, e := range typeObj.Examples.All() {
				example, err := resolution.Resolve(ctx, e, docInfo)
				if err != nil {
					return err
				}

				examples = append(examples, ast.NewExample(name, example.GetDescription(), resolveExampleValue(example)))
			}
		} else if typeObj.Example != nil {
			examples = append(examples, ast.NewExample("", "", typeObj.Example))
		}

		content = responseContent{
			statusToContentType: map[string]map[string]ast.SerializationMethod{},
			responseField:       resField,
			usageExample:        map[string]map[string]bool{},
			examples:            examples,
			mediaType:           typeObj,
		}
	}
	if _, ok := content.statusToContentType[statusCode]; !ok {
		content.statusToContentType[statusCode] = map[string]ast.SerializationMethod{}
	}
	if _, ok := content.usageExample[statusCode]; !ok {
		content.usageExample[statusCode] = map[string]bool{}
	}

	if err := content.responseField.Type.IsEqual(resField.Type); err != nil {
		// Attempt to disambiguate duplicate response field names by finding the best
		// discriminating labels from each type's identifiers (context stack, model namespace, etc.)
		existingSuffix, newSuffix := findResponseFieldDiscriminators(content.responseField.Type, resField.Type)
		if existingSuffix != "" || newSuffix != "" {
			// Rename the existing entry by appending its discriminator suffix,
			// but only if the target name is not already taken in the mapping.
			oldName := content.responseField.Name
			newExistingName := oldName + existingSuffix
			if !mapping.Has(newExistingName) {
				content.responseField.Name = newExistingName
				mapping.Delete(oldName)
				mapping.Set(content.responseField.Name, content)

				// Rename the new field and recurse to create a fresh entry
				resField.Name += newSuffix
				return g.registerContentType(ctx, mapping, resField, statusCode, contentType, serializationMethod, usageExample, typeObj, docInfo)
			}
			// If the discriminator-based name would collide with an existing entry,
			// fall through to the numeric fallback below.
		}

		// Fallback: number the new field when types have identical labels.
		// Find the next available number suffix (2, 3, 4, ...).
		n := 2
		for mapping.Has(resField.Name + strconv.Itoa(n)) {
			n++
		}
		resField.Name += strconv.Itoa(n)
		return g.registerContentType(ctx, mapping, resField, statusCode, contentType, serializationMethod, usageExample, typeObj, docInfo)
	}
	var err error
	content.statusToContentType[statusCode][contentType] = serializationMethod
	content.usageExample[statusCode][contentType] = usageExample

	sentinel, err := g.subsystem.Extensions.HandleSSESentinelExtension(typeObj.GetExtensions())
	if err != nil {
		return err
	}
	if sentinel != nil {
		g.subsystem.Features.RecordFeatureUsage(ctx, features.FeatureServerEventsSentinels)
		content.sseSentinel = sentinel.DataValue
	}

	mapping.Set(resField.Name, content)

	return nil
}

// findResponseFieldDiscriminators uses the name resolution machinery to find the
// minimal set of labels that disambiguate two types with the same response field name.
// Returns suffix strings to append to each field name, or empty strings if the types
// cannot be distinguished by their identifiers.
func findResponseFieldDiscriminators(existingType, newType *ast.TypeDef) (string, string) {
	types := [][]string{
		responseFieldLabels(existingType),
		responseFieldLabels(newType),
	}

	discriminators := namer.FindBestDiscriminators(types)

	return strings.Join(discriminators[0], ""), strings.Join(discriminators[1], "")
}

// responseFieldLabels collects identifying labels from a TypeDef's context stack
// for use with FindBestDiscriminators. All frames are included (regardless of Used
// flag) since the Used flag tracks type naming, not field naming.
func responseFieldLabels(t *ast.TypeDef) []string {
	var labels []string
	for _, frame := range t.ContextStack {
		name := frame.DisplayName()
		if name != "" {
			labels = append(labels, name)
		}
	}
	return labels
}

// buildSSEEnvelopeSchema constructs a synthetic SSE envelope schema from an itemSchema.
// The envelope has the standard SSE fields (data, event, id, retry) where the data field
// references the provided itemSchema type.
func buildSSEEnvelopeSchema(itemSchema *oas3.JSONSchema[oas3.Referenceable]) *oas3.JSONSchema[oas3.Referenceable] {
	props := sequencedmap.New[string, *oas3.JSONSchema[oas3.Referenceable]]()
	props.Set("data", itemSchema)
	props.Set("event", oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{
		Type: oas3.NewTypeFromString(oas3.SchemaTypeString),
	}))
	props.Set("id", oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{
		Type: oas3.NewTypeFromString(oas3.SchemaTypeString),
	}))
	props.Set("retry", oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{
		Type: oas3.NewTypeFromString(oas3.SchemaTypeInteger),
	}))

	return oas3.NewJSONSchemaFromSchema[oas3.Referenceable](&oas3.Schema{
		Type:       oas3.NewTypeFromString(oas3.SchemaTypeObject),
		Properties: props,
	})
}
