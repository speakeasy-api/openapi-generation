// cmd/regen provides an ergonomic entry point for regenerating an already-bootstrapped SDK.
//
// Instead of manually specifying -s, -o, and -l flags, point it at an SDK directory
// and it infers the language from gen.yaml and the schema from workflow.yaml or an
// openapi* file in the directory root.
//
// Usage:
//
//	go run cmd/regen/main.go [flags] [dir]
//
// Examples:
//
//	go run cmd/regen/main.go ~/sdks/my-go-sdk
//	go run cmd/regen/main.go -s override.yaml .
//	go run cmd/regen/main.go --set go.version=2.0.0 ~/sdks/my-go-sdk
//	go run cmd/regen/main.go --set generation.sdkClassName=MySDK --set go.flattenGlobalSecurity=true .
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	_ "github.com/KimMachineGun/automemlimit"
	"github.com/speakeasy-api/openapi-generation/v2/internal/generate"
	"github.com/speakeasy-api/openapi-generation/v2/internal/runner"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/speakeasy-api/sdk-gen-config/workflow"
	"gopkg.in/yaml.v3"
)

var logger = logging.NewLogger(logging.LevelFromEnv())

// setFlag collects repeated --set key=value pairs.
type setFlag []string

func (s *setFlag) String() string { return strings.Join(*s, ", ") }
func (s *setFlag) Set(v string) error {
	*s = append(*s, v)
	return nil
}

