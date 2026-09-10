package server

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	generationaccess "github.com/speakeasy-api/generation-context/access"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/filesystem"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/generate"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/mode"
	"github.com/speakeasy-api/openapi-generation/v2/pkg/oas"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/pointer"
	"github.com/tliron/glsp"
	protocol "github.com/tliron/glsp/protocol_3_16"
	"gopkg.in/yaml.v3"
)

type Server struct {
	//nolint:containedctx
	Context      context.Context
	Handler      glsp.Handler
	Log          Logger
	Timeout      time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

var (
	DefaultTimeout = time.Minute
	mu             sync.Mutex
)

const languageServerName = "speakeasy"

func newServer(handler glsp.Handler, log Logger) *Server {
	return &Server{
		Handler:      handler,
		Log:          log,
		Timeout:      DefaultTimeout,
		ReadTimeout:  DefaultTimeout,
		WriteTimeout: DefaultTimeout,
	}
}

func ptr[T any](v T) *T {
	return &v
}

type LanguageServer struct {
	Server        *Server
	fs            filesystem.FileSystem
	documents     map[string]*Document
	schemaManager *oas.SchemaManager // Global schema manager for efficient caching
}

type Document struct {
	URI               protocol.DocumentUri
	RunningDiagnostic bool
	Content           string
	ParsedDoc         *openapi.OpenAPI            // Parsed OpenAPI document
	Index             *openapi.Index              // Document index for finding references
	ParseErrors       []error                     // Errors from parsing the document
	ReservedChecker   *oas.DynamicReservedChecker // Dynamic reserved property checker
}

// parseOpenAPIDocument parses an OpenAPI document from YAML/JSON content and builds an index
func parseOpenAPIDocument(content string, uri string) (*openapi.OpenAPI, *openapi.Index, []error) {
	ctx := context.Background()
	parsedDoc, parseErrs, err := openapi.Unmarshal(ctx, bytes.NewReader([]byte(content)))
	var allParseErrors []error

	if err != nil {
		allParseErrors = append(allParseErrors, err)
	}
	allParseErrors = append(allParseErrors, parseErrs...)

	// Build index if we have a parsed document
	var idx *openapi.Index
	if parsedDoc != nil {
		idx = openapi.BuildIndex(ctx, parsedDoc, openapi.ResolveOptions{
			RootDocument:        parsedDoc,
			TargetDocument:      parsedDoc,
			TargetLocation:      uri,
			DisableExternalRefs: true, // Disable external refs for language server
		})

		// Add index errors to parse errors
		if idx.HasErrors() {
			allParseErrors = append(allParseErrors, idx.GetValidationErrors()...)
			allParseErrors = append(allParseErrors, idx.GetResolutionErrors()...)
			allParseErrors = append(allParseErrors, idx.GetCircularReferenceErrors()...)
		}
	}

	return parsedDoc, idx, allParseErrors
}

func NewSpeakeasyServer(version string, fs filesystem.FileSystem) *LanguageServer {
	handler := protocol.Handler{}
	server := newServer(&handler, nil)

	ls := &LanguageServer{
		Server:        server,
		documents:     map[string]*Document{},
		fs:            fs,
		schemaManager: oas.NewSchemaManager(),
	}
	handler.Initialize = func(context *glsp.Context, params *protocol.InitializeParams) (any, error) {
		if params.Trace != nil {
			protocol.SetTraceValue(*params.Trace)
		}

		serverCapabilities := handler.CreateServerCapabilities()
		serverCapabilities.TextDocumentSync = protocol.TextDocumentSyncKindIncremental
		serverCapabilities.CompletionProvider = &protocol.CompletionOptions{}
		serverCapabilities.HoverProvider = true
		serverCapabilities.DocumentSymbolProvider = true
		serverCapabilities.DefinitionProvider = true
		serverCapabilities.RenameProvider = &protocol.RenameOptions{
			PrepareProvider: ptr(true),
		}
		serverCapabilities.Workspace = &protocol.ServerCapabilitiesWorkspace{
			WorkspaceFolders: &protocol.WorkspaceFoldersServerCapabilities{
				Supported:           ptr(true),
				ChangeNotifications: &protocol.BoolOrString{Value: true},
			},
		}

		return protocol.InitializeResult{
			Capabilities: serverCapabilities,
			ServerInfo: &protocol.InitializeResultServerInfo{
				Name:    languageServerName,
				Version: &version,
			},
		}, nil
	}
	handler.Initialized = func(context *glsp.Context, params *protocol.InitializedParams) error {
		return nil
	}
	handler.Shutdown = func(context *glsp.Context) error {
		return nil
	}
	handler.SetTrace = func(context *glsp.Context, params *protocol.SetTraceParams) error {
		protocol.SetTraceValue(params.Value)
		return nil
	}
	handler.TextDocumentDidOpen = func(glspContext *glsp.Context, params *protocol.DidOpenTextDocumentParams) error {
		parsedDoc, idx, parseErrors := parseOpenAPIDocument(params.TextDocument.Text, params.TextDocument.URI)
		doc := &Document{
			URI:             params.TextDocument.URI,
			Content:         params.TextDocument.Text,
			ParsedDoc:       parsedDoc,
			Index:           idx,
			ParseErrors:     parseErrors,
			ReservedChecker: oas.NewDynamicReservedCheckerWithManager(ls.schemaManager),
		}
		ls.documents[params.TextDocument.URI] = doc
		go ls.runDiagnostic(doc, glspContext.Notify, false)
		return nil
	}
	handler.TextDocumentDidChange = func(glspContext *glsp.Context, params *protocol.DidChangeTextDocumentParams) error {
		doc, ok := ls.documents[params.TextDocument.URI]
		if !ok {
			return nil
		}
		for _, change := range params.ContentChanges {
			switch c := change.(type) {
			case protocol.TextDocumentContentChangeEvent:
				startIndex, endIndex := c.Range.IndexesIn(doc.Content)
				doc.Content = doc.Content[:startIndex] + c.Text + doc.Content[endIndex:]
			case protocol.TextDocumentContentChangeEventWhole:
				doc.Content = c.Text
			}
		}
		parsedDoc, idx, parseErrors := parseOpenAPIDocument(doc.Content, doc.URI)
		doc.ParsedDoc = parsedDoc
		doc.Index = idx
		doc.ParseErrors = parseErrors
		// Initialize dynamic checker if not present
		if doc.ReservedChecker == nil {
			doc.ReservedChecker = oas.NewDynamicReservedCheckerWithManager(ls.schemaManager)
		}
		go ls.runDiagnostic(doc, glspContext.Notify, true)
		return nil
	}
	handler.TextDocumentDidClose = func(context *glsp.Context, params *protocol.DidCloseTextDocumentParams) error {
		delete(ls.documents, params.TextDocument.URI)

		return nil
	}

	handler.TextDocumentCompletion = func(context *glsp.Context, params *protocol.CompletionParams) (any, error) {
		return nil, nil
	}

	handler.TextDocumentHover = func(context *glsp.Context, params *protocol.HoverParams) (*protocol.Hover, error) {
		return ls.HandleHover(params)
	}

	handler.TextDocumentPrepareRename = func(context *glsp.Context, params *protocol.PrepareRenameParams) (any, error) {
		return ls.HandlePrepareRename(params)
	}

	handler.TextDocumentRename = func(context *glsp.Context, params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
		return ls.HandleRename(params)
	}

	return ls
}

func (s *LanguageServer) runDiagnostic(doc *Document, notify glsp.NotifyFunc, _ bool) {
	mu.Lock()
	defer mu.Unlock()

	type openapi struct {
		OpenAPI string `yaml:"openapi"`
	}

	var oas openapi
	if err := yaml.Unmarshal([]byte(doc.Content), &oas); err != nil {
		notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         doc.URI,
			Diagnostics: fatalDiagnostic(fmt.Sprintf("Failed to unmarshal document: %v\n", err)),
		})
		fmt.Printf("Failed to unmarshal document: %v\n", err)
		return
	}
	if oas.OpenAPI == "" {
		notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
			URI:         doc.URI,
			Diagnostics: fatalDiagnostic("Not an OpenAPI document"),
		})

		fmt.Printf("Not an OpenAPI document, skipping validation\n")
		return
	}

	ctx := generationaccess.WithDirect(context.Background(), generationaccess.ElectAGPL())
	ctx = mode.SetSpeakeasyExecutionContextValidation(ctx)
	ctx = mode.SetSpeakeasyExecutionContextEmbedded(ctx)
	diagnostics := s.Validate(ctx, []byte(doc.Content))

	notify(protocol.ServerTextDocumentPublishDiagnostics, protocol.PublishDiagnosticsParams{
		URI:         doc.URI,
		Diagnostics: diagnostics,
	})
}

