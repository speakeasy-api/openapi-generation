package types

import (
	"strconv"
	"strings"
)

type Maturity string

const (
	MaturityGA    Maturity = "GA"
	MaturityBeta  Maturity = "Beta"
	MaturityAlpha Maturity = "Alpha"
)

type SupportLevel string

const (
	SupportLevelGA  SupportLevel = "GA"
	SupportLevelTwo SupportLevel = "Level 2"
	SupportLevelOne SupportLevel = "Level 1"
)

// TODO: this will eventually be replaced by a report from each target, but hard coded for now
var TargetMaturity = map[string]Maturity{
	"cli":            MaturityBeta,
	"csharp":         MaturityGA,
	"go":             MaturityGA,
	"java":           MaturityGA,
	"mcp-typescript": MaturityBeta,
	"php":            MaturityGA,
	"postman":        MaturityAlpha,
	"python":         MaturityGA,
	"ruby":           MaturityGA,
	"terraform":      MaturityGA,
	"typescript":     MaturityGA,
	"unity":          MaturityBeta,
}

var MethodologyPaths = map[string]string{
	"typescript": "/docs/languages/typescript/methodology-ts",
	"go":         "/docs/languages/golang/methodology-go",
	"java":       "/docs/languages/java/methodology-java",
	"python":     "/docs/languages/python/methodology-python",
	"csharp":     "/docs/languages/csharp/methodology-csharp",
	"terraform":  "/docs/create-terraform",
	"php":        "/docs/languages/php/methodology-php",
	"unity":      "/docs/languages/unity/methodology-unity",
	"ruby":       "/docs/languages/ruby/methodology-ruby",
}

var SupportLevels = map[string]SupportLevel{
	"cli":            SupportLevelTwo,
	"csharp":         SupportLevelGA,
	"go":             SupportLevelGA,
	"java":           SupportLevelGA,
	"php":            SupportLevelOne,
	"postman":        SupportLevelOne,
	"python":         SupportLevelGA,
	"ruby":           SupportLevelOne,
	"terraform":      SupportLevelTwo,
	"typescript":     SupportLevelGA,
	"unity":          SupportLevelOne,
	"mcp-typescript": SupportLevelTwo,
}

type Target struct {
	// Target is the name of the target language ie "go", "python", "typescript"
	Target string
	// Template is the name of the template folder to use for the target ie "go", "typescriptv2", "javav2"
	Template string
	Maturity Maturity
}

func NewTargetFromTemplate(template string) Target {
	m := MaturityAlpha

	target := getTargetFromTemplate(template)

	if maturity, ok := TargetMaturity[target]; ok {
		m = maturity
	}

	return Target{
		Target:   target,
		Template: template,
		Maturity: m,
	}
}

func getTargetFromTemplate(template string) string {
	target := template

	lastIdx := strings.LastIndex(template, "v")
	if lastIdx != -1 {
		suffix := template[lastIdx+1:]

		if _, err := strconv.Atoi(suffix); err == nil {
			target = template[:lastIdx]
		}
	}

	return target
}
