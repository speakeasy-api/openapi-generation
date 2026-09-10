package main

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/speakeasy-api/openapi-generation/v2/internal/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/changes"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var (
	// Flags
	lang            string
	outputDir       string
	oldPath         string
	newPath         string
	detailLevel     string
	openMainBrowser bool
)

var rootCmd = &cobra.Command{
	Use:   "changes",
	Short: "Compare OpenAPI spec changes between versions",
	Long:  `A tool to diff OpenAPI specifications and show SDK changes between versions.`,
	RunE:  runComparison,
}

func init() {
	// Flags
	rootCmd.Flags().StringVar(&lang, "lang", "typescript", "Target language")
	rootCmd.Flags().StringVar(&outputDir, "output-dir", "/tmp/debug-changelog", "Output directory for saving specs and changelog")
	rootCmd.Flags().StringVar(&oldPath, "old", "", "Path to the old OpenAPI spec (required)")
	rootCmd.Flags().StringVar(&newPath, "new", "", "Path to the new OpenAPI spec (required)")
	rootCmd.Flags().StringVar(&detailLevel, "detail-level", "compact", "Detail level for changelog output (compact or full)")
	rootCmd.Flags().BoolVar(&openMainBrowser, "open", false, "Open the HTML changelog in default browser")
	_ = rootCmd.MarkFlagRequired("old")
	_ = rootCmd.MarkFlagRequired("new")

	// Add subcommands
	rootCmd.AddCommand(registryCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runComparison(cmd *cobra.Command, args []string) error {
	logger := logging.NewLogger(logging.LevelFromEnv())

	// Create output directory
	if err := os.MkdirAll(outputDir, 0o755); err != nil {
		return fmt.Errorf("error creating output directory: %w", err)
	}

	// Copy specs to output directory (reformatted for easier debugging)
	oldSpecContent, err := os.ReadFile(oldPath)
	if err != nil {
		return fmt.Errorf("error reading old spec: %w", err)
	}
	newSpecContent, err := os.ReadFile(newPath)
	if err != nil {
		return fmt.Errorf("error reading new spec: %w", err)
	}

	// Reformat YAML for easier debugging
	oldFormatted, err := reformatYAML(oldSpecContent)
	if err != nil {
		return fmt.Errorf("error reformatting old spec: %w", err)
	}
	newFormatted, err := reformatYAML(newSpecContent)
	if err != nil {
		return fmt.Errorf("error reformatting new spec: %w", err)
	}

	if err := os.WriteFile(filepath.Join(outputDir, "old.openapi.yaml"), oldFormatted, 0o644); err != nil {
		return fmt.Errorf("error writing old spec: %w", err)
	}
	if err := os.WriteFile(filepath.Join(outputDir, "new.openapi.yaml"), newFormatted, 0o644); err != nil {
		return fmt.Errorf("error writing new spec: %w", err)
	}

	// Create GenerateOptions for both specs
	oldOptions := generate.GenerateOptions{
		Lang:       lang,
		SchemaPath: oldPath,
		OutDir:     outputDir,
		Logger:     logger,
		Verbose:    false,
	}

	newOptions := generate.GenerateOptions{
		Lang:       lang,
		SchemaPath: newPath,
		OutDir:     outputDir,
		Logger:     logger,
		Verbose:    false,
	}

	// Get the differences using the Changes function
	fmt.Fprintf(os.Stderr, "Generating diff...\n")
	context := context.Background()
	diff, err := changes.Changes(context, oldOptions, newOptions)
	if err != nil {
		return fmt.Errorf("failed to get changes. error: %s", err.Error())
	}

	// Generate and save markdown changelog
	changelogContent := changes.ToMarkdown(diff, changes.DetailLevel(detailLevel))
	if err := os.WriteFile(filepath.Join(outputDir, "changelog.md"), []byte(changelogContent), 0o644); err != nil {
		return fmt.Errorf("error writing changelog: %w", err)
	}

	// Generate and save HTML changelog
	htmlContent := changes.ToHTML(diff)
	htmlPath := filepath.Join(outputDir, "changelog.html")
	if err := os.WriteFile(htmlPath, htmlContent, 0o644); err != nil {
		return fmt.Errorf("error writing HTML changelog: %w", err)
	}

	// Open in browser if requested
	if openMainBrowser {
		if err := openInMainBrowser(htmlPath); err != nil {
			logger.Warn(fmt.Sprintf("Failed to open browser: %v", err))
		}
	}

	// Print the diff as JSON
	fmt.Printf("Changes between %s and %s:\n", oldPath, newPath)

	jsonOutput, err := diff.ToJSON()
	if err != nil {
		return fmt.Errorf("error converting diff to JSON: %w", err)
	}

	fmt.Println(string(jsonOutput))
	fmt.Fprintf(os.Stderr, "\nFiles saved to: %s\n", outputDir)
	fmt.Fprintf(os.Stderr, "HTML changelog: %s\n", htmlPath)
	return nil
}

func openInMainBrowser(path string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "darwin":
		cmd = "open"
		args = []string{path}
	case "linux":
		cmd = "xdg-open"
		args = []string{path}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", path}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	return exec.Command(cmd, args...).Start()
}

// reformatYAML parses and re-encodes YAML with consistent indentation for easier debugging
func reformatYAML(content []byte) ([]byte, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(content, &node); err != nil {
		return nil, err
	}

	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(2)
	if err := encoder.Encode(&node); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