func fatalDiagnostic(message string) []protocol.Diagnostic {
	severity := protocol.DiagnosticSeverityError
	code := protocol.IntegerOrString{Value: "SPEAKEASY_FATAL"}

	return []protocol.Diagnostic{
		{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      0,
					Character: 0,
				},
				End: protocol.Position{
					Line:      0,
					Character: ^uint32(0),
				},
			},
			Severity:        &severity,
			Source:          pointer.From("speakeasy"),
			Code:            &code,
			CodeDescription: &protocol.CodeDescription{HRef: "https://www.speakeasy.com/docs/openapi/validation"},
			Message:         message,
		},
	}
}

func (s *LanguageServer) Validate(ctx context.Context, schema []byte) []protocol.Diagnostic {
	opts := make([]generate.GeneratorOptions, 0, 3)
	opts = append(opts, generate.WithRunLocation("cli"))
	opts = append(opts, generate.WithDebuggingEnabled())
	opts = append(opts, generate.WithFileSystem(s.fs))

	g, err := generate.New(opts...)
	if err != nil {
		return nil
	}

	res, err := g.Validate(ctx, schema, "", false, "")
	if err != nil {
		fmt.Printf("Validation failed with error: %v\n", err)
		return nil
	}

	vErrs := res.GetValidationErrors()

	diagnostics := make([]protocol.Diagnostic, 0, len(vErrs))
	for _, err := range vErrs {
		vErr := errors.GetValidationErr(err)

		severity := protocol.DiagnosticSeverityError
		code := protocol.IntegerOrString{Value: "unknown"}

		if vErr.Rule == "resolving-references" {
			fmt.Printf("Skipping 'resolving-references' error\n")
			continue
		}

		lineNumber := vErr.GetLineNumber() - 1
		columnNumber := vErr.GetColumnNumber() - 1
		code = protocol.IntegerOrString{Value: vErr.Rule}

		switch vErr.Severity {
		case errors.SeverityError:
			severity = protocol.DiagnosticSeverityError
		case errors.SeverityWarn:
			severity = protocol.DiagnosticSeverityWarning
		case errors.SeverityHint:
			severity = protocol.DiagnosticSeverityHint
		}

		message := vErr.Message

		diagnostics = append(diagnostics, protocol.Diagnostic{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      protocol.UInteger(lineNumber),
					Character: protocol.UInteger(columnNumber),
				},
				End: protocol.Position{
					Line:      protocol.UInteger(lineNumber),
					Character: ^uint32(0),
				},
			},
			Severity:        &severity,
			Source:          pointer.From("speakeasy"),
			Code:            &code,
			CodeDescription: &protocol.CodeDescription{HRef: "https://www.speakeasy.com/docs/prep-openapi/linting#available-rules"},
			Message:         message,
		})
	}
	return diagnostics
}

