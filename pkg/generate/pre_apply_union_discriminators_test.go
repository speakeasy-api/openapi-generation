package generate

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/ast_post_processing/pre_apply_union_discriminators"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/stretchr/testify/require"
)

// findTypeByName finds a TypeDef by name in the AST and asserts exactly one is found
func findTypeByName(t *testing.T, astTree *ast.AST, name string) *ast.TypeDef {
	t.Helper()
	found := make(map[*ast.TypeDef]bool)
	_ = astTree.WalkSDK(astTree.MainSDK, []ast.Node{}, func(node ast.Node, parents []ast.Node, a *ast.AST) error {
		if ast.GetNodeType(node) != ast.NodeTypeTypeDef {
			return nil
		}
		td := node.(*ast.TypeDef)
		if td.Name == name {
			found[td] = true
		}
		return nil
	}, make(map[string]bool))

	require.Len(t, found, 1, "expected exactly one type named %q, found %d", name, len(found))
	for td := range found {
		return td
	}
	return nil
}

func doTestPreApplyUnionDiscriminators(t *testing.T, rawSpec string) *ast.AST {
	t.Helper()
	t.Setenv("SPEAKEASY_DEBUG_PRE_APPLY_UNION_DISCRIMINATORS", "true")

	additionalConfig := map[string]any{
		"preApplyUnionDiscriminators": true,
	}
	_, astTree, err := TestResolveAST(TestResolveASTInput{
		OpenAPIContents:  []byte(rawSpec),
		AdditionalConfig: &additionalConfig,
	})
	require.NoError(t, err)

	return astTree
}

func TestPreApplyUnionDiscriminators_ConstAlreadyApplied(t *testing.T) {
	// Spec with 1 op with oneOf: dog & cat, with discriminator mapping,
	// petKind already has const: dog,cat respectively
	// expect that IsDiscriminatorAlreadyApplied returns true for both
	rawSpec := utils.Dedent(`
		openapi: 3.1.0
		info:
		  title: Pet API
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet
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
		      discriminator:
		        propertyName: petKind
		        mapping:
		          dog: "#/components/schemas/Dog"
		          cat: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      required: [petKind]
		      properties:
		        petKind:
		          type: string
		          const: dog
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [petKind]
		      properties:
		        petKind:
		          type: string
		          const: cat
		        name:
		          type: string
	`)

	astTree := doTestPreApplyUnionDiscriminators(t, rawSpec)

	dogType := findTypeByName(t, astTree, "Dog")
	catType := findTypeByName(t, astTree, "Cat")

	// Verify IsDiscriminatorAlreadyApplied returns true for both types
	require.True(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(dogType, "petKind", "dog"),
		"Dog should have discriminator already applied with petKind=dog")
	require.True(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(catType, "petKind", "cat"),
		"Cat should have discriminator already applied with petKind=cat")

	// Verify DiscriminatorPreApplied is set
	require.Equal(t, "petKind", dogType.DiscriminatorPreApplied,
		"Dog.DiscriminatorPreApplied should be 'petKind'")
	require.Equal(t, "petKind", catType.DiscriminatorPreApplied,
		"Cat.DiscriminatorPreApplied should be 'petKind'")
}

