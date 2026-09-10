package main

import (
	"bytes"
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

func checkTestCoverage(lang string) {
	allTests := features.GetAllTests()

	f, err := features.New(types.NewTargetFromTemplate(lang), map[string]any{})
	if err != nil {
		panic(err)
	}

	testRecords, err := loadTestRecords(lang)
	if err != nil {
		panic(err)
	}

	var missingTests []string
	var coveredButSkippedTests []string

	for _, test := range allTests {
		id := test.String()
		_, ok := testRecords[id]

		skipped := f.IsTestSkipped(context.Background(), test)

		if !ok {
			if skipped {
				fmt.Printf("SKIPPED: Test %s is skipped for %s\n", id, lang)
			} else {
				fmt.Printf("MISSING: Test %s is not covered by %s\n", id, lang)
				missingTests = append(missingTests, id)
			}
		} else if skipped {
			fmt.Printf("COVERED: Test %s is covered but marked as skipped for %s\n", id, lang)
			coveredButSkippedTests = append(coveredButSkippedTests, id)
		}
		delete(testRecords, id)
	}

	if len(missingTests) > 0 {
		fmt.Printf("Test coverage is not complete, MISSING tests: %s\n", strings.Join(missingTests, ", "))
	}

	if len(coveredButSkippedTests) > 0 {
		fmt.Printf("Some tests are COVERED but still marked as skipped. Remove these from the skip list: %s\n", strings.Join(coveredButSkippedTests, ", "))
	}

	if len(testRecords) > 0 {
		fmt.Println("Some UNREGISTERED tests were recorded. Register new test IDs in internal/features/tests.go.")
		for id := range testRecords {
			fmt.Printf("  %s\n", id)
		}
	}

	if len(missingTests) > 0 || len(coveredButSkippedTests) > 0 || len(testRecords) > 0 {
		os.Exit(1)
	}
}

func loadTestRecords(lang string) (map[string]bool, error) {
	recordPaths := []string{}

	switch lang {
	case "csharp":
		files, _ := filepath.Glob("./testSDKs/sdk-csharp-*/Tests/bin/Debug/net*.0/test-csharp-record.txt")
		recordPaths = append(recordPaths, files...)
	case "cli":
		files, _ := filepath.Glob("./testSDKs/sdk-cli-*/tests/test-cli-record.txt")
		recordPaths = append(recordPaths, files...)
	case "go":
		files, _ := filepath.Glob("./testSDKs/sdk-go-*/tests/test-go-record.txt")
		recordPaths = append(recordPaths, files...)
	case "javav2":
		files, _ := filepath.Glob("./testSDKs/sdk-javav2-*/build/test-javav2-record.txt")
		recordPaths = append(recordPaths, files...)
	case "mcp-typescript":
		files, _ := filepath.Glob("./testSDKs/mcp-typescript-*/test-mcp-typescript-record.txt")
		recordPaths = append(recordPaths, files...)
	case "mockserver":
		files, _ := filepath.Glob("./testSDKs/sdk-mockserver-*/tests/test-mockserver-record.txt")
		recordPaths = append(recordPaths, files...)
	case "php":
		files, _ := filepath.Glob("./testSDKs/sdk-php-*/test-php-record.txt")
		recordPaths = append(recordPaths, files...)
	case "pythonv2":
		files, _ := filepath.Glob("./testSDKs/sdk-pythonv2-*/test-python-record.txt")
		recordPaths = append(recordPaths, files...)
	case "ruby":
		files, _ := filepath.Glob("./testSDKs/sdk-ruby-*/test-ruby-record.txt")
		recordPaths = append(recordPaths, files...)
	case "typescriptv2":
		files, _ := filepath.Glob("./testSDKs/sdk-typescriptv2-*/test-typescript-record.txt")
		recordPaths = append(recordPaths, files...)
	case "unity":
		if os.Getenv("WSL_DISTRO_NAME") != "" {
			files, _ := filepath.Glob("/mnt/c/workspace/sdk-unity-*-tests/test-unity-record.txt")
			recordPaths = append(recordPaths, files...)
		} else {
			files, _ := filepath.Glob("./testSDKs/sdk-unity-*-tests/test-unity-record.txt")
			recordPaths = append(recordPaths, files...)
		}
	}

	records := make(map[string]bool)
	for _, recordPath := range recordPaths {
		data, err := os.ReadFile(recordPath)
		if err != nil {
			return nil, err
		}

		for _, line := range bytes.Split(bytes.ReplaceAll(data, []byte{'\r', '\n'}, []byte{'\n'}), []byte{'\n'}) {
			if len(line) == 0 {
				continue
			}
			records[string(line)] = true
		}
	}

	return records, nil
}
