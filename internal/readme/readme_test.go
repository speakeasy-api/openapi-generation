package readme_test

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/readme"
	"github.com/stretchr/testify/assert"
)

const (
	defaultPrefix = "<!--"
	defaultSuffix = "-->"
)

func TestFormatSection(t *testing.T) {
	type args struct {
		sectionID    string
		sectionTitle string
		replacement  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "format legacy section",
			args: args{
				sectionTitle: "Legacy Section",
				sectionID:    "",
				replacement:  "## Example Usage\n\n...",
			},
			want: `<!-- Start Legacy Section -->
## Example Usage

...
<!-- End Legacy Section -->`,
		},
		{
			name: "format section with ID",
			args: args{
				sectionTitle: "New Section",
				sectionID:    "new",
				replacement:  "## Example Usage\n\n...",
			},
			want: `<!-- Start New Section [new] -->
## Example Usage

...
<!-- End New Section [new] -->`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents, replaced := readme.ReplaceBlock(
				defaultPrefix,
				"",
				tt.args.sectionID,
				defaultSuffix,
				"",
				tt.args.sectionTitle,
				tt.args.replacement,
			)
			assert.True(t, replaced)
			assert.Equal(t, tt.want, contents)
		})
	}
}

func TestSkipSection(t *testing.T) {
	type args struct {
		contents    string
		legacyTitle string
		sectionID   string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "skip invalid boundaries with no id",
			args: args{
				contents: `# README

This is a README file for the SDK.

<!-- False SDK Installation -->
Invalid Boundaries
<!-- End SDK Installation -->

<!-- Start SDK Example Usage -->
## Example Usage

<!-- Comment -->
<!-- End SDK Example Usage -->
`,
				legacyTitle: "SDK Installation",
				sectionID:   "installation",
			},
		},
		{
			name: "skip invalid boundaries with id",
			args: args{
				contents: `# README

This is a README file for the SDK.

<!-- Start SDK Installation [installation] -->
Invalid Boundaries
<!-- Not SDK Installation [installation] -->

<!-- Start SDK Example Usage -->
## Example Usage

<!-- Comment -->
<!-- End SDK Example Usage -->
`,
				legacyTitle: "SDK Installation",
				sectionID:   "installation",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents, replaced := readme.ReplaceBlock(
				defaultPrefix,
				tt.args.legacyTitle,
				tt.args.sectionID,
				defaultSuffix,
				tt.args.contents,
				"",
				"",
			)
			assert.False(t, replaced)
			assert.Equal(t, tt.args.contents, contents)
		})
	}
}