func TestPreApplyUnionDiscriminators_SingleValueEnumAlreadyApplied(t *testing.T) {
	// Spec with 1 op with oneOf: dog & cat, with discriminator mapping,
	// petKind is an enum with a single value (dog, cat respectively)
	// expect that IsDiscriminatorAlreadyApplied returns true for both
	rawSpec := utils.Dedent(`
		openapi: 3.1.0
		info:
		  title: Pet API
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet
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
		      discriminator:
		        propertyName: petKind
		        mapping:
		          dog: "#/components/schemas/Dog"
		          cat: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      required: [petKind]
		      properties:
		        petKind:
		          type: string
		          enum: [dog]
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [petKind]
		      properties:
		        petKind:
		          type: string
		          enum: [cat]
		        name:
		          type: string
	`)

	astTree := doTestPreApplyUnionDiscriminators(t, rawSpec)

	dogType := findTypeByName(t, astTree, "Dog")
	catType := findTypeByName(t, astTree, "Cat")

	// Verify IsDiscriminatorAlreadyApplied returns true for both types
	require.True(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(dogType, "petKind", "dog"),
		"Dog should have discriminator already applied with petKind=dog")
	require.True(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(catType, "petKind", "cat"),
		"Cat should have discriminator already applied with petKind=cat")

	// Verify DiscriminatorPreApplied is set
	require.Equal(t, "petKind", dogType.DiscriminatorPreApplied,
		"Dog.DiscriminatorPreApplied should be 'petKind'")
	require.Equal(t, "petKind", catType.DiscriminatorPreApplied,
		"Cat.DiscriminatorPreApplied should be 'petKind'")
}

func TestPreApplyUnionDiscriminators_PlainStringCoerced(t *testing.T) {
	// Spec with 1 op with oneOf: dog & cat, with discriminator mapping,
	// petKind is just a plain string (no const, no enum)
	// expect that the discriminator value is coerced onto the field
	rawSpec := utils.Dedent(`
		openapi: 3.1.0
		info:
		  title: Pet API
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet
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
		      discriminator:
		        propertyName: petKind
		        mapping:
		          dog: "#/components/schemas/Dog"
		          cat: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      required: [petKind]
		      properties:
		        petKind:
		          type: string
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [petKind]
		      properties:
		        petKind:
		          type: string
		        name:
		          type: string
	`)

	astTree := doTestPreApplyUnionDiscriminators(t, rawSpec)

	dogType := findTypeByName(t, astTree, "Dog")
	catType := findTypeByName(t, astTree, "Cat")

	// Verify IsDiscriminatorAlreadyApplied returns true for both types (after coercion)
	require.True(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(dogType, "petKind", "dog"),
		"Dog should have discriminator applied with petKind=dog after coercion")
	require.True(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(catType, "petKind", "cat"),
		"Cat should have discriminator applied with petKind=cat after coercion")

	// Verify DiscriminatorPreApplied is set
	require.Equal(t, "petKind", dogType.DiscriminatorPreApplied,
		"Dog.DiscriminatorPreApplied should be 'petKind'")
	require.Equal(t, "petKind", catType.DiscriminatorPreApplied,
		"Cat.DiscriminatorPreApplied should be 'petKind'")
}

func TestPreApplyUnionDiscriminators_PlainStringNotCoercedWhenUsedElsewhere(t *testing.T) {
	// Spec with 1 op with oneOf: dog & cat, with discriminator mapping,
	// petKind is just a plain string (no const, no enum)
	// BUT Dog is also used directly in another operation without the discriminator
	// expect that the discriminator value is NOT coerced (since Dog is used without discriminator context)
	rawSpec := utils.Dedent(`
		openapi: 3.1.0
		info:
		  title: Pet API
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet
		          content:
		            application/json:
		              schema:
		                $ref: "#/components/schemas/Pet"
		  /dog:
		    get:
		      summary: Get a dog directly
		      operationId: getDog
		      responses:
		        "200":
		          description: A dog
		          content:
		            application/json:
		              schema:
		                $ref: "#/components/schemas/Dog"
		components:
		  schemas:
		    Pet:
		      oneOf:
		        - $ref: "#/components/schemas/Dog"
		        - $ref: "#/components/schemas/Cat"
		      discriminator:
		        propertyName: petKind
		        mapping:
		          dog: "#/components/schemas/Dog"
		          cat: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      required: [petKind]
		      properties:
		        petKind:
		          type: string
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [petKind]
		      properties:
		        petKind:
		          type: string
		        name:
		          type: string
	`)

	astTree := doTestPreApplyUnionDiscriminators(t, rawSpec)

	dogType := findTypeByName(t, astTree, "Dog")
	catType := findTypeByName(t, astTree, "Cat")

	// Dog should NOT have discriminator applied because it's used directly in /dog endpoint
	require.False(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(dogType, "petKind", "dog"),
		"Dog should NOT have discriminator applied because it's used without discriminator context in /dog")
	// Cat should still have discriminator applied because it's only used in the union
	require.True(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(catType, "petKind", "cat"),
		"Cat should have discriminator applied with petKind=cat after coercion")

	// Verify DiscriminatorPreApplied
	require.Empty(t, dogType.DiscriminatorPreApplied,
		"Dog.DiscriminatorPreApplied should be empty because it's used without discriminator context")
	require.Equal(t, "petKind", catType.DiscriminatorPreApplied,
		"Cat.DiscriminatorPreApplied should be 'petKind'")
}

