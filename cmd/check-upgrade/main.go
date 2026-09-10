// check-upgrade analyzes a gen.yaml file from an SDK repository and compares
// its values against the defaults for new SDKs, identifying which settings
// differ from the recommended defaults.
package main

import (
	"flag"
	"fmt"
	"os"
	"sort"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/checkupgrade"
)

// ANSI color codes
const (
	colorRed   = "\033[31m"
	colorGreen = "\033[32m"
	colorReset = "\033[0m"
)

func main() {
	repoPath := flag.String("repo", "", "path to the SDK repository containing gen.yaml")
	showAll := flag.Bool("all", false, "show all config values, not just differences")
	flag.Parse()

	if *repoPath == "" {
		fmt.Fprintln(os.Stderr, "Usage: check-upgrade -repo <path-to-sdk-repo> [-all]")
		fmt.Fprintln(os.Stderr, "")
		fmt.Fprintln(os.Stderr, "This command finds the gen.yaml in the given repository and compares")
		fmt.Fprintln(os.Stderr, "its values against the defaults for new SDKs.")
		os.Exit(1)
	}

	result, err := checkupgrade.Check(*repoPath, checkupgrade.Options{
		IncludeMatches: *showAll,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Checking gen.yaml at: %s\n\n", result.GenYamlPath)

	// Print generation section
	if result.Generation != nil {
		fmt.Println("=== Generation Section ===")
		printSection(result.Generation, *showAll)
		fmt.Println()
	}

	// Print language sections in sorted order
	targets := make([]string, 0, len(result.Languages))
	for target := range result.Languages {
		targets = append(targets, target)
	}
	sort.Strings(targets)

	for _, target := range targets {
		section := result.Languages[target]
		fmt.Printf("=== %s Section ===\n", target)
		printSection(section, *showAll)
		fmt.Println()
	}
}

func printSection(section *checkupgrade.SectionResult, showAll bool) {
	if len(section.Differences) > 0 {
		fmt.Println("Differences from newSDK defaults:")
		for i, diff := range section.Differences {
			if i > 0 {
				fmt.Println() // Add newline between config keys
			}
			fmt.Printf("  %s:\n", diff.Key)
			if diff.Description != "" {
				fmt.Printf("    %s\n", diff.Description)
			}
			fmt.Printf("    %scurrent:  %v%s\n", colorRed, checkupgrade.FormatValue(diff.CurrentValue), colorReset)
			fmt.Printf("    %snewSDK:   %v%s\n", colorGreen, checkupgrade.FormatValue(diff.NewSDKValue), colorReset)
		}
	} else {
		fmt.Println(colorGreen + "All values match newSDK defaults" + colorReset)
	}

	if showAll && len(section.Matches) > 0 {
		fmt.Println("\nMatching values:")
		for _, match := range section.Matches {
			if match.CurrentValue != nil {
				fmt.Printf("  %s: %v (matches default)\n", match.Key, checkupgrade.FormatValue(match.CurrentValue))
			} else {
				fmt.Printf("  %s: (not set, using default: %v)\n", match.Key, checkupgrade.FormatValue(match.NewSDKValue))
			}
		}
	}
}
