package main

import (
	"context"
	"encoding/csv"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"strings"

	templates2 "github.com/speakeasy-api/openapi-generation/v2/pkg/templates"

	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/markdown"
)

func main() {
	_ = os.Setenv("SPEAKEASY_DEBUG", "true")

	templates := flag.String("t", "all", "comma separated templates to generate report for (defaults to all)")
	flag.Parse()

	if templates == nil {
		return
	}

	targets := []types.Target{}

	if *templates == "all" {
		tt := templates2.AllSupportedTargets()
		for _, t := range tt {
			if slices.Contains([]string{"java", "python", "typescript"}, t.Template) {
				continue
			}

			targets = append(targets, t)
		}
	} else if *templates != "" {
		tt := strings.Split(*templates, ",")

		slices.Sort(tt)

		for _, t := range tt {
			targets = append(targets, types.NewTargetFromTemplate(t))
		}
	}

	templateFeatures := []*features.Features{}

	for _, target := range targets {
		f, err := features.New(target, map[string]any{})
		if err != nil {
			log.Fatalf("failed to create feature for target %s: %v", target, err)
		}

		templateFeatures = append(templateFeatures, f)
	}

	ctx := context.Background()

	ratings := getTemplateRatingsReport(ctx, templateFeatures)
	implementedFeatures := getImplementedFeaturesReport(ctx, templateFeatures)
	implementedReadmeSections := getImplementedReadmeSectionsReport(ctx, templateFeatures)
	tests, err := getTestsReport(ctx, templateFeatures)
	if err != nil {
		log.Fatalf("failed to get tests report: %v", err)
	}

	if err := createMarkdownReport(ratings, implementedFeatures, implementedReadmeSections, tests); err != nil {
		log.Fatalf("failed to create markdown report: %v", err)
	}

	if err := createCSVReport("ratings.csv", ratings); err != nil {
		log.Fatalf("failed to create ratings CSV report: %v", err)
	}

	if err := createCSVReport("implemented-features.csv", implementedFeatures); err != nil {
		log.Fatalf("failed to create implemented features CSV report: %v", err)
	}

	if err := createCSVReport("implemented-readme-sections.csv", implementedReadmeSections); err != nil {
		log.Fatalf("failed to create implemented readme sections CSV report: %v", err)
	}

	if err := createCSVReport("tests.csv", tests); err != nil {
		log.Fatalf("failed to create tests CSV report: %v", err)
	}

	log.Println("Features report generated successfully")
}

func createMarkdownReport(ratings, implementedFeatures, implementedReadmeSections, tests [][]string) error {
	output := "# Features Report\n\n## Template Ratings\n\nLegend: Gold :dollar:, GA :shipit:, Beta :chart_with_upwards_trend:, Alpha :baby_bottle:, WIP :hammer:\n\n"

	output += markdown.CreateMarkdownTable(ratings)

	output += "\n\n## Implemented Features\n\nLegend: :white_check_mark: Implemented, :warning: Partially Implemented (missing Readme sections or tests), :no_entry: Not Implemented, :heavy_minus_sign: Ignored\n\n"

	output += markdown.CreateMarkdownTable(implementedFeatures)

	output += "\n\n## Implemented Readme Sections\n\nLegend: :white_check_mark: Implemented, :no_entry: Not Implemented, :heavy_minus_sign: Ignored\n\n"

	output += markdown.CreateMarkdownTable(implementedReadmeSections)

	output += "\n\n## Tests\n\nLegend: :white_check_mark: Implemented, :question_mark: Skipped, :no_entry: Not Implemented, :warning: Missing Data\n\n"

	output += markdown.CreateMarkdownTable(tests)

	return os.WriteFile("./reports/features-report.md", []byte(output), 0o644)
}

func createCSVReport(fileName string, content [][]string) error {
	file, err := os.Create(filepath.Join("./reports", fileName))
	if err != nil {
		return err
	}
	defer file.Close()

	w := csv.NewWriter(file)
	return w.WriteAll(content)
}

