package generate

import (
	"context"
	"fmt"
	"sync"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/js/template"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
)

type sanitizer struct {
	mu            sync.Mutex
	e             *template.Engine
	fileNameCache sync.Map // string -> string
}

func (s *sanitizer) GetEnumNames(t *ast.TypeDef) []string {
	s.mu.Lock()
	defer s.mu.Unlock()

	v, err := s.e.RunFunction(context.Background(), "getEnumNames", t)
	if err != nil {
		panic(fmt.Errorf("failed to get enum names: %w", err))
	}

	exported := v.Export().([]interface{})
	enumNames := make([]string, 0, len(exported))

	for _, v := range exported {
		enumNames = append(enumNames, v.(string))
	}

	return enumNames
}

func (s *sanitizer) SanitizeFileName(name string) string {
	if cached, ok := s.fileNameCache.Load(name); ok {
		return cached.(string)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring lock
	if cached, ok := s.fileNameCache.Load(name); ok {
		return cached.(string)
	}

	fileName, err := s.e.RunFunction(context.Background(), "sanitizeFileName", name)
	if err != nil {
		panic(fmt.Errorf("failed to sanitizeFileName: %w", err))
	}

	result := fileName.String()
	s.fileNameCache.Store(name, result)
	return result
}

func (g *Generator) getSanitizer(ctx context.Context, target types.Target) (*sanitizer, error) {
	e := g.getExecutor(ctx, target, g.onWriteFile(""), g.onReadFile(""))

	// Build a minimal config context for the sanitizer so it can access config values like FileNaming
	cfg := g.getBaseTemplateConfigs(target, g.subsystem.Config, g.lockFile)

	if err := e.Init(ctx, GlobalContext{
		Config: cfg,
	}); err != nil {
		return nil, fmt.Errorf("failed to initialize sanitizer: %w", err)
	}

	if err := e.RunScript(ctx, "common/templating.ts"); err != nil {
		return nil, fmt.Errorf("failed to run common/templating script: %w", err)
	}
	if err := e.RunScript(ctx, "includes/sanitization.ts"); err != nil {
		return nil, fmt.Errorf("failed to run sanitization script: %w", err)
	}

	return &sanitizer{
		e: e,
	}, nil
}
