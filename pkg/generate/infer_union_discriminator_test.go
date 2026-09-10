package generate

import (
	"fmt"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to test discriminator inference
func doTestDiscriminator(t *testing.T, name, rawSpec string, expectedPropertyName string, expectedMappings map[string]string) {
	t.Helper()
	t.Setenv("SPEAKEASY_DEBUG_INFER_DISCRIMINATORS", "true")

	_, astTree, err := TestResolveAST(TestResolveASTInput{OpenAPIContents: []byte(rawSpec)})
	require.NoError(t, err)

	// Find the union type (first one we encounter)
	var unionType *ast.TypeDef
	visited := make(map[string]bool)
	_ = astTree.WalkSDK(astTree.MainSDK, []ast.Node{}, func(node ast.Node, parents []ast.Node, a *ast.AST) error {
		if ast.GetNodeType(node) != ast.NodeTypeTypeDef {
			return nil
		}
		t := node.(*ast.TypeDef)
		if t.Type == ast.DataTypeUnion && unionType == nil {
			unionType = t
		}
		return nil
	}, visited)
	require.NotNil(t, unionType, "Union type should exist")

	fmt.Printf("\n ---- Testing discriminator for %q ----\n", name)

	if unionType.Discriminator == nil {
		fmt.Printf("❌ No discriminator inferred\n")
		if expectedPropertyName != "" {
			fmt.Printf("---- FAIL ----\n\n\n")
		} else {
			fmt.Printf("---- PASS (expected no discriminator) ----\n\n\n")
		}
	} else {
		fmt.Printf("✓ Discriminator: %q (inferred=%v)\n", unionType.Discriminator.TypePropertyName, unionType.Discriminator.Inferred)
		for i, mapping := range unionType.Discriminator.Mapping {
			fmt.Printf("  [%d] %q -> %s\n", i, mapping.Name, mapping.Type.Name)
		}
		fmt.Printf("\n")
	}

	// Verify discriminator
	if expectedPropertyName == "" {
		assert.Nil(t, unionType.Discriminator, "Expected no discriminator")
	} else {
		require.NotNil(t, unionType.Discriminator, "Discriminator should be inferred")
		require.True(t, unionType.Discriminator.Inferred, "Discriminator should be marked as inferred")

		if !assert.Equal(t, expectedPropertyName, unionType.Discriminator.TypePropertyName) {
			fmt.Printf("---- FAIL ----\n\n\n")
			return
		}

		if !assert.Len(t, unionType.Discriminator.Mapping, len(expectedMappings)) {
			fmt.Printf("---- FAIL ----\n\n\n")
			return
		}

		// Verify mappings
		for _, mapping := range unionType.Discriminator.Mapping {
			expectedTypeName, ok := expectedMappings[mapping.Name]
			if !assert.True(t, ok, "Unexpected mapping name: %s", mapping.Name) {
				fmt.Printf("---- FAIL ----\n\n\n")
				return
			}
			if !assert.Equal(t, expectedTypeName, mapping.Type.Name, "Mapping %s should point to %s", mapping.Name, expectedTypeName) {
				fmt.Printf("---- FAIL ----\n\n\n")
				return
			}
		}

		fmt.Printf("---- PASS ----\n\n\n")
	}
}

func TestInferDiscriminatorMinimalPetstore(t *testing.T) {
	rawSpec := utils.Dedent(`
		openapi: 3.0.3
		info:
		  title: Minimal Petstore
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet (dog or cat)
		          content:
		            application/json:
		              schema:
		                $ref: "#/components/schemas/Pet"
		components:
		  schemas:
		    Pet:
		      oneOf:
		        - $ref: "#/components/schemas/Dog"
		        - $ref: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		          enum: [dog]
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		          enum: [cat]
		        name:
		          type: string
	`)

	expectedMappings := map[string]string{
		"dog": "Dog",
		"cat": "Cat",
	}

	doTestDiscriminator(t, "Minimal Petstore", rawSpec, "kind", expectedMappings)
}

func TestInferDiscriminatorWithConst(t *testing.T) {
	rawSpec := utils.Dedent(`
		openapi: 3.0.3
		info:
		  title: Petstore with const
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet (dog or cat)
		          content:
		            application/json:
		              schema:
		                $ref: "#/components/schemas/Pet"
		components:
		  schemas:
		    Pet:
		      oneOf:
		        - $ref: "#/components/schemas/Dog"
		        - $ref: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		          const: dog
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		          const: cat
		        name:
		          type: string
	`)

	expectedMappings := map[string]string{
		"dog": "Dog",
		"cat": "Cat",
	}

	doTestDiscriminator(t, "Petstore with const", rawSpec, "kind", expectedMappings)
}

func TestInferDiscriminatorMultiEnum(t *testing.T) {
	rawSpec := utils.Dedent(`
		openapi: 3.0.3
		info:
		  title: Petstore with multi-value enums
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet (dog or cat)
		          content:
		            application/json:
		              schema:
		                $ref: "#/components/schemas/Pet"
		components:
		  schemas:
		    Pet:
		      oneOf:
		        - $ref: "#/components/schemas/Dog"
		        - $ref: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		          enum: [labrador, poodle]
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		          enum: [siamese, furry]
		        name:
		          type: string
	`)

	expectedMappings := map[string]string{
		"labrador": "Dog",
		"poodle":   "Dog",
		"siamese":  "Cat",
		"furry":    "Cat",
	}

	doTestDiscriminator(t, "Petstore with multi-value enums", rawSpec, "kind", expectedMappings)
}

func TestInferDiscriminatorNoConstraints(t *testing.T) {
	rawSpec := utils.Dedent(`
		openapi: 3.0.3
		info:
		  title: Petstore without constraints
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet (dog or cat)
		          content:
		            application/json:
		              schema:
		                $ref: "#/components/schemas/Pet"
		components:
		  schemas:
		    Pet:
		      oneOf:
		        - $ref: "#/components/schemas/Dog"
		        - $ref: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		        name:
		          type: string
	`)

	// No discriminator should be inferred when there are no literal value constraints
	doTestDiscriminator(t, "Petstore without constraints", rawSpec, "", nil)
}

func TestInferDiscriminatorMissingRequiredOnOneMember(t *testing.T) {
	rawSpec := utils.Dedent(`
		openapi: 3.0.3
		info:
		  title: Petstore with inconsistent required
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet (dog or cat)
		          content:
		            application/json:
		              schema:
		                $ref: "#/components/schemas/Pet"
		components:
		  schemas:
		    Pet:
		      oneOf:
		        - $ref: "#/components/schemas/Dog"
		        - $ref: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      properties:
		        kind:
		          type: string
		          const: dog
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		          const: cat
		        name:
		          type: string
	`)

	// No discriminator should be inferred when 'kind' is not required on all members
	doTestDiscriminator(t, "Petstore with inconsistent required", rawSpec, "", nil)
}

func TestInferDiscriminatorOverlappingEnums(t *testing.T) {
	rawSpec := utils.Dedent(`
		openapi: 3.0.3
		info:
		  title: Petstore with overlapping enums
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet (dog or cat)
		          content:
		            application/json:
		              schema:
		                $ref: "#/components/schemas/Pet"
		components:
		  schemas:
		    Pet:
		      oneOf:
		        - $ref: "#/components/schemas/Dog"
		        - $ref: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		          enum: [b, c]
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [kind]
		      properties:
		        kind:
		          type: string
		          enum: [a, b]
		        name:
		          type: string
	`)

	// No discriminator should be inferred when enum values overlap (b appears in both)
	doTestDiscriminator(t, "Petstore with overlapping enums", rawSpec, "", nil)
}

func TestInferDiscriminatorMixedTagProperties(t *testing.T) {
	rawSpec := utils.Dedent(`
		openapi: 3.0.3
		info:
		  title: Mixed tag properties
		  version: 1.0.0
		paths:
		  /record:
		    get:
		      summary: Get an record
		      operationId: getRecord
		      responses:
		        "200":
		          description: An record
		          content:
		            application/json:
		              schema:
		                type: object
		                properties:
		                  options:
		                    $ref: "#/components/schemas/RecordOptions"
		components:
		  schemas:
		    RecordOptions:
		      oneOf:
		        - $ref: "#/components/schemas/DraftOptions"
		        - $ref: "#/components/schemas/PublishedOptions"
		        - $ref: "#/components/schemas/LocalOptions"
		        - $ref: "#/components/schemas/RemoteOptions"
		    DraftOptions:
		      type: object
		      required: [type]
		      properties:
		        type:
		          type: string
		          const: draft
		    PublishedOptions:
		      type: object
		      required: [type]
		      properties:
		        type:
		          type: string
		          const: published
		    LocalOptions:
		      type: object
		      required: [request]
		      properties:
		        request:
		          type: string
		          const: local
		    RemoteOptions:
		      type: object
		      required: [request]
		      properties:
		        request:
		          type: string
		          const: remote
	`)

	// No single discriminator property exists across all members.
	doTestDiscriminator(t, "Mixed tag properties", rawSpec, "", nil)
}
