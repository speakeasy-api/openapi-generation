package validation

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"github.com/speakeasy-api/openapi-generation/v2/internal/extensions"
	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
)

type DuplicateOperationName struct {
	sdkConfig *config.Configuration
	target    types.Target
}

var _ Rule = (*DuplicateOperationName)(nil)
var _ configAwareRule = (*DuplicateOperationName)(nil)
var _ targetAwareRule = (*DuplicateOperationName)(nil)

func (r *DuplicateOperationName) SetConfig(cfg *config.Configuration) {
	r.sdkConfig = cfg
}

func (r *DuplicateOperationName) SetTarget(target types.Target) {
	r.target = target
}

func (r *DuplicateOperationName) ID() string {
	return "generator-duplicate-operation-name"
}

func (r *DuplicateOperationName) Category() string {
	return "validation"
}

func (r *DuplicateOperationName) Summary() string {
	return "Ensure no duplicate operation names collide after naming rules."
}

func (r *DuplicateOperationName) HowToFix() string {
	return "Ensure each operation resolves to a unique method name by adjusting operationId, name overrides, tags, or groups to avoid collisions after naming rules."
}

func (r *DuplicateOperationName) Description() string {
	return "Duplicate operation names can cause SDK method name collisions after naming rules are applied (group + method name). The method name comes from operationId (or x-speakeasy-name-override) and the group comes from x-speakeasy-group or tags, so duplicates here will collide."
}

func (r *DuplicateOperationName) Link() string {
	return ""
}

func (r *DuplicateOperationName) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *DuplicateOperationName) Versions() []string {
	return nil // Applies to all versions
}

func (r *DuplicateOperationName) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Index == nil {
		return nil
	}

	var validationErrors []error

	// Get target from SDK config
	target := r.getTarget()

	// Initialize extensions based on the target context
	exts := extensions.New(target)

	doc := docInfo.Document
	if doc == nil {
		return nil
	}

	// Handle extension rewrites (errors intentionally ignored)
	_ = exts.HandleRewriteExtension(extensions.WithDocumentExtensions(doc.GetExtensions()))

	// Parse global name overrides
	globalNameOverrides, _ := exts.HandleGlobalNameOverrideExtensions(doc)

	// Initialize checkers
	// Note: operationIDChecker is not needed - the built-in validation-operation-id-unique rule
	// already handles exact duplicate operationIds. We only check for sanitized name collisions.
	methodNameChecker := newMethodNameConflictChecker(target)
	filenameChecker := newOperationFilenameConflictChecker()

	// Iterate through operations from index
	for _, opIndexNode := range docInfo.Index.Operations {
		// Skip webhooks and callbacks - they are not SDK methods
		location := opIndexNode.Location.ToJSONPointer().String()
		if strings.HasPrefix(location, "/webhooks/") || strings.Contains(location, "/callbacks/") {
			continue
		}

		operation := opIndexNode.Node
		if operation == nil {
			continue
		}

		opCore := operation.GetCore()
		if opCore == nil {
			continue
		}

		// Check if operation should be ignored
		if ignored, _ := exts.HandleIgnoreExtension(operation.GetExtensions()); ignored != nil && *ignored {
			continue
		}

		// Build OperationInfo
		operationInfo := r.buildOperationNameInfo(
			opIndexNode,
			operation,
			exts,
			globalNameOverrides,
		)

		// Check for method name conflicts (sanitized name collisions)
		conflicts := methodNameChecker.Check(operationInfo, r)
		if len(conflicts) > 0 {
			for _, conflict := range conflicts {
				validationErrors = append(validationErrors, conflict)
			}
			continue
		}

		// Check for filename conflicts
		if conflict := filenameChecker.Check(operationInfo, r); conflict != nil {
			validationErrors = append(validationErrors, conflict)
			continue
		}
	}

	return validationErrors
}

// getTarget returns the target for SDK generation
func (r *DuplicateOperationName) getTarget() types.Target {
	if r.target.Template != "" {
		return r.target
	}
	// Default to "go" if no target was set
	// This should rarely happen - target should be set via SetTarget()
	return types.NewTargetFromTemplate("go")
}

// buildOperationNameInfo extracts operation naming information from an operation
func (r *DuplicateOperationName) buildOperationNameInfo(
	opIndexNode *openapi.IndexNode[*openapi.Operation],
	operation *openapi.Operation,
	exts *extensions.Extensions,
	globalNameOverrides []*extensions.NameOverride,
) operationNameInfo {
	var info operationNameInfo

	// Get path and method from index node location
	info.HTTPMethod, info.HTTPPath = openapi.ExtractMethodAndPath(opIndexNode.Location)

	opCore := operation.GetCore()
	opRootNode := operation.GetRootNode()

	// Initialize OperationNode
	info.OperationNode = opRootNode

	// operationId
	operationID := operation.GetOperationID()
	if operationID != "" {
		info.OperationID = operationID
		info.HasOperationID = true
		// Get the operationId node for precise location
		if opCore != nil {
			info.OperationIDNode = opCore.OperationID.GetKeyNodeOrRoot(opRootNode)
		} else {
			info.OperationIDNode = opRootNode
		}
	} else {
		info.OperationID = fmt.Sprintf("%s_%s", info.HTTPMethod, info.HTTPPath)
		info.OperationIDNode = opRootNode
	}

	// methodNameOverride from global
	for _, globalNameOverride := range globalNameOverrides {
		if globalNameOverride.OperationId != "" {
			re := regexp.MustCompile(globalNameOverride.OperationId)
			if re.MatchString(info.OperationID) {
				info.MethodNameOverride = globalNameOverride.GlobalMethodNameOverride
				info.MethodNameOverrideNode = opRootNode
			}
		}
	}

	// methodNameOverride from operation extension
	if nameOverride, hasOverride, _ := exts.HandleOperationMethodNameExtension(operation); hasOverride && nameOverride != nil {
		info.MethodNameOverride = nameOverride.Name
		info.MethodNameOverrideNode = nameOverride.Node
	}

	// methodGroupOverride
	if groups, groupNode, _ := exts.GetGroupsWithNode(operation); len(groups) > 0 {
		info.MethodGroupOverride = groups[0]
		info.MethodGroupOverrideNode = groupNode
	}

	// tags
	tags := operation.GetTags()
	info.Tags = tags
	if len(tags) > 0 && opCore != nil {
		// Get the first tag node for error reporting
		info.TagsNode = opCore.Tags.GetSliceValueNodeOrRoot(0, opRootNode)
	}

	return info
}
