package main

import (
	"bufio"
	"context"
	"fmt"
	"go/ast"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/ettle/strcase"
	"github.com/speakeasy-api/easytemplate"
)

const (
	testType = "Test"
)

type Test struct {
	SymbolName string
	ValueName  string
}

func generateTests(f *ast.File, wd string) error {
	tests := getTests(f)

	// Parse already generated tests
	existingTests := getGeneratedTests(wd)

	// Look for features.ts (relative to internal/features directory)
	featuresFiles, err := filepath.Glob(filepath.Join(wd, "../../templates/templates/*/features.ts"))
	if err != nil {
		return err
	}

	// Skip newly added tests by default
	for _, test := range tests {
		testID := test.ValueName
		if !existingTests[testID] {
			for _, filePath := range featuresFiles {
				if err := skipTest(filePath, testID); err != nil {
					return err
				}
			}
		}
	}

	if err := generateTestFiles(wd, tests); err != nil {
		return err
	}

	return nil
}

func getTests(f *ast.File) []Test {
	tests := []Test{}

	inspect(f, testType, func(name string, vs *ast.ValueSpec) {
		tests = append(tests, Test{
			SymbolName: name,
			ValueName:  sanitizeTest(name),
		})
	})

	slices.SortFunc(tests, func(a Test, b Test) int {
		return strings.Compare(a.SymbolName, b.SymbolName)
	})

	return tests
}

func generateTestFiles(outDir string, tests []Test) error {
	et := easytemplate.New(easytemplate.WithTemplateFuncs(map[string]any{}))
	ctx := context.Background()
	if err := et.Init(ctx, nil); err != nil {
		return err
	}

	if err := et.TemplateFile(ctx, "../../cmd/features/generate/templates/tests_generated.go.stmpl", filepath.Join(outDir, "tests_generated.go"), tests); err != nil {
		return err
	}

	return nil
}

func sanitizeTest(test string) string {
	name := strings.TrimPrefix(test, "Test")
	name = strcase.ToKebab(name)

	return name
}

// Collect existing tests from tests_generated.go
func getGeneratedTests(wd string) map[string]bool {
	existingTests := make(map[string]bool)

	file, err := os.Open(filepath.Join(wd, "tests_generated.go"))
	if err != nil {
		return existingTests
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	re := regexp.MustCompile(`return "([^"]+)"`)

	for scanner.Scan() {
		line := scanner.Text()
		if matches := re.FindStringSubmatch(line); len(matches) > 1 {
			existingTests[matches[1]] = true
		}
	}

	return existingTests
}

// Add testID to the isTestSkipped function in a features.ts file
func skipTest(featuresFilePath string, testID string) error {
	expectedSignature := "function isTestSkipped(test: string): boolean"
	expectedEndOfArray := "].includes(test);"

	content, err := os.ReadFile(featuresFilePath)
	if err != nil {
		return err
	}

	lines := strings.Split(string(content), "\n")
	for i, line := range lines {
		if strings.Contains(line, expectedSignature) {
			for j := i; j < len(lines); j++ {
				if strings.Contains(lines[j], expectedEndOfArray) {
					// Insert testID at the end of the array
					newTestLine := fmt.Sprintf("    \"%s\",", testID)
					lines = slices.Insert(lines, j, newTestLine)
					break
				}
			}
			updatedContent := strings.Join(lines, "\n")
			return os.WriteFile(featuresFilePath, []byte(updatedContent), 0644)
		}
	}

	return nil
}