func (s *LanguageServer) HandleHover(params *protocol.HoverParams) (*protocol.Hover, error) {
	doc, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return nil, nil
	}

	definition, err := lookupDefinition(doc, params)
	if err != nil {
		// Swallow the errors so that "No documentation found" doesn't surface to the console
		//nolint:nilerr
		return nil, nil
	}

	return &protocol.Hover{
		Contents: protocol.MarkupContent{
			Kind:  protocol.MarkupKindMarkdown,
			Value: definition.documentation,
		},
		Range: &protocol.Range{
			Start: protocol.Position{
				Line:      protocol.UInteger(definition.node.Line - 1),
				Character: protocol.UInteger(definition.node.Column - 1),
			},
			End: protocol.Position{
				Line:      protocol.UInteger(definition.node.Line - 1),
				Character: protocol.UInteger(definition.node.Column - 1 + len(definition.node.Value)),
			},
		},
	}, nil
}

func (s *LanguageServer) HandlePrepareRename(params *protocol.PrepareRenameParams) (*protocol.RangeWithPlaceholder, error) {
	doc, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return nil, nil
	}

	var root yaml.Node
	if err := yaml.Unmarshal([]byte(doc.Content), &root); err != nil {
		return nil, fmt.Errorf("unable to unmarshall doc.Content into yaml: %w", err)
	}

	line := int(params.Position.Line)
	character := int(params.Position.Character)
	node, path := findNodeAtPosition(&root, line, character, []string{})
	if node == nil {
		return nil, errors.New("no symbol found at position")
	}

	// Validate that this is a renameable symbol
	if !isRenameableSymbol(node, path, doc) {
		return nil, errors.New("symbol cannot be renamed")
	}

	// Return the range of the symbol
	return &protocol.RangeWithPlaceholder{
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      protocol.UInteger(node.Line - 1),
				Character: protocol.UInteger(node.Column - 1),
			},
			End: protocol.Position{
				Line:      protocol.UInteger(node.Line - 1),
				Character: protocol.UInteger(node.Column - 1 + len(node.Value)),
			},
		},
		Placeholder: node.Value,
	}, nil
}

