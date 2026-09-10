package fastAST

import (
	"encoding/json"
	"fmt"
	"strings"
)

// GroupNode represents a node in the group tree
type GroupNode struct {
	Name       string                `json:"name"`
	Operations []*FastOperation      `json:"operations,omitempty"`
	Children   map[string]*GroupNode `json:"children,omitempty"`
	Parent     *GroupNode            `json:"-"`
}

// GroupTree represents a tree structure for organizing operations into groups
type GroupTree struct {
	root *GroupNode
}

// MarshalJSON implements the json.Marshaler interface
func (gt *GroupTree) MarshalJSON() ([]byte, error) {
	if gt.root == nil {
		return json.Marshal(nil)
	}
	return json.Marshal(gt.createSerializableTree(gt.root))
}

// NewGroupTree creates a new GroupTree instance
func NewGroupTree() *GroupTree {
	return &GroupTree{
		root: &GroupNode{
			Name:       "root",
			Operations: []*FastOperation{},
			Children:   make(map[string]*GroupNode),
		},
	}
}

// AddOperation adds an operation to the specified group path
func (gt *GroupTree) AddOperation(groupPath string, operation *FastOperation) {
	// If no group path is specified, add to root
	if groupPath == "" {
		gt.root.Operations = append(gt.root.Operations, operation)
		return
	}

	parts := splitGroupPath(groupPath)
	currentNode := gt.root

	// Navigate or create the path
	for _, part := range parts {
		if _, exists := currentNode.Children[part]; !exists {
			currentNode.Children[part] = &GroupNode{
				Name:       part,
				Operations: []*FastOperation{},
				Children:   make(map[string]*GroupNode),
				Parent:     currentNode,
			}
		}
		currentNode = currentNode.Children[part]
	}

	// Add operation to the final node
	currentNode.Operations = append(currentNode.Operations, operation)
}

// GetGroup returns the group node at the specified path
func (gt *GroupTree) GetGroup(groupPath string) *GroupNode {
	parts := splitGroupPath(groupPath)
	currentNode := gt.root

	for _, part := range parts {
		if _, exists := currentNode.Children[part]; !exists {
			return nil
		}
		currentNode = currentNode.Children[part]
	}

	return currentNode
}

// GetAllOperations returns all operations in a group and its subgroups
func (gt *GroupTree) GetAllOperations(groupPath string) []*FastOperation {
	group := gt.GetGroup(groupPath)
	if group == nil {
		return []*FastOperation{}
	}

	operations := make([]*FastOperation, 0, len(group.Operations))
	operations = append(operations, group.Operations...)

	// Recursively get operations from children
	for _, child := range group.Children {
		childPath := groupPath
		if childPath != "" {
			childPath += "."
		}
		childPath += child.Name
		operations = append(operations, gt.GetAllOperations(childPath)...)
	}

	return operations
}

// ToJSON serializes the tree to a JSON string
func (gt *GroupTree) ToJSON() (string, error) {
	serializableTree := gt.createSerializableTree(gt.root)
	data, err := json.Marshal(serializableTree)
	if err != nil {
		return "", fmt.Errorf("failed to marshal tree: %w", err)
	}
	return string(data), nil
}

// FromJSON creates a GroupTree from a JSON string
func FromJSON(jsonString string) (*GroupTree, error) {
	tree := NewGroupTree()
	var data map[string]interface{}

	if err := json.Unmarshal([]byte(jsonString), &data); err != nil {
		return nil, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	if err := tree.reconstructTree(tree.root, data); err != nil {
		return nil, fmt.Errorf("failed to reconstruct tree: %w", err)
	}

	return tree, nil
}

// createSerializableTree creates a serializable version of the tree
func (gt *GroupTree) createSerializableTree(node *GroupNode) map[string]interface{} {
	serializableNode := map[string]interface{}{
		"name": node.Name,
	}

	// Only add operations if there are any
	if len(node.Operations) > 0 {
		// Create a simplified version of operations with only ID and jsonPath
		simplifiedOps := make([]map[string]string, len(node.Operations))
		for i, op := range node.Operations {
			simplifiedOps[i] = map[string]string{
				"operationID": op.OperationID,
				"jsonPath":    op.JSONPath,
			}
		}
		serializableNode["operations"] = simplifiedOps
	}

	// Only add children if there are any
	if len(node.Children) > 0 {
		children := make(map[string]interface{})
		// Recursively process children
		for key, child := range node.Children {
			children[key] = gt.createSerializableTree(child)
		}
		serializableNode["children"] = children
	}

	return serializableNode
}

// reconstructTree reconstructs the tree from serialized data
func (gt *GroupTree) reconstructTree(currentNode *GroupNode, data map[string]interface{}) error {
	currentNode.Name = data["name"].(string)

	// Reconstruct operations
	if ops, ok := data["operations"].([]interface{}); ok {
		currentNode.Operations = make([]*FastOperation, len(ops))
		for i, op := range ops {
			opMap, ok := op.(map[string]interface{})
			if !ok {
				return fmt.Errorf("invalid operation data at index %d", i)
			}
			opJSON, err := json.Marshal(opMap)
			if err != nil {
				return fmt.Errorf("failed to marshal operation: %w", err)
			}
			var sandboxOp FastOperation
			if err := json.Unmarshal(opJSON, &sandboxOp); err != nil {
				return fmt.Errorf("failed to unmarshal operation: %w", err)
			}
			currentNode.Operations[i] = &sandboxOp
		}
	}

	// Reconstruct children
	if children, ok := data["children"].(map[string]interface{}); ok {
		for key, childData := range children {
			childNode := &GroupNode{
				Name:       key,
				Operations: []*FastOperation{},
				Children:   make(map[string]*GroupNode),
				Parent:     currentNode,
			}
			currentNode.Children[key] = childNode
			if err := gt.reconstructTree(childNode, childData.(map[string]interface{})); err != nil {
				return err
			}
		}
	}

	return nil
}

// splitGroupPath splits a group path into its components
func splitGroupPath(path string) []string {
	if path == "" {
		return []string{}
	}
	return strings.Split(path, ".")
}
