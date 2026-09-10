package buckettypes

import (
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

// seenType keeps track of resolved type-model relationships
type seenType struct {
	Type  *ast.TypeDef
	Model string
}

func DeduplicateTypes(types ast.BucketedTypes) ast.BucketedTypes {
	dedupedTypes := ast.NewBucketedTypes()

	for location, models := range types.All() {
		dedupedModels := sequencedmap.New[string, ast.TypeDefs]()
		seenTypes := make(map[string]*seenType)

		for model, types := range models.All() {
			modelToUse := model

			if _, ok := dedupedModels.Get(modelToUse); !ok {
				dedupedModels.Set(modelToUse, ast.TypeDefs{})
			}

			for _, typ := range types {
				if _, ok := seenTypes[getRegistrationID(typ)]; !ok {
					seenTypes[getRegistrationID(typ)] = &seenType{
						Type:  typ,
						Model: modelToUse,
					}

					typ.ResolvedModel = modelToUse
					prev, ok := dedupedModels.Get(modelToUse)
					if !ok {
						prev = ast.TypeDefs{}
					}
					dedupedModels.Set(modelToUse, append(prev, typ))
				} else {
					currModel := seenTypes[getRegistrationID(typ)].Model

					// The new model is a better fit
					if modelToUse == typ.Name {
						seenTypes[typ.GetRegistrationID()].Model = modelToUse

						currModels, _ := dedupedModels.Get(currModel)
						for _, typ := range currModels {
							typ.ResolvedModel = modelToUse
						}
						dedupedModels.Set(currModel, ast.TypeDefs{})

						prev, _ := dedupedModels.Get(modelToUse)
						dedupedModels.Set(modelToUse, append(prev, currModels...))
					} else {
						otherModels, _ := dedupedModels.Get(modelToUse)
						for _, typ := range otherModels {
							typ.ResolvedModel = currModel
						}
						dedupedModels.Set(modelToUse, ast.TypeDefs{})

						prev, _ := dedupedModels.Get(currModel)
						dedupedModels.Set(currModel, append(prev, otherModels...))

						modelToUse = currModel
					}
				}
			}
		}

		dedupedTypes.Set(location, dedupedModels)
	}

	return dedupedTypes
}
