package extensions

import (
	"fmt"
)

// EntityMissingCodes describes HTTP status codes that indicate an entity is
// missing/deleted in the API. Used by Terraform target to call RemoveResource()
// during Read operations.
type EntityMissingCodes []int

// HandleEntityMissingCodesExtension handles parsing of the
// x-speakeasy-entity-missing-codes extension from the given OpenAPI extensions map.
func (e *Extensions) HandleEntityMissingCodesExtension(extensions OAExtensions) (EntityMissingCodes, error) {
	if extensions.Len() == 0 {
		return nil, nil
	}

	yamlNode, ok := e.findExtension(extensions, ExtEntityMissingCodes)
	if !ok {
		return nil, nil
	}

	var codes []int
	if err := yamlNode.Decode(&codes); err != nil {
		return nil, fmt.Errorf("failed to parse %s: %w", ExtEntityMissingCodes.Name(), err)
	}

	return codes, nil
}