func getTemplateRatingsReport(ctx context.Context, templateFeatures []*features.Features) [][]string {
	headings := []string{"Template", "Tier"}

	content := make([][]string, 0, 1+len(templateFeatures))
	content = append(content, headings)

	for _, f := range templateFeatures {
		coreFeatures := features.GetFeaturesByTier(features.TierCore)
		betaFeatures := features.GetFeaturesByTier(features.TierBeta)
		gaFeatures := features.GetFeaturesByTier(features.TierGA)
		goldFeatures := features.GetFeaturesByTier(features.TierGold)

		allCoreFeaturesImplemented := true

		for _, feature := range coreFeatures {
			if f.IsFeatureIgnored(ctx, feature) {
				continue
			}

			if !f.IsFeatureSupported(ctx, feature) {
				allCoreFeaturesImplemented = false
				break
			}
		}

		allBetaFeaturesImplemented := allCoreFeaturesImplemented

		for _, feature := range betaFeatures {
			if f.IsFeatureIgnored(ctx, feature) {
				continue
			}

			if !f.IsFeatureSupported(ctx, feature) {
				allBetaFeaturesImplemented = false
				break
			}
		}

		allGAFeaturesImplemented := allBetaFeaturesImplemented

		for _, feature := range gaFeatures {
			if f.IsFeatureIgnored(ctx, feature) {
				continue
			}

			if !f.IsFeatureSupported(ctx, feature) {
				allGAFeaturesImplemented = false
				break
			}
		}

		allGoldFeaturesImplemented := allGAFeaturesImplemented

		for _, feature := range goldFeatures {
			if f.IsFeatureIgnored(ctx, feature) {
				continue
			}

			if !f.IsFeatureSupported(ctx, feature) {
				allGoldFeaturesImplemented = false
				break
			}
		}

		tier := "WIP :hammer:"

		switch {
		case allGoldFeaturesImplemented:
			tier = features.TierGold + " :dollar:"
		case allGAFeaturesImplemented:
			tier = features.TierGA + " :shipit:"
		case allBetaFeaturesImplemented:
			tier = features.TierBeta + " :chart_with_upwards_trend:"
		case allCoreFeaturesImplemented:
			tier = "Alpha :baby_bottle:"
		}

		content = append(content, []string{f.Target().Template, tier})
	}

	slices.SortFunc(content[1:], func(a, b []string) int {
		switch {
		case a[1] == b[1]:
			return strings.Compare(a[0], b[0])
		case strings.Contains(a[1], features.TierGold):
			return -1
		case strings.Contains(b[1], features.TierGold):
			return 1
		case strings.Contains(a[1], features.TierGA):
			return -1
		case strings.Contains(b[1], features.TierGA):
			return 1
		case strings.Contains(a[1], features.TierBeta):
			return -1
		case strings.Contains(b[1], features.TierBeta):
			return 1
		case strings.Contains(a[1], "Alpha"):
			return -1
		case strings.Contains(b[1], "Alpha"):
			return 1
		}

		return strings.Compare(a[0], b[0])
	})

	return content
}

func getImplementedFeaturesReport(ctx context.Context, templateFeatures []*features.Features) [][]string {
	allFeatures := features.GetAllFeatures()

	headings := make([]string, 0, 1+len(allFeatures))
	headings = append(headings, "Template")

	slices.SortFunc(allFeatures, func(a, b features.Feature) int {
		tierA := features.GetFeatureTier(a)
		tierB := features.GetFeatureTier(b)

		if tierA == tierB {
			return strings.Compare(a.String(), b.String())
		}

		switch {
		case tierA == features.TierCore:
			return -1
		case tierB == features.TierCore:
			return 1
		case tierA == features.TierBeta:
			return -1
		case tierB == features.TierBeta:
			return 1
		case tierA == features.TierGA:
			return -1
		case tierB == features.TierGA:
			return 1
		case tierA == features.TierGold:
			return -1
		case tierB == features.TierGold:
			return 1
		default:
			return strings.Compare(a.String(), b.String())
		}
	})

	for _, feature := range allFeatures {
		headings = append(headings, fmt.Sprintf("%s (%s)", feature, features.GetFeatureTier(feature)))
	}

	content := make([][]string, 0, 1+len(templateFeatures))
	content = append(content, headings)

	for _, f := range templateFeatures {
		content = append(content, getImplementedFeaturesRow(ctx, f, allFeatures))
	}

	return transposeContent(content)
}

