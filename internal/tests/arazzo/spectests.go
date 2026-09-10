package arazzo

import (
	"context"
	"fmt"
	"net/http"
	"slices"
	"strconv"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/tests"
	"github.com/speakeasy-api/openapi/arazzo"
	"github.com/speakeasy-api/openapi/sequencedmap"
	config "github.com/speakeasy-api/sdk-gen-config"
	"gopkg.in/yaml.v3"
)

type generateTestsFromSpecOptions struct {
	AST                *ast.AST
	ArazzoDoc          *arazzo.Arazzo
	RegenerateTests    map[string]bool
	OpenAPIDocPath     string
	LockFile           *config.LockFile
	IsInternalTestSpec bool
	Config             *configuration.Config
}

func generateTestsFromSpec(ctx context.Context, opts generateTestsFromSpecOptions) (*arazzo.Arazzo, error) {
	tests, existingTests := getTestsFromSDK(getTestsFromSDKOptions{
		SDK:                opts.AST.MainSDK,
		IsInternalTestSpec: opts.IsInternalTestSpec,
		Config:             opts.Config,
		LockFile:           opts.LockFile,
	})

	arazzoDoc := opts.ArazzoDoc

	if arazzoDoc == nil {
		summary := "Created from " + opts.OpenAPIDocPath
		if existingTests {
			summary = "Migrated from " + opts.OpenAPIDocPath
		}

		arazzoDoc = createArazzoDoc(summary, opts.OpenAPIDocPath)
	}

	if err := workflowsFromTests(ctx, workflowsFromTestsOptions{
		AST:                opts.AST,
		Tests:              tests,
		RegenerateTests:    opts.RegenerateTests,
		ArazzoDoc:          arazzoDoc,
		OpenAPIDocPath:     opts.OpenAPIDocPath,
		LockFile:           opts.LockFile,
		Config:             opts.Config,
		IsInternalTestSpec: opts.IsInternalTestSpec,
	}); err != nil {
		return nil, err
	}

	return arazzoDoc, nil
}

type getTestsFromSDKOptions struct {
	SDK                *ast.SDK
	IsInternalTestSpec bool
	Config             *configuration.Config
	LockFile           *config.LockFile
}

func getTestsFromSDK(opts getTestsFromSDKOptions) (*sequencedmap.Map[string, []test], bool) {
	existingTests := false

	sdkTests := sequencedmap.New[string, []test]()

	for _, op := range opts.SDK.Operations {
		if op.TestExplicitlyDisabled || !tests.IsOperationSupported(op, opts.Config) || op == nil {
			continue
		}

		opTests := getTestsFromOp(op, opts.LockFile.Examples)

		tts, ok := sdkTests.Get(op.ID)
		if !ok {
			tts = []test{}
		}
		sdkTests.Set(op.ID, append(tts, opTests...))
	}

	for _, subSDK := range opts.SDK.SubSDKs {
		subSDKTests, subExistingTests := getTestsFromSDK(getTestsFromSDKOptions{
			SDK:                subSDK,
			IsInternalTestSpec: opts.IsInternalTestSpec,
			Config:             opts.Config,
			LockFile:           opts.LockFile,
		})
		if subExistingTests {
			existingTests = true
		}

		for operationID, stts := range subSDKTests.All() {
			tts, ok := sdkTests.Get(operationID)
			if !ok {
				tts = []test{}
			}
			sdkTests.Set(operationID, append(tts, stts...))
		}
	}

	return sdkTests, existingTests
}

func getResponseCodeAsInt(code string) int {
	if code == "default" {
		return http.StatusPermanentRedirect // TODO we ideally want to chose a status code that is not taken by the operation already
	}
	if strings.HasSuffix(strings.ToLower(code), "xx") {
		code = code[0:1] + "00"
	}
	c, err := strconv.Atoi(code)
	if err != nil {
		panic(fmt.Errorf("failed to parse code %s: %w", code, err))
	}
	return c
}

func getTestsFromOp(op *ast.Operation, preCalculateExamples config.Examples) []test {
	internalID := ""
	internalIDVal, ok := op.Extensions.All["x-speakeasy-test-internal-id"]
	if ok {
		internalID, _ = internalIDVal.(string)
	}

	opExamples, ok := preCalculateExamples.Get(op.OriginalID)
	if !ok {
		return nil
	}

	tests := []test{}

	for exampleName, opExample := range opExamples.All() {
		t := test{
			Name: exampleName,
		}

		populateTestFromRequestBody(op, &t, exampleName, &opExample)

		if op.Request != nil && op.Request.Params != nil && opExample.Parameters != nil {
			populateTestFromParams(op.Request.Params.QueryParams, "query", &t, exampleName, opExample.Parameters.Query)
			populateTestFromParams(op.Request.Params.PathParams, "path", &t, exampleName, opExample.Parameters.Path)
			populateTestFromParams(op.Request.Params.HeaderParams, "header", &t, exampleName, opExample.Parameters.Header)
		}

		if !populateTestFromResponses(op, &t, exampleName, &opExample) {
			// If we didn't populate a response this isn't a valid test
			continue
		}

		tests = append(tests, t)
	}

	for i, test := range tests {
		test.InternalID = internalID

		if len(tests) > 1 && test.InternalID != "" {
			test.InternalID = fmt.Sprintf("%s-%s", test.InternalID, test.Name)
		}

		tests[i] = test
	}

	return tests
}

