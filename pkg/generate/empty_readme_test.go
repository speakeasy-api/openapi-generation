package generate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEmptyReadmeGeneration tests that empty README files are not generated for
// intermediate groups that have no operations when using hierarchical grouping
// with x-speakeasy-group.
func TestEmptyReadmeGeneration(t *testing.T) {
	tests := []struct {
		name                  string
		openAPIYAML           string
		language              string
		expectedReadmeFiles   []string // Files that SHOULD be generated
		unexpectedReadmeFiles []string // Files that should NOT be generated
	}{
		{
			name: "hierarchical grouping should not create empty intermediate READMEs",
			openAPIYAML: `openapi: 3.0.3
info:
  title: Empty README Test API
  description: Test API to demonstrate empty README generation issue
  version: 1.0.0

tags:
  - name: alpha
    description: Alpha namespace (should not create empty README)
  - name: beta
    description: Beta namespace (should not create empty README)

paths:
  /alpha/response/get:
    get:
      operationId: getAlphaResponse
      summary: Get alpha response
      tags:
        - alpha
      x-speakeasy-group: alpha.response
      responses:
        '200':
          description: Success
          content:
            application/json:
              schema:
                type: object
                properties:
                  message:
                    type: string

  /beta/data/list:
    get:
      operationId: listBetaData
      summary: List beta data
      tags:
        - beta
      x-speakeasy-group: beta.data.items
      responses:
        '200':
          description: Success

  /beta/data/special/process:
    post:
      operationId: processSpecialBetaData
      summary: Process special beta data
      tags:
        - beta
      x-speakeasy-group: beta.data.special
      responses:
        '200':
          description: Success`,
			language: "typescript",
			expectedReadmeFiles: []string{
				"docs/sdks/response/README.md",
				"docs/sdks/items/README.md",
				"docs/sdks/special/README.md",
			},
			unexpectedReadmeFiles: []string{
				"docs/sdks/alpha/README.md", // Should not exist - intermediate group with no operations
				"docs/sdks/beta/README.md",  // Should not exist - intermediate group with no operations
				"docs/sdks/data/README.md",  // Should not exist - intermediate group with no operations
			},
		},
		{
			name: "non-hierarchical grouping should still generate README files normally",
			openAPIYAML: `openapi: 3.0.3
info:
  title: Control Test API
  description: Normal API without hierarchical grouping
  version: 1.0.0

tags:
  - name: users
    description: User operations
  - name: orders
    description: Order operations

paths:
  /users:
    get:
      operationId: getUsers
      tags: [users]
      responses:
        '200':
          description: Success

  /orders:
    get:
      operationId: getOrders
      tags: [orders]
      responses:
        '200':
          description: Success`,
			language: "typescript",
			expectedReadmeFiles: []string{
				"docs/sdks/users/README.md",
				"docs/sdks/orders/README.md",
			},
			unexpectedReadmeFiles: []string{
				// No unexpected files for this test case
			},
		},
		{
			name: "mixed hierarchical and flat groups",
			openAPIYAML: `openapi: 3.0.3
info:
  title: Mixed Grouping Test API
  description: API with both hierarchical and flat grouping
  version: 1.0.0

tags:
  - name: simple
    description: Simple flat group
  - name: complex
    description: Complex hierarchical group

paths:
  /simple:
    get:
      operationId: getSimple
      tags: [simple]
      responses:
        '200':
          description: Success

  /complex/nested/deep:
    get:
      operationId: getComplexNestedDeep
      tags: [complex]
      x-speakeasy-group: complex.nested.deep
      responses:
        '200':
          description: Success`,
			language: "go",
			expectedReadmeFiles: []string{
				"docs/sdks/simple/README.md",
				"docs/sdks/deep/README.md",
			},
			unexpectedReadmeFiles: []string{
				"docs/sdks/complex/README.md", // Should not exist - intermediate group with no operations
				"docs/sdks/nested/README.md",  // Should not exist - intermediate group with no operations
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary directory for SDK generation
			tempDir := t.TempDir()

			// Create generator
			g, err := New()
			require.NoError(t, err, "Failed to create generator")

			// Generate the SDK
			errs := g.Generate(
				generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL()),
				[]byte(tt.openAPIYAML),
				"empty_readme_test.yaml",
				tt.language,
				tempDir,
				false, // debug
				false, // installDependencies
			)
			require.Empty(t, errs, "SDK generation should succeed")

			// Verify expected README files exist and have content
			for _, expectedFile := range tt.expectedReadmeFiles {
				filePath := filepath.Join(tempDir, expectedFile)

				// File should exist
				assert.FileExists(t, filePath, "Expected README file should exist: %s", expectedFile)

				// File should have substantial content (more than just headers)
				if _, err := os.Stat(filePath); err == nil {
					content, err := os.ReadFile(filePath)
					require.NoError(t, err)

					// File should have meaningful content (more than just a basic header structure)
					// We expect at least some operation documentation
					assert.Greater(t, len(content), 200,
						"Expected README file should have substantial content: %s", expectedFile)

					// Should contain operation documentation markers
					contentStr := string(content)
					assert.Contains(t, contentStr, "Available Operations",
						"Expected README should contain operations section: %s", expectedFile)
				}
			}

			// Verify unexpected README files do NOT exist
			for _, unexpectedFile := range tt.unexpectedReadmeFiles {
				filePath := filepath.Join(tempDir, unexpectedFile)

				// File should NOT exist
				assert.NoFileExists(t, filePath,
					"Intermediate group README should not exist: %s", unexpectedFile)
			}

			// Additional verification: Check main README doesn't have broken links
			mainReadmePath := filepath.Join(tempDir, "README.md")
			if _, err := os.Stat(mainReadmePath); err == nil {
				content, err := os.ReadFile(mainReadmePath)
				require.NoError(t, err)

				contentStr := string(content)

				// Verify no links to non-existent intermediate README files
				for _, unexpectedFile := range tt.unexpectedReadmeFiles {
					linkPattern := fmt.Sprintf("(%s)", unexpectedFile)
					assert.NotContains(t, contentStr, linkPattern,
						"Main README should not link to non-existent file: %s", unexpectedFile)
				}

				// Verify links to expected README files exist
				for _, expectedFile := range tt.expectedReadmeFiles {
					linkPattern := fmt.Sprintf("(%s)", expectedFile)
					assert.Contains(t, contentStr, linkPattern,
						"Main README should link to expected file: %s", expectedFile)
				}
			}
		})
	}
}
