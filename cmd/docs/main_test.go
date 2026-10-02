package main

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var reportFixture = [][]string{
	{"Template", "typescriptv2", "mcp-typescript", "postman", "cli", "unity"},
	{"core (Core)", ":white_check_mark:", ":warning:", ":no_entry:", ":heavy_minus_sign:", ":white_check_mark:"},
	{"reactQueryHooks (Experimental)", ":white_check_mark:", ":white_check_mark:", ":white_check_mark:", ":white_check_mark:", ":white_check_mark:"},
}

func TestRender_FeatureTable(t *testing.T) {
	t.Parallel()

	out := render(reportFixture)

	assert.Contains(t, out, `## Customization Basics

<Table
  columns={[
    { key: "feature", header: "Feature" },
    { key: "typescript", header: "TypeScript" },
    { key: "python", header: "Python" },
    { key: "go", header: "Go" },
    { key: "java", header: "Java" },
    { key: "csharp", header: "C#" },
    { key: "php", header: "PHP" },
    { key: "ruby", header: "Ruby" },
    { key: "terraform", header: "Terraform" },
    { key: "mcp-typescript", header: "MCP Typescript" },
    { key: "postman", header: "Postman" },
    { key: "cli", header: "CLI" },
  ]}
  data={[
`)

	assert.Contains(t, out, "    {\n"+
		"      feature: \"`core`\",\n"+
		"      typescript: \"✅\",\n"+
		"      python: \"⛔\",\n"+
		"      go: \"⛔\",\n"+
		"      java: \"⛔\",\n"+
		"      csharp: \"⛔\",\n"+
		"      php: \"⛔\",\n"+
		"      ruby: \"⛔\",\n"+
		"      terraform: \"⛔\",\n"+
		"      \"mcp-typescript\": \"⚠️\",\n"+
		"      postman: \"⛔\",\n"+
		"      cli: \"➖\",\n"+
		"    },\n")

	// Categories without any features in the report are left out entirely, and
	// features that are not mapped to a category are not published.
	assert.NotContains(t, out, "## Structure")
	assert.NotContains(t, out, "reactQueryHooks")
}

func TestRender_SupportLevels(t *testing.T) {
	t.Parallel()

	out := render(reportFixture)

	assert.Contains(t, out, `<Table
  columns={[
    { key: "target", header: "Target" },
    { key: "maturity", header: "Maturity level" },
    { key: "support", header: "Feature support level" },
  ]}
  data={[
    {
      target: "[TypeScript](/docs/sdks/languages/typescript/methodology-ts)",
      maturity: "GA",
      support: "GA",
    },
`)

	assert.Contains(t, out, `    {
      target: "[CLI](/docs/cli-generation/create-cli)",
      maturity: "Beta",
      support: "Level 2",
    },
  ]}
/>
`)
}

// The docs site is Prettier formatted, so a page Prettier would rewrite creates
// formatting churn on top of the real changes on every regeneration.
func TestRender_StaysWithinPrettierPrintWidth(t *testing.T) {
	t.Parallel()

	for _, line := range strings.Split(render(reportFixture), "\n") {
		trimmed := strings.TrimLeft(line, " ")
		if !strings.HasSuffix(trimmed, ",") || !strings.Contains(trimmed, ": ") {
			continue
		}

		require.LessOrEqualf(t, len(line), prettierPrintWidth, "property line exceeds the print width: %s", line)
	}
}

// reportFixture is shared by parallel tests, so rendering must leave the
// records it is handed untouched.
func TestRender_DoesNotMutateRecords(t *testing.T) {
	t.Parallel()

	records := [][]string{
		{"Template", "typescriptv2", "mcp-typescript"},
		{"core (Core)", ":white_check_mark:", ":warning:"},
	}
	untouched := [][]string{
		{"Template", "typescriptv2", "mcp-typescript"},
		{"core (Core)", ":white_check_mark:", ":warning:"},
	}

	render(records)

	assert.Equal(t, untouched, records)
}

func TestRender_EndsWithSingleNewline(t *testing.T) {
	t.Parallel()

	out := render(reportFixture)

	assert.True(t, strings.HasSuffix(out, "/>\n"))
	assert.False(t, strings.HasSuffix(out, "\n\n"))
}
