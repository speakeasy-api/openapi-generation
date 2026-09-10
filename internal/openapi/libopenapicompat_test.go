package openapi_test

import (
	"testing"

	config "github.com/speakeasy-api/sdk-gen-config"

	"github.com/speakeasy-api/openapi-generation/v2/internal/configuration"
	internalOpenAPI "github.com/speakeasy-api/openapi-generation/v2/internal/openapi"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
	"github.com/stretchr/testify/assert"
)

func newConfig(maintainOrder bool) *configuration.Config {
	return configuration.New(&config.Configuration{
		Generation: config.Generation{
			MaintainOpenAPIOrder: maintainOrder,
		},
	}, "go")
}

func newPathItem() *openapi.PathItem {
	return &openapi.PathItem{
		Map: sequencedmap.New[openapi.HTTPMethod, *openapi.Operation](),
	}
}

func collectMethods(pathItem *openapi.PathItem, cfg *configuration.Config) []openapi.HTTPMethod {
	methods := make([]openapi.HTTPMethod, 0, pathItem.Len())
	for method := range internalOpenAPI.SortOperationsLikeLibOpenAPI(pathItem, cfg) {
		methods = append(methods, method)
	}
	return methods
}

func TestSortOperationsLikeLibOpenAPI_QueryMethod(t *testing.T) {
	t.Run("includes QUERY method", func(t *testing.T) {
		pathItem := newPathItem()
		pathItem.Set(openapi.HTTPMethodGet, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodQuery, &openapi.Operation{})

		methods := collectMethods(pathItem, newConfig(true))

		assert.Contains(t, methods, openapi.HTTPMethodGet)
		assert.Contains(t, methods, openapi.HTTPMethodQuery)
		assert.Len(t, methods, 2)
	})

	t.Run("QUERY only", func(t *testing.T) {
		pathItem := newPathItem()
		pathItem.Set(openapi.HTTPMethodQuery, &openapi.Operation{})

		methods := collectMethods(pathItem, newConfig(true))

		assert.Equal(t, []openapi.HTTPMethod{openapi.HTTPMethodQuery}, methods)
	})
}

func TestSortOperationsLikeLibOpenAPI_AdditionalOperations(t *testing.T) {
	t.Run("includes additional operations", func(t *testing.T) {
		pathItem := newPathItem()
		pathItem.Set(openapi.HTTPMethodGet, &openapi.Operation{})

		additionalOps := sequencedmap.New[string, *openapi.Operation]()
		additionalOps.Set("COPY", &openapi.Operation{})
		additionalOps.Set("PURGE", &openapi.Operation{})
		pathItem.AdditionalOperations = additionalOps

		methods := collectMethods(pathItem, newConfig(true))

		assert.Contains(t, methods, openapi.HTTPMethodGet)
		assert.Contains(t, methods, openapi.HTTPMethod("COPY"))
		assert.Contains(t, methods, openapi.HTTPMethod("PURGE"))
		assert.Len(t, methods, 3)
	})

	t.Run("additional operations only", func(t *testing.T) {
		pathItem := newPathItem()

		additionalOps := sequencedmap.New[string, *openapi.Operation]()
		additionalOps.Set("LOCK", &openapi.Operation{})
		pathItem.AdditionalOperations = additionalOps

		methods := collectMethods(pathItem, newConfig(true))

		assert.Equal(t, []openapi.HTTPMethod{openapi.HTTPMethod("LOCK")}, methods)
	})

	t.Run("nil additional operations", func(t *testing.T) {
		pathItem := newPathItem()
		pathItem.Set(openapi.HTTPMethodPost, &openapi.Operation{})

		methods := collectMethods(pathItem, newConfig(true))

		assert.Equal(t, []openapi.HTTPMethod{openapi.HTTPMethodPost}, methods)
	})

	t.Run("skips nil operations in additional operations", func(t *testing.T) {
		pathItem := newPathItem()

		additionalOps := sequencedmap.New[string, *openapi.Operation]()
		additionalOps.Set("COPY", &openapi.Operation{})
		additionalOps.Set("INVALID", nil)
		additionalOps.Set("PURGE", &openapi.Operation{})
		pathItem.AdditionalOperations = additionalOps

		methods := collectMethods(pathItem, newConfig(true))

		assert.Contains(t, methods, openapi.HTTPMethod("COPY"))
		assert.Contains(t, methods, openapi.HTTPMethod("PURGE"))
		assert.NotContains(t, methods, openapi.HTTPMethod("INVALID"))
		assert.Len(t, methods, 2)
	})
}

func TestSortOperationsLikeLibOpenAPI_NonMaintainOrder(t *testing.T) {
	t.Run("additional operations included when not maintaining order", func(t *testing.T) {
		pathItem := newPathItem()
		pathItem.Set(openapi.HTTPMethodPost, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodGet, &openapi.Operation{})

		additionalOps := sequencedmap.New[string, *openapi.Operation]()
		additionalOps.Set("PURGE", &openapi.Operation{})
		pathItem.AdditionalOperations = additionalOps

		// maintainOrder=false uses alphabetical ordering via chainAdditionalOperations
		methods := collectMethods(pathItem, newConfig(false))

		assert.Contains(t, methods, openapi.HTTPMethodGet)
		assert.Contains(t, methods, openapi.HTTPMethodPost)
		assert.Contains(t, methods, openapi.HTTPMethod("PURGE"))
		assert.Len(t, methods, 3)
	})

	t.Run("no additional operations when not maintaining order", func(t *testing.T) {
		pathItem := newPathItem()
		pathItem.Set(openapi.HTTPMethodGet, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodPost, &openapi.Operation{})

		methods := collectMethods(pathItem, newConfig(false))

		// Should be alphabetically sorted (get < post)
		assert.Equal(t, []openapi.HTTPMethod{openapi.HTTPMethodGet, openapi.HTTPMethodPost}, methods)
	})
}

func TestSortOperationsLikeLibOpenAPI_AllMethodTypes(t *testing.T) {
	t.Run("all standard methods plus QUERY plus additional", func(t *testing.T) {
		pathItem := newPathItem()
		pathItem.Set(openapi.HTTPMethodGet, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodPut, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodPost, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodDelete, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodOptions, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodHead, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodPatch, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodTrace, &openapi.Operation{})
		pathItem.Set(openapi.HTTPMethodQuery, &openapi.Operation{})

		additionalOps := sequencedmap.New[string, *openapi.Operation]()
		additionalOps.Set("COPY", &openapi.Operation{})
		pathItem.AdditionalOperations = additionalOps

		methods := collectMethods(pathItem, newConfig(true))

		assert.Len(t, methods, 10)
		assert.Contains(t, methods, openapi.HTTPMethodGet)
		assert.Contains(t, methods, openapi.HTTPMethodPut)
		assert.Contains(t, methods, openapi.HTTPMethodPost)
		assert.Contains(t, methods, openapi.HTTPMethodDelete)
		assert.Contains(t, methods, openapi.HTTPMethodOptions)
		assert.Contains(t, methods, openapi.HTTPMethodHead)
		assert.Contains(t, methods, openapi.HTTPMethodPatch)
		assert.Contains(t, methods, openapi.HTTPMethodTrace)
		assert.Contains(t, methods, openapi.HTTPMethodQuery)
		assert.Contains(t, methods, openapi.HTTPMethod("COPY"))
	})
}
