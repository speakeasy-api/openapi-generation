package changelogs

import (
	"context"
	"embed"
	"fmt"
	"io"
	"maps"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/templates"
	"github.com/speakeasy-api/openapi/sequencedmap"

	"github.com/hashicorp/go-version"
	"github.com/speakeasy-api/openapi-generation/v2/internal/features"
)

//go:embed **/*.md
var changeLogs embed.FS

var (
	SplitRegex        = regexp.MustCompile(`(?m)^## `)
	HeaderRegex       = regexp.MustCompile(`(?m)^## (.*?) - (\d{4}-\d{2}-\d{2})`)
	VersionMatchRegex = regexp.MustCompile(`(?:##|//) ([a-zA-Z0-9]+): ([0-9]+\.[0-9]+\.[0-9]+)`)
)

func GetLatestVersions(target string) (map[string]string, error) {
	ctx := context.Background()

	t, err := templates.GetTargetFromTargetString(target)
	if err != nil {
		return nil, fmt.Errorf("failed to get target from target string %s: %w", target, err)
	}

	f, err := features.New(t, map[string]any{})
	if err != nil {
		return nil, fmt.Errorf("failed to initialize features for language %s: %w", t.Target, err)
	}

	versions := map[string]string{}

	featureList := f.GetImplementedFeaturesList(ctx)
	for _, feature := range featureList {
		versions[feature.String()] = f.GetFeatureVersion(ctx, feature)
	}

	return versions, nil
}

func GetChangeLog(lang string, targetVersions, previousVersions map[string]string) (string, error) {
	sections, err := getFeatureSections(lang, targetVersions, previousVersions)
	if err != nil {
		return "", err
	}

	return strings.Join(sections, "\n\n"), nil
}

func GetTemplateChangeLog(templateVersion string, targetVersions, previousVersions map[string]string) (string, error) {
	sections, err := getFeatureSections(templateVersion, targetVersions, previousVersions)
	if err != nil {
		return "", err
	}

	return strings.Join(sections, "\n\n"), nil
}

func getFeatureSections(lang string, targetVersions map[string]string, previousVersions map[string]string) ([]string, error) {
	changelog, err := CompileChangelog(lang)
	if err != nil {
		return nil, fmt.Errorf("failed to compile changelog for language %s: %w", lang, err)
	}

	versions := getChangelogVersions(changelog)

	sections := []string{}
	sortedFeatures := slices.Collect(maps.Keys(targetVersions))
	slices.Sort(sortedFeatures)

	for _, feature := range sortedFeatures {
		if _, ok := versions[feature]; !ok {
			continue
		}

		featureVersions := versions[feature].(*sequencedmap.Map[string, string])

		targetVersion := targetVersions[feature]
		targetVersionFound := false

		for version, section := range featureVersions.All() {
			if version == targetVersion {
				targetVersionFound = true
			}

			if previousVersions != nil && previousVersions[feature] != "" && version == previousVersions[feature] {
				break
			}

			if targetVersionFound {
				sections = append(sections, section)

				if previousVersions == nil || previousVersions[feature] == "" {
					break
				}
			}
		}
	}
	return sections, nil
}

// getChangelogFolders returns all changelog directories that exist for a given
// language. For example, for lang "python", it will return ["pythonv2"] if only
// pythonv2/ exists, or ["python", "pythonv2"] if both exist.
func getChangelogFolders(lang string) []string {
	otherTemplateVersions := templates.GetAvailableTemplates()
	candidates := make([]string, 0, len(otherTemplateVersions)+1)
	candidates = append(candidates, lang)

	for _, otherTemplate := range otherTemplateVersions {
		if strings.HasPrefix(otherTemplate, lang) && otherTemplate != lang {
			candidates = append(candidates, otherTemplate)
		}
	}

	folders := make([]string, 0, len(candidates))

	for _, folder := range candidates {
		if _, err := changeLogs.ReadDir(folder); err == nil {
			folders = append(folders, folder)
		}
	}

	return folders
}

func getChangelogVersions(languageChangelog string) map[string]any {
	versions := map[string]any{}

	sections := SplitRegex.Split(languageChangelog, -1)

	for _, section := range sections {
		section = "## " + section

		if VersionMatchRegex.MatchString(section) {
			matches := VersionMatchRegex.FindAllStringSubmatch(section, -1)

			for _, subMatch := range matches {
				feature := subMatch[1]
				version := subMatch[2]

				if _, ok := versions[feature]; !ok {
					versions[feature] = sequencedmap.New[string, string]()
				}

				versions[feature].(*sequencedmap.Map[string, string]).Set(version, section)
			}
		}
	}

	return versions
}

