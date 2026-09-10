package ast

import "slices"

type SecurityScheme string

// A security requirement can be either:
// - a single scheme
// - a composition of multiple schemes (AND)
type SecurityRequirement []SecurityScheme

func AreEquivalentRequirements(a, b []SecurityRequirement) bool {
	a = sanitizeRequirements(a)
	b = sanitizeRequirements(b)
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if !areEquivalentSchemes(a[i], b[i]) {
			return false
		}
	}
	return true
}

func sanitizeRequirements(reqs []SecurityRequirement) []SecurityRequirement {
	result := make([]SecurityRequirement, 0, len(reqs))
	seenEmpty := false
	for _, req := range reqs {
		if len(req) == 0 {
			if seenEmpty {
				continue
			}
			seenEmpty = true
		}
		result = append(result, req)
	}
	return result
}

func areEquivalentSchemes(a, b []SecurityScheme) bool {
	if len(a) != len(b) {
		return false
	}
	sortedA := slices.Sorted(slices.Values(a))
	sortedB := slices.Sorted(slices.Values(b))
	return slices.Equal(sortedA, sortedB)
}

// Security represents a security configuration block
type Security struct {
	// Security represents the field added either to the SDK or an operation representing how its security is configured
	Security *FieldDef
	// Requirements are stored to determine if operation security is a subset of global security.
	Requirements []SecurityRequirement
	// SecurityConfig represents the security configuration options for this security block
	SecurityConfig SecurityConfig
}

type SecurityConfig struct {
	// OptionalityReason represents the reason why this security block is optional
	OptionalityReason SecurityOptionalityReason
	// OAuth2Config configures a OAuth2 flow in this security block, used to collect OAuth2Scopes and enable hooks
	OAuth2Config OAuth2Config
	// Disabled indicates whether this security block was explicitly disabled
	Disabled bool
	// HoistedSecurityConfig is populated when the operation security was hoisted to the global level.
	// Nil when the operation defines its own non-matching security or no operation-level security exists.
	HoistedSecurityConfig *HoistedSecurityConfig
}

// HoistedSecurityConfig describes how an operation's security was hoisted to global security.
type HoistedSecurityConfig struct {
	// Equivalent is true when the operation's security requirements functionally match
	// global security (identical schemes, same OR-order, and matching optionality).
	// When false, templates must filter global security fields using Fields.
	Equivalent bool
	// Fields maps operation-level security requirements to their corresponding global security
	// fields, ordered according to the operation-level security definition.
	Fields []HoistedSecurityField
}

// HoistedSecurityField keeps track of a hoisted operation-level security requirement.
type HoistedSecurityField struct {
	// The global security field name corresponding to the hoisted security requirement
	Name string `yaml:",omitempty"`
	// The position of the hoisted field in the global security field slice
	Index int `yaml:",omitempty"`
	// The index of the requirement group, used to identify composite requirements (AND).
	// For flattened schemes (e.g. Username/Password): Group will always be 0.
	// For simple OR configurations (unflattened): Index and Group will match.
	Group int `yaml:",omitempty"`
}

// IsSubsetOfGlobalSecurity returns true if the operation's security requirements are a subset of (or equivalent to)
// the global security requirements. In this case the operation-security gets hoisted and global security is reused.
// Note for this function to return true, the operation security does not have to be a *strict* (or proper) subset.
func (c SecurityConfig) IsSubsetOfGlobalSecurity() bool {
	return c.HoistedSecurityConfig != nil && len(c.HoistedSecurityConfig.Fields) > 0
}

type OAuth2Flow string

const (
	OAuth2FlowNone              OAuth2Flow = "none"
	OAuth2FlowClientCredentials OAuth2Flow = "client_credentials"
	OAuth2FlowPassword          OAuth2Flow = "password"
	OAuth2FlowImplicit          OAuth2Flow = "implicit"
	OAuth2FlowAuthorizationCode OAuth2Flow = "authorization_code"
)

type OAuth2Scope struct {
	Name     string
	Comments Comment
}

type OAuth2FlowConfig struct {
	Flow            OAuth2Flow
	Enabled         bool
	Comments        Comment
	RequiredScopes  []string
	AvailableScopes []OAuth2Scope
}

type OAuth2Config map[string]OAuth2FlowConfig

type SecurityOptionalityReason string

const (
	// SecOptReasonNotOptional indicates that the security has been determined to be required
	SecOptReasonNotOptional       SecurityOptionalityReason = "not-optional"
	SecOptReasonOptionalScheme    SecurityOptionalityReason = "optional-scheme"
	SecOptReasonOperationOverride SecurityOptionalityReason = "operation-override"
	SecOptReasonEnvVar            SecurityOptionalityReason = "env-var"
)