func (s *LanguageServer) HandleRename(params *protocol.RenameParams) (*protocol.WorkspaceEdit, error) {
	doc, ok := s.documents[params.TextDocument.URI]
	if !ok {
		return nil, nil
	}

	var root yaml.Node
	if err := yaml.Unmarshal([]byte(doc.Content), &root); err != nil {
		return nil, fmt.Errorf("unable to unmarshall doc.Content into yaml: %w", err)
	}

	line := int(params.Position.Line)
	character := int(params.Position.Character)
	node, path := findNodeAtPosition(&root, line, character, []string{})
	if node == nil {
		return nil, errors.New("no symbol found at position")
	}

	// Validate that this is a renameable symbol
	if !isRenameableSymbol(node, path, doc) {
		return nil, errors.New("symbol cannot be renamed")
	}

	// Use index-based approach if available
	var textEdits []protocol.TextEdit
	if doc.Index != nil {
		textEdits = s.findReferencesWithRolodex(doc, node, path, params.NewName)
	} else {
		// Fallback to manual traversal
		references := findReferences(&root, node.Value, path)
		// Extract just the name part from user input in case they provided a full reference path
		actualNewName := extractNameFromUserInput(params.NewName)
		textEdits = s.createTextEditsFromReferences(references, node, path, actualNewName)
	}

	return &protocol.WorkspaceEdit{
		Changes: map[protocol.DocumentUri][]protocol.TextEdit{
			params.TextDocument.URI: textEdits,
		},
	}, nil
}

// findReferencesWithRolodex uses libopenapi's rolodex to find references more accurately
func (s *LanguageServer) findReferencesWithRolodex(doc *Document, node *yaml.Node, path []string, newName string) []protocol.TextEdit {
	var textEdits []protocol.TextEdit
	seen := make(map[string]bool) // Track already processed positions

	// Index is already built when the document was parsed

	// Start with the original node
	originalPos := fmt.Sprintf("%d:%d", node.Line-1, node.Column-1)
	var originalNewText string

	line := node.Line
	column := node.Column

	// Since we don't allow renaming references directly, this should be a definition name
	originalNewText = newName

	originalEdit := protocol.TextEdit{
		Range: protocol.Range{
			Start: protocol.Position{
				Line:      protocol.UInteger(line - 1),
				Character: protocol.UInteger(column - 1),
			},
			End: protocol.Position{
				Line:      protocol.UInteger(line - 1),
				Character: protocol.UInteger(column - 1 + len(node.Value)),
			},
		},
		NewText: originalNewText,
	}
	textEdits = append(textEdits, originalEdit)
	seen[originalPos] = true

	// Handle different types of references using the rolodex
	var symbolName string
	// Extract just the name part from user input in case they provided a full reference path
	actualNewName := extractNameFromUserInput(newName)

	// Check what type of symbol this is based on the path context
	if isComponentReference(path) {
		symbolName = node.Value
		s.findSchemaReferencesAndDefinition(doc, symbolName, actualNewName, &textEdits, seen)
	} else if isInlinePathParameter(path) {
		// Handle inline path parameters
		symbolName = node.Value
		pathTemplate := getPathFromInlineParameter(path)
		s.findInlinePathParameterReferences(doc, symbolName, pathTemplate, actualNewName, &textEdits, seen)
	}

	// If no references found via rolodex, fall back to manual search
	if len(textEdits) == 1 {

		var root yaml.Node
		if err := yaml.Unmarshal([]byte(doc.Content), &root); err == nil {
			references := findReferences(&root, node.Value, path)
			for _, ref := range references {
				refPos := fmt.Sprintf("%d:%d", ref.Line-1, ref.Column-1)

				// Skip already processed positions
				if seen[refPos] {
					continue
				}

				// Only process valid references with proper position info
				if ref.Column > 0 && len(ref.Value) > 0 {
					var newText string
					if strings.HasPrefix(ref.Value, "#/") {
						// Handle both schema and parameter references
						newText = strings.Replace(ref.Value, node.Value, actualNewName, 1)
					} else {
						newText = actualNewName
					}

					// Calculate end character position
					endChar := ref.Column + len(ref.Value)

					fallbackEdit := protocol.TextEdit{
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      protocol.UInteger(ref.Line - 1),
								Character: protocol.UInteger(ref.Column),
							},
							End: protocol.Position{
								Line:      protocol.UInteger(ref.Line - 1),
								Character: protocol.UInteger(endChar),
							},
						},
						NewText: newText,
					}
					textEdits = append(textEdits, fallbackEdit)
					seen[refPos] = true
				}
			}
		}
	}

	return textEdits
}

