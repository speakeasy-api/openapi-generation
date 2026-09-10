package openapigeneration

import (
	_ "embed"
	"regexp"
	"strings"

	"github.com/speakeasy-api/openapi/sequencedmap"
)

//go:embed CHANGELOG.md
var ChangeLog string

type Option func(*options)

func WithPreviousVersion(version string) Option {
	return func(o *options) {
		o.previousVersion = version
	}
}

func WithTargetVersion(version string) Option {
	return func(o *options) {
		o.targetVersion = version
	}
}

func WithSpecificVersion(version string) Option {
	return func(o *options) {
		o.specificVersion = version
	}
}

type options struct {
	previousVersion string
	targetVersion   string
	specificVersion string
}

func GetChangeLog(opts ...Option) string {
	o := &options{}
	for _, opt := range opts {
		opt(o)
	}

	versions := getChangelogVersions()

	targetVersionFound := false

	sections := []string{}

	for version, section := range versions.All() {
		if version == o.specificVersion || (version == o.targetVersion && o.specificVersion == "") || (o.targetVersion == "" && o.specificVersion == "") {
			targetVersionFound = true
		}

		if o.previousVersion != "" && version == o.previousVersion && o.specificVersion == "" {
			break
		}

		if targetVersionFound {
			sections = append(sections, section)

			if o.specificVersion != "" {
				break
			}
		}
	}

	return strings.Join(sections, "\n\n")
}

func GetLatestVersion() string {
	versions := getChangelogVersions()

	return versions.First().GetKey()
}

var versionMatchRegex = regexp.MustCompile(`^\[(v[0-9]+\.[0-9]+\.[0-9]+)\]`)

func getChangelogVersions() *sequencedmap.Map[string, string] {
	changelog := strings.ReplaceAll(ChangeLog, "\r\n", "\n")
	bodyParts := strings.Split(changelog, "\n\n\n[")

	versions := sequencedmap.New[string, string]()

	if len(bodyParts) != 2 {
		return versions
	}

	footer := "[" + bodyParts[1]
	links := strings.Split(footer, "\n")

	sections := strings.Split(bodyParts[0], "\n\n## ")

	for i, section := range sections {
		if versionMatchRegex.MatchString(section) {
			version := versionMatchRegex.FindStringSubmatch(section)[1]

			block := "## " + section

			linkIdx := len(links) - i
			if linkIdx < len(links) {
				block += "\n\n" + links[linkIdx]
			}

			versions.Set(version, block)
		}
	}

	return versions
}
