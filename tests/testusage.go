package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"go.uber.org/zap/zapcore"
)

var logger = logging.NewLogger(zapcore.DebugLevel)

// Number of usage examples generated in the `USAGE.md` file
//
// This corresponds to the number operations that define the
// `x-speakeasy-usage-example extension` with either:
// - no associated tag
// - the `usage` tag
//
// if the `x-speakeasy-usage-example` extension is not used,
// only a single snippet is expected in `USAGE.md`.
var expectedUsageSnippetCount = 3

func testUsage(lang, group string, all bool) {
	testUsagePath := fmt.Sprintf("./testusages/%s/%s", lang, group)

	buildCmd := buildCommand(lang, group, !all)
	if buildCmd == "" {
		logger.Warn("No build command set for testUsage in " + lang)
		return
	}

	// TODO: enable usage example execution for PHP
	if lang != "php" {
		runCommand(testUsagePath, installCommand(lang, group))
	}

	if all {
		for _, mdFile := range collectReadmeFiles(lang, group) {
			exampleCodes := collectUsageExampleCodes(mdFile)
			numCodes := len(exampleCodes)
			if numCodes == 0 {
				continue
			}

			subsdk := filepath.Base(filepath.Dir(mdFile))
			logger.Debug(fmt.Sprintf("Compiling %d usage example(s) from %s", numCodes, subsdk))
			itt := 0
			for _, exampleCode := range collectUsageExampleCodes(mdFile) {
				filename := fmt.Sprintf("%s%d", subsdk, itt)
				outFile := fmt.Sprintf("%s/%s", testUsagePath, getOutputFile(lang, filename))
				writeExampleCode(outFile, exampleCode)
				logger.Debug(fmt.Sprintf("compiling usage %d: %s , working directory: %s", itt, buildCmd, testUsagePath))
				runCommand(testUsagePath, buildCmd)
				itt++
			}
		}
	} else {
		usageFile := sdkDirectory(lang, group) + "/USAGE.md"
		exampleCodes := collectUsageExampleCodes(usageFile)
		numCodes := len(exampleCodes)

		expected := expectedUsageSnippetCount
		if lang == "pythonv2" {
			expected *= 2 // python has two snippets per usage example due to sync/async methods
		}

		if numCodes != expected {
			panic(fmt.Sprintf("Unexpected number of usage examples in %s. Found %d. Expected %d.",
				usageFile, numCodes, expected))
		}

		// TODO: enable usage example execution for PHP
		if lang != "php" {
			logger.Debug(fmt.Sprintf("Running %d usage example(s) from %s", numCodes, usageFile))
			itt := 0
			for _, exampleCode := range exampleCodes {
				outFile := fmt.Sprintf("%s/%s", testUsagePath, getOutputFile(lang, "Program"))
				writeExampleCode(outFile, exampleCode)
				logger.Debug(fmt.Sprintf("running usage %d: %s , working directory: %s", itt, buildCmd, testUsagePath))
				runCommand(testUsagePath, buildCmd)
				itt++
			}
		}
	}
}

func collectUsageExampleCodes(file string) []string {
	data, err := os.ReadFile(file)
	if err != nil {
		panic(err)
	}
	exampleCodeRegex := regexp.MustCompile("(?s)```[a-z]+\\n(.*?)```\\n")
	matches := exampleCodeRegex.FindAllStringSubmatch(string(data), -1)

	exampleCodes := make([]string, 0, len(matches))
	for _, match := range matches {
		exampleCodes = append(exampleCodes, match[1])
	}

	return exampleCodes
}

func collectReadmeFiles(lang, group string) []string {
	mdFilesRegex := regexp.MustCompile(`docs/sdks/[^ ]*/README\.md$`)
	var files []string

	err := filepath.Walk(sdkDirectory(lang, group),
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !info.IsDir() && mdFilesRegex.MatchString(path) {
				files = append(files, path)
			}
			return nil
		})
	if err != nil {
		panic(fmt.Errorf("Could not collect README files in %s.", sdkDirectory(lang, group)))
	}

	return files
}

func getOutputFile(lang, filename string) string {
	switch lang {
	case "go":
		return "main.go"
	case "csharp", "unity":
		return filename + ".cs"
	case "javav2":
		return "src/main/java/hello/world/Application.java"
	case "php":
		return "src/main.php"
	case "pythonv2":
		return "main.py"
	case "ruby":
		return "app.rb"
	case "typescriptv2":
		return "src/index.ts"
	case "postman":
		return ""
	case "cli":
		return "main.go"
	default:
		panic(fmt.Sprintf("unknown language %q when getting output file", lang))
	}
}

func writeExampleCode(outFile, exampleCode string) {
	if err := os.MkdirAll(filepath.Dir(outFile), 0o755); err != nil {
		panic(err)
	}

	if err := os.WriteFile(outFile, []byte(exampleCode), 0o644); err != nil {
		panic(err)
	}
}

func sdkDirectory(lang, group string) string {
	if group == "review" {
		return "./zSDKs/sdk-" + lang
	}

	return fmt.Sprintf("testSDKs/sdk-%s-%s", lang, group)
}

func getPythonPackageManager(lang, group string) string {
	if lang != "pythonv2" {
		return ""
	}

	sdkDir := sdkDirectory(lang, group)
	return detectPythonPackageManagerFromProject(sdkDir)
}

