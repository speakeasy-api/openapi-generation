package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/hashicorp/go-version"
	"github.com/speakeasy-api/openapi-generation/v2/internal/changeset"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	templatesConfig "github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
)

var errFeatureNotSupportedForTarget = errors.New("feature not supported for target")

func main() {
	dryRun := false
	for _, arg := range os.Args[1:] {
		if arg == "--dry-run" {
			dryRun = true
		}
	}

	if err := applyChangesets(dryRun); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func applyChangesets(dryRun bool) error {
	changesets, err := changeset.ReadAll()
	if err != nil {
		return err
	}

	if len(changesets) == 0 {
		fmt.Println("No changesets to apply")
		return nil
	}

	fmt.Printf("Found %d changeset(s) to apply\n", len(changesets))

	supportedTargets := templatesConfig.GetAvailableTemplates()
	supportedFeatures := features.GetAllFeatures()

	for _, cs := range changesets {
		for _, targetName := range cs.Targets {
			if !slices.Contains(supportedTargets, targetName) {
				return fmt.Errorf("invalid target %q in changeset %s", targetName, cs.ID)
			}

			t := types.NewTargetFromTemplate(targetName)

			for _, featureName := range cs.Features {
				feature := features.FeatureFromString(featureName)
				if !slices.Contains(supportedFeatures, feature) {
					return fmt.Errorf("invalid feature %q in changeset %s", featureName, cs.ID)
				}

				currentVersion, err := readFeatureVersion(feature, t)
				if err != nil {
					if errors.Is(err, errFeatureNotSupportedForTarget) {
						fmt.Printf("  Skipping %s for %s (not supported)\n", featureName, targetName)
						continue
					}

					return fmt.Errorf("error reading version for %s on %s: %w", featureName, targetName, err)
				}

				v, err := version.NewVersion(currentVersion)
				if err != nil {
					return fmt.Errorf("error parsing version %s for feature %s on %s: %w", currentVersion, featureName, targetName, err)
				}

				major := v.Segments()[0]
				minor := v.Segments()[1]
				patch := v.Segments()[2]

				switch cs.Bump {
				case "major":
					major++
					minor = 0
					patch = 0
				case "minor":
					minor++
					patch = 0
				default:
					patch++
				}

				newVersion := fmt.Sprintf("%d.%d.%d", major, minor, patch)

				if err := bumpFeatureVersion(feature, t, newVersion, dryRun); err != nil {
					return fmt.Errorf("error bumping version for %s on %s: %w", featureName, targetName, err)
				}

				if err := bumpReviewSDKFeatureVersion(feature, t, newVersion, dryRun); err != nil {
					return fmt.Errorf("error bumping review gen.lock for %s on %s: %w", featureName, targetName, err)
				}

				entry := createChangelogEntry(cs, feature, newVersion)
				changelogFile := fmt.Sprintf("changelogs/%s/%s-%s.md", t.Template, feature, newVersion)

				if dryRun {
					if _, err := os.Stat(filepath.Dir(changelogFile)); err != nil {
						return fmt.Errorf("changelog directory missing for %s: %w", targetName, err)
					}

					fmt.Printf("  [dry-run] Would bump %s for %s: %s -> %s\n", featureName, targetName, currentVersion, newVersion)
					continue
				}

				if err := os.WriteFile(changelogFile, []byte(entry), 0o644); err != nil {
					return fmt.Errorf("error writing changelog %s: %w", changelogFile, err)
				}

				fmt.Printf("  %s for %s: %s -> %s (%s)\n", featureName, targetName, currentVersion, newVersion, changelogFile)
			}
		}
	}

	if !dryRun {
		if err := changeset.DeleteAll(); err != nil {
			return err
		}
		fmt.Println("Deleted consumed changesets")
	}

	return nil
}

func readFeatureVersion(feature features.Feature, target types.Target) (string, error) {
	file := fmt.Sprintf("templates/templates/%s/features.ts", target.Template)

	data, err := os.ReadFile(file)
	if err != nil {
		return "", fmt.Errorf("error reading features file: %w", err)
	}

	r := regexp.MustCompile(fmt.Sprintf(`%s: "([0-9]+\.[0-9]+\.[0-9]+)"`, feature))
	matches := r.FindStringSubmatch(string(data))
	if len(matches) < 2 {
		return "", fmt.Errorf("%w: feature %s not found in %s", errFeatureNotSupportedForTarget, feature, file)
	}

	return matches[1], nil
}

func bumpFeatureVersion(feature features.Feature, target types.Target, newVersion string, dryRun bool) error {
	versionMatchRegex := fmt.Sprintf(`%s: "(.*?)",`, feature)

	file := fmt.Sprintf("templates/templates/%s/features.ts", target.Template)

	featureData, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error reading feature file: %w", err)
	}

	r := regexp.MustCompile(versionMatchRegex)
	updated := r.ReplaceAllString(string(featureData), fmt.Sprintf(`%s: "%s",`, feature, newVersion))

	if dryRun {
		return nil
	}

	if err := os.WriteFile(file, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("error writing feature file: %w", err)
	}

	return nil
}

