package ast

// Describes Terraform per-operation server configuration. This is defined when
// the underlying operations have path or operation server URLs defined. Only
// a single server configuration is supported, even if multiple operations have
// differing server URLs defined. This is intentional to simplify the generated
// Terraform code and avoid complexity around per-operation server URL
// configuration for Terraform consumers.
type TerraformServer struct {
	// Name of the configurable attribute in the schema.
	AttributeName string `json:"attributeName" yaml:"attributeName"`

	// Server description.
	Description string `json:"description" yaml:"description"`

	// Server URL.
	URL string `json:"url" yaml:"url"`

	// TODO: In the future, may need to add support for server URL variables.
	// Variables []*ServerVariable `json:"variables" yaml:"variables"`
}
