package arazzo

import (
	"fmt"

	"github.com/speakeasy-api/openapi/sequencedmap"
	"gopkg.in/yaml.v3"
)

type test struct {
	Name            string
	Description     string
	Targets         []string
	Server          string
	Security        yaml.Node
	Parameters      *parameters
	RequestBody     *sequencedmap.Map[string, yaml.Node]
	Responses       *sequencedmap.Map[string, yaml.Node]
	InternalID      string
	TestGroups      []string
	InternalEnvVars *sequencedmap.Map[string, string]
}

type parameters struct {
	Path   *sequencedmap.Map[string, yaml.Node]
	Query  *sequencedmap.Map[string, yaml.Node]
	Header *sequencedmap.Map[string, yaml.Node]
}

func (t test) GetResponse(statusCode string) (*sequencedmap.Map[string, yaml.Node], bool, error) {
	if t.Responses == nil {
		return nil, false, nil
	}
	resNode, ok := t.Responses.Get(statusCode)
	if !ok {
		return nil, false, nil
	}

	var res *sequencedmap.Map[string, yaml.Node]
	if err := resNode.Decode(&res); err == nil {
		return res, false, nil
	}

	var assertStatusCode bool
	if err := resNode.Decode(&assertStatusCode); err != nil {
		return nil, false, fmt.Errorf("failed to decode response body: %w", err)
	}

	return nil, assertStatusCode, nil
}
