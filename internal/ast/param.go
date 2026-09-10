package ast

type Param struct {
	Field           *FieldDef `yaml:",omitempty"` // The field that the parameter is associated with
	Examples        Examples  `yaml:",omitempty"` // The named examples available for the parameter
	Hidden          bool      `yaml:",omitempty"` // Whether the parameter is hidden as it is only accepted as a global parameter
	AllowEmptyValue bool      `yaml:",omitempty"` // Whether to send the parameter with an empty value in the query string (e.g., "?param=")
}

// HasMatchConfigUsePriorState returns true if the parameter's field type has
// the MatchConfig.UsePriorState flag set.
func (p *Param) HasMatchConfigUsePriorState() bool {
	return p != nil && p.Field != nil && p.Field.Type != nil &&
		p.Field.Type.Extensions != nil && p.Field.Type.Extensions.MatchConfig != nil &&
		p.Field.Type.Extensions.MatchConfig.UsePriorState
}

func (p *Param) Match(matchers Matchers) error {
	if matchers.Param != nil {
		return matchers.Param(p)
	}

	return nil
}
