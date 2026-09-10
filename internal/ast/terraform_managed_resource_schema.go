package ast

import (
	"fmt"

	"github.com/speakeasy-api/openapi-generation/v2/internal/terraform"
)

// Describes a Terraform managed resource schema.
type TerraformManagedResourceSchema struct {
	// Schema version for the managed resource. This relates to Terraform's
	// managed resource state upgrade mechanism. By default and in most cases,
	// this is never set and left as the default value of 0. Sourced from
	// x-speakeasy-entity-version configuration, if available.
	Version int64 `json:"version" yaml:"version"`
}

// Creates a new Terraform managed resource schema, safely initializing
// underlying fields.
func NewTerraformManagedResourceSchema() *TerraformManagedResourceSchema {
	return &TerraformManagedResourceSchema{}
}

// Sets the version.
func (s *TerraformManagedResourceSchema) SetVersion(version int64) error {
	if !terraform.IsResourceSchemaVersionValid(version) {
		return fmt.Errorf("invalid schema version: %d", version)
	}

	// Highest value wins.
	if s.Version > version {
		return nil
	}

	s.Version = version

	return nil
}