func main() {
	// Flags that override inferred values
	schemaOverride := flag.String("s", "", "override: path to the schema file (inferred from workflow.yaml or openapi* file)")
	langOverride := flag.String("l", "", "override: language to generate (inferred from gen.yaml)")

	// Pass-through flags
	testGroup := flag.String("t", "", "test group to generate")
	usageGroup := flag.String("u", "", "usage snippet namespace (or 'all')")
	published := flag.Bool("p", false, "is the SDK published")
	installationURL := flag.String("i", "", "installation url for the SDK")
	repoURL := flag.String("r", "", "repo url for the SDK")
	repoSubDirectory := flag.String("b", "", "repo subdirectory for the SDK")
	skipCompile := flag.Bool("skip-compile", false, "skip compiling the SDK")
	traceDest := flag.String("trace", "", "enable OpenTelemetry tracing (stderr, grpc, file://<path>)")
	verbose := flag.Bool("v", false, "verbose output")
	watchTemplatesLocation := flag.String("watch-templates-location", "", "watch templates directory for changes")
	watchTemplates := flag.Bool("watch-templates", false, "watch templates (requires WATCH_TEMPLATES_LOCATION env)")
	profile := flag.String("profile", "", "save profiling data to the given directory")
	validateIntegrity := flag.Bool("validate-integrity", false, "after generation, validate that gen.lock checksums match actual file contents on disk")
	licenseTokenPath := flag.String("license-token", "", "path to a Speakeasy license token file (a registry-signed JWT); a valid token generates with authenticated commercial access. Defaults to the SPEAKEASY_LICENSE_TOKEN environment variable (the raw token) when unset")
	licenseElection := flag.String("license", "", "license election for generated output: 'agpl-3.0-only' accepts AGPL-3.0-only licensing, 'commercial' asserts that a license token must be present. Defaults to the SPEAKEASY_GENERATED_LICENSE environment variable. Without an election a valid license token is required — output is never silently licensed")

	// gen.yaml patching
	var sets setFlag
	flag.Var(&sets, "set", "patch gen.yaml value (e.g. --set go.version=2.0.0 --set generation.sdkClassName=Foo)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: regen [flags] [dir]\n\n")
		fmt.Fprintf(os.Stderr, "Regenerate an already-bootstrapped SDK. Language and schema are inferred\n")
		fmt.Fprintf(os.Stderr, "from the gen.yaml and workflow.yaml in the target directory.\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	// Positional arg: SDK directory (defaults to ".")
	outDir := "."
	if flag.NArg() > 0 {
		outDir = flag.Arg(0)
	}
	outDir, err := filepath.Abs(outDir)
	if err != nil {
		logger.Fatal("failed to resolve directory: " + err.Error())
	}

	// --- Infer language from gen.yaml ---
	lang := inferLang(outDir, langOverride)

	// --- Infer schema path ---
	schemaPath := inferSchema(outDir, lang, schemaOverride)

	// --- Apply --set patches to gen.yaml ---
	if len(sets) > 0 {
		applyGenYAMLPatches(outDir, sets)
	}

	// --- Print what we inferred ---
	logger.Info(fmt.Sprintf("regen: dir=%s lang=%s schema=%s", outDir, lang, schemaPath))
	if len(sets) > 0 {
		logger.Info(fmt.Sprintf("regen: applied %d gen.yaml patch(es)", len(sets)))
	}

	// --- Delegate to shared runner ---
	runner.Run(runner.RunConfig{
		SpanName:                 "regen",
		TraceDest:                runner.ResolveTraceDest(*traceDest),
		ProfileDir:               *profile,
		ValidateIntegrity:        *validateIntegrity,
		Logger:                   logger,
		GenerationContextFactory: runner.ResolveGenerationContextFactory(*licenseElection, *licenseTokenPath, os.Getenv),
		GenOpts: generate.GenerateOptions{
			Lang:                   lang,
			SchemaPath:             schemaPath,
			OutDir:                 outDir,
			TestGroup:              *testGroup,
			UsageGroup:             *usageGroup,
			Published:              *published,
			InstallationURL:        *installationURL,
			RepoURL:                *repoURL,
			RepoSubDirectory:       *repoSubDirectory,
			SkipCompile:            *skipCompile,
			Verbose:                *verbose,
			WatchTemplatesLocation: runner.ResolveWatchTemplatesDir(*watchTemplatesLocation, *watchTemplates, logger),
		},
	})
}

// inferLang determines the target language. If an explicit override is given, use it.
// Otherwise, parse gen.yaml and find the language key (the single top-level key that
// isn't "configVersion" or "generation").
func inferLang(outDir string, override *string) string {
	if override != nil && *override != "" {
		return *override
	}

	configRes, err := config.FindConfigFile(outDir, nil)
	if err != nil || configRes.Data == nil {
		logger.Fatal("No gen.yaml found in " + outDir + " — is this a bootstrapped SDK directory? Use -l to specify language explicitly.")
	}

	var cfg config.Configuration
	if err := yaml.Unmarshal(configRes.Data, &cfg); err != nil {
		logger.Fatal("Failed to parse gen.yaml: " + err.Error())
	}

	for lang := range cfg.Languages {
		return lang
	}

	logger.Fatal("gen.yaml has no language section. Use -l to specify language explicitly.")
	return "" // unreachable
}

// inferSchema determines the schema path. Priority:
//  1. Explicit -s override
//  2. workflow.yaml source for the target language
//  3. An openapi* file in the SDK directory root
func inferSchema(outDir, lang string, override *string) string {
	if override != nil && *override != "" {
		return *override
	}

	// Try workflow.yaml
	if path := schemaFromWorkflow(outDir, lang); path != "" {
		return path
	}

	// Scan for openapi* files in the SDK directory and up to 3 parent directories.
	// This handles the common layout where the spec sits alongside SDK dirs:
	//   minimal-sdks/
	//     openapi.yaml
	//     sdk-go/
	//     sdk-python/
	if path := findOpenAPIFile(outDir, 3); path != "" {
		return path
	}

	logger.Fatal("Could not infer schema path. No workflow.yaml or openapi* file found in " + outDir + " or parent directories. Use -s to specify explicitly.")
	return "" // unreachable
}

// findOpenAPIFile searches dir and up to maxUp parent directories for an openapi* file.
func findOpenAPIFile(dir string, maxUp int) string {
	for i := 0; i <= maxUp; i++ {
		entries, err := os.ReadDir(dir)
		if err != nil {
			break
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasPrefix(e.Name(), "openapi") {
				return filepath.Join(dir, e.Name())
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break // reached filesystem root
		}
		dir = parent
	}
	return ""
}

// schemaFromWorkflow tries to load workflow.yaml and find the source spec for the given language.
func schemaFromWorkflow(outDir, lang string) string {
	// Only check within the SDK directory itself — don't recurse up to parent dirs
	// (workflow.Load uses recursive search which could find ~/.speakeasy/workflow.yaml)
	wfPath := filepath.Join(outDir, ".speakeasy", "workflow.yaml")
	wfData, err := os.ReadFile(wfPath)
	if err != nil {
		// Also try .gen/ directory
		wfPath = filepath.Join(outDir, ".gen", "workflow.yaml")
		wfData, err = os.ReadFile(wfPath)
		if err != nil {
			return "" // no workflow.yaml — that's fine, fall through
		}
	}
	var wf workflow.Workflow
	if err := yaml.Unmarshal(wfData, &wf); err != nil {
		return ""
	}

	wfDir := filepath.Dir(filepath.Dir(wfPath)) // go up from .speakeasy/workflow.yaml to project root

	// Find a target matching our language
	for _, target := range wf.Targets {
		if target.Target != lang {
			continue
		}
		sourceName := target.Source
		source, ok := wf.Sources[sourceName]
		if !ok || len(source.Inputs) == 0 {
			continue
		}
		loc := source.Inputs[0].Location.Resolve()
		if loc == "" {
			continue
		}
		// If it's a relative path, resolve relative to the workflow.yaml's project root
		if !filepath.IsAbs(loc) && !strings.HasPrefix(loc, "http://") && !strings.HasPrefix(loc, "https://") && !strings.HasPrefix(loc, "registry.") {
			loc = filepath.Join(wfDir, loc)
		}
		return loc
	}
	return ""
}

// applyGenYAMLPatches reads gen.yaml, applies --set key=value patches, and writes it back.
// Keys use dot-separated YAML paths, e.g. "generation.sdkClassName" or "go.version".
func applyGenYAMLPatches(outDir string, patches []string) {
	configRes, err := config.FindConfigFile(outDir, nil)
	if err != nil || configRes.Data == nil {
		logger.Fatal("Cannot apply --set: no gen.yaml found in " + outDir)
	}

	// Parse into an ordered tree so we preserve comments/ordering as best we can
	var doc yaml.Node
	if err := yaml.Unmarshal(configRes.Data, &doc); err != nil {
		logger.Fatal("Failed to parse gen.yaml for patching: " + err.Error())
	}

	for _, patch := range patches {
		keyPath, value, ok := strings.Cut(patch, "=")
		if !ok {
			logger.Fatal("Invalid --set syntax (expected key=value): " + patch)
		}
		parts := strings.Split(keyPath, ".")
		if len(parts) == 0 {
			logger.Fatal("Invalid --set key: " + keyPath)
		}
		setYAMLNode(&doc, parts, value)
	}

	out, err := yaml.Marshal(&doc)
	if err != nil {
		logger.Fatal("Failed to marshal patched gen.yaml: " + err.Error())
	}

	if err := os.WriteFile(configRes.Path, out, 0o644); err != nil {
		logger.Fatal("Failed to write patched gen.yaml: " + err.Error())
	}
}

// setYAMLNode navigates into a yaml.Node tree by the given path segments and sets the leaf value.
// Creates intermediate mapping nodes as needed.
func setYAMLNode(root *yaml.Node, path []string, value string) {
	// The root document node wraps the actual content
	node := root
	if node.Kind == yaml.DocumentNode && len(node.Content) > 0 {
		node = node.Content[0]
	}

	for i, key := range path {
		isLast := i == len(path)-1

		if node.Kind != yaml.MappingNode {
			logger.Fatal(fmt.Sprintf("--set path %q: expected mapping at %q, got kind %d", strings.Join(path, "."), key, node.Kind))
		}

		// Search for the key in the mapping's Content pairs
		found := false
		for j := 0; j < len(node.Content)-1; j += 2 {
			if node.Content[j].Value == key {
				if isLast {
					// Overwrite the value node, preserving the existing tag
					// so that integers stay integers, bools stay bools, etc.
					existingTag := node.Content[j+1].Tag
					node.Content[j+1] = &yaml.Node{
						Kind:  yaml.ScalarNode,
						Value: value,
						Tag:   existingTag,
					}
				} else {
					node = node.Content[j+1]
				}
				found = true
				break
			}
		}

		if !found {
			if isLast {
				// Append new key-value pair
				node.Content = append(node.Content,
					&yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"},
					&yaml.Node{Kind: yaml.ScalarNode, Value: value, Tag: inferYAMLTag(value)},
				)
			} else {
				// Create intermediate mapping
				newMap := &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
				node.Content = append(node.Content,
					&yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"},
					newMap,
				)
				node = newMap
			}
		}
	}
}

// inferYAMLTag is a fallback for new keys that don't already exist in gen.yaml.
// For existing keys, setYAMLNode preserves the original tag from the parsed YAML.
// Booleans and integers are detected; floats are treated as strings to avoid
// misclassifying version strings like "2.0".
func inferYAMLTag(v string) string {
	switch strings.ToLower(v) {
	case "true", "false":
		return "!!bool"
	}
	if _, err := strconv.Atoi(v); err == nil {
		return "!!int"
	}
	return "!!str"
}
