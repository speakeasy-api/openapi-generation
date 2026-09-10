package cms

import (
	"fmt"
	"sync"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/stretchr/testify/assert"
)

// Reproduces a nondeterministic doc-comment flip: an SDK class `Search` (no
// comments) and the method `Search` (with operation docs) shared the CMS key
// `<ns>|Search`. Whether the class rendered the method's docs depended on
// which render job ran first.
func TestResolveComment_KindScopesKeys(t *testing.T) {
	c := New("sdk")

	methodComment := c.ResolveComment("openapi", "v4", "method", "Search", &ast.Comment{
		Summary:     "search.v1",
		Description: "Elasticsearch.v1 query engine",
	})
	assert.Equal(t, "search.v1", methodComment.Summary)

	typeComment := c.ResolveComment("openapi", "v4", "type", "Search", nil)
	assert.Empty(t, typeComment.Summary, "type Search must not inherit the method's summary")
	assert.Empty(t, typeComment.Description, "type Search must not inherit the method's description")
}

// Even when two entities resolve the same key, each call must return a merge
// of its own input (plus user overrides) — never another entity's earlier
// registration. An empty registration claims the source slot.
func TestResolveComment_ReadsOwnWriteOnSharedKey(t *testing.T) {
	c := New("sdk")

	first := c.ResolveComment("openapi", "v4", "", "Search", &ast.Comment{Summary: "from method"})
	assert.Equal(t, "from method", first.Summary)

	second := c.ResolveComment("openapi", "v4", "", "Search", nil)
	assert.Empty(t, second.Summary, "empty registration must claim the slot, not read stale content")

	third := c.ResolveComment("openapi", "v4", "", "Search", &ast.Comment{Summary: "from method"})
	assert.Equal(t, "from method", third.Summary)
}

// Claiming the openapi slot must not clear user-authored comments; user
// content still takes precedence in the merged result.
func TestResolveComment_PreservesUserOverride(t *testing.T) {
	c := New("sdk")

	c.RegisterComment("", "v4", "Search", &ast.Comment{Summary: "user edit"})

	merged := c.ResolveComment("openapi", "v4", "", "Search", nil)
	assert.Equal(t, "user edit", merged.Summary)

	merged = c.ResolveComment("openapi", "v4", "", "Search", &ast.Comment{Summary: "from openapi"})
	assert.Equal(t, "user edit", merged.Summary, "user comments outrank openapi comments")
}

// Tag-targeted extended comments must behave the same as RegisterComment.
func TestResolveComment_AppliesExtendedCommentTags(t *testing.T) {
	c := New("sdk")

	merged := c.ResolveComment("openapi", "v4", "method", "Search", &ast.Comment{
		Summary: "generic",
		ExtendedComments: map[string]*ast.ExtendedComment{
			"sdk": {Summary: "sdk specific"},
		},
	})
	assert.Equal(t, "sdk specific", merged.Summary)
}

// Legacy register-then-read pair must keep working for custom templates.
func TestRegisterComment_GetCommentRoundTrip(t *testing.T) {
	c := New("sdk")

	fullName := c.RegisterComment("openapi", "v4", "Search", &ast.Comment{Summary: "s"})
	assert.Equal(t, "s", c.GetComment(fullName).Summary)

	// empty registration is a no-op for the legacy API
	c.RegisterComment("openapi", "v4", "Search", nil)
	assert.Equal(t, "s", c.GetComment(fullName).Summary)
}

func TestCommentTracker_ZeroValue(t *testing.T) {
	t.Run("get before write", func(t *testing.T) {
		var c CommentTracker
		assert.Equal(t, Comment{}, c.GetComment("v4|Missing"))
	})

	t.Run("first write via RegisterComment", func(t *testing.T) {
		var c CommentTracker
		fullName := c.RegisterComment("openapi", "v4", "Search", &ast.Comment{Summary: "registered"})

		assert.Equal(t, "v4|Search", fullName)
		assert.Equal(t, "registered", c.GetComment(fullName).Summary)
	})

	t.Run("first write via ResolveComment", func(t *testing.T) {
		var c CommentTracker
		resolved := c.ResolveComment("openapi", "v4", "var", "ServerList", &ast.Comment{Summary: "resolved"})

		assert.Equal(t, "resolved", resolved.Summary)
		assert.Equal(t, "resolved", c.GetComment("v4|var:ServerList").Summary)
	})
}

// Render jobs run on parallel workers sharing one tracker. Each concurrent
// resolve must observe its own registration atomically: every goroutine
// writes a distinct value while groups of 25 hammer the same key, so a
// non-atomic register-then-read would observe another goroutine's summary.
func TestResolveComment_ConcurrentResolvesAreAtomic(t *testing.T) {
	c := New("sdk")

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Go(func() {
			want := fmt.Sprintf("summary-%d", i)
			merged := c.ResolveComment("openapi", "v4", "method", fmt.Sprintf("Op%d", i%4), &ast.Comment{Summary: want})
			if merged.Summary != want {
				t.Errorf("got %q want %q", merged.Summary, want)
			}
		})
	}
	wg.Wait()
}
