package buckettypes

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

const (
	SplitModelTypeThreshold = 200
	SplitModelChunkSize     = 150
)

// SplitLargeModels splits models that exceed SplitModelTypeThreshold into
// smaller chunks of SplitModelChunkSize. Each chunk becomes a new model named
// after its last (most root-like) type after topological sorting. This prevents
// compiler complexity limits in languages like Python (pyright) and TypeScript (tsc).
func (b *Bucketer) SplitLargeModels(
	types ast.BucketedTypes,
	operationServers *sequencedmap.Map[string, *ast.Servers],
) ast.BucketedTypes {
	result := ast.NewBucketedTypes()

	threshold := b.getSplitThreshold()
	chunkSize := b.getSplitChunkSize()

	for location, models := range types.All() {
		newModels := sequencedmap.New[string, ast.TypeDefs]()

		// Collect all existing model names in this location for dedup.
		allModelNames := make(map[string]bool)
		for modelName := range models.All() {
			allModelNames[modelName] = true
		}

		for model, typeDefs := range models.All() {
			if len(typeDefs) <= threshold {
				newModels.Set(model, typeDefs)
				continue
			}

			// The original model is being replaced by chunks, so free its name
			// for reuse by a chunk (typically the last/root chunk).
			delete(allModelNames, model)

			sorted := ast.TopologicalSortTypeDefs(typeDefs, false)
			chunks := chunkTypeDefs(sorted, chunkSize)

			chunkNames := make([]string, len(chunks))
			for i, chunk := range chunks {
				name := chunk[len(chunk)-1].Name

				finalName := name
				if allModelNames[finalName] || b.isReservedFileName(finalName) {
					suffix := 1
					for {
						candidate := fmt.Sprintf("%s_%d", name, suffix)
						if !allModelNames[candidate] && !b.isReservedFileName(candidate) {
							finalName = candidate
							break
						}
						suffix++
					}
				}

				allModelNames[finalName] = true
				chunkNames[i] = finalName
			}

			for i, chunk := range chunks {
				chunkName := chunkNames[i]
				for _, td := range chunk {
					td.ResolvedModel = chunkName
				}
				newModels.Set(chunkName, chunk)
			}

			if servers, ok := operationServers.Get(model); ok {
				operationServers.Delete(model)
				operationServers.Set(chunkNames[0], servers)
			}
		}

		result.Set(location, newModels)
	}

	return result
}

// isReservedFileName checks whether the given name, once sanitized, collides
// with a reserved model filename (e.g. __init__, index).
func (b *Bucketer) isReservedFileName(name string) bool {
	if len(b.reservedModelFileNames) == 0 {
		return false
	}
	sanitized := b.subsystem.Sanitizer.SanitizeFileName(name)
	return b.reservedModelFileNames[sanitized]
}

func chunkTypeDefs(types ast.TypeDefs, size int) []ast.TypeDefs {
	var chunks []ast.TypeDefs
	for i := 0; i < len(types); i += size {
		end := i + size
		if end > len(types) {
			end = len(types)
		}
		chunks = append(chunks, types[i:end])
	}
	return chunks
}
