// Copyright 2022 Dave Shanley / Quobix
// SPDX-License-Identifier: MIT
package report

import (
	"errors"
	"html"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/speakeasy-api/openapi/validation"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestHTMLReportGenerateReportEscapesUntrustedContent(t *testing.T) {
	const (
		sourcePayload  = `<img src=x onerror=alert('SEC-50-source')>`
		scriptPayload  = `"><script>alert('SEC-50-message')</script>`
		hostileRule    = `validation-sec-50"><img src=x onerror=alert('SEC-50-rule')>`
		hostileVersion = `v1"><script>alert('SEC-50-version')</script>`
	)

	sourceImageLine := `  description: "` + sourcePayload + `"`
	sourceScriptLine := `  x-message: '` + scriptPayload + `'`
	specBytes := []byte(strings.Join([]string{
		"openapi: 3.0.3",
		"info:",
		"  title: SEC-50 regression",
		sourceImageLine,
		sourceScriptLine,
		"  version: 1.0.0",
		"paths: {}",
	}, "\n"))

	message := "validation failed: " + sourcePayload + " " + scriptPayload
	metadata := &RuleMetadata{
		Summary:     `summary"><img src=x onerror=alert('SEC-50-summary')>`,
		Description: `description"><script>alert('SEC-50-description')</script>`,
		HowToFix:    `fix"><img src=x onerror=alert('SEC-50-fix')>`,
	}
	report := NewHTMLReport(&ValidationReportData{
		Results: []*validation.Error{
			{
				UnderlyingError: errors.New(message),
				Node:            &yaml.Node{Line: 4, Column: 3},
				Severity:        validation.SeverityError,
				Rule:            hostileRule,
			},
		},
		Statistics: &ReportStatistics{
			TotalErrors:   1,
			TotalWarnings: 2,
			TotalHints:    3,
		},
		SpecBytes: specBytes,
		Generated: time.Date(2026, time.August, 4, 12, 0, 0, 0, time.UTC),
		RuleMetadata: map[string]*RuleMetadata{
			hostileRule: metadata,
		},
	})

	output := string(report.GenerateReport(false, hostileVersion))

	require.NotEmpty(t, output)
	require.NotContains(t, output, "failed to render")
	require.NotContains(t, output, "ZgotmplZ")
	for _, payload := range []string{
		sourcePayload,
		scriptPayload,
		hostileRule,
		hostileVersion,
		metadata.Summary,
		metadata.Description,
		metadata.HowToFix,
	} {
		require.NotContains(t, output, payload)
	}

	for name, value := range map[string]string{
		"rule ID":          hostileRule,
		"version":          hostileVersion,
		"metadata summary": metadata.Summary,
		"metadata detail":  metadata.Description,
		"metadata fix":     metadata.HowToFix,
	} {
		require.Contains(t, output, html.EscapeString(value), "%s should be rendered in escaped form", name)
	}

	require.Contains(t, output, `<pre tabindex="0" class="chroma"><code>`)
	require.Contains(t, output, `<span class="cl">`+html.EscapeString(sourceImageLine)+`</span>`)
	require.Contains(t, output, `<span class="cl">`+html.EscapeString(sourceScriptLine)+`</span>`)

	messageAttribute := regexp.MustCompile(`\bmessage="([^"]*)"`).FindStringSubmatch(output)
	require.Len(t, messageAttribute, 2)
	require.Equal(t, message, html.UnescapeString(messageAttribute[1]))

	require.Contains(t, output, `window.statistics = {"TotalErrors":1,"TotalWarnings":2,"TotalHints":3}`)
	require.NotContains(t, strings.ToLower(bundledJS), "</script")
	require.NotContains(t, strings.ToLower(hydrateJS), "</script")
	require.NotContains(t, strings.ToLower(reportCSS), "</style")
	require.Contains(t, output, bundledJS, "bundled JavaScript should be embedded verbatim")
	require.Contains(t, output, hydrateJS, "hydration JavaScript should be embedded verbatim")
	require.Contains(t, output, reportCSS, "report CSS should be embedded verbatim")
}
