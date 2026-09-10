package ast

// Scope represents within which scope a type is defined
type Scope string

const (
	ScopeShared     Scope = "shared"     // The shared scope contains types that are generally components used by multiple operations
	ScopeOperations Scope = "operations" // The operations scope contains request/response types and operations
	ScopeUtils      Scope = "utils"      // The utils scope represents types/methods that are provided by utils packages
	ScopeSDK        Scope = "sdk"        // The sdk scope represents the sdk classes
	ScopeWebhooks   Scope = "webhooks"   // The webhooks scope represents the webhook types
	ScopeCallbacks  Scope = "callbacks"  // The callbacks scope represents the callback types
	ScopeErrors     Scope = "errors"     // The errors scope represents the error types
	ScopeGlobals    Scope = "globals"    // The globals scope represents the global variables
)
