package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"

	"github.com/speakeasy-api/jsonpath/pkg/jsonpath"
	"github.com/speakeasy-api/jsonpath/pkg/jsonpath/config"
	"github.com/speakeasy-api/jsonpath/pkg/jsonpath/token"
	"github.com/speakeasy-api/openapi/overlay"
	"gopkg.in/yaml.v3"
)

func CalculateOverlay(originalYAML, targetYAML, existingOverlay string) (string, error) {
	var orig yaml.Node
	err := yaml.Unmarshal([]byte(originalYAML), &orig)
	if err != nil {
		return "", fmt.Errorf("failed to parse source schema: %w", err)
	}
	var target yaml.Node
	err = yaml.Unmarshal([]byte(targetYAML), &target)
	if err != nil {
		return "", fmt.Errorf("failed to parse target schema: %w", err)
	}

	// we go from the original to a new version, then look at the extra overlays on top
	// of that, then add that to the existing overlay
	var existingOverlayDocument overlay.Overlay
	err = yaml.Unmarshal([]byte(existingOverlay), &existingOverlayDocument)
	if err != nil {
		return "", fmt.Errorf("failed to parse overlay schema in CalculateOverlay: %w", err)
	}
	existingOverlayDocument.JSONPathVersion = "rfc9535" // force this in the playground.
	// now modify the original using the existing overlay
	err = existingOverlayDocument.ApplyTo(&orig)
	if err != nil {
		return "", fmt.Errorf("failed to apply existing overlay: %w", err)
	}

	newOverlay, err := overlay.Compare("example overlay", &orig, target)
	if err != nil {
		return "", fmt.Errorf("failed to compare schemas: %w", err)
	}
	// For each new action, check if there's an existing action with the same target and replace it
	// special case, is there only one action and it targets the same as the last overlayDocument.Actions item entry, we'll just replace it.
	if len(newOverlay.Actions) == 1 && len(existingOverlayDocument.Actions) > 0 && newOverlay.Actions[0].Target == existingOverlayDocument.Actions[len(existingOverlayDocument.Actions)-1].Target {
		existingOverlayDocument.Actions[len(existingOverlayDocument.Actions)-1] = newOverlay.Actions[0]
	} else {
		// Otherwise, we'll just append the new overlay to the existing overlay
		existingOverlayDocument.Actions = append(existingOverlayDocument.Actions, newOverlay.Actions...)
	}

	// Use standard YAML marshaling - formatting will be normalized separately
	result, err := yaml.Marshal(existingOverlayDocument)
	if err != nil {
		return "", fmt.Errorf("failed to marshal overlay: %w", err)
	}

	return string(result), nil
}

// FormatYAML takes a YAML string, parses it, and re-marshals it with consistent formatting.
// This normalizes YAML formatting to eliminate spurious diffs caused by formatting differences.
func FormatYAML(yamlContent string) (string, error) {
	if strings.TrimSpace(yamlContent) == "" {
		return yamlContent, nil
	}

	var node overlay.Overlay
	err := yaml.Unmarshal([]byte(yamlContent), &node)
	if err != nil {
		return "", fmt.Errorf("failed to parse YAML: %w", err)
	}

	result, err := yaml.Marshal(&node)
	if err != nil {
		return "", fmt.Errorf("failed to marshal YAML: %w", err)
	}

	return string(result), nil
}

func GetInfo(originalYAML string) (string, error) {
	var orig yaml.Node
	err := yaml.Unmarshal([]byte(originalYAML), &orig)
	if err != nil {
		return "", fmt.Errorf("failed to parse source schema: %w", err)
	}

	titlePath, err := jsonpath.NewPath("$.info.title")
	if err != nil {
		return "", err
	}
	versionPath, err := jsonpath.NewPath("$.info.version")
	if err != nil {
		return "", err
	}
	descriptionPath, err := jsonpath.NewPath("$.info.version")
	if err != nil {
		return "", err
	}
	toString := func(node []*yaml.Node) string {
		if len(node) == 0 {
			return ""
		}
		return node[0].Value
	}

	return `{
    "title": "` + toString(titlePath.Query(&orig)) + `",
    "version": "` + toString(versionPath.Query(&orig)) + `",
    "description": "` + toString(descriptionPath.Query(&orig)) + `"
}`, nil
}

type ApplyOverlaySuccess struct {
	Type   string `json:"type"`
	Result string `json:"result"`
}

