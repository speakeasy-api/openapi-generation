module example.com/openapi-go-sdk

go 1.26.8

require (
	github.com/ericlagergren/decimal v0.0.0-20221120152707-495c53812d05
	github.com/google/uuid v1.6.0
	github.com/spyzhov/ajson v0.9.6
	github.com/stretchr/testify v1.11.1
)

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

retract (
	v1.0.1
	v1.0.0 // Published accidentally
)