func TestPreApplyUnionDiscriminators_MultiValueEnumNotCoercedWhenUsedElsewhere(t *testing.T) {
	// Spec with 1 op with oneOf: dog & cat, with discriminator mapping,
	// petKind has multiple enum values for each type
	// AND Dog is also used directly in another operation without the discriminator
	// expect that neither type gets coerced (Dog used elsewhere, Cat has multiple enum values)
	rawSpec := utils.Dedent(`
		openapi: 3.1.0
		info:
		  title: Pet API
		  version: 1.0.0
		paths:
		  /pet:
		    get:
		      summary: Get a pet
		      operationId: getPet
		      responses:
		        "200":
		          description: A pet
		          content:
		            application/json:
		              schema:
		                $ref: "#/components/schemas/Pet"
		  /dog:
		    get:
		      summary: Get a dog directly
		      operationId: getDog
		      responses:
		        "200":
		          description: A dog
		          content:
		            application/json:
		              schema:
		                $ref: "#/components/schemas/Dog"
		components:
		  schemas:
		    Pet:
		      oneOf:
		        - $ref: "#/components/schemas/Dog"
		        - $ref: "#/components/schemas/Cat"
		      discriminator:
		        propertyName: petKind
		        mapping:
		          labrador: "#/components/schemas/Dog"
		          poodle: "#/components/schemas/Dog"
		          siamese: "#/components/schemas/Cat"
		          persian: "#/components/schemas/Cat"
		    Dog:
		      type: object
		      required: [petKind]
		      properties:
		        petKind:
		          type: string
		          enum: [labrador, poodle]
		        name:
		          type: string
		    Cat:
		      type: object
		      required: [petKind]
		      properties:
		        petKind:
		          type: string
		          enum: [siamese, persian]
		        name:
		          type: string
	`)

	astTree := doTestPreApplyUnionDiscriminators(t, rawSpec)

	dogType := findTypeByName(t, astTree, "Dog")
	catType := findTypeByName(t, astTree, "Cat")

	// Dog should NOT have discriminator applied because it's used directly in /dog endpoint
	require.False(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(dogType, "petKind", "labrador"),
		"Dog should NOT have discriminator applied because it's used without discriminator context in /dog")
	require.False(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(dogType, "petKind", "poodle"),
		"Dog should NOT have discriminator applied because it's used without discriminator context in /dog")
	// Cat should NOT have discriminator applied because it has multiple mappings (siamese, persian)
	require.False(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(catType, "petKind", "siamese"),
		"Cat should NOT have discriminator applied because it has multiple discriminator mappings")
	require.False(t, pre_apply_union_discriminators.IsDiscriminatorAlreadyAppliedAsConst(catType, "petKind", "persian"),
		"Cat should NOT have discriminator applied because it has multiple discriminator mappings")

	// Verify DiscriminatorPreApplied is empty for both
	require.Empty(t, dogType.DiscriminatorPreApplied,
		"Dog.DiscriminatorPreApplied should be empty because it's used without discriminator context")
	require.Empty(t, catType.DiscriminatorPreApplied,
		"Cat.DiscriminatorPreApplied should be empty because it has multiple discriminator mappings")
}
