package ast

import (
	"fmt"
	"hash/fnv"

	"github.com/speakeasy-api/openapi-generation/v2/internal/contenttypes"
)

// BaseOperation represents an operation that could be a method, webhook or callback
type BaseOperation struct {
	ID                        string               `yaml:",omitempty"` // The unique ID of the operation
	OriginalID                string               `yaml:",omitempty"` // The original ID of the operation this won't be mutated by having multiple request bodies for example
	Request                   *Request             `yaml:",omitempty"` // The request of the operation
	Response                  *Response            `yaml:",omitempty"` // The response of the operation
	UsesUserAgentHeader       bool                 `yaml:",omitempty"` // Whether the operation uses the User-Agent header
	SerializationMethod       *SerializationMethod `yaml:",omitempty"` // The serialization method for the operation
	SerializationMethodSuffix string               `yaml:",omitempty"` // The serialization method suffix for the operation if its been split into multiple methods
}

// Operation represents an operation within an SDK
type Operation struct {
	BaseOperation
	Path                   string                 `yaml:",omitempty"` // The API path the operation is associated with
	Method                 string                 `yaml:",omitempty"` // The HTTP method the operation uses
	Security               *FieldDef              `yaml:",omitempty"` // The security definition for the operation (if any)
	GlobalSecurity         *FieldDef              `yaml:",omitempty"` // The global security definition for the SDK (if any)
	HoistedSecurityConfig  *HoistedSecurityConfig `yaml:",omitempty"` // Only set if the operation security was hoisted to the global level
	OAuth2Config           *OAuth2Config          `yaml:",omitempty"` // The operation-specific OAuth2 configuration override
	Scope                  Scope                  `yaml:",omitempty"` // The scope of the operation
	Servers                *Servers               `yaml:",omitempty"` // The list of servers that are specific to the operation
	Comments               *Comment               `yaml:",omitempty"` // The comments associated with the operation
	Tags                   []string               `yaml:",omitempty"` // The OpenAPI tags associated with the operation
	UsesUserAgentHeader    bool                   `yaml:",omitempty"` // Whether the operation uses the User-Agent header
	Callbacks              []*TypeDef             `yaml:",omitempty"` // The list of callbacks associated with the operation
	OwningSDK              *SDK                   `yaml:"-"`          // The SDK that owns the operation
	Extensions             *OperationExtensions   `yaml:",omitempty"` // The extensions associated with the operation
	Globals                *TypeDef               `yaml:",omitempty"` // The global variables associated with the operation
	MaxMethodParams        int                    `yaml:",omitempty"` // The maximum number of parameters the method can have
	Arguments              *Arguments             `yaml:",omitempty"` // The method arguments which include parameters and request body fields
	TestExplicitlyDisabled bool                   `yaml:",omitempty"` // Whether the test configuration was explicitly disabled
	Webhook                *Webhook               `yaml:",omitempty"` // Whether the operation is a webhook
	Location               *OpenAPILocation       `yaml:"-"`          // The location of the operation in the OpenAPI document
	exampleSeed            int                    `yaml:"-"`          // The cached example seed for the operation
}

// Returns true if the Operation contains a truncated (circular reference) type
// for the Request or Response.
func (o *Operation) ContainsTruncated() bool {
	if o.Request != nil && o.Request.ContainsTruncated() {
		return true
	}

	if o.Response != nil && o.Response.ContainsTruncated() {
		return true
	}

	return false
}

func (o *Operation) Match(matchers Matchers) error {
	if matchers.Operation != nil {
		return matchers.Operation(o)
	}

	return nil
}

func (o Operation) GetID() string {
	if o.Extensions.MethodNameOverride != "" {
		return fmt.Sprintf("%s%s", o.Extensions.MethodNameOverride, o.SerializationMethodSuffix)
	}

	return o.ID
}

func (o *Operation) GetExampleSeed() int {
	if o.exampleSeed != 0 {
		return o.exampleSeed
	}

	// convert OriginalID string to int
	hash := fnv.New32a()
	_, _ = hash.Write([]byte(o.OriginalID))

	o.exampleSeed = int(hash.Sum32())
	return o.exampleSeed
}

func (o Operation) GetSerializationMethod() SerializationMethod {
	if o.SerializationMethod == nil {
		return ""
	}

	return *o.SerializationMethod
}

func (o Operation) GetAcceptTypes() []string {
	successAccepts := []string{}
	errorAccepts := []string{}

	for _, response := range o.Response.Responses {
		for _, content := range response.Content {
			if response.Error {
				errorAccepts = append(errorAccepts, content.ContentType)
			} else {
				successAccepts = append(successAccepts, content.ContentType)
			}
		}
	}

	accept := successAccepts
	if len(successAccepts) == 0 {
		accept = errorAccepts
	}

	if len(accept) == 0 {
		accept = append(accept, "*/*")
	}

	accept = contenttypes.SortAcceptTypes(contenttypes.DeDupeAcceptTypes(accept))

	if len(accept) > 1 {
		increments := 1.0 / float64(len(accept))

		for i, acceptType := range accept {
			var quality string

			switch {
			case i == 0:
				quality = "1"
			case i == len(accept)-1:
				quality = "0"
			case len(accept) > 9:
				quality = fmt.Sprintf("%.2f", 1.0-(float64(i)*increments))
			default:
				quality = fmt.Sprintf("%.1f", 1.0-(float64(i)*increments))
			}

			accept[i] = acceptType + ";q=" + quality
		}
	}

	return accept
}