// findInlinePathParameterReferences finds all references to an inline path parameter and updates the path template
func (s *LanguageServer) findInlinePathParameterReferences(doc *Document, parameterName string, pathTemplate string, newName string, textEdits *[]protocol.TextEdit, seen map[string]bool) {
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(doc.Content), &root); err != nil {
		return
	}

	// Find all inline parameters with the same name in the same path
	s.findInlineParametersInPath(&root, []string{}, pathTemplate, parameterName, newName, textEdits, seen)

	// Find and update the path template
	s.updatePathTemplate(&root, []string{}, pathTemplate, parameterName, newName, textEdits, seen)
}

// findInlineParametersInPath finds all inline parameters with the given name in the specified path
func (s *LanguageServer) findInlineParametersInPath(node *yaml.Node, currentPath []string, targetPath string, parameterName string, newName string, textEdits *[]protocol.TextEdit, seen map[string]bool) {
	if node == nil {
		return
	}

	// Handle DocumentNode by traversing into its content
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) > 0 {
			s.findInlineParametersInPath(node.Content[0], currentPath, targetPath, parameterName, newName, textEdits, seen)
		}
		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			newPath := append(currentPath, keyNode.Value)

			// Check if we're looking at a parameter name in the target path
			if len(newPath) >= 4 && newPath[0] == "paths" && newPath[1] == targetPath {
				// Look for parameter names in this path
				if keyNode.Value == "name" && valueNode.Value == parameterName {
					// This is a parameter name we want to rename
					paramPos := fmt.Sprintf("%d:%d", valueNode.Line-1, valueNode.Column-1)
					if !seen[paramPos] {
						paramEdit := protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      protocol.UInteger(valueNode.Line - 1),
									Character: protocol.UInteger(valueNode.Column - 1),
								},
								End: protocol.Position{
									Line:      protocol.UInteger(valueNode.Line - 1),
									Character: protocol.UInteger(valueNode.Column - 1 + len(valueNode.Value)),
								},
							},
							NewText: newName,
						}
						*textEdits = append(*textEdits, paramEdit)
						seen[paramPos] = true
					}
				}
			}

			// Continue recursively
			s.findInlineParametersInPath(keyNode, newPath, targetPath, parameterName, newName, textEdits, seen)
			s.findInlineParametersInPath(valueNode, newPath, targetPath, parameterName, newName, textEdits, seen)
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			newPath := append(currentPath, fmt.Sprintf("[%d]", i))
			s.findInlineParametersInPath(child, newPath, targetPath, parameterName, newName, textEdits, seen)
		}
	}
}

