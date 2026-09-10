package validation

import (
	"context"
	goErrs "errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	generatorErrors "github.com/speakeasy-api/openapi-generation/v2/pkg/errors"
	"github.com/speakeasy-api/openapi/linter"
	"github.com/speakeasy-api/openapi/openapi"
	"github.com/speakeasy-api/openapi/validation"
	"gopkg.in/yaml.v3"
)

const (
	ErrInvalidPath         = generatorErrors.Error("invalid path")
	ErrUnencodedCharacters = generatorErrors.Error("unencoded characters")
)

var templateRegex = regexp.MustCompile(`({[^#/?]+?})`)

type ValidatePaths struct{}

var _ Rule = (*ValidatePaths)(nil)

func (r *ValidatePaths) ID() string {
	return "generator-validate-paths"
}

func (r *ValidatePaths) Category() string {
	return "validation"
}

func (r *ValidatePaths) Summary() string {
	return "Validate paths use RFC 3986 and valid URI templates."
}

func (r *ValidatePaths) HowToFix() string {
	return "Update path strings to be valid RFC 3986 URIs and correct URI template syntax, removing unencoded characters and invalid templates."
}

func (r *ValidatePaths) Description() string {
	return "Validate paths conform to RFC 3986 and use valid URI template syntax. This prevents illegal characters and ambiguous route matching."
}

func (r *ValidatePaths) Link() string {
	return ""
}

func (r *ValidatePaths) DefaultSeverity() validation.Severity {
	return validation.SeverityError
}

func (r *ValidatePaths) Versions() []string {
	return nil // Applies to all versions
}

func (r *ValidatePaths) Run(ctx context.Context, docInfo *linter.DocumentInfo[*openapi.OpenAPI], config *linter.RuleConfig) []error {
	if docInfo == nil || docInfo.Document == nil {
		return nil
	}

	doc := docInfo.Document
	if doc.Paths == nil {
		return nil
	}

	var validationErrors []error

	// Iterate through actual paths in the document
	for pathStr, pathItem := range doc.Paths.All() {

		// Skip OpenAPI extensions (keys starting with "x-")
		if strings.HasPrefix(pathStr, "x-") {
			continue
		}

		errs := validatePath(pathStr)

		if len(errs) > 0 {
			// Get the node for error reporting
			var node *yaml.Node
			if pathItem != nil && pathItem.GetCore() != nil {
				node = pathItem.GetRootNode()
			}

			if goErrs.Is(errs[0], ErrInvalidPath) {
				validationErrors = append(validationErrors, &validation.Error{
					Rule:            r.ID(),
					Severity:        r.DefaultSeverity(),
					Node:            node,
					UnderlyingError: goErrs.New("path is invalid unable to be parsed"),
				})
			} else {
				unencodedCharacters := []string{}

				for _, err := range errs {
					if goErrs.Is(err, ErrUnencodedCharacters) {
						unencodedCharacters = append(unencodedCharacters, strings.Split(goErrs.Unwrap(err).Error(), ",")...)
					}
				}

				if len(unencodedCharacters) > 0 {
					validationErrors = append(validationErrors, &validation.Error{
						Rule:            r.ID(),
						Severity:        r.DefaultSeverity(),
						Node:            node,
						UnderlyingError: fmt.Errorf("path contains unencoded characters [ `%s` ]", strings.Join(unencodedCharacters, " ")),
					})
				}
			}
		}
	}

	return validationErrors
}

func validatePath(path string) []error {
	// 1. Handle Templates
	// We replace templates with "test" so we can validate the rest of the syntax.
	path = templateRegex.ReplaceAllString(path, "test")

	// 2. Parse check
	// We still run url.Parse to ensure it's a structural URI, but we won't use
	// u.Path for validation segments because it auto-decodes (ruining %20 checks).
	u, err := url.Parse(path)
	if err != nil {
		return []error{ErrInvalidPath.Wrap(err)}
	}

	if path == "" || path[0] != '/' {
		return []error{ErrInvalidPath}
	}

	errors := []error{}

	// 3. Extract Raw Path (Prevent decoding)
	// We strip Query (?) and Fragment (#) manually to keep the path raw.
	rawPath := path
	if idx := strings.IndexAny(rawPath, "?#"); idx != -1 {
		rawPath = rawPath[:idx]
	}

	// 4. Validate Path Segments
	segments := strings.Split(rawPath, "/")
	for _, segment := range segments {
		if segment == "" {
			continue // Skip empty segments (e.g. leading /)
		}
		if err := validateSegment(segment); err != nil {
			errors = append(errors, err)
		}
	}

	// 5. Validate Query & Fragment
	// We use RawQuery to ensure we check encoded values properly.
	if u.RawQuery != "" {
		if err := validateSegment(u.RawQuery); err != nil {
			errors = append(errors, err)
		}
	}

	// Fragment is not stored raw in u.RawFragment (it doesn't exist), so we pull it manually.
	if idx := strings.Index(path, "#"); idx != -1 {
		// +1 to skip the '#'
		rawFragment := path[idx+1:]
		if err := validateSegment(rawFragment); err != nil {
			errors = append(errors, err)
		}
	}

	return errors
}

// validateSegment checks if the string contains only allowed RFC 3986 characters.
// This is much more lenient (and correct) than checking against url.PathEscape.
func validateSegment(segment string) error {
	unencodedCharacters := []string{}

	for i := 0; i < len(segment); i++ {
		c := segment[i]

		// 1. Allowed: Unreserved Characters (A-Z, a-z, 0-9, -, ., _, ~)
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || (c >= '0' && c <= '9') ||
			c == '-' || c == '.' || c == '_' || c == '~' {
			continue
		}

		// 2. Allowed: Sub-delimiters (! $ & ' ( ) * + , ; =)
		if c == '!' || c == '$' || c == '&' || c == '\'' || c == '(' || c == ')' ||
			c == '*' || c == '+' || c == ',' || c == ';' || c == '=' {
			continue
		}

		// 3. Allowed: Other pchar (: @)
		if c == ':' || c == '@' {
			continue
		}

		// 4. Allowed: Valid Percent Encoding (% followed by 2 Hex Digits)
		if c == '%' {
			if i+2 < len(segment) && isHex(segment[i+1]) && isHex(segment[i+2]) {
				i += 2 // Skip the next two characters as they are part of the sequence
				continue
			}
		}

		// If none of the above, it's invalid (e.g., Space, <, >, {, }, etc.)
		unencodedCharacters = append(unencodedCharacters, string(c))
	}

	if len(unencodedCharacters) > 0 {
		return ErrUnencodedCharacters.Wrap(goErrs.New(strings.Join(unencodedCharacters, ",")))
	}

	return nil
}

func isHex(c byte) bool {
	return (c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')
}
