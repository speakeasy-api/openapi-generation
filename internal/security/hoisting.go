package security

import (
	"context"
	"slices"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/document"
	"github.com/speakeasy-api/openapi-generation/v2/internal/resolution"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

type securityUsage struct {
	count        int
	requirements *openapi.SecurityRequirement
}

// getHoistingCandidate will return the security requirement that is most used in the document by operations,
// this is used to hoist security requirements to the global level if no global security is defined
func getHoistingCandidate(ctx context.Context, docInfo *document.DocumentInfo, opts *Opts) (*openapi.SecurityRequirement, error) {
	if docInfo.Doc.Paths == nil {
		return nil, nil
	}

	var securitySchemes *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme]
	if docInfo.Doc.Components != nil {
		securitySchemes = docInfo.Doc.Components.SecuritySchemes
	}
	nameOverrides, err := BuildSchemeNameOverrides(opts.Subsystem.Extensions, securitySchemes)
	if err != nil {
		return nil, err
	}

	usedSchemes := sequencedmap.New[string, securityUsage]()

	for _, pi := range docInfo.Doc.Paths.AllOrdered(opts.Subsystem.Config.GetSequencedMapIterationOrder()) {
		pathItem, err := resolution.Resolve(ctx, pi, docInfo)
		if err != nil {
			return nil, err
		}
		for _, op := range pathItem.AllOrdered(opts.Subsystem.Config.GetSequencedMapIterationOrder()) {
			if op.Security == nil {
				continue
			}

			for _, sec := range op.Security {
				schemeKeys := []string{}
				for scheme := range sec.All() {
					schemeKeys = append(schemeKeys, nameOverrides.ResolveSchemeKey(scheme))
				}
				slices.Sort(schemeKeys)

				key := strings.Join(schemeKeys, "-")

				schemaUsage, ok := usedSchemes.Get(key)
				if !ok {
					schemaUsage = securityUsage{
						count:        0,
						requirements: sec,
					}
				}

				usedSchemes.Set(key, securityUsage{
					count:        schemaUsage.count + 1,
					requirements: schemaUsage.requirements,
				})
			}
		}
	}

	mostUsedCount := 0
	var mostUsedSecurity *openapi.SecurityRequirement

	for usage := range usedSchemes.Values() {
		if usage.count > mostUsedCount {
			mostUsedCount = usage.count
			mostUsedSecurity = usage.requirements
		}
	}

	return mostUsedSecurity, nil
}