// updatePathTemplate finds and updates the path template to replace {oldName} with {newName}
func (s *LanguageServer) updatePathTemplate(node *yaml.Node, currentPath []string, targetPath string, parameterName string, newName string, textEdits *[]protocol.TextEdit, seen map[string]bool) {
	if node == nil {
		return
	}

	// Handle DocumentNode by traversing into its content
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) > 0 {
			s.updatePathTemplate(node.Content[0], currentPath, targetPath, parameterName, newName, textEdits, seen)
		}
		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			newPath := append(currentPath, keyNode.Value)

			// Check if this is the path template key we're looking for
			if len(newPath) == 2 && newPath[0] == "paths" && keyNode.Value == targetPath {
				// This is the path template key - check if it contains our parameter
				if strings.Contains(keyNode.Value, "{"+parameterName+"}") {
					pathPos := fmt.Sprintf("%d:%d", keyNode.Line-1, keyNode.Column-1)
					if !seen[pathPos] {
						// Replace {oldName} with {newName} in the path template
						newPathTemplate := strings.ReplaceAll(keyNode.Value, "{"+parameterName+"}", "{"+newName+"}")

						pathEdit := protocol.TextEdit{
							Range: protocol.Range{
								Start: protocol.Position{
									Line:      protocol.UInteger(keyNode.Line - 1),
									Character: protocol.UInteger(keyNode.Column - 1),
								},
								End: protocol.Position{
									Line:      protocol.UInteger(keyNode.Line - 1),
									Character: protocol.UInteger(keyNode.Column - 1 + len(keyNode.Value)),
								},
							},
							NewText: newPathTemplate,
						}
						*textEdits = append(*textEdits, pathEdit)
						seen[pathPos] = true
					}
				}
			}

			// Continue recursively
			s.updatePathTemplate(keyNode, newPath, targetPath, parameterName, newName, textEdits, seen)
			s.updatePathTemplate(valueNode, newPath, targetPath, parameterName, newName, textEdits, seen)
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			newPath := append(currentPath, fmt.Sprintf("[%d]", i))
			s.updatePathTemplate(child, newPath, targetPath, parameterName, newName, textEdits, seen)
		}
	}
}

// findAndRenameSchemaDefinition finds and renames the schema definition
func (s *LanguageServer) findAndRenameSchemaDefinition(root *yaml.Node, schemaName string, newName string, textEdits *[]protocol.TextEdit, seen map[string]bool) {
	// Look for the schema definition at components.schemas.{schemaName}
	s.findAndRenameDefinitionRecursive(root, []string{}, "schemas", schemaName, newName, textEdits, seen)
}

// findAndRenameDefinitionRecursive recursively searches for a definition to rename
func (s *LanguageServer) findAndRenameDefinitionRecursive(node *yaml.Node, currentPath []string, componentType string, symbolName string, newName string, textEdits *[]protocol.TextEdit, seen map[string]bool) {
	if node == nil {
		return
	}

	// Handle DocumentNode by traversing into its content
	if node.Kind == yaml.DocumentNode {
		if len(node.Content) > 0 {
			s.findAndRenameDefinitionRecursive(node.Content[0], currentPath, componentType, symbolName, newName, textEdits, seen)
		}
		return
	}

	switch node.Kind {
	case yaml.MappingNode:
		for i := 0; i < len(node.Content); i += 2 {
			keyNode := node.Content[i]
			valueNode := node.Content[i+1]

			newPath := append(currentPath, keyNode.Value)

			// Check if we found the definition we're looking for
			if len(newPath) == 3 && newPath[0] == "components" && newPath[1] == componentType && newPath[2] == symbolName {
				// This is the definition key node - rename it
				defPos := fmt.Sprintf("%d:%d", keyNode.Line-1, keyNode.Column-1)
				if !seen[defPos] {
					defEdit := protocol.TextEdit{
						Range: protocol.Range{
							Start: protocol.Position{
								Line:      protocol.UInteger(keyNode.Line - 1),
								Character: protocol.UInteger(keyNode.Column - 1),
							},
							End: protocol.Position{
								Line:      protocol.UInteger(keyNode.Line - 1),
								Character: protocol.UInteger(keyNode.Column - 1 + len(keyNode.Value)),
							},
						},
						NewText: newName,
					}
					*textEdits = append(*textEdits, defEdit)
					seen[defPos] = true
				}
			}

			// Continue recursively
			s.findAndRenameDefinitionRecursive(keyNode, newPath, componentType, symbolName, newName, textEdits, seen)
			s.findAndRenameDefinitionRecursive(valueNode, newPath, componentType, symbolName, newName, textEdits, seen)
		}
	case yaml.SequenceNode:
		for i, child := range node.Content {
			newPath := append(currentPath, fmt.Sprintf("[%d]", i))
			s.findAndRenameDefinitionRecursive(child, newPath, componentType, symbolName, newName, textEdits, seen)
		}
	}
}

