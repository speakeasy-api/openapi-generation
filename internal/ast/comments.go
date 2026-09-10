package ast

// ExternalDocs represents a link to external documentation
type ExternalDocs struct {
	Description string `yaml:",omitempty"`
	URL         string `yaml:",omitempty"`
}

// Clone creates a deep copy of the ExternalDocs
func (e *ExternalDocs) Clone() *ExternalDocs {
	if e == nil {
		return nil
	}

	return &ExternalDocs{
		Description: e.Description,
		URL:         e.URL,
	}
}

func (e *ExternalDocs) Match(matchers Matchers) error {
	if matchers.ExternalDocs != nil {
		return matchers.ExternalDocs(e)
	}

	return nil
}

// Merges the given ExternalDocs into this ExternalDocs. The algorithm adds data
// from the given ExternalDocs to this ExternalDocs where it is undefined. Where
// there is conflicting data, this ExternalDocs's data is preserved or otherwise
// delegated to data-specific merge functionality. It does not remove any data
// from this ExternalDocs.
func (e *ExternalDocs) Merge(other *ExternalDocs) {
	if e == nil || other == nil {
		return
	}

	if e.Description == "" {
		e.Description = other.Description
	}

	if e.URL == "" {
		e.URL = other.URL
	}
}

// ExtendedComment allows for the definition of special comments using the
// x-speakeasy-docs extension.
type ExtendedComment struct {
	Summary     string `yaml:",omitempty"`
	Description string `yaml:",omitempty"`
}

// Clone creates a deep copy of the ExtendedComment
func (e *ExtendedComment) Clone() *ExtendedComment {
	if e == nil {
		return nil
	}

	return &ExtendedComment{
		Description: e.Description,
		Summary:     e.Summary,
	}
}

func (e *ExtendedComment) Match(matchers Matchers) error {
	if matchers.ExtendedComment != nil {
		return matchers.ExtendedComment(e)
	}

	return nil
}

// Merges the given ExtendedComment into this ExtendedComment. The algorithm
// adds data from the given ExtendedComment to this ExtendedComment where it is
// undefined. Where there is conflicting data, this ExtendedComment's data is
// preserved or otherwise delegated to data-specific merge functionality. It
// does not remove any data from this ExtendedComment.
func (e *ExtendedComment) Merge(other *ExtendedComment) {
	if e == nil || other == nil {
		return
	}

	if e.Description == "" {
		e.Description = other.Description
	}

	if e.Summary == "" {
		e.Summary = other.Summary
	}
}

// Comment represents a comment that is associated with a sdk, type or operation
type Comment struct {
	Summary                string                      `yaml:",omitempty"`
	Description            string                      `yaml:",omitempty"`
	ExternalDocs           *ExternalDocs               `yaml:",omitempty"`
	ExtendedComments       map[string]*ExtendedComment `yaml:",omitempty"`
	Deprecated             bool                        `yaml:",omitempty"`
	DeprecationReplacement string                      `yaml:",omitempty"`
	DeprecationMessage     string                      `yaml:",omitempty"`
}

// Clone creates a deep copy of the Comment
func (c *Comment) Clone() *Comment {
	if c == nil {
		return nil
	}

	var extendedComments map[string]*ExtendedComment

	if c.ExtendedComments != nil {
		extendedComments = make(map[string]*ExtendedComment, len(c.ExtendedComments))
		for k, v := range c.ExtendedComments {
			extendedComments[k] = v.Clone()
		}
	}

	return &Comment{
		Deprecated:             c.Deprecated,
		DeprecationMessage:     c.DeprecationMessage,
		DeprecationReplacement: c.DeprecationReplacement,
		Description:            c.Description,
		ExtendedComments:       extendedComments,
		ExternalDocs:           c.ExternalDocs.Clone(),
		Summary:                c.Summary,
	}
}

func (c *Comment) Match(matchers Matchers) error {
	if matchers.Comment != nil {
		return matchers.Comment(c)
	}

	return nil
}

// Merges the given Comment into this Comment. The algorithm adds data from the
// given Comment to this Comment where it is undefined. Where there is
// conflicting data, this Comment's data is preserved or otherwise delegated to
// data-specific merge functionality. It does not remove any data from this
// Comment.
func (c *Comment) Merge(other *Comment) {
	if c == nil || other == nil {
		return
	}

	if !c.Deprecated {
		c.Deprecated = other.Deprecated
	}

	if c.DeprecationMessage == "" {
		c.DeprecationMessage = other.DeprecationMessage
	}

	if c.DeprecationReplacement == "" {
		c.DeprecationReplacement = other.DeprecationReplacement
	}

	if c.Description == "" {
		c.Description = other.Description
	}

	if c.Summary == "" {
		c.Summary = other.Summary
	}

	if c.ExternalDocs == nil {
		c.ExternalDocs = other.ExternalDocs
	} else if other.ExternalDocs != nil {
		c.ExternalDocs.Merge(other.ExternalDocs)
	}

	if c.ExtendedComments == nil && other.ExtendedComments != nil {
		c.ExtendedComments = make(map[string]*ExtendedComment, len(other.ExtendedComments))
	}

	for otherKey, otherExtendedComment := range other.ExtendedComments {
		cExtendedComment, ok := c.ExtendedComments[otherKey]

		if !ok {
			c.ExtendedComments[otherKey] = otherExtendedComment
		} else {
			cExtendedComment.Merge(otherExtendedComment)
		}
	}
}

// IsEmpty returns true if the comment has no meaningful content
func (c *Comment) IsEmpty() bool {
	if c == nil {
		return true
	}
	return c.Summary == "" &&
		c.Description == "" &&
		c.ExternalDocs == nil &&
		len(c.ExtendedComments) == 0 &&
		!c.Deprecated &&
		c.DeprecationReplacement == "" &&
		c.DeprecationMessage == ""
}
