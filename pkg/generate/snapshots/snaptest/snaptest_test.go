package snaptest

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAnalyzeGenerateErrors_ClassifiesCompileAndLint(t *testing.T) {
	t.Parallel()

	lintErr := errors.New("lint failed for sdk")
	compileErr := errors.New("compilation failed for sdk")
	generationErr := errors.New("plain generation failure")

	report := analyzeGenerateErrors([]error{
		lintErr,
		compileErr,
		generationErr,
	}, true, nil)

	assert.Equal(t, []error{generationErr}, report.GenerationErrors)
	assert.Equal(t, []error{compileErr}, report.CompilationErrors)
	assert.Equal(t, []error{lintErr}, report.LintErrors)
	assert.Empty(t, report.IgnoredCompile)
	assert.Empty(t, report.IgnoredLint)
	assert.True(t, report.HasFailures())
}

func TestAnalyzeGenerateErrors_AcceptsExpectedErrors(t *testing.T) {
	t.Parallel()

	report := analyzeGenerateErrors([]error{
		errors.New("merge conflicts detected in 1 file(s)"),
	}, false, []string{"merge conflicts detected"})

	assert.Empty(t, report.MissingExpected)
	assert.False(t, report.HasFailures())
}

func TestAnalyzeGenerateErrors_DuplicateExpectedConsumesDistinctErrors(t *testing.T) {
	t.Parallel()

	report := analyzeGenerateErrors([]error{
		errors.New("merge conflicts detected in pkg/foo/foo.go"),
		errors.New("merge conflicts detected in pkg/bar/bar.go"),
	}, false, []string{"merge conflicts detected", "merge conflicts detected"})

	assert.Empty(t, report.MissingExpected)
	assert.Empty(t, report.GenerationErrors)
	assert.False(t, report.HasFailures())
}

func TestAnalyzeGenerateErrors_SingleExpectedDoesNotSwallowExtraMatches(t *testing.T) {
	t.Parallel()

	extra := errors.New("merge conflicts detected in pkg/bar/bar.go")
	report := analyzeGenerateErrors([]error{
		errors.New("merge conflicts detected in pkg/foo/foo.go"),
		extra,
	}, false, []string{"merge conflicts detected"})

	assert.Empty(t, report.MissingExpected)
	assert.Equal(t, []error{extra}, report.GenerationErrors)
	assert.True(t, report.HasFailures())
}

func TestAnalyzeGenerateErrors_OverlappingExpectedNotConsumedTwice(t *testing.T) {
	t.Parallel()

	report := analyzeGenerateErrors([]error{
		errors.New("merge conflicts detected in pkg/foo/foo.go"),
	}, false, []string{"pkg/foo/foo.go", "merge conflicts detected"})

	assert.Equal(t, []string{"merge conflicts detected"}, report.MissingExpected)
	assert.True(t, report.HasFailures())
}

func TestAnalyzeGenerateErrors_ReportsMissingExpectedErrors(t *testing.T) {
	t.Parallel()

	report := analyzeGenerateErrors(nil, false, []string{"expected failure"})

	assert.Equal(t, []string{"expected failure"}, report.MissingExpected)
	assert.True(t, report.HasFailures())
}

func TestMatchExcludeGlob(t *testing.T) {
	t.Parallel()

	tests := []struct {
		glob    string
		rel     string
		want    bool
		wantErr bool
	}{
		{glob: "docs/**", rel: "docs/models/widget.md", want: true},
		{glob: "docs/**", rel: "docs/README.md", want: true},
		{glob: "docs/**", rel: "docs", want: true},
		{glob: "docs/**", rel: "docsx/widget.md", want: false},
		{glob: "docs/**", rel: "internal/docs/widget.md", want: false},
		{glob: "docs/*.md", rel: "docs/widget.md", want: true},
		{glob: "docs/*.md", rel: "docs/models/widget.md", want: false},
		{glob: "README.md", rel: "README.md", want: true},
		{glob: "README.md", rel: "docs/README.md", want: false},
		{glob: "docs/[", rel: "docs/widget.md", wantErr: true},
	}

	for _, tt := range tests {
		got, err := matchExcludeGlob(tt.glob, tt.rel)
		if (err != nil) != tt.wantErr {
			t.Errorf("matchExcludeGlob(%q, %q) error = %v, wantErr %v", tt.glob, tt.rel, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("matchExcludeGlob(%q, %q) = %v, want %v", tt.glob, tt.rel, got, tt.want)
		}
	}
}