func populateTestFromRequestBody(op *ast.Operation, test *test, exampleName string, opExample *config.OperationExamples) {
	if opExample.RequestBody == nil || op.Request == nil || op.Request.RequestBody == nil {
		return
	}

	example := op.Request.Examples.FindByName(exampleName)
	if example == nil {
		return
	}

	anno := op.Request.RequestBody.Annotations.Get(ast.AnnotationTypeRequest)
	if anno == nil {
		return
	}

	reqAnno := anno.(*ast.RequestAnnotation)
	if test.RequestBody == nil {
		test.RequestBody = sequencedmap.New[string, yaml.Node]()
	}

	test.RequestBody.Set(reqAnno.MediaType, *example.Value)
}

func populateTestFromParams(params []*ast.Param, paramType string, test *test, exampleName string, paramExamples *sequencedmap.Map[string, yaml.Node]) {
	if len(params) == 0 || paramExamples.Len() == 0 {
		return
	}

	for paramName := range paramExamples.All() {
		for _, param := range params {
			if paramName != param.Field.OriginalName {
				continue
			}

			example := param.Examples.FindByName(exampleName)
			if example == nil {
				continue
			}

			if test.Parameters == nil {
				test.Parameters = &parameters{}
			}

			switch paramType {
			case "path":
				if test.Parameters.Path == nil {
					test.Parameters.Path = sequencedmap.New[string, yaml.Node]()
				}

				test.Parameters.Path.Set(param.Field.OriginalName, *example.Value)
			case "query":
				if test.Parameters.Query == nil {
					test.Parameters.Query = sequencedmap.New[string, yaml.Node]()
				}

				test.Parameters.Query.Set(param.Field.OriginalName, *example.Value)
			case "header":
				if test.Parameters.Header == nil {
					test.Parameters.Header = sequencedmap.New[string, yaml.Node]()
				}

				test.Parameters.Header.Set(param.Field.OriginalName, *example.Value)
			}
		}
	}
}

func populateTestFromResponses(op *ast.Operation, test *test, exampleName string, opExample *config.OperationExamples) bool {
	if opExample.Responses.Len() == 0 || op.Response == nil || len(op.Response.Responses) == 0 {
		return false
	}

	resCodes := slices.Collect(opExample.Responses.Keys())

	// Filter out response codes that are no longer present in the operation's response AST
	validResCodes := make([]string, 0, len(resCodes))
	for _, code := range resCodes {
		// Check if this response code exists in the operation's responses
		for _, resp := range op.Response.Responses {
			if slices.Contains(resp.Code, code) {
				validResCodes = append(validResCodes, code)
				break
			}
		}
	}

	codes := sequencedmap.New[int, string]()
	codesAsInt := []int{}

	hasCodes := false

	for _, code := range validResCodes {
		intCode := getResponseCodeAsInt(code)
		hasCodes = true

		// TODO We don't currently support tests for non-successful responses
		if intCode < 200 || intCode >= 300 {
			continue
		}

		codes.Set(intCode, code)
		codesAsInt = append(codesAsInt, intCode)
	}
	slices.Sort(codesAsInt)

	// No successful response defined in examples so we can't generate a test for this operation
	if hasCodes && len(codesAsInt) == 0 {
		return false
	}

	responses := sequencedmap.New[string, yaml.Node]()

	if len(codesAsInt) == 0 {
		codesAsInt = append(codesAsInt, 200)
		codes.Set(200, "200")
	}

	code := codes.GetOrZero(codesAsInt[0])

	var contentType string
	var example *yaml.Node

	for _, resp := range op.Response.Responses {
		if !slices.Contains(resp.Code, code) {
			continue
		}

		// TODO we should really be trying to select the right content type based on accept headers or have a test per content type
		// but just use the first found content type for now
		for _, content := range resp.Content {
			ex := content.Examples.FindByName(exampleName)
			if ex == nil {
				continue
			}

			contentType = content.ContentType
			example = ex.Value
			break
		}
	}

	if example == nil {
		responses.Set(code, yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!bool",
			Value: "true",
		})
	} else {
		responseBodyExamples := yaml.Node{
			Kind: yaml.MappingNode,
			Tag:  "!!map",
		}

		responseBodyExamples.Content = append(responseBodyExamples.Content, &yaml.Node{
			Kind:  yaml.ScalarNode,
			Tag:   "!!str",
			Value: contentType,
			Style: yaml.DoubleQuotedStyle,
		}, example)

		responses.Set(code, responseBodyExamples)
	}

	test.Responses = responses

	return true
}
