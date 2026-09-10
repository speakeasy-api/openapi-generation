package cms

import (
	"sync"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
)

type CommentSource string

const (
	CSUser    CommentSource = ""
	CSSuggest CommentSource = "suggest"
	CSOpenAPI CommentSource = "openapi"
	CSBuiltIn CommentSource = "builtin"
)

type (
	SourceComments struct {
		Comment `yaml:",inline"`

		OpenAPI *AutoComment `yaml:"openapi,omitempty"`
		Builtin *AutoComment `yaml:"builtin,omitempty"`
		Suggest *AutoComment `yaml:"suggest,omitempty"`
	}
)

type Comment struct {
	Summary     string `yaml:"summary"`
	Description string `yaml:"description"`
}

var _ SettableComment = &Comment{}

func (c *Comment) SetSummary(s string) {
	c.Summary = s
}

func (c *Comment) SetDescription(s string) {
	c.Description = s
}

type AutoComment struct {
	Summary     string `yaml:"summary,omitempty"`
	Description string `yaml:"description,omitempty"`
}

var _ SettableComment = &AutoComment{}

func (c *AutoComment) SetSummary(s string) {
	c.Summary = s
}

func (c *AutoComment) SetDescription(s string) {
	c.Description = s
}

func (c *AutoComment) ToComment() *Comment {
	if c == nil {
		return nil
	}

	return &Comment{
		Summary:     c.Summary,
		Description: c.Description,
	}
}

type SettableComment interface {
	SetSummary(string)
	SetDescription(string)
}

type CommentTracker struct {
	mu       sync.Mutex
	comments map[string]*SourceComments // keyed by "namespace|name"
	tags     []string
}

func New(tags ...string) *CommentTracker {
	return &CommentTracker{
		comments: map[string]*SourceComments{},
		tags:     tags,
	}
}

func (c *CommentTracker) SetTags(tags ...string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.tags = tags
}

// RegisterComment stores a comment under (namespace, name) for the given
// source. Empty comments are ignored; use ResolveComment when an empty
// registration should claim the slot.
func (c *CommentTracker) RegisterComment(source, namespace, name string, comment *ast.Comment) string {
	fullName, _, ok := c.upsert(source, namespace, name, comment, false)
	if !ok {
		return ""
	}
	return fullName
}

// ResolveComment atomically registers a comment under (namespace, kind, name)
// for the given source and returns the merged comment for that key. An empty
// (or nil) comment claims the source slot, so the result reflects this
// registration — not another entity's — regardless of render-job order.
func (c *CommentTracker) ResolveComment(source, namespace, kind, name string, comment *ast.Comment) Comment {
	if kind != "" && name != "" {
		name = kind + ":" + name
	}
	_, merged, _ := c.upsert(source, namespace, name, comment, true)
	return merged
}

func (c *CommentTracker) GetComment(fullName string) Comment {
	c.mu.Lock()
	defer c.mu.Unlock()

	sc, ok := c.comments[fullName]
	if !ok {
		return Comment{}
	}

	return sc.merged()
}

// upsert writes the comment's effective content to the source slot under one
// lock and returns the merged result. When claimEmpty is false, empty content
// leaves the slot untouched (legacy RegisterComment behavior). ok is false
// for an empty name or unknown source.
func (c *CommentTracker) upsert(source, namespace, name string, comment *ast.Comment, claimEmpty bool) (string, Comment, bool) {
	if name == "" {
		return "", Comment{}, false
	}

	if namespace == "" {
		namespace = "__GLOBAL__"
	}
	fullName := namespace + "|" + name

	c.mu.Lock()
	defer c.mu.Unlock()

	if c.comments == nil {
		c.comments = map[string]*SourceComments{}
	}

	sc, ok := c.comments[fullName]
	if !ok {
		sc = &SourceComments{}
		c.comments[fullName] = sc
	}

	summary, description := c.effectiveComment(comment)
	if summary != "" || description != "" || claimEmpty {
		cc := sc.slot(CommentSource(source))
		if cc == nil {
			return "", sc.merged(), false
		}
		cc.SetSummary(summary)
		cc.SetDescription(description)
	}

	return fullName, sc.merged(), true
}

// effectiveComment extracts the summary/description from a comment, honoring
// tag-targeted extended comments. Callers must hold c.mu.
func (c *CommentTracker) effectiveComment(comment *ast.Comment) (string, string) {
	if comment == nil {
		return "", ""
	}

	for _, tag := range c.tags {
		if extComment, hasExtComment := comment.ExtendedComments[tag]; hasExtComment {
			return extComment.Summary, extComment.Description
		}
	}

	return comment.Summary, comment.Description
}

func (sc *SourceComments) slot(source CommentSource) SettableComment {
	switch source {
	case CSUser:
		return &sc.Comment
	case CSOpenAPI:
		if sc.OpenAPI == nil {
			sc.OpenAPI = &AutoComment{}
		}
		return sc.OpenAPI
	case CSBuiltIn:
		if sc.Builtin == nil {
			sc.Builtin = &AutoComment{}
		}
		return sc.Builtin
	case CSSuggest:
		if sc.Suggest == nil {
			sc.Suggest = &AutoComment{}
		}
		return sc.Suggest
	default:
		return nil
	}
}

// merged returns the highest-precedence non-empty comment: user, then
// suggest, openapi, builtin.
func (sc *SourceComments) merged() Comment {
	for _, cc := range []*Comment{&sc.Comment, sc.Suggest.ToComment(), sc.OpenAPI.ToComment(), sc.Builtin.ToComment()} {
		if cc != nil && (cc.Summary != "" || cc.Description != "") {
			return *cc
		}
	}

	return Comment{}
}