func bumpReviewSDKFeatureVersion(feature features.Feature, target types.Target, newVersion string, dryRun bool) error {
	file, ok := reviewGenLockPath(target.Template)
	if !ok {
		return nil
	}

	data, err := os.ReadFile(file)
	if err != nil {
		return fmt.Errorf("error reading review gen.lock file: %w", err)
	}

	updated := replaceReviewSDKFeatureVersion(string(data), target.Target, feature, newVersion)

	if dryRun {
		return nil
	}

	if err := os.WriteFile(file, []byte(updated), 0o644); err != nil {
		return fmt.Errorf("error writing review gen.lock file: %w", err)
	}

	return nil
}

func reviewGenLockPath(template string) (string, bool) {
	switch template {
	case "mcp-typescript":
		return "zSDKs/mcp-typescript/.speakeasy/gen.lock", true
	case "mockserver":
		return "", false
	case "terraform":
		return "zSDKs/terraform-provider-testing/.speakeasy/gen.lock", true
	default:
		return fmt.Sprintf("zSDKs/sdk-%s/.speakeasy/gen.lock", template), true
	}
}

func replaceReviewSDKFeatureVersion(genLock string, target string, feature features.Feature, newVersion string) string {
	lines := strings.Split(genLock, "\n")
	inFeatures := false
	inTarget := false

	for i, line := range lines {
		trimmed := strings.TrimSpace(line)

		if !inFeatures {
			if trimmed == "features:" {
				inFeatures = true
			}

			continue
		}

		if !strings.HasPrefix(line, " ") && trimmed != "" {
			break
		}

		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "    ") && strings.HasSuffix(trimmed, ":") {
			inTarget = strings.TrimSuffix(trimmed, ":") == target
			continue
		}

		if !inTarget || !strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "      ") {
			continue
		}

		key, _, ok := strings.Cut(trimmed, ": ")
		if !ok || key != feature.String() {
			continue
		}

		lines[i] = fmt.Sprintf("    %s: %s", feature, newVersion)
		return strings.Join(lines, "\n")
	}

	return genLock
}

var changeTypes = map[string]struct {
	header string
	icon   string
}{
	"feat":     {header: "New Features", icon: ":bee:"},
	"feature":  {header: "New Features", icon: ":bee:"},
	"fix":      {header: "Bug Fixes", icon: ":bug:"},
	"bugfix":   {header: "Bug Fixes", icon: ":bug:"},
	"perf":     {header: "Performance Improvements", icon: ":zap:"},
	"refactor": {header: "Refactors", icon: ":recycle:"},
	"test":     {header: "Tests", icon: ":white_check_mark:"},
	"tests":    {header: "Tests", icon: ":white_check_mark:"},
	"build":    {header: "Build System", icon: ":construction_worker:"},
	"ci":       {header: "Build System", icon: ":construction_worker:"},
	"doc":      {header: "Documentation Changes", icon: ":memo:"},
	"docs":     {header: "Documentation Changes", icon: ":memo:"},
	"style":    {header: "Code Style Changes", icon: ":art:"},
	"chore":    {header: "Chores", icon: ":wrench:"},
	"other":    {header: "Other Changes", icon: ":flying_saucer:"},
}

func createChangelogEntry(cs *changeset.Changeset, feature features.Feature, ver string) string {
	entry := fmt.Sprintf("## %s: %s - %s\n", feature, ver, cs.Date)

	info, ok := changeTypes[cs.Type]
	if !ok {
		info = changeTypes["other"]
	}

	entry += fmt.Sprintf("### %s %s\n", info.icon, info.header)
	entry += fmt.Sprintf("- %s *(commit by [@%s](https://github.com/%s))*\n", cs.Description, cs.Author, cs.Author)

	return entry
}