// findSchemaReferencesAndDefinition finds both the schema definition and all references to it
func (s *LanguageServer) findSchemaReferencesAndDefinition(doc *Document, schemaName string, newName string, textEdits *[]protocol.TextEdit, seen map[string]bool) {
	// First, find the schema definition by traversing the YAML document
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(doc.Content), &root); err == nil {
		s.findAndRenameSchemaDefinition(&root, schemaName, newName, textEdits, seen)
	}

	// Then find all references using the index
	if doc.Index != nil {
		allRefs := doc.Index.GetAllReferences()
		for _, indexNode := range allRefs {
			if indexNode == nil || indexNode.Node == nil {
				continue
			}
			// Check if this reference points to our schema
			refString := indexNode.Node.GetReference().String()
			if strings.Contains(refString, "#/components/schemas/"+schemaName) {
				// Find the node for this reference
				if refNode := indexNode.Node.GetRootNode(); refNode != nil && refNode.Value != "" {
					refPos := fmt.Sprintf("%d:%d", refNode.Line-1, refNode.Column-1)
					if !seen[refPos] {
						// Only process actual $ref nodes, not the schema definition itself
						if strings.HasPrefix(refNode.Value, "#/components/schemas/") {
							// Create the new reference path by replacing schema name in the reference value
							newRefPath := strings.Replace(refNode.Value, schemaName, newName, 1)

							// Calculate end character position based on the actual reference value length
							endChar := refNode.Column - 1 + len(refNode.Value)

							refEdit := protocol.TextEdit{
								Range: protocol.Range{
									Start: protocol.Position{
										Line:      protocol.UInteger(refNode.Line - 1),
										Character: protocol.UInteger(refNode.Column - 1),
									},
									End: protocol.Position{
										Line:      protocol.UInteger(refNode.Line - 1),
										Character: protocol.UInteger(endChar),
									},
								},
								NewText: newRefPath,
							}
							*textEdits = append(*textEdits, refEdit)
							seen[refPos] = true
						}
					}
				}
			}
		}
	}
}

// createTextEditsFromReferences creates text edits from manual reference finding (fallback)
func (s *LanguageServer) createTextEditsFromReferences(references []*yaml.Node, node *yaml.Node, path []string, newName string) []protocol.TextEdit {
	textEdits := make([]protocol.TextEdit, 0, len(references))

	for _, ref := range references {
		var newText string
		var startChar, endChar int

		// Check if this is a JSON reference that needs to be updated
		if strings.HasPrefix(ref.Value, "#/") && isComponentReference(path) {
			// For JSON references, we need to replace only the schema name part
			oldSchemaName := node.Value
			newText = strings.Replace(ref.Value, oldSchemaName, newName, 1)

			// Calculate the precise range - replace the entire reference value
			// Make sure we have valid column information
			if ref.Column > 0 && len(ref.Value) > 0 {
				startChar = ref.Column - 1
				endChar = ref.Column - 1 + len(ref.Value)
			} else {
				// If column info is invalid, skip this reference
				continue
			}
		} else {
			// For direct references, just use the new name
			newText = newName

			// Calculate the range for the direct symbol
			if ref.Column > 0 && len(ref.Value) > 0 {
				startChar = ref.Column - 1
				endChar = ref.Column - 1 + len(ref.Value)
			} else {
				// If column info is invalid, skip this reference
				continue
			}
		}

		textEdits = append(textEdits, protocol.TextEdit{
			Range: protocol.Range{
				Start: protocol.Position{
					Line:      protocol.UInteger(ref.Line - 1),
					Character: protocol.UInteger(startChar),
				},
				End: protocol.Position{
					Line:      protocol.UInteger(ref.Line - 1),
					Character: protocol.UInteger(endChar),
				},
			},
			NewText: newText,
		})
	}

	return textEdits
}
