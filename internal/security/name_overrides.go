package security

import (
	"fmt"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/namer"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/sequencedmap"
)

type SchemeNameOverrides struct {
	originalToOverride map[string]string
	overrideToOriginal map[string]string
	schemeNamespaces   map[string]string // schemeKey -> model namespace
}

func newSchemeNameOverrides() SchemeNameOverrides {
	return SchemeNameOverrides{
		originalToOverride: map[string]string{},
		overrideToOriginal: map[string]string{},
		schemeNamespaces:   map[string]string{},
	}
}

func BuildSchemeNameOverrides(exts *extensions.Extensions, securitySchemes *sequencedmap.Map[string, *openapi.ReferencedSecurityScheme]) (SchemeNameOverrides, error) {
	overrides := newSchemeNameOverrides()
	if exts == nil || securitySchemes == nil {
		return overrides, nil
	}

	schemeKeySet := map[string]struct{}{}
	for schemeKey := range securitySchemes.Keys() {
		schemeKeySet[schemeKey] = struct{}{}
	}

	for schemeKey, schemeRef := range securitySchemes.All() {
		if schemeRef == nil {
			continue
		}

		scheme := schemeRef.GetResolvedObject()
		if scheme == nil {
			scheme = schemeRef.Object
		}
		if scheme == nil {
			continue
		}
		if scheme.GetExtensions().Len() == 0 {
			continue
		}

		override, err := exts.GetResolvedSchemaName(scheme.GetExtensions(), schemeKey)
		if err != nil {
			return overrides, err
		}

		modelNamespace, _ := exts.GetModelNamespace(scheme.GetExtensions())
		if modelNamespace != "" {
			overrides.schemeNamespaces[schemeKey] = modelNamespace
		}

		if override == "" || override == schemeKey {
			continue
		}

		if _, exists := overrides.originalToOverride[schemeKey]; !exists {
			overrides.originalToOverride[schemeKey] = override
		}

		if existingOriginal, exists := overrides.overrideToOriginal[override]; exists && existingOriginal != schemeKey {
			// Only flag collision if both schemes are in the same namespace (or both have no namespace)
			existingNamespace := overrides.schemeNamespaces[existingOriginal]
			if existingNamespace == modelNamespace {
				return overrides, errors.NewValidationError(
					fmt.Sprintf("x-speakeasy-name-override collision: %q is already used by security scheme %q", override, existingOriginal),
					scheme.GetRootNode(),
					nil,
				)
			}
			// Different namespaces — not a collision, but overrideToOriginal can only store one.
			// The primary direction (originalToOverride) is unique per original key and is always correct.
		}

		if _, exists := schemeKeySet[override]; exists {
			return overrides, errors.NewValidationError(
				fmt.Sprintf("x-speakeasy-name-override collision: %q conflicts with existing security scheme key", override),
				scheme.GetRootNode(),
				nil,
			)
		}

		overrides.overrideToOriginal[override] = schemeKey
	}

	// Disambiguate cross-namespace collisions: group by resolved override name,
	// find groups where multiple originals map to the same override with different namespaces,
	// and use FindBestDiscriminators to produce unique field names.
	overrideToOriginals := map[string][]string{} // override -> list of original keys
	for original, override := range overrides.originalToOverride {
		overrideToOriginals[override] = append(overrideToOriginals[override], original)
	}

	for override, originals := range overrideToOriginals {
		if len(originals) < 2 {
			continue
		}

		// Build label lists for each scheme using namespace as discriminator
		types := make([][]string, len(originals))
		for i, orig := range originals {
			ns := overrides.schemeNamespaces[orig]
			types[i] = []string{ns}
		}

		discriminators := namer.FindBestDiscriminators(types)

		// Delete the shared override→original mapping before adding disambiguated ones
		delete(overrides.overrideToOriginal, override)

		for i, orig := range originals {
			prefix := strings.Join(discriminators[i], "")
			if prefix == "" {
				// No discriminator (e.g. non-namespaced scheme) — restore the reverse mapping
				overrides.overrideToOriginal[override] = orig
				continue
			}
			// Capitalize first letter of override to maintain camelCase when prefixed
			capitalizedOverride := strings.ToUpper(override[:1]) + override[1:]
			disambiguated := prefix + capitalizedOverride

			// Guard against the disambiguated name shadowing an existing scheme key
			if _, exists := schemeKeySet[disambiguated]; exists {
				return overrides, errors.NewValidationError(
					fmt.Sprintf("disambiguated override %q conflicts with existing security scheme key", disambiguated),
					nil,
					nil,
				)
			}

			overrides.originalToOverride[orig] = disambiguated
			overrides.overrideToOriginal[disambiguated] = orig
		}
	}

	return overrides, nil
}

func (o SchemeNameOverrides) ResolveSchemeKey(originalKey string) string {
	if override, ok := o.originalToOverride[originalKey]; ok {
		return override
	}
	return originalKey
}

func (o SchemeNameOverrides) ResolveOriginalKey(schemeKey string) string {
	if original, ok := o.overrideToOriginal[schemeKey]; ok {
		return original
	}
	return schemeKey
}
