package validation_test

import (
	"context"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/types"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/speakeasy-api/openapi-generation/v2/internal/validation"
	config "github.com/speakeasy-api/sdk-gen-config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_DuplicatePathParams_Success(t *testing.T) {
	type args struct {
		schema string
	}
	tests := []struct {
		name     string
		args     args
		wantErrs []string
		ruleset  string
	}{
		{

			name: "There should be warning errors when RulesetSpeakeasyGeneration is used and there are duplicate parameters defined inside parameter object ",
			args: args{
				schema: utils.Dedent(`
        openapi: 3.1.0
        info:
          title: FastAPI
          version: 0.1.0
          description: stop complaining
        paths:
          /musical/{melody}/{pizza}/:
            parameters:
              - name: melody
                in: path
                required: true
              - name: melody
                in: path
                required: true
            get:
              parameters:
                - in: path
                  name: pizza
                  required: true
              responses:
                '200':
                  description: OK`),
			},
			wantErrs: []string{"validation warn: [line 12:9] validation-operation-parameters - parameter \"melody\" is duplicated in path \"/musical/{melody}/{pizza}/\""},
			ruleset:  validation.RulesetSpeakeasyGeneration,
		},
		{

			name: "There should be warnings when RulesetSpeakeasyGeneration is used and there are duplicate parameters defined inside parameter object ",
			args: args{
				schema: utils.Dedent(`
		openapi: 3.1.0
		info:
		  title: FastAPI
		  version: 0.1.0
		  description: stop complaining
		paths:
		  /musical/{melody}/{pizza}/:
		    parameters:
		      - name: melody
		        in: path
		        required: true
		      - name: melody
		        in: path
		        required: true
		    get:
		      parameters:
		        - in: path
		          name: pizza
		          required: true
		      responses:
		        '200':
		          description: OK`),
			},
			wantErrs: []string{"validation warn: [line 12:9] validation-operation-parameters - parameter \"melody\" is duplicated in path \"/musical/{melody}/{pizza}/\""},
			ruleset:  validation.RulesetSpeakeasyRecommended,
		},
		{

			name: "There should be warning validation errors when parameter names are duplicated inside path uri and ruleset is RulesetSpeakeasyGeneration",
			args: args{
				schema: utils.Dedent(`
		  openapi: 3.1.0
		  info:
		    title: FastAPI
		    version: 0.1.0
		    description: stop complaining
		  paths:
		    /musical/{melody}/{melody}/:
		      parameters:
		        - name: melody
		          in: path
		          required: true
		      get:
		        responses:
		          '200':
		            description: OK`)},
			wantErrs: []string{"validation warn: [line 8:5] generator-duplicate-path-params - path `/musical/{melody}/{melody}/` must not use the parameter `melody` multiple times"},
			ruleset:  validation.RulesetSpeakeasyGeneration,
		},
		{

			name: "There should be warning validation errors when parameter names are duplicated inside path uri and pathparameter object, and ruleset is RulesetSpeakeasyGeneration",
			args: args{
				schema: utils.Dedent(`
		  openapi: 3.1.0
		  info:
		    title: FastAPI
		    version: 0.1.0
		    description: stop complaining
		  paths:
		    /musical/{melody}/{melody}/:
		      parameters:
		        - name: melody
		          in: path
		          required: true
		        - name: melody
		          in: path
		          required: true
		      get:
		        responses:
		          '200':
		            description: OK`)},
			wantErrs: []string{
				"validation warn: [line 8:5] generator-duplicate-path-params - path `/musical/{melody}/{melody}/` must not use the parameter `melody` multiple times",
				"validation warn: [line 12:9] validation-operation-parameters - parameter \"melody\" is duplicated in path \"/musical/{melody}/{melody}/\"",
			},
			ruleset: validation.RulesetSpeakeasyGeneration,
		},
		{
			name: "There should be warning validations when parameter names are duplicated inside path uri and ruleset is RulesetSpeakeasyRecommended",
			args: args{
				schema: utils.Dedent(`
		  openapi: 3.1.0
		  info:
		    title: FastAPI
		    version: 0.1.0
		    description: stop complaining
		  paths:
		    /musical/{melody}/{melody}/:
		      parameters:
		        - name: melody
		          in: path
		          required: true
		      get:
		        responses:
		          '200':
		            description: OK`)},
			wantErrs: []string{"validation warn: [line 8:5] generator-duplicate-path-params - path `/musical/{melody}/{melody}/` must not use the parameter `melody` multiple times"},
			ruleset:  validation.RulesetSpeakeasyRecommended,
		},
		{
			name: "There should be no errors when parameter names are not duplicated inside path uri ",
			args: args{
				schema: utils.Dedent(`
		      openapi: 3.1.0
		      info:
		        title: FastAPI
		        version: 0.1.0
		        description: stop complaining
		      paths:
		        /musical/{melody}/:
		          parameters:
		            - name: melody
		              in: path
		              required: true
		          get:
		            responses:
		              '200':
		                description: OK`)},
			wantErrs: []string{},
			ruleset:  validation.RulesetSpeakeasyGeneration,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("SPEAKEASY_DEBUG", "true")
			ruleset := tt.ruleset
			if ruleset == "" {
				panic("Ruleset is not set. Set ruleset in the test case")
			}
			v, err := validation.NewValidator(&config.Configuration{}, tt.ruleset, validation.WithFilteredRules([]string{"generator-duplicate-path-params", "validation-operation-parameters"}))
			require.NoError(t, err)

			res := validateSpec(v, context.Background(), []byte(tt.args.schema), "", types.NewTargetFromTemplate("go"))
			errs := res.GetValidationErrors()
			errStrs := []string{}
			for _, err := range errs {
				errStrs = append(errStrs, err.Error())
			}

			assert.ElementsMatch(t, tt.wantErrs, errStrs)
		})
	}
}
