package ast_test

import (
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/stretchr/testify/assert"
)

func TestOperation_GetAcceptTypes(t *testing.T) {
	type args struct {
		op ast.Operation
	}
	tests := []struct {
		name string
		args args
		want []string
	}{
		{
			name: "Returns accept types with single digit quality values",
			args: args{
				op: ast.Operation{
					BaseOperation: ast.BaseOperation{
						Response: &ast.Response{
							Responses: []*ast.SubResponse{
								{
									Content: []*ast.ResponseBodyContent{
										{
											ContentType: "application/json",
										},
										{
											ContentType: "application/*",
										},
										{
											ContentType: "application/yaml",
										},
										{
											ContentType: "*/*",
										},
										{
											ContentType: "application/csv",
										},
									},
								},
							},
						},
					},
				},
			},
			want: []string{
				"application/json;q=1", "application/yaml;q=0.8", "application/csv;q=0.6", "application/*;q=0.4", "*/*;q=0",
			},
		},
		{
			name: "Returns accept types with double digit quality values",
			args: args{
				op: ast.Operation{
					BaseOperation: ast.BaseOperation{
						Response: &ast.Response{
							Responses: []*ast.SubResponse{
								{
									Content: []*ast.ResponseBodyContent{
										{
											ContentType: "application/json",
										},
										{
											ContentType: "application/*",
										},
										{
											ContentType: "application/yaml",
										},
										{
											ContentType: "*/*",
										},
										{
											ContentType: "application/csv",
										},
									},
								},
								{
									Content: []*ast.ResponseBodyContent{
										{
											ContentType: "text/json; charset=utf-8",
										},
										{
											ContentType: "text/*",
										},
										{
											ContentType: "text/plain",
										},
										{
											ContentType: "text/csv",
										},
										{
											ContentType: "application/csv; charset=utf-8",
										},
										{
											ContentType: "multipart/form-data",
										},
										{
											ContentType: "text/json",
										},
									},
								},
							},
						},
					},
				},
			},
			want: []string{
				"text/json; charset=utf-8;q=1",
				"application/json;q=0.92",
				"text/json;q=0.83",
				"application/yaml;q=0.75",
				"application/csv; charset=utf-8;q=0.67",
				"application/csv;q=0.58",
				"text/csv;q=0.50",
				"text/plain;q=0.42",
				"multipart/form-data;q=0.33",
				"text/*;q=0.25",
				"application/*;q=0.17",
				"*/*;q=0",
			},
		},
		{
			name: "Only considers content types for success responses",
			args: args{
				op: ast.Operation{
					BaseOperation: ast.BaseOperation{
						Response: &ast.Response{
							Responses: []*ast.SubResponse{
								{
									Content: []*ast.ResponseBodyContent{
										{
											ContentType: "text/event-stream",
										},
									},
								},
								{
									Error: true,
									Content: []*ast.ResponseBodyContent{
										{
											ContentType: "application/json",
										},
										{
											ContentType: "application/custom+json",
										},
									},
								},
								{
									Error: true,
									Content: []*ast.ResponseBodyContent{
										{
											ContentType: "text/plain",
										},
									},
								},
							},
						},
					},
				},
			},
			want: []string{"text/event-stream"},
		},
		{
			name: "Only considers error content types if no success types defined",
			args: args{
				op: ast.Operation{
					BaseOperation: ast.BaseOperation{
						Response: &ast.Response{
							Responses: []*ast.SubResponse{
								{
									Error: true,
									Content: []*ast.ResponseBodyContent{
										{
											ContentType: "application/json",
										},
										{
											ContentType: "application/custom+json",
										},
									},
								},
								{
									Error: true,
									Content: []*ast.ResponseBodyContent{
										{
											ContentType: "text/plain",
										},
									},
								},
							},
						},
					},
				},
			},
			want: []string{"application/custom+json;q=1", "application/json;q=0.7", "text/plain;q=0"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.args.op.GetAcceptTypes()
			assert.Equal(t, tt.want, got)
		})
	}
}
