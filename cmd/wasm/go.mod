module github.com/speakeasy-api/openapi-generation/v2/cmd/wasm

go 1.26.2

replace github.com/speakeasy-api/openapi-generation/v2 => ../../

replace github.com/testcontainers/testcontainers-go => ./internal/testcontainers

replace github.com/testcontainers/testcontainers-go/wait => ./internal/testcontainers/wait

replace github.com/pb33f/doctor => github.com/speakeasy-api/doctor v0.20.0-fixschemawalk

replace github.com/daveshanley/vacuum => github.com/speakeasy-api/vacuum v0.16.16

replace github.com/pb33f/libopenapi => github.com/speakeasy-api/libopenapi v0.21.8-wasm