func getImplementedFeaturesRow(ctx context.Context, f *features.Features, allFeatures []features.Feature) []string {
	implementedFeatures := f.GetImplementedFeaturesList(ctx)

	row := []string{f.Target().Template}

	for _, feature := range allFeatures {
		readmeSections := features.GetReadmeSectionsByFeature(feature)

		allReadmeSectionsImplemented := true
		for _, section := range readmeSections {
			if !f.IsReadmeSectionImplemented(ctx, section) && !f.IsReadmeSectionIgnored(ctx, section) {
				allReadmeSectionsImplemented = false
				break
			}
		}

		switch {
		case slices.Contains(implementedFeatures, feature):
			if allReadmeSectionsImplemented {
				row = append(row, ":white_check_mark:")
			} else {
				row = append(row, ":warning:")
			}
		case f.IsFeatureIgnored(ctx, feature):
			row = append(row, ":heavy_minus_sign:")
		default:
			row = append(row, ":no_entry:")
		}
	}

	return row
}

func getImplementedReadmeSectionsReport(ctx context.Context, templateFeatures []*features.Features) [][]string {
	allReadmeSections := features.GetAllReadmeSections()
	slices.Sort(allReadmeSections)

	headings := make([]string, 0, 1+len(allReadmeSections))
	headings = append(headings, "Template")
	headings = append(headings, allReadmeSections...)

	content := [][]string{
		headings,
	}

	for _, f := range templateFeatures {
		if slices.Contains([]string{"postman", "terraform"}, f.Target().Template) {
			continue
		}

		content = append(content, getImplementedReadmeSectionsRow(ctx, f))
	}

	return transposeContent(content)
}

func getImplementedReadmeSectionsRow(ctx context.Context, f *features.Features) []string {
	allReadmeSections := features.GetAllReadmeSections()
	slices.Sort(allReadmeSections)

	row := []string{f.Target().Template}

	for _, section := range allReadmeSections {
		switch {
		case f.IsReadmeSectionImplemented(ctx, section):
			row = append(row, ":white_check_mark:")
		case f.IsReadmeSectionIgnored(ctx, section):
			row = append(row, ":heavy_minus_sign:")
		default:
			row = append(row, ":no_entry:")
		}
	}

	return row
}

func getTestsReport(ctx context.Context, templateFeatures []*features.Features) ([][]string, error) {
	allTests := features.GetAllTests()
	slices.Sort(allTests)

	headings := make([]string, 0, 1+len(allTests))
	headings = append(headings, "Template")
	for _, test := range allTests {
		headings = append(headings, test.String())
	}

	content := [][]string{
		headings,
	}

	for _, f := range templateFeatures {
		if slices.Contains([]string{"postman", "terraform"}, f.Target().Template) {
			continue
		}

		row, err := getTestsRow(ctx, f)
		if err != nil {
			return nil, err
		}

		content = append(content, row)
	}

	return transposeContent(content), nil
}

func getTestsRow(ctx context.Context, f *features.Features) ([]string, error) {
	allTests := features.GetAllTests()
	recordedTests, err := f.GetTestRecords(ctx)
	if err != nil {
		return nil, err
	}

	row := []string{f.Target().Template}

	for _, test := range allTests {
		testID := test.String()

		switch {
		case slices.Contains(recordedTests, testID):
			row = append(row, ":white_check_mark:")
		case f.IsTestSkipped(ctx, test):
			row = append(row, ":question_mark:")
		default:
			row = append(row, ":warning:")
		}
	}

	return row, nil
}

func transposeContent(content [][]string) [][]string {
	rows := len(content)
	cols := len(content[0])

	transposed := make([][]string, cols)
	for i := range transposed {
		transposed[i] = make([]string, rows)
	}

	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			transposed[j][i] = content[i][j]
		}
	}

	return transposed
}
