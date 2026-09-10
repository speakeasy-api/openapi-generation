package readme_test

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/readme"
	"github.com/stretchr/testify/assert"
	"strings"
)

func TestGenerateTableOfContents(t *testing.T) {
	type args struct {
		contents string
		minLevel int
		maxLevel int
		exclude  []string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{

			name: "No contents",
			args: args{
				minLevel: 1,
				maxLevel: 2,
				exclude:  []string{},
				contents: "",
			},
			want: "",
		},
		{
			name: "Contents with no headers",
			args: args{
				minLevel: 1,
				maxLevel: 2,
				exclude:  []string{},
				contents: `
#1: this text has no headers
length = 2  # mm
`,
			},
			want: "",
		},
		{
			name: "Level selection",
			args: args{
				minLevel: 2,
				maxLevel: 2,
				exclude:  []string{},
				contents: `
# Level 1
## Level 2
### Level 3
`,
			},
			want: "* [Level 2](#level-2)\n",
		},
		{
			name: "Minimun level too high",
			args: args{
				minLevel: 3,
				maxLevel: 4,
				exclude:  []string{},
				contents: `
# Level 1
## Level 2
`,
			},
			want: "",
		},
		{
			name: "Maximum level ignored",
			args: args{
				minLevel: 2,
				maxLevel: 1,
				exclude:  []string{},
				contents: `
# Level 1
## Level 2
### Level 3
`,
			},
			want: "* [Level 2](#level-2)\n",
		},
		{
			name: "Headers with leading whitespace",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: `
Level 0
# Level A1
## Level A2
### Level A3
 # Level B1
  ## Level B2
   ### Level B3
`,
			},
			want: `
* [Level A1](#level-a1)
  * [Level A2](#level-a2)
    * [Level A3](#level-a3)
`,
		},
		{
			name: "Headers with whitespace",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: `
Level 0
# # Level 1
##  ##  Level 2  ##  ##
### Level 3 ##
`,
			},
			want: `
* [Level 1](#level-1)
  * [Level 2](#level-2)
    * [Level 3](#level-3)
`,
		},
		{
			name: "Slug sanitization",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: `
# Level-1!
## Level-_-2
### Level---<3
`,
			},
			want: `
* [Level-1!](#level-1)
  * [Level-_-2](#level-2)
    * [Level---<3](#level-3)
`,
		},
		{
			name: "Slug deduplication",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: `
# Level 1 #
## Level 2 ##
# Level-1
## Level-2
### Level-3
# Level_1
## Level 2
`,
			},
			want: `
* [Level 1](#level-1)
  * [Level 2](#level-2)
* [Level-1](#level-1-1)
  * [Level-2](#level-2-1)
    * [Level-3](#level-3)
* [Level_1](#level1)
  * [Level 2](#level-2-2)
`,
		},
		{
			name: "Exclude headers",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude: []string{
					"# Section X1",
					"## Section X2",
					"### Section X3",
				},
				contents: `
# Section X1
# Section A1
## Section A2
### Section A3
### Section X3

# Section B1
## Section B2
## Section X2
### Section B3
`,
			},
			want: `
* [Section A1](#section-a1)
  * [Section A2](#section-a2)
    * [Section A3](#section-a3)
* [Section B1](#section-b1)
  * [Section B2](#section-b2)
    * [Section B3](#section-b3)
`,
		},
		{
			name: "Ignore code blocks",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Examples
## Example 1

^^^python
# this is a python comment
not = "a header"  # it should be skipped
^^^
## Example 2

^^^
## Another comment ##
^^^
`, "^", "`"),
			},
			want: `
* [Examples](#examples)
  * [Example 1](#example-1)
  * [Example 2](#example-2)
`,
		},
		{
			name: "Ignore indented code blocks nested in lists",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Commands
## Create

* generate - text generation

  ^^^bash
  # Generate with the default model
  cli generate "hello"
  ^^^

* image - image generation

  ^^^bash
  # Generate an image
  cli image "a lighthouse"
  ^^^

## Manage

^^^bash
# Fetch a single page
cli list
^^^
`, "^", "`"),
			},
			want: `
* [Commands](#commands)
  * [Create](#create)
  * [Manage](#manage)
`,
		},
		{
			name: "Deeply indented literal fence inside a block does not close it",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Examples

^^^text
    ^^^
# Still inside the code block
^^^

## Real section
`, "^", "`"),
			},
			want: `
* [Examples](#examples)
  * [Real section](#real-section)
`,
		},
		{
			name: "Indented fence prefix with trailing text does not close block",

			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Examples

^^^text
  ^^^not a closing fence
# Still inside the code block
^^^

## Real section
`, "^", "`"),
			},
			want: `
* [Examples](#examples)
  * [Real section](#real-section)
`,
		},
		{
			name: "Longer fence is not closed by a shorter fence inside it",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Examples

^^^^markdown
^^^bash
# Still inside the code block
^^^
# Still inside the code block
^^^^

## Real section
`, "^", "`"),
			},
			want: `
* [Examples](#examples)
  * [Real section](#real-section)
`,
		},
		{
			name: "Inline code line with backticks in the info string is not an opening fence",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Examples

^^^inline^^^ some text

## Real section

^^^bash
# comment inside a real fence
^^^

## Another section
`, "^", "`"),
			},
			want: `
* [Examples](#examples)
  * [Real section](#real-section)
  * [Another section](#another-section)
`,
		},
		{
			name: "Tilde fence with backticks in the info string still opens a block",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Examples

~~~with ^backticks^
# Still inside the code block
~~~

## Real section
`, "^", "`"),
			},
			want: `
* [Examples](#examples)
  * [Real section](#real-section)
`,
		},
		{
			name: "Fence-shaped line indented 4+ spaces is indented code, not an opening fence",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Examples

Fence syntax shown as indented code:

    ^^^

## Section A

^^^bash
# comment inside a real fence
^^^

## Section B
`, "^", "`"),
			},
			want: `
* [Examples](#examples)
  * [Section A](#section-a)
  * [Section B](#section-b)
`,
		},
		{
			name: "Unmatched indented fence-shaped line does not swallow the rest of the document",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Usage

To open a fenced code block write:

    ^^^python

## Installation
`, "^", "`"),
			},
			want: `
* [Usage](#usage)
  * [Installation](#installation)
`,
		},
		{
			name: "Tab-indented fence-shaped line is indented code, not an opening fence",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Usage

	^^^

## Installation
`, "^", "`"),
			},
			want: `
* [Usage](#usage)
  * [Installation](#installation)
`,
		},
		{
			name: "Fence opener indented exactly 3 spaces is still a fence",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Examples

   ^^^bash
# comment inside the fence
   ^^^

## Real section
`, "^", "`"),
			},
			want: `
* [Examples](#examples)
  * [Real section](#real-section)
`,
		},
		{
			name: "Tilde fence is not closed by a backtick fence inside it",
			args: args{
				minLevel: 1,
				maxLevel: 3,
				exclude:  []string{},
				contents: strings.ReplaceAll(`
# Examples

~~~markdown
^^^
# Still inside the code block
^^^
~~~

## Real section
`, "^", "`"),
			},
			want: `
* [Examples](#examples)
  * [Real section](#real-section)
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			toc := readme.GenerateTableOfContents(tt.args.contents, tt.args.minLevel, tt.args.maxLevel, tt.args.exclude)
			assert.Equal(t, strings.TrimLeft(tt.want, "\n"), toc)
		})
	}
}
