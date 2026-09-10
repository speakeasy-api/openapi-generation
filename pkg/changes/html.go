package changes

import (
	"fmt"

	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
)

// ToHTML converts the SDKDiff to HTML format with full detail
func ToHTML(d SDKDiff) []byte {
	// Get full-detail markdown
	md := ToMarkdown(d, DetailLevelFull)

	// Convert markdown to HTML
	extensions := parser.CommonExtensions | parser.AutoHeadingIDs
	p := parser.NewWithExtensions(extensions)
	doc := p.Parse([]byte(md))

	htmlFlags := html.CommonFlags | html.HrefTargetBlank
	opts := html.RendererOptions{Flags: htmlFlags}
	renderer := html.NewRenderer(opts)

	htmlBody := markdown.Render(doc, renderer)

	// Wrap in styled HTML document
	lang := ""
	if d.OldSubsystem != nil && d.OldSubsystem.Target.Target != "" {
		lang = d.OldSubsystem.Target.Target
	}

	return wrapInHTMLDocument(htmlBody, lang)
}

func wrapInHTMLDocument(body []byte, lang string) []byte {
	template := `<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta name="color-scheme" content="light dark">
    <title>SDK Changelog - %s</title>
    <style>
        :root {
            --bg-color: #ffffff;
            --text-color: #1a1a1a;
            --text-secondary: #666666;
            --code-bg: #f4f4f4;
            --border-color: #e0e0e0;
            --added-color: #22863a;
            --removed-color: #cb2431;
            --changed-color: #6f42c1;
            --breaking-color: #d73a49;
        }
        @media (prefers-color-scheme: dark) {
            :root {
                --bg-color: #1a1a1a;
                --text-color: #e6e6e6;
                --text-secondary: #a0a0a0;
                --code-bg: #2d2d2d;
                --border-color: #404040;
                --added-color: #85e89d;
                --removed-color: #f97583;
                --changed-color: #b392f0;
                --breaking-color: #f97583;
            }
        }
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
            max-width: 900px;
            margin: 40px auto;
            padding: 20px;
            line-height: 1.7;
            background-color: var(--bg-color);
            color: var(--text-color);
        }
        code {
            background: var(--code-bg);
            padding: 2px 8px;
            border-radius: 4px;
            font-family: 'SF Mono', 'Menlo', 'Monaco', 'Courier New', monospace;
            font-size: 0.9em;
        }
        h2 {
            color: var(--text-color);
            border-bottom: 2px solid var(--border-color);
            padding-bottom: 12px;
            margin-top: 32px;
        }
        ul {
            padding-left: 24px;
            list-style-type: disc;
            list-style-position: outside;
        }
        ul ul {
            padding-left: 28px;
            margin-top: 6px;
            margin-bottom: 6px;
            list-style-type: circle;
        }
        li {
            margin: 10px 0;
            display: list-item;
        }
        li li {
            margin: 4px 0;
            color: var(--text-secondary);
            display: list-item;
        }
        li li code {
            font-size: 0.85em;
        }
        /* Ensure nested list items always show bullets */
        ul > li {
            list-style: disc;
        }
        ul ul > li {
            list-style: circle;
        }
        strong {
            font-weight: 600;
        }
        /* Semantic coloring for change types */
        strong:contains('Added'), li:has(strong:contains('Added')) strong:first-of-type {
            color: var(--added-color);
        }
        /* Warning emoji styling */
        li {
            list-style-position: outside;
        }
    </style>
</head>
<body>%s</body>
</html>`
	return []byte(fmt.Sprintf(template, lang, string(body)))
}
