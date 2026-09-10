package main

import (
	"context"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

const openapiHeaderDiscriminator = `openapi: 3.1.0
info:
  title: NestedOneOfDiscriminator
  version: 0.1.0
servers:
  - url: http://localhost:35123
    description: The default server.
paths:
  /test:
    get:
      responses:
        '200':
          description: OK
`

func TestNestedOneOfDiscriminator_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name string
		args args
	}{
		{
			name: "nested oneOf with discriminator succeeds",
			args: args{
				schema: openapiHeaderDiscriminator + utils.Dedent(`
					components:
					  schemas:
					    Pet:
					      x-speakeasy-include: true
					      oneOf:
					        - $ref: '#/components/schemas/Dog'
					        - $ref: '#/components/schemas/Cat'
					      discriminator:
					        propertyName: petType
					    Dog:
					      oneOf:
					        - $ref: '#/components/schemas/Doberman'
					        - $ref: '#/components/schemas/Labrador'
					      discriminator:
					        propertyName: breed
					    Cat:
					      type: object
					      properties:
					        petType:
					          type: string
					        name:
					          type: string
					      required:
					        - petType
					    Doberman:
					      type: object
					      properties:
					        petType:
					          type: string
					        breed:
					          type: string
					        aggressiveness:
					          type: integer
					      required:
					        - petType
					        - breed
					    Labrador:
					      type: object
					      properties:
					        petType:
					          type: string
					        breed:
					          type: string
					        friendliness:
					          type: integer
					      required:
					        - petType
					        - breed
				`),
			},
		},
		{
			name: "deeply nested oneOf with discriminator succeeds",
			args: args{
				schema: openapiHeaderDiscriminator + utils.Dedent(`
					components:
					  schemas:
					    Animal:
					      x-speakeasy-include: true
					      oneOf:
					        - $ref: '#/components/schemas/Pet'
					        - $ref: '#/components/schemas/WildAnimal'
					      discriminator:
					        propertyName: animalType
					    Pet:
					      oneOf:
					        - $ref: '#/components/schemas/Dog'
					        - $ref: '#/components/schemas/Cat'
					      discriminator:
					        propertyName: petType
					    Dog:
					      oneOf:
					        - $ref: '#/components/schemas/Doberman'
					        - $ref: '#/components/schemas/Labrador'
					      discriminator:
					        propertyName: breed
					    Cat:
					      type: object
					      properties:
					        animalType:
					          type: string
					        petType:
					          type: string
					        name:
					          type: string
					      required:
					        - animalType
					        - petType
					    WildAnimal:
					      type: object
					      properties:
					        animalType:
					          type: string
					        habitat:
					          type: string
					      required:
					        - animalType
					    Doberman:
					      type: object
					      properties:
					        animalType:
					          type: string
					        petType:
					          type: string
					        breed:
					          type: string
					        aggressiveness:
					          type: integer
					      required:
					        - animalType
					        - petType
					        - breed
					    Labrador:
					      type: object
					      properties:
					        animalType:
					          type: string
					        petType:
					          type: string
					        breed:
					          type: string
					        friendliness:
					          type: integer
					      required:
					        - animalType
					        - petType
					        - breed
				`),
			},
		},
		{
			name: "mixed nested oneOf with discriminator and regular objects succeeds",
			args: args{
				schema: openapiHeaderDiscriminator + `
components:
  schemas:
    Vehicle:
      x-speakeasy-include: true
      oneOf:
        - $ref: '#/components/schemas/Car'
        - $ref: '#/components/schemas/Bicycle'
      discriminator:
        propertyName: vehicleType
    Car:
      oneOf:
        - $ref: '#/components/schemas/Sedan'
        - $ref: '#/components/schemas/SUV'
      discriminator:
        propertyName: carType
    Bicycle:
      type: object
      properties:
        vehicleType:
          type: string
        brand:
          type: string
        gears:
          type: integer
      required:
        - vehicleType
    Sedan:
      type: object
      properties:
        vehicleType:
          type: string
        carType:
          type: string
        doors:
          type: integer
      required:
        - vehicleType
        - carType
    SUV:
      type: object
      properties:
        vehicleType:
          type: string
        carType:
          type: string
        fourWheelDrive:
          type: boolean
      required:
        - vehicleType
        - carType
`,
			},
		},
		{
			name: "nested anyOf with discriminator succeeds",
			args: args{
				schema: openapiHeaderDiscriminator + `
components:
  schemas:
    Device:
      x-speakeasy-include: true
      oneOf:
        - $ref: '#/components/schemas/Computer'
        - $ref: '#/components/schemas/Phone'
      discriminator:
        propertyName: deviceType
    Computer:
      anyOf:
        - $ref: '#/components/schemas/Laptop'
        - $ref: '#/components/schemas/Desktop'
      discriminator:
        propertyName: computerType
    Phone:
      type: object
      properties:
        deviceType:
          type: string
        brand:
          type: string
      required:
        - deviceType
    Laptop:
      type: object
      properties:
        deviceType:
          type: string
        computerType:
          type: string
        portability:
          type: integer
      required:
        - deviceType
        - computerType
    Desktop:
      type: object
      properties:
        deviceType:
          type: string
        computerType:
          type: string
        powerConsumption:
          type: integer
      required:
        - deviceType
        - computerType
`,
			},
		},
		{
			name: "subschema with union property succeeds",
			args: args{
				schema: openapiHeaderDiscriminator + `
components:
  schemas:
    Shape:
      x-speakeasy-include: true
      oneOf:
        - $ref: '#/components/schemas/Rectangle'
        - $ref: '#/components/schemas/Circle'
      discriminator:
        propertyName: shapeType
    Rectangle:
      type: object
      properties:
        shapeType:
          type: string
        dimensions:
          oneOf:
            - $ref: '#/components/schemas/TwoDimensions'
            - $ref: '#/components/schemas/ThreeDimensions'
          discriminator:
            propertyName: dimensionType
        area:
          type: number
      required:
        - shapeType
        - dimensions
    Circle:
      type: object
      properties:
        shapeType:
          type: string
        radius:
          type: number
        color:
          anyOf:
            - $ref: '#/components/schemas/RGB'
            - $ref: '#/components/schemas/HSL'
      required:
        - shapeType
        - radius
    TwoDimensions:
      type: object
      properties:
        dimensionType:
          type: string
        width:
          type: number
        height:
          type: number
      required:
        - dimensionType
        - width
        - height
    ThreeDimensions:
      type: object
      properties:
        dimensionType:
          type: string
        width:
          type: number
        height:
          type: number
        depth:
          type: number
      required:
        - dimensionType
        - width
        - height
        - depth
    RGB:
      type: object
      properties:
        red:
          type: integer
          minimum: 0
          maximum: 255
        green:
          type: integer
          minimum: 0
          maximum: 255
        blue:
          type: integer
          minimum: 0
          maximum: 255
      required:
        - red
        - green
        - blue
    HSL:
      type: object
      properties:
        hue:
          type: integer
          minimum: 0
          maximum: 360
        saturation:
          type: integer
          minimum: 0
          maximum: 100
        lightness:
          type: integer
          minimum: 0
          maximum: 100
      required:
        - hue
        - saturation
        - lightness
`,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs := testGenerate(t, []byte(tt.args.schema))
			require.Empty(t, errs)
		})
	}
}

