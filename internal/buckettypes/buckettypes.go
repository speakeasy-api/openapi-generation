package buckettypes

import (
	"context"
	"fmt"
	"os"
	"slices"
	"sync"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	"github.com/speakeasy-api/openapi-generation/v2/internal/env"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

type Sanitizer interface {
	SanitizeFileName(name string) string
}

type Bucketer struct {
	subsystem          *subsystem.Subsystem
	operationModelsMap sync.Map

	// Configurable thresholds for splitting large models.
	// Zero values mean "use the package-level defaults".
	splitThreshold int
	splitChunkSize int

	// reservedModelFileNames contains sanitized filenames (without extension)
	// that must not be used as chunk names when splitting large models.
	// These typically correspond to hardcoded template files (e.g. __init__, index).
	reservedModelFileNames map[string]bool
}

func New(subsystem *subsystem.Subsystem) *Bucketer {
	return &Bucketer{
		subsystem:          subsystem,
		operationModelsMap: sync.Map{},
	}
}

// SetSplitThresholds overrides the default split-model thresholds.
func (b *Bucketer) SetSplitThresholds(threshold, chunkSize int) {
	b.splitThreshold = threshold
	b.splitChunkSize = chunkSize
}

// SetReservedModelFileNames sets the list of sanitized filenames that must not
// be used as chunk names when splitting large models.
func (b *Bucketer) SetReservedModelFileNames(names []string) {
	b.reservedModelFileNames = make(map[string]bool, len(names))
	for _, n := range names {
		b.reservedModelFileNames[n] = true
	}
}

func (b *Bucketer) getSplitThreshold() int {
	if b.splitThreshold > 0 {
		return b.splitThreshold
	}
	return SplitModelTypeThreshold
}

func (b *Bucketer) getSplitChunkSize() int {
	if b.splitChunkSize > 0 {
		return b.splitChunkSize
	}
	return SplitModelChunkSize
}

// BucketTypes buckets types using the improved ordered maps.
func (b *Bucketer) BucketTypes(ctx context.Context, sdk *ast.AST) (ast.BucketedTypes, *sequencedmap.Map[string, *ast.Servers]) {
	start := time.Now()

	types := ast.NewBucketedTypes()
	operationServers := sequencedmap.New[string, *ast.Servers]()

	mainSDK := sdk.MainSDK
	globals := mainSDK.Globals
	if globals != nil {
		for _, global := range globals.Fields {
			types = mergeTypes(types, getModelTypes(global.Type.Name, global.Type, nil))
		}
	}

	for _, operation := range mainSDK.Operations {
		types = mergeTypes(types, b.collectOperationModels(operation, operationServers))
	}

	for _, subSDK := range mainSDK.SubSDKs {
		for _, operation := range flattenOperationsPerSDK(subSDK) {
			types = mergeTypes(types, b.collectOperationModels(operation, operationServers))
		}
	}

	// Collect types from webhook operations
	for _, ops := range sdk.Webhooks.All() {
		for i := range ops {
			types = mergeTypes(types, b.collectOperationModels(&ops[i], operationServers))
		}
	}

	types = mergeTypes(types, collectSecurityModels(mainSDK))

	for _, typ := range mainSDK.AdditionalTypes {
		types = mergeTypes(types, getModelTypes(typ.Name, typ, nil))
	}

	if !b.subsystem.Features.CollectGlobalsAndServersModels(ctx) {
		return types, operationServers
	}

	types = mergeTypes(types, b.collectOperationGlobalsModels(mainSDK))
	types = mergeTypes(types, b.collectAllServerModels(mainSDK))

	// TODO: Move to language config rather than inline here
	skipDeduplicationForTargets := []string{"java", "csharp", "php", "unity"}

	if !slices.Contains(skipDeduplicationForTargets, b.subsystem.Target.Target) {
		types = DeduplicateTypes(types)
	}

	types = b.SplitLargeModels(types, operationServers)

	logging.From(ctx).Debug(fmt.Sprintf("BucketTypes completed in '%s'", time.Since(start)))

	return types, operationServers
}

func getRegistrationID(typeDef *ast.TypeDef) string {
	return typeDef.GetRegistrationID()
}

func AddTypeToBucket(types ast.BucketedTypes, model string, typeDef *ast.TypeDef) ast.BucketedTypes {
	typeDef.ResolvedModel = model
	bucket, ok := types.Get(typeDef.OutputLocation)
	if !ok {
		bucket = sequencedmap.New[string, ast.TypeDefs]()
		types.Set(typeDef.OutputLocation, bucket)
	}
	bucket.Set(model, ast.TypeDefs{typeDef})
	return types
}

func getModelTypes(model string, typeDef *ast.TypeDef, visited map[string]*ast.TypeDef) ast.BucketedTypes {
	if typeDef == nil {
		if env.IsDebug() {
			fmt.Fprintf(os.Stderr, "DEBUG buckettypes: getModelTypes called with nil typeDef for model %q\n", model)
		}
		return ast.NewBucketedTypes()
	}

	if visited == nil {
		visited = make(map[string]*ast.TypeDef)
	}

	if typeDef.IsComponent {
		model = typeDef.Name
	}

	types := ast.NewBucketedTypes()

	if typeDef.IsCustomType() {
		if typeDef.Truncated || visited[getRegistrationID(typeDef)] != nil {
			if typeDef != visited[getRegistrationID(typeDef)] {
				// This probably means, something went wrong when registering types.
				// Somehow the AST is pointing at a type which is not registered.
				// One of these two types is not registered.
				// Maybe one was registered but then later replaced but the AST has a dangling
				// pointer to the old type?
				panic(
					fmt.Sprintf("Multiple types for the same registration ID %q", getRegistrationID(typeDef)),
				)
			}
			return types
		}
		visited[getRegistrationID(typeDef)] = typeDef
	}

	switch typeDef.Type {
	case "error", "class":
		types = AddTypeToBucket(types, model, typeDef)
		// Reverse fields iteration over a local alias of the slice. This only snapshots
		// the slice header/length; callers must ensure the AST is not mutated concurrently.
		fields := typeDef.Fields
		for i := len(fields) - 1; i >= 0; i-- {
			if fields[i] == nil {
				if env.IsDebug() {
					fmt.Fprintf(os.Stderr, "DEBUG buckettypes: nil field at index %d in typeDef %q (model %q, type %q)\n", i, typeDef.Name, model, typeDef.Type)
				}
				continue
			}
			if fields[i].Type == nil {
				if env.IsDebug() {
					fmt.Fprintf(os.Stderr, "DEBUG buckettypes: nil field.Type at index %d (field %q) in typeDef %q (model %q, type %q)\n", i, fields[i].Name, typeDef.Name, model, typeDef.Type)
				}
				continue
			}
			types = mergeTypes(types, getModelTypes(model, fields[i].Type, visited))
		}
	case "event-stream", "array", "set", "map", "jsonl":
		if typeDef.ItemType != nil {
			types = mergeTypes(types, getModelTypes(model, typeDef.ItemType, visited))
		}
	case "enum", "union":
		types = AddTypeToBucket(types, model, typeDef)
	}

	for i, typ := range typeDef.AssociatedTypes {
		if typ == nil {
			if env.IsDebug() {
				fmt.Fprintf(os.Stderr, "DEBUG buckettypes: nil AssociatedType at index %d in typeDef %q (model %q, type %q)\n", i, typeDef.Name, model, typeDef.Type)
			}
			continue
		}
		types = mergeTypes(types, getModelTypes(model, typ, visited))
	}

	return types
}

func mergeTypes(outTypes, inTypes ast.BucketedTypes) ast.BucketedTypes {
	for location, models := range inTypes.All() {
		existingModels, ok := outTypes.Get(location)
		if !ok {
			existingModels = sequencedmap.New[string, ast.TypeDefs]()
			outTypes.Set(location, existingModels)
		}

		for model, typeDefs := range models.All() {
			existingTypeDefs, ok := existingModels.Get(model)
			if !ok {
				existingModels.Set(model, typeDefs)
			} else {
				combined := append(typeDefs, existingTypeDefs...)
				uniqueCombined := uniqueBy(combined, getRegistrationID)
				existingModels.Set(model, uniqueCombined)
			}
		}
	}
	return outTypes
}

func uniqueBy(slice ast.TypeDefs, keyFunc func(*ast.TypeDef) string) ast.TypeDefs {
	seen := make(map[string]bool)
	result := ast.TypeDefs{}
	for _, item := range slice {
		key := keyFunc(item)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return result
}

func collectSecurityModels(sdk *ast.SDK) ast.BucketedTypes {
	types := ast.NewBucketedTypes()
	if sdk.Security != nil {
		types = mergeTypes(types, getModelTypes(sdk.Security.Name, sdk.Security.Type, nil))
	}
	return types
}

func flattenOperationsPerSDK(sdk *ast.SDK) []*ast.Operation {
	operations := append([]*ast.Operation{}, sdk.Operations...)
	for _, subSDK := range sdk.SubSDKs {
		operations = append(operations, flattenOperationsPerSDK(subSDK)...)
	}
	return operations
}

func (b *Bucketer) collectOperationModels(operation *ast.Operation, operationServers *sequencedmap.Map[string, *ast.Servers]) ast.BucketedTypes {
	types := b.GetOperationModels(operation, false, false, false)

	operationServers.Set(b.GetDeduplicatedOperationModelName(operation), operation.Servers)

	return types
}

func getTopLevelResponseType(model string, typeDef *ast.TypeDef) ast.BucketedTypes {
	if typeDef == nil {
		if env.IsDebug() {
			fmt.Fprintf(os.Stderr, "DEBUG buckettypes: getTopLevelResponseType called with nil typeDef for model %q\n", model)
		}
		return ast.NewBucketedTypes()
	}
	if typeDef.IsComponent {
		model = typeDef.Name
	}

	types := ast.NewBucketedTypes()
	if typeDef.Truncated {
		return types
	}

	return AddTypeToBucket(types, model, typeDef)
}

func GetOperationModelName(subsystem *subsystem.Subsystem, operation *ast.Operation) string {
	model := operation.ID
	imports := subsystem.Config.GetLanguageConfigValue("imports").(configuration.ImportConfig)
	if imports.GetOperationsPath() == imports.GetSharedPath() {
		model = operation.ID + "Op"
	}
	return model
}

func (b *Bucketer) GetDeduplicatedOperationModelName(operation *ast.Operation) string {
	model := GetOperationModelName(b.subsystem, operation)

	sanitized := b.subsystem.Sanitizer.SanitizeFileName(model)

	originalModelName, ok := b.operationModelsMap.Load(sanitized)

	if !ok {
		originalModelName = model
		b.operationModelsMap.Store(sanitized, originalModelName)
	}

	return originalModelName.(string)
}

func (b *Bucketer) GetOperationModels(operation *ast.Operation, ignoreResponseSubTypes bool, ignoreResponseTypes bool, considerFlattening bool) ast.BucketedTypes {
	types := ast.NewBucketedTypes()
	model := b.GetDeduplicatedOperationModelName(operation)

	if !ignoreResponseTypes && operation.Response != nil {
		if ignoreResponseSubTypes && operation.Response.Type != nil {
			types = mergeTypes(types, getTopLevelResponseType(model, operation.Response.Type))
		} else {
			if operation.Response.Type != nil {
				types = mergeTypes(types, getModelTypes(model, operation.Response.Type, nil))
			}
			for _, response := range operation.Response.Responses {
				for _, content := range response.Content {
					types = mergeTypes(types, getModelTypes(model, content.Content.Type, nil))
				}
			}
		}
	}

	if operation.Request != nil && (operation.Request.Field == nil || operation.Request.Field.Type == nil) {
		if env.IsDebug() {
			fmt.Fprintf(os.Stderr, "DEBUG buckettypes: operation %q has Request but Request.Field or Request.Field.Type is nil\n", operation.ID)
		}
	}
	if operation.Request != nil && operation.Request.Field != nil && operation.Request.Field.Type != nil {
		requestModels := ast.NewBucketedTypes()
		if !considerFlattening || !operationParametersFlattened(operation) {
			requestModels = getModelTypes(model, operation.Request.Field.Type, nil)
		} else {
			if operation.Request.RequestBody != nil && operation.Request.RequestBody.Type != nil {
				requestModels = mergeTypes(requestModels, getModelTypes(model, operation.Request.RequestBody.Type, nil))
			}
			if operation.Request.Params != nil {
				for _, param := range operation.Request.Params.QueryParams {
					requestModels = mergeTypes(requestModels, getModelTypes(model, param.Field.Type, nil))
				}
				for _, param := range operation.Request.Params.PathParams {
					requestModels = mergeTypes(requestModels, getModelTypes(model, param.Field.Type, nil))
				}
				for _, param := range operation.Request.Params.HeaderParams {
					requestModels = mergeTypes(requestModels, getModelTypes(model, param.Field.Type, nil))
				}
			}
		}
		types = mergeTypes(types, requestModels)
	}

	if operation.Security != nil && operation.Security.Type == nil {
		if env.IsDebug() {
			fmt.Fprintf(os.Stderr, "DEBUG buckettypes: operation %q has Security but Security.Type is nil\n", operation.ID)
		}
	}
	if operation.Security != nil && operation.Security.Type != nil {
		types = mergeTypes(types, getModelTypes(model, operation.Security.Type, nil))
	}

	for _, callback := range operation.Callbacks {
		if callback == nil {
			continue
		}
		types = mergeTypes(types, getModelTypes(model, callback, nil))
	}

	return types
}

func operationParametersFlattened(operation *ast.Operation) bool {
	return operation.Request != nil &&
		operation.Request.Field.Type.Type == "class" &&
		operation.MaxMethodParams > 0 &&
		operation.Request.Params != nil &&
		len(operation.Request.Field.Type.Fields) <= operation.MaxMethodParams
}

func (b *Bucketer) collectOperationGlobalsModels(sdk *ast.SDK) ast.BucketedTypes {
	types := ast.NewBucketedTypes()

	for _, op := range sdk.Operations {
		if op.Globals != nil {
			types = mergeTypes(types, AddTypeToBucket(ast.NewBucketedTypes(), b.GetDeduplicatedOperationModelName(op), op.Globals))
		}
	}

	for _, subSDK := range sdk.SubSDKs {
		types = mergeTypes(types, b.collectOperationGlobalsModels(subSDK))
	}

	return types
}

func (b *Bucketer) collectAllServerModels(sdk *ast.SDK) ast.BucketedTypes {
	types := ast.NewBucketedTypes()

	for _, subSDK := range sdk.SubSDKs {
		types = mergeTypes(types, b.collectAllServerModels(subSDK))
	}

	opsPath := b.subsystem.Config.GetLanguageConfigValue("imports").(configuration.ImportConfig).GetOperationsPath()

	for _, op := range sdk.Operations {
		if op.Servers != nil {
			model := b.GetDeduplicatedOperationModelName(op)

			serverTypes := ast.NewBucketedTypes()

			// Create a bucket for the operations path and set the model with an empty slice.
			bucket := sequencedmap.New[string, ast.TypeDefs]()
			bucket.Set(model, []*ast.TypeDef{})
			serverTypes.Set(opsPath, bucket)

			types = mergeTypes(types, serverTypes)
		}
	}
	return types
}
