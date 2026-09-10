package changes

import (
	"encoding/json"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/subsystem"
)

// SDKDiff represents the differences between two SDK configurations
type SDKDiff struct {
	Changes      []MethodDiff         `json:"changes"`
	OldAST       *ast.AST             `json:"-"` // Don't serialize ASTs
	NewAST       *ast.AST             `json:"-"` // Don't serialize ASTs
	OldSubsystem *subsystem.Subsystem `json:"-"`
	NewSubsystem *subsystem.Subsystem `json:"-"`
}

// ToJSON converts the SDKDiff to JSON format
func (d *SDKDiff) ToJSON() ([]byte, error) {
	return json.Marshal(d)
}

// Filter returns a new SDKDiff containing only changes for which the predicate
// returns true. AST and subsystem references are shared with the original.
func (d SDKDiff) Filter(predicate func(MethodDiff) bool) SDKDiff {
	var filtered []MethodDiff
	for _, change := range d.Changes {
		if predicate(change) {
			filtered = append(filtered, change)
		}
	}
	return SDKDiff{
		Changes:      filtered,
		OldAST:       d.OldAST,
		NewAST:       d.NewAST,
		OldSubsystem: d.OldSubsystem,
		NewSubsystem: d.NewSubsystem,
	}
}
