module github.com/speakeasy-api/openapi-generation/v2/cmd/wasm/overlay

go 1.26.2

require (
	github.com/speakeasy-api/jsonpath v0.6.3
	github.com/speakeasy-api/openapi v1.25.1
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/dprotaso/go-yit v0.0.0-20191028211022-135eb7262960 // indirect
	github.com/vmware-labs/yaml-jsonpath v0.3.2 // indirect
)

replace github.com/dop251/goja => github.com/speakeasy-api/goja v0.0.0-20260223084236-ed0328a0a462

replace github.com/dop251/goja/debugger => github.com/speakeasy-api/goja/debugger v0.0.0-20260223084236-ed0328a0a462

replace github.com/speakeasy-api/openapi-generation/v2 => ../../..
