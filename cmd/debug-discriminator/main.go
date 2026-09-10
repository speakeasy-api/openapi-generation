package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/logging"
)

func main() {
	schema := flag.String("s", "", "path to the schema file or directory")
	flag.Parse()

	if *schema == "" {
		fmt.Println("Usage: go run cmd/debug-discriminator/main.go -s <path-to-schema-or-directory>")
		os.Exit(1)
	}

	_ = os.Setenv("SPEAKEASY_DEBUG", "true")
	_ = os.Setenv("SPEAKEASY_DEBUG_INFER_DISCRIMINATORS", "true")

	logger := logging.NewLogger(logging.LevelFromEnv())
	ctx := logging.With(context.Background(), logger)

	if err := debugDiscriminator(ctx, *schema); err != nil {
		logger.Fatal(err.Error())
	}
}

func pickRandomOpenAPIFile(dir string) (string, error) {
	var files []string

	// Clean the directory path to handle trailing slashes
	dir = filepath.Clean(dir)

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			ext := strings.ToLower(filepath.Ext(path))
			// Only consider yaml files in the root directory (not subdirectories)
			if (ext == ".yaml" || ext == ".yml") && filepath.Dir(path) == dir {
				// Check if it looks like an OpenAPI file
				content, err := os.ReadFile(path)
				if err == nil {
					// Skip files containing "petstore" in contents
					if strings.Contains(strings.ToLower(string(content)), "petstore") {
						return nil
					}
					files = append(files, path)
				}
			}
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", errors.New("no OpenAPI files found in directory")
	}

	return files[rand.Intn(len(files))], nil
}

func debugDiscriminator(ctx context.Context, schemaPath string) error {
	logger := logging.From(ctx)
	fmt.Println("Debugging Union Discriminator Inference")

	// Check if path is a directory and pick a random file if so
	info, err := os.Stat(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to stat path %s: %w", schemaPath, err)
	}

	if info.IsDir() {
		schemaPath, err = pickRandomOpenAPIFile(schemaPath)
		if err != nil {
			return fmt.Errorf("failed to find OpenAPI file in directory: %w", err)
		}
		fmt.Printf("Selected file: %s\n\n", schemaPath)
	}

	schema, err := os.ReadFile(schemaPath)
	if err != nil {
		return fmt.Errorf("failed to read schema file %s: %w", schemaPath, err)
	}

	// Use TestResolveAST to build the AST without full generation
	g, _, err := generate.TestResolveAST(generate.TestResolveASTInput{
		OpenAPIContents: schema,
		OpenAPIPath:     schemaPath,
	})
	if err != nil {
		return fmt.Errorf("failed to resolve AST: %w", err)
	}

	// Get the subsystem to access the type register
	subsystem := g.GetSubsystem()
	if subsystem == nil || subsystem.Register.Types == nil {
		return errors.New("subsystem or type register not initialized")
	}

	types := subsystem.Register.AllTypes()

	logger.Info(fmt.Sprintf("Analyzed %d types for union discriminators", len(types)))

	// Count results
	unionsFound := 0
	discriminatorsInferred := 0
	for _, t := range types {
		if t.Type == "union" {
			unionsFound++
			if t.Discriminator != nil && t.Discriminator.Inferred {
				discriminatorsInferred++
			}
		}
	}

	fmt.Printf("\nResults:\n")
	fmt.Printf("  Total types: %d\n", len(types))
	fmt.Printf("  Union types: %d\n", unionsFound)
	fmt.Printf("  Discriminators inferred: %d\n", discriminatorsInferred)

	return nil
}