func ApplyOverlay(originalYAML, overlayYAML string) (string, error) {
	var orig yaml.Node
	err := yaml.Unmarshal([]byte(originalYAML), &orig)
	if err != nil {
		return "", fmt.Errorf("failed to parse original schema: %w", err)
	}

	var overlay overlay.Overlay
	err = yaml.Unmarshal([]byte(overlayYAML), &overlay)
	if err != nil {
		return "", fmt.Errorf("failed to parse overlay schema in ApplyOverlay: %w", err)
	}
	err = overlay.Validate()
	if err != nil {
		return "", fmt.Errorf("failed to validate overlay schema in ApplyOverlay: %w", err)
	}
	hasFilterExpression := false
	// check to see if we have an overlay with an error, or a partial overlay: i.e. any overlay actions are missing an update or remove
	for i, action := range overlay.Actions {
		tokenized := token.NewTokenizer(action.Target, config.WithPropertyNameExtension()).Tokenize()
		for _, tok := range tokenized {
			if tok.Token == token.FILTER {
				hasFilterExpression = true
				break
			}
		}
		parsed, pathErr := jsonpath.NewPath(action.Target, config.WithPropertyNameExtension())

		var node *yaml.Node
		if pathErr != nil {
			node, err = lookupOverlayActionTargetNode(overlayYAML, i)
			if err != nil {
				return "", err
			}

			return applyOverlayJSONPathError(pathErr, node)
		}
		if reflect.ValueOf(action.Update).IsZero() && action.Remove == false {
			result := parsed.Query(&orig)

			node, err = lookupOverlayActionTargetNode(overlayYAML, i)
			if err != nil {
				return "", err
			}

			return applyOverlayJSONPathIncomplete(result, node)
		}
	}
	if hasFilterExpression && overlay.JSONPathVersion != "rfc9535" {
		return "", fmt.Errorf("invalid overlay schema: must have `x-speakeasy-jsonpath: rfc9535`")
	}

	err = overlay.ApplyTo(&orig)
	if err != nil {
		return "", fmt.Errorf("failed to apply overlay: %w", err)
	}

	// Unwrap the document node if it exists and has only one content node
	if orig.Kind == yaml.DocumentNode && len(orig.Content) == 1 {
		orig = *orig.Content[0]
	}

	out, err := yaml.Marshal(&orig)
	if err != nil {
		return "", fmt.Errorf("failed to marshal result: %w", err)
	}

	out, err = json.Marshal(ApplyOverlaySuccess{
		Type:   "success",
		Result: string(out),
	})

	return string(out), err
}

type IncompleteOverlayErrorMessage struct {
	Type   string `json:"type"`
	Line   int    `json:"line"`
	Col    int    `json:"col"`
	Result string `json:"result"`
}

func applyOverlayJSONPathIncomplete(result []*yaml.Node, node *yaml.Node) (string, error) {
	yamlResult, err := yaml.Marshal(&result)
	if err != nil {
		return "", err
	}

	out, err := json.Marshal(IncompleteOverlayErrorMessage{
		Type:   "incomplete",
		Line:   node.Line,
		Col:    node.Column,
		Result: string(yamlResult),
	})

	return string(out), err
}

type JSONPathErrorMessage struct {
	Type       string `json:"type"`
	Line       int    `json:"line"`
	Col        int    `json:"col"`
	ErrMessage string `json:"error"`
}

func applyOverlayJSONPathError(err error, node *yaml.Node) (string, error) {
	// first lets see if we can find a target expression
	out, err := json.Marshal(JSONPathErrorMessage{
		Type:       "error",
		Line:       node.Line,
		Col:        node.Column,
		ErrMessage: err.Error(),
	})
	return string(out), err
}

func lookupOverlayActionTargetNode(overlayYAML string, i int) (*yaml.Node, error) {
	var node struct {
		Actions []struct {
			Target yaml.Node `yaml:"target"`
		} `yaml:"actions"`
	}
	err := yaml.Unmarshal([]byte(overlayYAML), &node)
	if err != nil {
		return nil, fmt.Errorf("failed to parse overlay schema in lookupOverlayActionTargetNode: %w", err)
	}
	if len(node.Actions) <= i {
		return nil, fmt.Errorf("no action at index %d", i)
	}
	if reflect.ValueOf(node.Actions[i].Target).IsZero() {
		return nil, fmt.Errorf("no target at index %d", i)
	}
	return &node.Actions[i].Target, nil
}

func Query(currentYAML, path string) (string, error) {
	var orig yaml.Node
	err := yaml.Unmarshal([]byte(currentYAML), &orig)
	if err != nil {
		return "", fmt.Errorf("failed to parse original schema in Query: %w", err)
	}
	parsed, err := jsonpath.NewPath(path, config.WithPropertyNameExtension())
	if err != nil {
		return "", err
	}
	result := parsed.Query(&orig)
	// Marshal it back out
	out, err := yaml.Marshal(result)
	if err != nil {
		return "", err
	}
	return string(out), nil
}