func CompileChangelog(lang string) (string, error) {
	folders := getChangelogFolders(lang)
	if len(folders) == 0 {
		return "", fmt.Errorf("no changelogs found for language %s", lang)
	}

	files := make([]string, 0, len(folders))

	for _, folder := range folders {
		templateFiles, _ := changeLogs.ReadDir(folder)
		for _, file := range templateFiles {
			files = append(files, fmt.Sprintf("%s/%s", folder, file.Name()))
		}
	}

	// combine with other templates to get all available versions

	allEntries := []string{}
	alreadyRead := map[string]bool{}

	for _, path := range files {
		f, err := changeLogs.Open(path)
		if err != nil {
			return "", fmt.Errorf("failed to open changelog file %s: %w", path, err)
		}

		data, err := io.ReadAll(f)
		if err != nil {
			return "", fmt.Errorf("failed to read changelog file %s: %w", path, err)
		}
		if _, ok := alreadyRead[string(data)]; ok {
			continue
		} else {
			alreadyRead[string(data)] = true
		}

		fileEntries := SplitRegex.Split(string(data), -1)

		for _, entry := range fileEntries {
			if entry == "" {
				continue
			}
			subEntry := "## " + strings.TrimSpace(entry)
			matchesHeader := HeaderRegex.FindStringSubmatch(subEntry)
			if len(matchesHeader) < 3 {
				return "", fmt.Errorf("failed to parse header (%v) in %v for changelog entry: \"%s\"", HeaderRegex, path, subEntry)
			}
			allEntries = append(allEntries, subEntry)
		}
	}

	slices.SortFunc(allEntries, func(a, b string) int {
		aMatches := HeaderRegex.FindStringSubmatch(a)
		bMatches := HeaderRegex.FindStringSubmatch(b)

		aTime, aErr := time.Parse("2006-01-02", aMatches[2])
		bTime, bErr := time.Parse("2006-01-02", bMatches[2])

		if aErr == nil && bErr == nil {
			if aTime.Before(bTime) {
				return 1
			} else if aTime.After(bTime) {
				return -1
			}
		}

		aFeatures := strings.Split(aMatches[1], " // ")
		bFeatures := strings.Split(bMatches[1], " // ")

		for _, aFeature := range aFeatures {
			aParts := strings.Split(aFeature, ": ")
			if len(aParts) != 2 {
				continue
			}

			for _, bFeature := range bFeatures {
				bParts := strings.Split(bFeature, ": ")
				if len(bParts) != 2 {
					continue
				}

				if aParts[0] == bParts[0] {
					aV, aErr := version.NewVersion(aParts[1])
					bV, bErr := version.NewVersion(bParts[1])

					if aErr == nil && bErr == nil {
						if aV.LessThan(bV) {
							return 1
						} else if aV.GreaterThan(bV) {
							return -1
						}
					}
				}
			}
		}

		return strings.Compare(aMatches[1], bMatches[1])
	})

	return strings.Join(allEntries, "\n\n"), nil
}

func MustGenerate(template string, targetVersions map[string]string, previousVersions map[string]string) (bool, error) {
	sections, err := getFeatureSections(template, targetVersions, previousVersions)
	if err != nil {
		return false, err
	}

	for _, section := range sections {
		// if a section has [force-gen] within, it is important to generate this target.
		if strings.Contains(section, "[force-gen]") {
			return true, nil
		}
	}
	return false, nil
}

func ChangelogEntriesBetweenVersions(lang string, feature string, newVersion string, oldVersion string) ([]string, error) {
	folders := getChangelogFolders(lang)
	if len(folders) == 0 {
		return nil, fmt.Errorf("no changelogs found for language %s", lang)
	}

	returnValues := []string{}
	for _, folder := range folders {
		files, _ := changeLogs.ReadDir(folder)
		for _, file := range files {
			fileName := file.Name()
			// split it into feature + version
			fileParts := strings.Split(fileName, "-")
			if len(fileParts) != 2 {
				continue
			}
			featurePart := fileParts[0]
			// versionPart is fileParts[1] with the extension trimmed
			versionPart, _ := strings.CutSuffix(fileParts[1], ".md")
			if featurePart != feature {
				continue
			}
			// if semver is between (inclusive) newVersion and oldVersion, add it to the return values
			newSemVer, err := version.NewVersion(newVersion)
			if err != nil {
				return nil, err
			}
			oldSemVer, err := version.NewVersion(oldVersion)
			if err != nil {
				return nil, err
			}
			fileSemVer, err := version.NewVersion(versionPart)
			if err != nil {
				return nil, err
			}
			if fileSemVer.GreaterThanOrEqual(oldSemVer) && fileSemVer.LessThanOrEqual(newSemVer) {
				returnValues = append(returnValues, fileName)
			}
		}
	}
	return returnValues, nil
}
