package generate

import (
	"context"
	"fmt"
	"time"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
)

// Generates a.TerraformAST based on a.MainSDK. Terraform generation requires
// both the SDK AST and Terraform AST.
func (g *Generator) generateTerraformAST(ctx context.Context, a *ast.AST) error {
	if g.target.Target != "terraform" {
		return nil
	}

	start := time.Now()
	g.log.Debug("starting generateTerraformAST")
	_, span := g.tracer.Start(ctx, "Generator.generateTerraformAST")
	defer span.End()

	generationConfig := g.subsystem.Config.GetLanguageConfig(g.target.Target).Cfg
	terraformProvider := ast.NewTerraformProvider()
	terraformProvider.TerraformTypeName = terraform.ProviderTypeName(generationConfig)

	entityOperations, err := a.EntityOperations()
	if err != nil {
		return fmt.Errorf("failed to get Terraform operations: %w", err)
	}

	for operationID, operation := range entityOperations.All() {
		if operation.ContainsTruncated() {
			return fmt.Errorf("operation %s data contains circular reference, which is not supported in Terraform schemas", operationID)
		}

		if err := terraformProvider.AddOperation(generationConfig, operation); err != nil {
			return err
		}
	}

	excludeEmptyObjectSchemas, _ := generationConfig["excludeEmptyObjectSchemas"].(bool)

	if err := terraformProvider.AssembleSchemas(ctx, excludeEmptyObjectSchemas); err != nil {
		return err
	}

	enableTypeDeduplication, _ := generationConfig["enableTypeDeduplication"].(bool)

	if err := terraformProvider.AnnotateTerraformSymbols(ctx, enableTypeDeduplication); err != nil {
		return err
	}

	elapsed := time.Since(start)
	g.log.Debug(fmt.Sprintf("generateTerraformAST completed in '%s'", elapsed))

	a.TerraformProvider = terraformProvider

	return nil
}