func TestNestedOneOfDiscriminator_Errors(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name         string
		args         args
		wantWarnings []string
	}{
		{
			name: "nested oneOf missing discriminator property logs warning",
			args: args{
				schema: openapiHeaderDiscriminator + `
components:
  schemas:
    Pet:
      x-speakeasy-include: true
      oneOf:
        - $ref: '#/components/schemas/Dog'
        - $ref: '#/components/schemas/Cat'
      discriminator:
        propertyName: petType
    Dog:
      oneOf:
        - $ref: '#/components/schemas/Doberman'
        - $ref: '#/components/schemas/Labrador'
      discriminator:
        propertyName: breed
    Cat:
      type: object
      properties:
        petType:
          type: string
        name:
          type: string
      required:
        - petType
    Doberman:
      type: object
      properties:
        petType:
          type: string
        # Missing breed property required by discriminator
        aggressiveness:
          type: integer
      required:
        - petType
    Labrador:
      type: object
      properties:
        petType:
          type: string
        breed:
          type: string
        friendliness:
          type: integer
      required:
        - petType
        - breed
`,
			},
			wantWarnings: []string{"object must have a property named breed to match discriminator propertyName"},
		},
		{
			name: "nested oneOf discriminator property wrong type logs warning",
			args: args{
				schema: openapiHeaderDiscriminator + `
components:
  schemas:
    Pet:
      x-speakeasy-include: true
      oneOf:
        - $ref: '#/components/schemas/Dog'
        - $ref: '#/components/schemas/Cat'
      discriminator:
        propertyName: petType
    Dog:
      oneOf:
        - $ref: '#/components/schemas/Doberman'
        - $ref: '#/components/schemas/Labrador'
      discriminator:
        propertyName: breed
    Cat:
      type: object
      properties:
        petType:
          type: string
        name:
          type: string
      required:
        - petType
    Doberman:
      type: object
      properties:
        petType:
          type: string
        breed:
          type: integer  # Wrong type - should be string or enum
        aggressiveness:
          type: integer
      required:
        - petType
        - breed
    Labrador:
      type: object
      properties:
        petType:
          type: string
        breed:
          type: string
        friendliness:
          type: integer
      required:
        - petType
        - breed
`,
			},
			wantWarnings: []string{"discriminator property named breed must be of type string or enum"},
		},
		{
			name: "deeply nested oneOf missing discriminator property logs warning",
			args: args{
				schema: openapiHeaderDiscriminator + `
components:
  schemas:
    Animal:
      x-speakeasy-include: true
      oneOf:
        - $ref: '#/components/schemas/Pet'
        - $ref: '#/components/schemas/WildAnimal'
      discriminator:
        propertyName: animalType
    Pet:
      oneOf:
        - $ref: '#/components/schemas/Dog'
        - $ref: '#/components/schemas/Cat'
      discriminator:
        propertyName: petType
    Dog:
      oneOf:
        - $ref: '#/components/schemas/Doberman'
        - $ref: '#/components/schemas/Labrador'
      discriminator:
        propertyName: breed
    Cat:
      type: object
      properties:
        animalType:
          type: string
        petType:
          type: string
        name:
          type: string
      required:
        - animalType
        - petType
    WildAnimal:
      type: object
      properties:
        # Missing animalType property required by discriminator
        habitat:
          type: string
    Doberman:
      type: object
      properties:
        animalType:
          type: string
        petType:
          type: string
        breed:
          type: string
        aggressiveness:
          type: integer
      required:
        - animalType
        - petType
        - breed
    Labrador:
      type: object
      properties:
        animalType:
          type: string
        petType:
          type: string
        breed:
          type: string
        friendliness:
          type: integer
      required:
        - animalType
        - petType
        - breed
`,
			},
			wantWarnings: []string{"object must have a property named animalType to match discriminator propertyName"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			errs, warnings := testGenerateWithWarnings(t, []byte(tt.args.schema))

			// These cases now log warnings instead of returning errors
			// Generation should succeed
			require.Empty(t, errs)

			// Check that expected warnings were logged
			var warningMessages []string
			for _, w := range warnings {
				warningMessages = append(warningMessages, w.Error())
			}

			found := false
			for _, warningMsg := range warningMessages {
				for _, expectedWarning := range tt.wantWarnings {
					if containsIgnoringLineNumbers(warningMsg, expectedWarning) {
						found = true
						break
					}
				}
				if found {
					break
				}
			}

			if !found {
				t.Errorf("Expected warning containing one of %v, but got warnings: %v", tt.wantWarnings, warningMessages)
			}
		})
	}
}

func testGenerateWithWarnings(t *testing.T, contents []byte) ([]error, []error) {
	t.Helper()

	ctx := generationaccess.WithDirect(logging.With(context.Background(), logger), generationaccess.ElectAGPL())

	outDir := t.TempDir()
	schemaPath := filepath.Join(outDir, "openapi.yaml")
	if err := os.WriteFile(schemaPath, contents, 0o644); err != nil {
		return []error{err}, nil
	}

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return []error{err}, nil
	}

	g, err := generate.New(
		generate.WithLogger(logger),
		generate.WithRunLocation("cli"),
		generate.WithDebuggingEnabled(),
	)
	if err != nil {
		return []error{err}, nil
	}

	errs := g.Generate(ctx, schema, schemaPath, "go", outDir, false, false)
	warnings := g.GetWarnings()

	return errs, warnings
}

// containsIgnoringLineNumbers checks if the actual error contains the expected error message,
// ignoring line number information like "[line 41]"
func containsIgnoringLineNumbers(actual, expected string) bool {
	// Remove line number patterns like "[line 41] " from the actual error
	re := regexp.MustCompile(`\[line \d+\] `)
	cleanActual := re.ReplaceAllString(actual, "")
	return strings.Contains(cleanActual, expected)
}
