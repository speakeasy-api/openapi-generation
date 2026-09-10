package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/overlay/loader"
	"gopkg.in/yaml.v3"
)

type arrayFlag []string

var _ flag.Value = &arrayFlag{}

// Set implements the flag.Value interface.
func (i *arrayFlag) Set(value string) error {
	*i = append(*i, value)
	return nil
}

// String implements the flag.Value interface.
func (i *arrayFlag) String() string {
	return fmt.Sprintf("%v", *i)
}

type joinConflictStrategy string

const (
	joinConflictCounter  joinConflictStrategy = "counter"
	joinConflictFilePath joinConflictStrategy = "filepath"
)

func (s joinConflictStrategy) toOpenAPI() (openapi.JoinConflictStrategy, error) {
	switch s {
	case "", joinConflictCounter:
		return openapi.JoinConflictCounter, nil
	case joinConflictFilePath:
		return openapi.JoinConflictFilePath, nil
	default:
		return 0, fmt.Errorf("unsupported join conflict strategy %q", s)
	}
}

func encodeYAMLDocument(y *yaml.Node) ([]byte, error) {
	var res bytes.Buffer

	encoder := yaml.NewEncoder(&res)
	encoder.SetIndent(2)

	if err := encoder.Encode(y); err != nil {
		return nil, fmt.Errorf("failed to encode YAML: %w", err)
	}

	return res.Bytes(), nil
}

func loadOpenAPIDocument(ctx context.Context, path string) (*openapi.OpenAPI, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open OpenAPI document at %s: %w", path, err)
	}
	defer file.Close()

	doc, _, err := openapi.Unmarshal(ctx, file, openapi.WithSkipValidation())
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI document at %s: %w", path, err)
	}

	return doc, nil
}

func joinSpecifications(ctx context.Context, sourceSpecPath string, joinPaths []string, strategy joinConflictStrategy) (*yaml.Node, error) {
	if len(joinPaths) == 0 {
		node, err := loader.LoadSpecification(sourceSpecPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load source OpenAPI Specification document at %s: %w", sourceSpecPath, err)
		}
		return node, nil
	}

	mainDoc, err := loadOpenAPIDocument(ctx, sourceSpecPath)
	if err != nil {
		return nil, err
	}

	documents := make([]openapi.JoinDocumentInfo, 0, len(joinPaths))
	for _, joinPath := range joinPaths {
		doc, err := loadOpenAPIDocument(ctx, joinPath)
		if err != nil {
			return nil, err
		}

		documents = append(documents, openapi.JoinDocumentInfo{
			Document: doc,
			FilePath: joinPath,
		})
	}

	openapiStrategy, err := strategy.toOpenAPI()
	if err != nil {
		return nil, err
	}

	if err := openapi.Join(ctx, mainDoc, documents, openapi.JoinOptions{
		ConflictStrategy: openapiStrategy,
	}); err != nil {
		return nil, fmt.Errorf("failed to join OpenAPI specifications: %w", err)
	}

	var buf bytes.Buffer
	if err := openapi.Marshal(ctx, mainDoc, &buf); err != nil {
		return nil, fmt.Errorf("failed to marshal joined OpenAPI specification: %w", err)
	}

	node, err := loader.LoadSpecificationFromReader(bytes.NewReader(buf.Bytes()))
	if err != nil {
		return nil, fmt.Errorf("failed to parse joined OpenAPI specification: %w", err)
	}

	return node, nil
}

func applyOverlays(specNode *yaml.Node, overlayPaths []string, strict bool) error {
	for _, overlayPath := range overlayPaths {
		overlayDoc, err := loader.LoadOverlay(overlayPath)
		if err != nil {
			return fmt.Errorf("error loading OpenAPI Overlay at %s: %w", overlayPath, err)
		}

		if err := overlayDoc.Validate(); err != nil {
			return fmt.Errorf("failed to validate OpenAPI Overlay at %s: %w", overlayPath, err)
		}

		if strict {
			warnings, err := overlayDoc.ApplyToStrict(specNode)
			for _, warning := range warnings {
				log.Printf("overlay apply warning: %s", warning)
			}
			if err != nil {
				return fmt.Errorf("failed to apply OpenAPI Overlay %s: %w", overlayPath, err)
			}
			continue
		}

		if err := overlayDoc.ApplyTo(specNode); err != nil {
			return fmt.Errorf("failed to apply OpenAPI Overlay %s: %w", overlayPath, err)
		}
	}

	return nil
}

func writeSpecification(path string, specNode *yaml.Node) error {
	destinationSpecBytes, err := encodeYAMLDocument(specNode)
	if err != nil {
		return err
	}

	destinationSpecFile, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("failed to create destination OpenAPI Specification document at %s: %w", path, err)
	}
	defer destinationSpecFile.Close()

	if _, err := io.Copy(destinationSpecFile, bytes.NewReader(destinationSpecBytes)); err != nil {
		return fmt.Errorf("failed to write to destination OpenAPI Specification document at %s: %w", path, err)
	}

	return nil
}

func main() {
	ctx := context.Background()
	var joins arrayFlag
	var overlays arrayFlag

	source := flag.String("s", "", "Path to the source OpenAPI Specification document")
	destination := flag.String("out", "", "Path to the destination OpenAPI Specification document")
	joinStrategy := flag.String("join-conflict-strategy", string(joinConflictCounter), "Conflict strategy for joined OpenAPI specs: counter or filepath")
	flag.Var(&joins, "join", "Path to an additional OpenAPI Specification document to join before applying overlays. May be specified multiple times.")
	flag.Var(&overlays, "overlay", "Path to OpenAPI Overlay document. May be specified multiple times.")
	disableStrict := flag.Bool("disable-strict", false, "Disable strict overlay application checks.")

	flag.Parse()

	if source == nil || *source == "" {
		log.Fatalf("source OpenAPI Specification document path (-s PATH) is required")
	}

	if destination == nil || *destination == "" {
		log.Fatalf("destination OpenAPI Specification document path (-out PATH) is required")
	}

	if len(joins) == 0 && len(overlays) == 0 {
		log.Fatalf("at least one OpenAPI Overlay (-overlay PATH) or joined spec (-join PATH) is required")
	}

	specNode, err := joinSpecifications(ctx, *source, []string(joins), joinConflictStrategy(*joinStrategy))
	if err != nil {
		log.Fatalf("failed to prepare OpenAPI Specification document: %s", err)
	}

	if len(overlays) > 0 {
		if err := applyOverlays(specNode, []string(overlays), !*disableStrict); err != nil {
			log.Fatalf("failed to apply OpenAPI Overlay documents: %s", err)
		}
	}

	if err := writeSpecification(*destination, specNode); err != nil {
		log.Fatalf("failed to write OpenAPI Specification document: %s", err)
	}
}