// detectPythonPackageManagerFromProject detects the package manager by examining
// the generated pyproject.toml file structure rather than parsing gen.yaml
func detectPythonPackageManagerFromProject(projectDir string) string {
	pyprojectPath := filepath.Join(projectDir, "pyproject.toml")
	data, err := os.ReadFile(pyprojectPath)
	if err != nil {
		logger.Warn(fmt.Sprintf("Could not read pyproject.toml file at %s, defaulting to uv", pyprojectPath))
		return "uv"
	}

	content := string(data)

	// Poetry projects have [tool.poetry] sections
	if strings.Contains(content, "[tool.poetry]") {
		return "poetry"
	}

	// uv projects use [dependency-groups] instead of [tool.poetry.group.dev.dependencies]
	if strings.Contains(content, "[dependency-groups]") {
		return "uv"
	}

	// Fallback: if neither pattern is found, default to uv
	logger.Warn("Could not detect package manager from pyproject.toml structure, defaulting to uv")
	return "uv"
}

func sourceDirectory(lang, group string) string {
	sdkDir := sdkDirectory(lang, group)

	switch lang {
	case "csharp":
		projectDir := ""
		csprojRegex := regexp.MustCompile(`\.csproj$`)
		err := filepath.Walk(sdkDir,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					panic(err)
				}
				if info.IsDir() && info.Name() == "Tests" {
					return filepath.SkipDir
				}
				if !info.IsDir() && csprojRegex.MatchString(info.Name()) {
					projectDir = filepath.Dir(path)
				}
				return nil
			})
		if err != nil || projectDir == "" {
			panic(fmt.Errorf("Could not find directory containing .csproj file in %s", sdkDirectory(lang, group)))
		}

		return projectDir
	case "php":
		projectDir := ""
		composerJsonRegex := regexp.MustCompile(`composer\.json$`)
		err := filepath.Walk(sdkDir,
			func(path string, info os.FileInfo, err error) error {
				if err != nil {
					panic(err)
				}
				if info.IsDir() && info.Name() == "vendor" {
					return filepath.SkipDir
				}
				if !info.IsDir() && composerJsonRegex.MatchString(info.Name()) {
					projectDir = filepath.Dir(path)
				}
				return nil
			})
		if err != nil || projectDir == "" {
			panic(fmt.Errorf("Could not find directory containing composer.json file in %s", sdkDirectory(lang, group)))
		}

		return projectDir
	}

	return sdkDir
}

func installCommand(lang, group string) string {
	srcDir := sourceDirectory(lang, group)

	switch lang {
	case "csharp":
		return "dotnet add reference ../../../" + srcDir
	case "go":
		return fmt.Sprintf("cp ../../../%s/go.sum .", srcDir)
	case "php":
		file, err := os.Open(srcDir + "/composer.json")
		if err != nil {
			panic(fmt.Errorf("Could not open composer.json. %w", err))
		}
		defer file.Close()

		var data map[string]interface{}
		decoder := json.NewDecoder(file)
		err = decoder.Decode(&data)
		if err != nil {
			panic(fmt.Errorf("Could not parse composer.json. %w", err))
		}

		return fmt.Sprintf(`composer install && composer require "%s @dev"`, data["name"])
	case "pythonv2":
		packageManager := getPythonPackageManager(lang, group)
		if packageManager == "poetry" {
			return "poetry add --editable ../../../" + srcDir
		}
		return "uv add --editable ../../../" + srcDir
	case "typescriptv2":
		return "npm install --no-save --ignore-scripts ../../../" + srcDir
	}

	return ""
}

func buildCommand(lang, group string, run bool) string {
	command := ""

	switch lang {
	case "csharp":
		command = "dotnet build"
		if run {
			command += " && dotnet run"
		}
	case "go":
		command = "export GOROOT= && go build"
		if run {
			command += " && go run ."
		}
	case "javav2":
		command = "gradle build"
		if run {
			command += " && gradle run"
		}
	case "php":
		command = "vendor/bin/phpstan analyse src --level 7 --memory-limit 1G --no-progress"
		if run {
			command += " && php src/main.php"
		}
	case "pythonv2":
		packageManager := getPythonPackageManager(lang, group)
		if packageManager == "poetry" {
			command = "poetry install && poetry run python -m compileall -q . && poetry run pylint main.py"
			if run {
				command += " && poetry run python main.py"
			}
		} else {
			command = "uv sync && uv run python -m compileall -q . && uv run pylint main.py"
			if run {
				command += " && uv run python main.py"
			}
		}
	case "ruby":
		command = "bundle install"
	case "typescriptv2":
		command = "npm install --ignore-scripts && npm run check"

		if run {
			command += " && npm run prepare && node dist/index.js"
		}
	case "unity":
		if run {
			command = "Unity -batchmode -runTests -projectPath ../unity-test-project"
		}
	case "postman":
		command = ""
	case "cli":
		command = ""
	default:
		panic(fmt.Errorf("unknown language %q when getting build command", lang))
	}

	return command
}

func runCommand(path, command string) {
	if command == "" {
		return
	}

	cmd := exec.Command("/bin/sh", "-c", command)
	cmd.Dir = path

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		panic(fmt.Errorf("Failed running command %s. %w", cmd, err))
	}
}
