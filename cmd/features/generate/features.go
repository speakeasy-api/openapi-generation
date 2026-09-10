package main

import (
	"context"
	"go/ast"
	"path/filepath"
	"slices"
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/easytemplate"
)

const (
	featureType = "Feature"
)

type Feature struct {
	SymbolName     string
	ValueName      string
	Tier           string
	ReadmeSections []string
}

func generateFeatures(f *ast.File, wd string) error {
	features := getFeatures(f)

	if err := generateFeatureFiles(wd, features); err != nil {
		return err
	}

	return nil
}

func getFeatures(f *ast.File) []Feature {
	features := []Feature{}

	inspect(f, featureType, func(name string, vs *ast.ValueSpec) {
		readmeSections := []string{}
		tier := "Core"

		if vs.Doc != nil {
			for _, doc := range vs.Doc.List {
				comment := strings.TrimSpace(strings.TrimPrefix(doc.Text, "//"))

				switch {
				case strings.HasPrefix(comment, "ReadmeSections:"):
					readmeSections = strings.Split(strings.TrimSpace(strings.TrimPrefix(comment, "ReadmeSections:")), ",")
				case strings.HasPrefix(comment, "Tier:"):
					tier = strings.TrimSpace(strings.TrimPrefix(comment, "Tier:"))
				}
			}
		}

		f := Feature{
			SymbolName:     name,
			ValueName:      sanitizeFeature(name),
			Tier:           tier,
			ReadmeSections: readmeSections,
		}

		features = append(features, f)
	})

	slices.SortFunc(features, func(a Feature, b Feature) int {
		return strings.Compare(a.SymbolName, b.SymbolName)
	})

	return features
}

func sanitizeFeature(feature string) string {
	name := strings.TrimPrefix(feature, "Feature")
	name = strcase.ToGoCamel(name)

	return fixFeatureName(name)
}

// Unfortunately, the casing does not handle some cases correctly because of either bugs or inconsistent initial human naming.
func fixFeatureName(name string) string {
	if strings.Contains(name, "UrLs") {
		name = strings.ReplaceAll(name, "UrLs", "URLs")
	}

	if strings.HasPrefix(name, "oA") {
		name = strings.ReplaceAll(name, "oA", "oa")
	}

	if name == "bigInt" {
		name = "bigint"
	}

	return name
}

func generateFeatureFiles(outDir string, features []Feature) error {
	et := easytemplate.New(easytemplate.WithTemplateFuncs(map[string]any{
		"getFeaturesByTier":          getFeaturesByTier,
		"getReadmeSectionsByFeature": getReadmeSectionsByFeature,
	}))
	ctx := context.Background()
	if err := et.Init(ctx, nil); err != nil {
		return err
	}

	if err := et.TemplateFile(ctx, "../../cmd/features/generate/templates/features_generated.go.stmpl", filepath.Join(outDir, "features_generated.go"), features); err != nil {
		return err
	}

	if err := et.TemplateFile(ctx, "../../cmd/features/generate/templates/features.d.ts.stmpl", "../../templates/templates/common/common/features.d.ts", features); err != nil {
		return err
	}

	return nil
}

func getFeaturesByTier(features []Feature) map[string][]Feature {
	tiers := map[string][]Feature{}

	for _, f := range features {
		if _, ok := tiers[f.Tier]; !ok {
			tiers[f.Tier] = []Feature{}
		}

		tiers[f.Tier] = append(tiers[f.Tier], f)
	}

	return tiers
}

func getReadmeSectionsByFeature(features []Feature) map[string][]string {
	sections := map[string][]string{}

	for _, f := range features {
		sections[f.SymbolName] = f.ReadmeSections
	}

	return sections
}