func TestReplaceSection(t *testing.T) {
	type args struct {
		contents     string
		legacyTitle  string
		sectionTitle string
		sectionID    string
		replacement  string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "replace empty section using title",
			args: args{
				contents: `# README

This is a README file for the SDK.

<!-- Start Section -->
<!-- End Section -->

<!-- Start SDK Installation -->
<!-- End SDK Installation -->

<!-- Start SDK Example Usage -->
## Example Usage

<!-- Comment -->
<!-- End SDK Example Usage -->
`,
				legacyTitle:  "SDK Installation",
				sectionTitle: "New Installation",
				sectionID:    "installation",
				replacement:  "go get github.com/something/something",
			},
			want: `# README

This is a README file for the SDK.

<!-- Start Section -->
<!-- End Section -->

<!-- Start New Installation [installation] -->
go get github.com/something/something
<!-- End New Installation [installation] -->

<!-- Start SDK Example Usage -->
## Example Usage

<!-- Comment -->
<!-- End SDK Example Usage -->
`,
		},
		{
			name: "replace empty section using ID",
			args: args{
				contents: `# README

This is a README file for the SDK.

<!-- Start Section [section] -->
<!-- End Section [section] -->

<!-- Start SDK Installation [installation] -->
<!-- End SDK Installation [installation] -->

<!-- Start SDK Example Usage -->
## Example Usage

  ...
<!-- End SDK Example Usage -->
`,
				legacyTitle:  "Deprecated Title",
				sectionTitle: "A New Title",
				sectionID:    "installation",
				replacement:  "go get github.com/something/something",
			},
			want: `# README

This is a README file for the SDK.

<!-- Start Section [section] -->
<!-- End Section [section] -->

<!-- Start A New Title [installation] -->
go get github.com/something/something
<!-- End A New Title [installation] -->

<!-- Start SDK Example Usage -->
## Example Usage

  ...
<!-- End SDK Example Usage -->
`,
		},
		{
			name: "replace section using title",
			args: args{
				contents: `# README

This is a README file for the SDK.

<!-- Start SDK Installation -->
<!-- End SDK Installation -->

<!-- Start SDK Example Usage -->
## Example Usage

  ...
<!-- End SDK Example Usage -->

<!-- Start Authentication [security] -->
<!-- End Authentication [security] -->
`,
				legacyTitle:  "SDK Example Usage",
				sectionTitle: "NEW Example Usage",
				sectionID:    "usage",
				replacement:  "## New Usage\n\n  ...",
			},
			want: `# README

This is a README file for the SDK.

<!-- Start SDK Installation -->
<!-- End SDK Installation -->

<!-- Start NEW Example Usage [usage] -->
## New Usage

  ...
<!-- End NEW Example Usage [usage] -->

<!-- Start Authentication [security] -->
<!-- End Authentication [security] -->
`,
		},
		{
			name: "replace section using ID",
			args: args{
				contents: `# README

This is a README file for the SDK.

<!-- Start SDK Installation [installation] -->
go get github.com/something/something
<!-- End SDK Installation [installation] -->

<!-- Start SDK Example Usage [usage] -->
## Example Usage

  ...
<!-- End SDK Example Usage [usage] -->

<!-- Start Authentication [security] -->
<!-- End Authentication [security] -->
`,
				legacyTitle:  "Deprecated Title",
				sectionTitle: "A New Title",
				sectionID:    "usage",
				replacement:  "## New Usage\n\n  ...",
			},
			want: `# README

This is a README file for the SDK.

<!-- Start SDK Installation [installation] -->
go get github.com/something/something
<!-- End SDK Installation [installation] -->

<!-- Start A New Title [usage] -->
## New Usage

  ...
<!-- End A New Title [usage] -->

<!-- Start Authentication [security] -->
<!-- End Authentication [security] -->
`,
		},
		{
			name: "replace section with no whitespace",
			args: args{
				contents: `# README

This is a README file for the SDK.

<!-- Start SDK Installation --><!-- End SDK Installation -->

<!-- Start SDK Example Usage -->
## Example Usage
  ...
<!-- End SDK Example Usage -->
`,
				legacyTitle:  "SDK Installation",
				sectionTitle: "NEW Installation",
				sectionID:    "installation",
				replacement:  "\n",
			},
			want: `# README

This is a README file for the SDK.

<!-- Start NEW Installation [installation] -->


<!-- End NEW Installation [installation] -->

<!-- Start SDK Example Usage -->
## Example Usage
  ...
<!-- End SDK Example Usage -->
`,
		},
		{
			name: "preserve empty lines",
			args: args{
				contents: `# README

<!-- Start Empty Section [test] -->
<!-- End Empty Section [test] -->
`,
				legacyTitle:  "Deprecated Title",
				sectionTitle: "Padded Section",
				sectionID:    "test",
				replacement:  "\n\nContent\n\n",
			},
			want: `# README

<!-- Start Padded Section [test] -->


Content


<!-- End Padded Section [test] -->
`,
		},
		{
			name: "remove unneeded section using title",
			args: args{
				contents: `# README

This is a README file for the SDK.

<!-- Start Pagination -->
This section is no longer needed
<!-- End Pagination -->

<!-- Start SDK Example Usage -->
## Example Usage
  ...
<!-- End SDK Example Usage -->
`,
				legacyTitle:  "Pagination",
				sectionTitle: "",
				sectionID:    "",
				replacement:  "",
			},
			want: `# README

This is a README file for the SDK.

<!-- Start SDK Example Usage -->
## Example Usage
  ...
<!-- End SDK Example Usage -->
`,
		},
		{
			name: "remove unneeded section using ID",
			args: args{
				contents: `# README

This is a README file for the SDK.

<!-- Start Pagination [pagination] -->
This section is no longer needed
<!-- End Pagination [pagination] -->

<!-- Start SDK Example Usage -->
## Example Usage
<!-- Usage Snippet -->

<!-- End SDK Example Usage -->
`,
				legacyTitle:  "",
				sectionTitle: "",
				sectionID:    "pagination",
				replacement:  "",
			},
			want: `# README

This is a README file for the SDK.

<!-- Start SDK Example Usage -->
## Example Usage
<!-- Usage Snippet -->

<!-- End SDK Example Usage -->
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents, replaced := readme.ReplaceBlock(
				defaultPrefix,
				tt.args.legacyTitle,
				tt.args.sectionID,
				defaultSuffix,
				tt.args.contents,
				tt.args.sectionTitle,
				tt.args.replacement,
			)
			assert.True(t, replaced)

			assert.Equal(t, tt.want, contents)
		})
	}
}

func TestReplaceDocs(t *testing.T) {
	type args struct {
		contents    string
		title       string
		replacement string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "replace docs section with no whitespace",
			args: args{
				contents: `
{/* Start Go Installation */}{/* End Go Installation */}
`,
				title:       "Go Installation",
				replacement: "go get github.com/something/something",
			},
			want: `
{/* Start Go Installation */}
go get github.com/something/something
{/* End Go Installation */}
`,
		},
		{
			name: "replace docs section with empty lines",
			args: args{
				contents: `
{/* Start Go Installation */}
go get github.com/something/something
{/* End Go Installation */}
`,
				title:       "Go Installation",
				replacement: "\ngo get github.com/something/else\n",
			},
			want: `
{/* Start Go Installation */}

go get github.com/something/else

{/* End Go Installation */}
`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			contents, replaced := readme.ReplaceBlock(
				"{/*",
				tt.args.title,
				"",
				"*/}",
				tt.args.contents,
				tt.args.title,
				tt.args.replacement,
			)
			assert.True(t, replaced)

			assert.Equal(t, tt.want, contents)
		})
	}
}

func TestBlockDisabled(t *testing.T) {
	type args struct {
		contents  string
		title     string
		sectionID string
	}
	tests := []struct {
		name string
		args args
		want bool
	}{
		{
			name: "Section marked as no action (with id)",
			args: args{
				title:     "Pagination",
				sectionID: "pagination",
				contents: `
<!-- Start SDK Example Usage [usage] -->
## How to paginate
   ...

<!-- End SDK Example Usage [usage] -->

<!-- No Pagination [pagination] -->
`,
			},
			want: true,
		},
		{
			name: "Section marked as no action (legacy)",
			args: args{
				title:     "Pagination",
				sectionID: "pagination",
				contents: `# README
<!-- Start SDK Example Usage -->
## How to paginate
   ...

<!-- End SDK Example Usage -->

<!-- No Pagination -->
`,
			},
			want: true,
		},
		{
			name: "False positives",
			args: args{
				title:     "Pagination",
				sectionID: "pagination",
				contents: `
<!-- No Pagination [wrong-id] -->
No Pagination
No Pagination [pagination]
<!-- No Pagi nation -->
`,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			disabled := readme.IsBlockDisabled(
				defaultPrefix,
				tt.args.title,
				tt.args.sectionID,
				defaultSuffix,
				tt.args.contents,
			)
			assert.Equal(t, tt.want, disabled)
		})
	}
}

func TestParseBlocks(t *testing.T) {
	type args struct {
		contents string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "No Start block boundary",
			args: args{
				contents: `
# README
<!-- This README has no blocks -->
`,
			},
			want: "",
		},
		{
			name: "Block IDs ordering",
			args: args{
				contents: `
<!-- Start SDK Installation [installation] -->
## SDK Installation
   ...
<!-- End SDK Installation [installation] -->

<!-- Start Authentication [security] -->
Anything could happen next
<!-- End Block Boundary [is-never-parsed] -->

<!-- Start Error Handling [error-handling] -->
`,
			},
			want: "installation,security,error-handling",
		},
		{
			name: "Duplicate block boundary",
			args: args{
				contents: `
<!-- Start SDK Installation [installation] -->
## SDK Installation
   ...
<!-- Start Authentication [security] -->

<!-- Start SDK Installation [installation] -->
`,
			},
			want: "installation,security",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ids := readme.ParseBlockIDs(
				defaultPrefix,
				defaultSuffix,
				tt.args.contents,
			)
			assert.Equal(t, tt.want, ids)
		})
	}
}
