package generate

import (
	"fmt"
	"testing"

	"github.com/speakeasy-api/openapi-generation/v2/internal/ast"
	"github.com/speakeasy-api/openapi-generation/v2/internal/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Helper function to run a single naming test
func doTestNaming(t *testing.T, name, rawSpec string, expected []string, additionalConfig *map[string]any) {
	t.Helper()
	t.Setenv("SPEAKEASY_DEBUG", "true")
	t.Setenv("SPEAKEASY_DEBUG_NAMING", "true")

	_, astTree, err := TestResolveAST(TestResolveASTInput{
		OpenAPIContents:  []byte(rawSpec),
		AdditionalConfig: additionalConfig,
	})
	require.NoError(t, err)

	got := collectTypeNameInfo(astTree)

	gotNames := make([]string, 0)
	for _, line := range got {
		if line.NameWithDepth != "" && line.Type.Scope != ast.ScopeSDK {
			gotNames = append(gotNames, line.NameWithDepth)
		}
	}

	fmt.Printf("\n ---- Got names for %q ----\n", name)
	for i := 0; i < len(gotNames); i++ {
		fmt.Printf("%q,\n", gotNames[i])
	}

	fmt.Printf("\n\n")

	if !assert.Equal(t, expected, gotNames) {
		fmt.Printf("---- FAIL ----\n\n\n")
	} else {
		fmt.Printf("---- PASS ----\n\n\n")
	}
}

// Each test below was extracted from the table in the original code.

func TestNamingBasic(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/test": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
						"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"$ref": "#/components/schemas/Pet"
									}
								}
							}
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"Pet": {
					"type": "object",
					"properties": {
						"id": {
							"type": "string"
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"Pet (id: string)",
	}
	doTestNaming(t, "Basic", rawSpec, expected, nil)
}

func TestNamingInlineResponseWithTitle(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/test": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
						"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"title": "Pet",
										"type": "object",
										"properties": {
											"id": { "type": "string" }
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"Pet (id: string)",
	}
	doTestNaming(t, "InlineResponseWithTitle", rawSpec, expected, nil)
}

func TestNamingOneOfIndexIsPreserved(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/test": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"oneOf": [
											{ "type": "object" },
											{ "type": "object", "properties": { "id": { "type": "string" } } }
										]
									}
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"GetPetResponse (union)",
		" ResponseBody1 (empty)",
		" ResponseBody2 (id: string)",
	}
	doTestNaming(t, "OneOf: index is preserved", rawSpec, expected, nil)
}

func TestNamingOneOfTitleIsUsed(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/test": {
			"get": { "operationId": "getPet", "responses": { "200": { "description": "OK", "content": { "application/json": { 
					"schema": {
						"type": "object",
						"properties": {
							"pet": {
							"oneOf": [
								{ "type": "object", "title": "Dog" },
								{ "type": "object", "title": "Cat" }
							]
							} } } } } } } } } } }
`)
	expected := []string{
		"GetPetResponse (pet: union)",
		" Pet (union)",
		"  Dog (empty)",
		"  Cat (empty)",
	}
	doTestNaming(t, "OneOf: index is preserved", rawSpec, expected, nil)
}

func TestNamingOneOfExplicitDiscriminator(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/test": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": { "application/json": { "schema": { "$ref": "#/components/schemas/Pet" } } }
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"Pet": {
					"oneOf": [
						{ "$ref": "#/components/schemas/Cat" },
						{ "$ref": "#/components/schemas/Dog" }
					],
					"discriminator": {
						"propertyName": "petType",
						"mapping": {
							"catty": "#/components/schemas/Cat",
							"doggy": "#/components/schemas/Dog"
						}
					}
				},
				"Cat": {
					"type": "object",
					"properties": {
						"petType": {
							"type": "string",
							"enum": ["catty"]
						},
						"name": { "type": "string" },
						"meows": { "type": "boolean" }
					},
					"required": ["petType", "name", "meows"]
				},
				"Dog": {
					"type": "object",
					"properties": {
						"petType": {
							"type": "string",
							"enum": ["doggy"]
						},
						"name": { "type": "string" },
						"barks": { "type": "boolean" }
					},
					"required": ["petType", "name", "barks"]
				}
			}
		}
	}`)
	expected := []string{
		"Pet (union)",
		" Cat (petType: enum, name: string, meows: boolean)",
		"  CatPetType (enum: catty)",
		" Dog (petType: enum, name: string, barks: boolean)",
		"  DogPetType (enum: doggy)",
	}
	doTestNaming(t, "OneOf: explicit discriminator", rawSpec, expected, nil)
}

func TestNamingOneOfInferredDiscriminatorEnumInlineUnion(t *testing.T) {
	rawSpec := utils.Dedent(` {
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pets": {
				"get": {
					"summary": "List all pets",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"oneOf": [
											{
												"type": "object",
												"properties": {
													"kind": { "type": "string", "enum": ["catty"] },
													"name": { "type": "string" },
													"meows": { "type": "boolean" }
												},
												"required": ["type", "name", "meows"]
											},
											{
												"type": "object",
												"properties": {
													"kind": { "type": "string", "enum": ["doggy"] },
													"name": { "type": "string" },
													"barks": { "type": "boolean" }
												},
												"required": ["kind", "name", "barks"]
											}
										]
									}
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"GetPetsResponse (union)",
		" Catty (kind: enum, name: string, meows: boolean)",
		"  KindCatty (enum: catty)",
		" Doggy (kind: enum, name: string, barks: boolean)",
		"  KindDoggy (enum: doggy)",
	}
	doTestNaming(t, "OneOf: inferred discriminator via enum inline Union<A,B>", rawSpec, expected, nil)
}

func TestNamingOneOfWithNoRequiredConstProp(t *testing.T) {
	rawSpec := utils.Dedent(` {
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pets": {
				"get": {
					"summary": "List all pets",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"oneOf": [
											{
												"type": "object",
												"properties": {
													"kind": { "type": "string", "const": "catty" },
													"foo": { "type": "string", "const": "bar" }
												}
											},
											{
												"type": "object",
												"properties": {
													"kind": { "type": "string", "const": "doggy" },
													"foo": { "type": "string", "const": "baz" }
												}
											}
											]
										} } } } } } } } } }`)
	expected := []string{
		"GetPetsResponse (union)",
		" ResponseBody1 (kind: string, foo: string)",
		" ResponseBody2 (kind: string, foo: string)",
	}
	doTestNaming(t, "OneOf: with merged parent property", rawSpec, expected, nil)
}

func TestNamingOneOfWithRequiredConstInheritedFromParentOneOf(t *testing.T) {
	rawSpec := utils.Dedent(` {
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pets": {
			"get": {
				"summary": "List all pets",
				"responses": {
				"200": {
					"description": "OK",
					"content": {
					"application/json": {
						"schema": {
						"anyOf": [
							{
							"type": "object",
							"properties": {
								"kind": { "type": "string", "const": "catty" },
								"foo": { "type": "string", "const": "bar" }
							}
							},
							{
							"type": "object",
							"properties": {
								"kind": { "type": "string", "const": "doggy" },
								"foo": { "type": "string", "const": "baz" }
							}
							}
						],
						"required": ["kind"]
						}
					} } } } } } } }`)
	expected := []string{
		"GetPetsResponse (union)",
		" Catty (kind: string, foo: string)",
		" Doggy (kind: string, foo: string)",
	}
	doTestNaming(t, "OneOf: with required const inherited from parent oneOf", rawSpec, expected, nil)
}

func TestNamingOneOfInferredDiscriminatorViaEnumInlineArray(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pets": {
				"get": {
					"summary": "List all pets",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"type": "array",
										"items": {
											"oneOf": [
												{
													"type": "object",
													"properties": {
														"type": {
															"type": "string",
															"enum": ["catty"]
														},
														"name": {
															"type": "string"
														},
														"meows": {
															"type": "boolean"
														}
													},
													"required": ["type", "name", "meows"]
												},
												{
													"type": "object",
													"properties": {
														"type": {
															"type": "string",
															"enum": ["doggy"]
														},
														"name": {
															"type": "string"
														},
														"barks": {
															"type": "boolean"
														}
													},
													"required": ["type", "name", "barks"]
												}
											]
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		" GetPetsResponse (union)",
		"  Catty (type: enum, name: string, meows: boolean)",
		"   TypeCatty (enum: catty)",
		"  Doggy (type: enum, name: string, barks: boolean)",
		"   TypeDoggy (enum: doggy)",
	}
	doTestNaming(t, "OneOf: inferred discriminator via enum inline Array<Union<A,B>>", rawSpec, expected, nil)
}

func TestNamingOneOfInferredDiscriminatorViaConstPropInline(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/const": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"oneOf": [
											{
												"type": "object",
												"properties": {
													"kind": {
														"type": "string",
														"const": "gamma"
													},
													"name": {
														"type": "string"
													}
												},
												"required": ["kind", "name"]
											},
											{
												"type": "object",
												"properties": {
													"kind": {
														"type": "string",
														"const": "delta"
													},
													"name": {
														"type": "string"
													},
													"extra": {
														"type": "boolean"
													}
												},
												"required": ["kind", "name", "extra"]
											}
										]
									}
								}
							}
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		}
	}`)
	expected := []string{
		"GetPetResponse (union)",
		" Gamma (kind: string, name: string)",
		" Delta (kind: string, name: string, extra: boolean)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "OneOf: inferred discriminator via const prop inline", rawSpec, expected, nil)
}

func TestSingularizeArrayItems(t *testing.T) {
	rawSpec := utils.Dedent(`{ "openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pet": {
			"get": {
				"operationId": "getPet",
				"responses": {
				"200": {
					"description": "OK",
					"content": {
					"application/json": {
						"schema": {
						"type": "array",
						"items": { "type": "object" }
						}
					}
					}
				}
				}
			} } } } `)
	expected := []string{
		" GetPetResponse (empty)",
	}
	doTestNaming(t, "Singularize array items", rawSpec, expected, nil)
}

func TestSingularizeArrayItems2(t *testing.T) {
	rawSpec := utils.Dedent( /*json*/ `{ "openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pet": {
			"get": {
				"operationId": "getPet",
				"responses": {
				"200": {
					"description": "OK",
					"content": {
					"application/json": {
						"schema": {
						"type": "object",
						"properties": {
						"pets": {
							"type": "array",
							"items": { "type": "object" }
						}}}}
					}
				}
				}
			} } } } `)
	expected := []string{
		"GetPetResponse (pets: array)",
		"  Pet (empty)",
	}
	doTestNaming(t, "Singularize array items", rawSpec, expected, nil)
}

func TestNamingOneOfEnumWithAllowUnknown(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pet": { "get": { "operationId": "getPet", "responses": {
						"200": { "description": "OK", "content": { "application/json": { "schema": {
										"oneOf": [
											{
												"type": "object",
												"properties": {
													"kind": { 
														"type": "string", 
														"x-speakeasy-unknown-values": "allow",
														"enum": ["alpha"] 
													},
													"name": { "type": "string" }
												},
												"required": ["kind", "name"]
											},
											{
												"type": "object",
												"properties": {
													"kind": { 
														"type": "string", 
														"x-speakeasy-unknown-values": "allow",
														"enum": ["beta"] 
													},
													"name": { "type": "string" },
												},
												"required": ["kind", "name"]
											}
										] } } } } } } } } }`)
	expected := []string{
		"GetPetResponse (union)",
		" ResponseBody1 (kind: enum, name: string)",
		"  Kind1 (enum: alpha)",
		" ResponseBody2 (kind: enum, name: string)",
		"  Kind2 (enum: beta)",
	}
	doTestNaming(t, "OneOf: enum with allow-unknown", rawSpec, expected, nil)
}

func TestNamingOneOfInlineOptionsWithTitles(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"oneOf": [
											{
												"title": "OptionA",
												"type": "object",
												"properties": {
													"kind": {
														"type": "string",
														"enum": ["X"]
													},
													"data": {
														"type": "number"
													}
												},
												"required": ["kind", "data"]
											},
											{
												"title": "OptionB",
												"type": "object",
												"properties": {
													"kind": {
														"type": "string",
														"enum": ["Y"]
													},
													"info": {
														"type": "string"
													}
												},
												"required": ["kind", "info"]
											}
										]
									}
								}
							}
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		}
	}`)
	expected := []string{
		"GetPetResponse (union)",
		" OptionA (kind: enum, data: number)",
		"  KindX (enum: X)",
		" OptionB (kind: enum, info: string)",
		"  KindY (enum: Y)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "OneOf: inline options with \"title\"s", rawSpec, expected, nil)
}

func TestNamingOneOfDeep(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pet": {
				"post": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"type": "object",
										"properties": {
											"a": {
												"type": "object",
												"properties": {
													"b": {
														"type": "object",
														"properties": {
															"c": {
																"type": "object",
																"properties": {
																	"d": {
																		"type": "object",
																		"oneOf": [
																			{
																				"type": "object",
																				"properties": {
																					"e": {
																						"type": "object"
																					}
																				}
																			},
																			{
																				"type": "object",
																				"properties": {
																					"f": {
																						"type": "object"
																					}
																				}
																			}
																		]
																	}
																}
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		}
	}`)
	expected := []string{
		"GetPetResponse (a: class)",
		" A (b: class)",
		"  B (c: class)",
		"   C (d: union)",
		"    DUnion (union)",
		"     D1 (e: class)",
		"      E (empty)",
		"     D2 (f: class)",
		"      F (empty)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "OneOf: deep", rawSpec, expected, nil)
}

func TestNamingOneOfWithFactoredOutProperties(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [ { "url": "https://api.foo.com" } ],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"title": "PetOptions",
										"oneOf": [
											{
												"title": "PetOptionsOneOf",
												"oneOf": [
													{ "$ref": "#/components/schemas/Dog" },
													{ "$ref": "#/components/schemas/Cat" }
												],
												"description": "A union of two types."
											},
											{ "$ref": "#/components/schemas/Pet" }
										]
									}
								}
							}
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"Pet": {
					"oneOf": [
						{ "$ref": "#/components/schemas/Dog" },
						{ "$ref": "#/components/schemas/Cat" }
					],
					"properties": {
						"foo": { "type": "string" }
					}
				},
				"Dog": {
					"type": "object",
					"properties": {
						"breed": { "type": "string" },
						"barks": { "type": "boolean" }
					}
				},
				"Cat": {
					"type": "object",
					"properties": {
						"color": { "type": "string" },
						"meows": { "type": "boolean" }
					}
				}
			}
		}
	}`)
	expected := []string{
		"PetOptions (union)",
		" PetOptionsOneOf (union)",
		"  Dog (breed: string, barks: boolean)",
		"  Cat (color: string, meows: boolean)",
		" Pet (union)",
		"  PetDog (breed: string, barks: boolean, foo: string)",
		"  PetCat (color: string, meows: boolean, foo: string)",
	}
	doTestNaming(t, "OneOf: with factored out properties", rawSpec, expected, nil)
}

func TestNamingConflictTwoDeepObjectsDifferentOperations(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/shared": {
				"get": {
					"operationId": "getShared",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"type": "object",
										"properties": {
											"a": {
												"type": "object",
												"properties": {
													"b": {
														"type": "object",
														"properties": {
															"c": {
																"type": "object"
															}
														}
													}
												}
											}
										}
									}
								}
							}
						}
					}
				},
				"post": {
					"operationId": "postShared",
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"type": "object",
									"properties": {
										"a": {
											"type": "object",
											"properties": {
												"b": {
													"type": "object",
													"properties": {
														"c": {
															"type": "object"
														}
													}
												}
											}
										}
									}
								}
							}
						}
					},
					"responses": { "200": { "description": "OK" } }
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"requestBody": { "content": { "application/json": { "schema": { "type": "object" } } } },
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		}
	}`)
	expected := []string{
		"GetSharedResponse (a: class)",
		" GetSharedA (b: class)",
		"  GetSharedB (c: class)",
		"   GetSharedC (empty)",
		"PostSharedRequest (a: class)",
		" PostSharedA (b: class)",
		"  PostSharedB (c: class)",
		"   PostSharedC (empty)",
		"GetOwnerRequest (empty)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "Conflict: Two very deep objects in different operations (same JSON path)", rawSpec, expected, nil)
}

func TestNamingErrorsTwoOpsOneOpWithInlineErrors(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/foo": {
				"get": {
					"operationId": "getFoo",
					"responses": {
						"200": {
							"description": "OK",
							"content": { "application/json": { "schema": { "type": "object", "properties": { "message": { "type": "string" } } } } }
						},
						"400": {
							"description": "Error",
							"content": { "application/json": { "schema": { "type": "object", "properties": { "message": { "type": "string" } } } } }
						},
						"404": {
							"description": "Error",
							"content": { "application/json": { "schema": { "type": "object", "properties": { "message": { "type": "string" } } } } }
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		}
	}`)
	expected := []string{
		"GetFooResponse (message: string)",
		"BadRequestError (error)",
		"NotFoundError (error)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "Errors: 2 operations, 1 op with inline errors", rawSpec, expected, nil)
}

func TestNamingErrorWithEnum(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/foo": {
				"get": {
					"operationId": "getFoo",
					"responses": {
						"200": {
							"description": "OK",
							"content": { "application/json": { "schema": { "type": "object", "properties": { "message": { "type": "string" } } } } }
						},
						"400": {
							"description": "Error",
							"content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrBadRequest" } } }
						},
						"403": {
							"description": "Error",
							"content": { "application/json": { "schema": { "$ref": "#/components/schemas/ErrForbidden" } } }
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"ErrBadRequest": {
					"type": "object",
					"properties": {
						"error": {
							"type": "object",
							"properties": {
								"code": { "type": "string", "enum": ["INVALID_ARGUMENT"] }
							},
							"required": ["code"]
						}
					}
				},
				"ErrForbidden": {
					"type": "object",
					"properties": {
						"error": {
							"type": "object",
							"properties": {
								"code": {
									"type": "string",
									"enum": ["PERMISSION_DENIED"],
									"description": "A machine readable error code.",
									"example": "FORBIDDEN"
								}
							},
							"required": ["code"]
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"GetFooResponse (message: string)",
		"ErrBadRequest (error)",
		" ErrBadRequestError (code: enum)",
		"  ErrBadRequestCode (enum: INVALID_ARGUMENT)",
		"ErrForbidden (error)",
		" ErrForbiddenError (code: enum)",
		"  ErrForbiddenCode (enum: PERMISSION_DENIED)",
	}
	doTestNaming(t, "Errors: 2 operations, 1 op with inline errors", rawSpec, expected, nil)
}

func TestNamingErrorsTwoOpsWithInlineErrors(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/foo": {
				"get": {
					"operationId": "getFoo",
					"responses": {
						"400": {
							"description": "Error",
							"content": { "application/json": { "schema": { "type": "object", "properties": { "message": { "type": "string" } } } } }
						},
						"404": {
							"description": "Error",
							"content": { "application/json": { "schema": { "type": "object", "properties": { "message": { "type": "string" } } } } }
						}
					}
				},
				"post": {
					"operationId": "postFoo",
					"responses": {
						"400": {
							"description": "Error",
							"content": { "application/json": { "schema": { "type": "object", "properties": { "message": { "type": "string" } } } } }
						},
						"404": {
							"description": "Error",
							"content": { "application/json": { "schema": { "type": "object", "properties": { "message": { "type": "string" } } } } }
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"BadRequestError (error)",
		"NotFoundError (error)",
	}
	doTestNaming(t, "Errors: 2 operations, 2 ops with inline errors", rawSpec, expected, nil)
}

func TestNamingErrorsComponent(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"400": {
							"description": "Error",
							"content": { "application/json": { "schema": { "$ref": "#/components/schemas/BadRequest" } } }
						},
						"404": {
							"description": "Error",
							"content": { "application/json": { "schema": { "$ref": "#/components/schemas/NotFound" } } }
						}
					}
				},
				"post": {
					"operationId": "postPet",
					"responses": {
						"400": {
							"description": "Error",
							"content": { "application/json": { "schema": { "$ref": "#/components/schemas/BadRequest" } } }
						},
						"404": {
							"description": "Error",
							"content": { "application/json": { "schema": { "$ref": "#/components/schemas/NotFound" } } }
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"BadRequest": {
					"type": "object",
					"properties": {
						"message": { "type": "string" },
						"code": { "type": "string", "enum": ["BAD_REQUEST"] }
					 }
				},
				"NotFound": {
					"type": "object",
					"properties": {
						"message": { "type": "string" },
						"code": { "type": "string", "enum": ["NOT_FOUND"] }
					}
				}
			}
		}
	}`)
	expected := []string{
		"BadRequestError (error)",
		" BadRequestCode (enum: BAD_REQUEST)",
		"NotFoundError (error)",
		" NotFoundCode (enum: NOT_FOUND)",
	}
	doTestNaming(t, "Errors: component", rawSpec, expected, nil)
}

func TestNamingErrorsComponentsWithDupeStructure(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"400": { "description": "Error", "content": { "application/json": { "schema": {
							"$ref": "#/components/schemas/BadRequest" } } }
						},
						"404": { "description": "Error", "content": { "application/json": { "schema": {
							"$ref": "#/components/schemas/NotFound" } } }
						},
						"429": { "description": "Error", "content": { "application/json": { "schema": {
							"$ref": "#/components/schemas/GetPetRateLimitError" } } }
						}
					}
				},
				"post": {
					"operationId": "postPet",
					"responses": {
						"400": { "description": "Error", "content": { "application/json": { "schema": {
							"$ref": "#/components/schemas/BadRequest" } } }
						},
						"404": { "description": "Error", "content": { "application/json": { "schema": {
							"$ref": "#/components/schemas/NotFound" } } }
						},
						"429": { "description": "Error", "content": { "application/json": { "schema": {
							"$ref": "#/components/schemas/PostPetRateLimitError" } } }
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"GetPetRateLimitError": {
					"type": "object",
					"properties": { 
						"message": { "type": "string" },
					 }
				},
				"PostPetRateLimitError": {
					"type": "object",
					"properties": { 
						"message": { "type": "string" },
					 }
				},
				"BadRequest": {
					"type": "object",
					"properties": { 
						"message": { "type": "string" },
					 }
				},
				"NotFound": {
					"type": "object",
					"properties": {
						"message": { "type": "string" },
					}
				}
			}
		}
	}`)
	expected := []string{
		"BadRequestError (error)",
		"NotFoundError (error)",
		"GetPetRateLimitError (error)",
		"PostPetRateLimitError (error)",
	}
	doTestNaming(t, "Errors: components with dupe structure", rawSpec, expected, nil)
}

func TestNamingErrorsSameOperationTwoStatusCodes(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"400": { "description": "Error", "content": { "application/json": { "schema": {
							"$ref": "#/components/schemas/BadRequest" } } }
						},
						"404": { "description": "Error", "content": { "application/json": { "schema": {
							"type": "object",
							"properties": {
								"message": { "type": "string" }
							}}}}},
						"429": { "description": "Error", "content": { "application/json": { "schema": {
							"type": "object",
							"properties": {
								"message": { "type": "string" }
							}}}}}
					}
				},
				"post": {
					"operationId": "postPet",
					"responses": {
						"400": { "description": "Error", "content": { "application/json": { "schema": {
							"$ref": "#/components/schemas/BadRequest" } } }
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"BadRequest": {
					"type": "object",
					"properties": { 
						"message": { "type": "string" },
						"code": { "type": "string", "enum": ["BAD_REQUEST"] }
					 }
				}
			}
		}
	}`)
	expected := []string{
		"BadRequestError (error)",
		" Code (enum: BAD_REQUEST)",
		"NotFoundError (error)",
		"TooManyRequestsError (error)",
	}
	doTestNaming(t, "Errors: same operation two status codes", rawSpec, expected, nil)
}

func TestNamingModelConflictsWithSubSDK(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"x-speakeasy-group": "Pets",
					"operationId": "getPet",
					"responses": {
						"200": { "description": "OK", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Pets" } } } }
					}
				},
				"post": {
					"tags": ["Pets"],
					"operationId": "postPet",
					"responses": {
						"200": { "description": "OK", "content": { "application/json": { "schema": { "$ref": "#/components/schemas/Pets" } } } }
					}
				}
			}
		},
		"components": {
			"schemas": {
				"Pets": {
					"type": "object",
					"properties": { "pets": { "type": "string", "enum": ["dog", "cat"] } }
				}
			}
		}
	}`)

	expected := []string{
		"Pets (pets: enum)",
		" PetsEnum (enum: dog, cat)",
	}
	doTestNaming(t, "Model conflicts with sub-sdk", rawSpec, expected, nil)
}

func TestNamingInputOutput(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
			"post": {
				"operationId": "createPet",
				"requestBody": {
				"content": {
					"application/json": {
					"schema": { "$ref": "#/components/schemas/Pet" }
					}
				}
				},
				"responses": {
				"200": {
					"description": "OK",
					"content": {
					"application/json": {
						"schema": { "$ref": "#/components/schemas/Pet" }
					}
					}
				}
				}
			}
			},
			"/dog": {
			"post": {
				"operationId": "createDog",
				"requestBody": {
				"content": {
					"application/json": {
					"schema": { "$ref": "#/components/schemas/Dog" }
					}
				}
				},
				"responses": {
				"200": {
					"description": "OK",
					"content": {
					"application/json": {
						"schema": { "$ref": "#/components/schemas/Dog" }
					}
					}
				}
				}
			}
			},
			"/cat": {
			"post": {
				"operationId": "createCat",
				"requestBody": {
				"content": {
					"application/json": {
					"schema": { "$ref": "#/components/schemas/Cat" }
					}
				}
				},
				"responses": {
				"200": {
					"description": "OK",
					"content": {
					"application/json": {
						"schema": { "$ref": "#/components/schemas/Cat" }
					}
					}
				}
				}
			}
			},
			"/rabbit": {
			"post": {
				"operationId": "createRabbit",
				"requestBody": {
				"content": {
					"application/json": {
					"schema": { "$ref": "#/components/schemas/Rabbit" }
					}
				}
				},
				"responses": {
				"200": {
					"description": "OK",
					"content": {
					"application/json": {
						"schema": { "$ref": "#/components/schemas/Rabbit" }
					}
					}
				}
				}
			}
			}
		},
		"components": {
			"schemas": {
			"Pet": {
				"type": "object",
				"properties": { "name": { "type": "string", "writeOnly": true } }
			},
			"Dog": {
				"type": "object",
				"properties": {
				"name": { "type": "string" },
				"id": { "type": "string", "writeOnly": true }
				}
			},
			"Cat": {
				"type": "object",
				"properties": {
				"name": { "type": "string" },
				"id": { "type": "string", "writeOnly": true }
				}
			},
			"Rabbit": {
				"type": "object",
				"properties": {
				"name": { "type": "string" },
				"id": { "type": "string", "readOnly": true }
				}
			}
			}
		}
		}
`)

	expected := []string{
		"Pet (name: string)",
		"PetOutput (empty)",
		"Dog (name: string, id: string)",
		"DogOutput (name: string)",
		"Cat (name: string, id: string)",
		"CatOutput (name: string)",
		"RabbitInput (name: string)",
		"Rabbit (name: string, id: string)",
	}
	doTestNaming(t, "Input output", rawSpec, expected, nil)
}

func TestNamingErrorsPolymorphicSameStatusCode(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"400": { "description": "Error", "content": { "application/json": { "schema": {
							"type": "object",
							"properties": {
								"message": { "type": "string" }
							}}}}}
					}
				},
				"post": {
					"operationId": "postPet",
					"responses": {
						"400": { "description": "Error", "content": { "application/json": { "schema": {
							"type": "object",
							"properties": {
								"code": { "type": "string" }
							}}}}}
					}
				}
			}
		}
	}`)
	expected := []string{
		"GetPetBadRequestError (error)",
		"PostPetBadRequestError (error)",
	}
	doTestNaming(t, "Errors: polymorphic same status code", rawSpec, expected, nil)
}

func TestNamingErrorsPolymorphicSameStatusCode2(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"400": { "description": "Error", "content": { "application/json": { "schema": {
							"type": "object",
							"properties": {
								"message": { "type": "string" }
							}}}}},
						"429": { "description": "Error", "content": { "application/json": { "schema": {
							"type": "object",
							"properties": {
								"code": { "type": "string" }
							}}}}}
					}
				},
				"post": {
					"operationId": "postPet",
					"responses": {
						"400": { "description": "Error", "content": { "application/json": { "schema": {
							"type": "object",
							"properties": {
								"code": { "type": "string" }
							}}}}},
						"429": { "description": "Error", "content": { "application/json": { "schema": {
							"type": "object",
							"properties": {
								"message": { "type": "string" }
							}}}}}
					}
				}
			}
		}
	}`)
	expected := []string{
		"GetPetBadRequestError (error)",
		"GetPetTooManyRequestsError (error)",
		"PostPetBadRequestError (error)",
		"PostPetTooManyRequestsError (error)",
	}
	doTestNaming(t, "Errors: polymorphic same status code 2", rawSpec, expected, nil)
}

func TestNamingErrorsDeduplicationAllOf(t *testing.T) {
	rawSpec := utils.Dedent(`{
			"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [ { "url": "https://api.foo.com" } ],
			"paths": {
				"/pet": {
					"get": {
						"operationId": "getPet",
						"responses": {
							"400": {
								"description": "Error",
								"content": {
									"application/json": {
										"schema": {
											"allOf": [
												{
													"$ref": "#/components/schemas/BaseError"
												},
												{
													"properties": {
														"code": {
															"type": "string",
															"const": "GET_PET_BAD_REQUEST"
														}
													}
												} ] } } } } } },
					"post": {
						"operationId": "postPet",
						"responses": {
							"400": {
								"description": "Error",
								"content": {
									"application/json": {
										"schema": {
											"allOf": [
												{
													"$ref": "#/components/schemas/BaseError"
												},
												{
													"properties": {
														"code": {
															"type": "string",
															"const": "POST_PET_BAD_REQUEST"
														}
													}
												} ] } } } } } } }
			},
			"components": {
				"schemas": {
					"BaseError": {
						"type": "object",
						"properties": {
							"message": {
								"type": "string"
							}
						}
					}
				}
			}
		}`)
	expected := []string{
		"GetPetBadRequestError (error)",
		"PostPetBadRequestError (error)",
	}
	doTestNaming(t, "Errors: deduplication allOf", rawSpec, expected, nil)
}

func TestNamingErrorsDeepError(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": { "application/json": { "schema": { "type": "object" } } }
						},
						"400": {
							"$ref": "#/components/responses/BadRequestResponse"
						}
					}
				}
			}
		},
		"components": {
			"responses": {
				"BadRequestResponse": {
					"description": "Error",
					"content": { "application/json": { "schema": {
						"type": "object",
						"properties": {
							"detail": {
								"anyOf": [
									{ "type": "string" },
									{ "type": "object", "properties": { "missing": { "type": "string" } } }
								]
							}
						}
					} } }
				}
			}
		}
	}`)
	expected := []string{
		"GetPetResponse (empty)",
		"BadRequestResponseError (error)",
		" DetailUnion (union)",
		"  Detail (missing: string)",
	}
	doTestNaming(t, "Errors: deep error", rawSpec, expected, nil)
}

func TestNamingErrorsInlineErrorWithTitleIncludingError(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"422": {
							"description": "Too Many Requests",
							"content": {
								"application/json": {
									"schema": {
										"title": "RateLimitError",
										"type": "object",
										"properties": {
											"detail": {
												"type": "string"
											}
										}
									}
								}
							}
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		}
	}`)
	expected := []string{
		"RateLimitError (error)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "Errors: inline error with title including \"error\"", rawSpec, expected, nil)
}

func TestNamingErrorsRefErrorWithRefTitleIncludingError(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/ref-error": {
				"get": {
					"operationId": "getRefError",
					"responses": {
						"429": {
							"description": "Too Many Requests",
							"content": {
								"application/json": {
									"schema": {
										"$ref": "#/components/schemas/RefError"
									}
								}
							}
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		},
		"components": {
			"schemas": {
				"RefError": {
					"title": "RefErrorTitle",
					"type": "object",
					"properties": {
						"code": {
							"type": "integer"
						},
						"message": {
							"type": "string"
						}
					},
					"required": ["code", "message"]
				}
			}
		}
	}`)
	expected := []string{
		"RefError (error)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "Errors: ref error with ref title including \"error\"", rawSpec, expected, nil)
}

func TestNamingErrorsInlineErrorWithoutTitle(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": { "schema": { "type": "object", "properties": { "info": { "type": "string" } } } }
							}
						},
						"429": {
							"description": "Too Many Requests",
							"content": {
								"application/json": {
									"schema": {
										"type": "object",
										"properties": {
											"info": {
												"type": "string"
											}
										}
									}
								}
							}
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		}
	}`)
	expected := []string{
		"GetPetResponse (info: string)",
		"TooManyRequestsError (error)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "Errors: inline error without title", rawSpec, expected, nil)
}

func TestNamingErrorsRefErrorWithoutRefTitleIncludingError(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"500": {
							"description": "Internal Server Error",
							"content": {
								"application/json": {
									"schema": {
										"$ref": "#/components/schemas/GenericError"
									}
								}
							}
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		},
		"components": {
			"schemas": {
				"GenericError": {
					"type": "object",
					"properties": {
						"errorDetail": {
							"type": "string"
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"GenericError (error)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "Errors: ref error without ref title including \"error\"", rawSpec, expected, nil)
}

func TestNamingOperationIDRequestWithJSONAndFormURLEncoded(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"post": {
					"operationId": "createPet",
					"requestBody": {
						"content": {
							"application/json": {
								"schema": {
									"type": "object",
									"properties": {
										"jsonField": {
											"type": "string"
										}
									}
								}
							},
							"application/x-www-form-urlencoded": {
								"schema": {
									"type": "object",
									"properties": {
										"formField": {
											"type": "integer"
										}
									}
								}
							}
						}
					},
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"type": "object",
										"properties": {
											"result": {
												"type": "string"
											}
										}
									}
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"CreatePetRequest (jsonField: string)",
		"CreatePetResponse (result: string)",
		"CreatePetFormRequest (formField: integer)",
		"CreatePetFormResponse (result: string)",
	}
	doTestNaming(t, "OperationID: Request with JSON and Form URL encoded options", rawSpec, expected, nil)
}

func TestNamingMultipleResponses(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": { "type": "object", "properties": { "data": { "type": "string" } } }
								},
								"text/json": { "schema": { "type": "object" } }
							}
						},
						"2XX": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": { "type": "object", "properties": { "foo": { "type": "string" } } }
								}
							}
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		}
	}`)
	expected := []string{
		"GetPetResponse (union)",
		" GetPetResponseBody1 (data: string)",
		" GetPetResponseBody2 (empty)",
		" GetPetResponseBody3 (foo: string)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "Multiple responses", rawSpec, expected, nil)
}

func TestNamingWithNestedDuplicateProperties(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
			"get": {
				"operationId": "getPet",
				"responses": {
				"2XX": {
					"description": "OK",
					"content": {
					"application/json": {
						"schema": {
						"type": "object",
						"properties": {
							"foo": {
							"type": "object",
							"properties": {
								"foo": {
								"type": "object",
								"properties": {
									"foo": {
									"type": "object",
									"properties": {
										"foo": { "type": "object" }
									}
									} } } } } } } } } } } } } } }

			`)
	expected := []string{
		"GetPetResponse (foo: class)",
		" Foo (foo: class)",
		"  FooFoo (foo: class)",
		"   FooFooFoo (foo: class)",
		"    FooFooFooFoo (empty)",
	}
	doTestNaming(t, "Multiple responses", rawSpec, expected, nil)
}

func TestNamingWithRefConflict(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
			"get": {
				"operationId": "getPet",
				"responses": {
				"2XX": {
					"description": "OK",
					"content": {
					"application/json": {
						"schema": {
						"oneOf": [
							{
							"$ref": "#/components/schemas/Pet"
							},
							{
							"$ref": "#/components/schemas/Dog"
							}
						]
						}
					}
					}
				}
				}
			}
			}
		},
		"components": {
			"schemas": {
			"Pet": {
				"type": "object",
				"properties": { 
					"petColor": { "type": "string", "enum": ["brown", "black"] },
					"foo": {
						"oneOf": [
							{ "type": "string", "title": "Hey", "enum": ["bar", "baz"] },
							{ "type": "string", "enum": ["qux", "quux"] }
						]
					}
				}
			},
			"Dog": {
				"type": "object",
				"properties": {
				"dogColor": { "$ref": "#/components/schemas/Pet/properties/petColor" },
				"asdf": { "$ref": "#/components/schemas/Pet/properties/foo/oneOf/0" },
				"asdf2": { "$ref": "#/components/schemas/Pet/properties/foo/oneOf/1" },
				}
			}
			}
		}
		}

			`)
	expected := []string{
		// TODO: These namings are totally undesirable. It's caused by how we interpret
		// 		 ref name and ref type inaccurately when we reset the context stack for
		//       components. Also when we see a $ref it's tricky to construct the
		//       context stack accurately. I decided to abandon fixing this at this time
		//       but I got pretty far:
		//       The parser context should be revisited before changing these expectations.
		"GetPetResponse (union)",
		" Pet (petColor: enum, foo: union)",
		"  PetPetColor (enum: brown, black)",
		"  Foo (union)",
		"   Hey (enum: bar, baz)",
		"   FooEnum (enum: qux, quux)",
		" Dog (dogColor: petColor_Properties, asdf: 0, asdf2: 1)",
		"  PetColorProperties (enum: brown, black)",
		"  0 (enum: bar, baz)",
		"  1 (enum: qux, quux)",
	}
	doTestNaming(t, "Naming: ref conflict", rawSpec, expected, nil)
}

func TestNamingOperationIDWithXSpeakeasyNameOverrideAndGroup(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"paths": {
			"/pet": {
				"post": {
					"operationId": "createPet",
					"x-speakeasy-name-override": "make",
					"x-speakeasy-group": "animals",
					"requestBody": {
						"content": {
							"application/json": {
								"schema": { "type": "object", "properties": { "payload": { "type": "string" } } }
							}
						}
					},
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": { "type": "object", "properties": { "success": { "type": "boolean" } } }
								}
							}
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"x-speakeasy-name-override": "get",
					"x-speakeasy-group": "owners",
					"responses": { "200": { "description": "OK", "content": { "application/json": { "schema": { "type": "object" } } } } }
				}
			}
		}
	}`)
	expected := []string{
		"CreatePetRequest (payload: string)",
		"CreatePetResponse (success: boolean)",
		"GetOwnerResponse (empty)",
	}
	doTestNaming(t, "OperationID: with x-speakeasy-name-override and x-speakeasy-group", rawSpec, expected, nil)
}

func TestNamingConflictBetweenOperations(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": { "type": "object", "properties": { "value": { "type": "string" } } }
								}
							}
						}
					}
				}
			},
			"/owner": {
				"get": {
					"operationId": "getOwner",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": { "type": "object", "properties": { "value": { "type": "number" } } }
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"GetPetResponse (value: string)",
		"GetOwnerResponse (value: number)",
	}
	doTestNaming(t, "Conflict: between operations", rawSpec, expected, nil)
}

func TestNamingResponseFormatEnvelope(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": { "application/json": { "schema": { "type": "object", "properties": { "value": { "type": "string" } } } } }
						}
					}
				}
			}
		}
	}`)
	additionalConfig := &map[string]any{"responseFormat": "envelope"}
	expected := []string{
		"GetPetResponse (ContentType: string, StatusCode: int32, RawResponse: response ...)",
		" GetPetResponseBody (value: string)",
	}
	doTestNaming(t, "responseFormat=envelope", rawSpec, expected, additionalConfig)
}

func TestNamingResponseFormatEnvelopeHTTP(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": { "application/json": { "schema": { "type": "object", "properties": { "value": { "type": "string" } } } } }
						}
					}
				}
			}
		}
	}`)
	additionalConfig := &map[string]any{"responseFormat": "envelope-http"}
	expected := []string{
		"GetPetResponse (HttpMeta: HTTPMetadata, object: class)",
		" HttpMetadata (Response: response, Request: request)",
		" GetPetResponseBody (value: string)",
	}
	doTestNaming(t, "responseFormat=envelope-http", rawSpec, expected, additionalConfig)
}

func TestNamingRequestJustBody(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [ { "url": "https://api.foo.com" } ],
		"paths": {
			"/pet": {
				"post": {
					"operationId": "createPet",
					"requestBody": {
						"content": {
							"application/json": {
								"schema": { "type": "object", "properties": { "value": { "type": "string" } } }
							}
						}
					},
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": { "type": "object", "properties": { "value": { "type": "string" } } }
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"CreatePetRequest (value: string)",
		"CreatePetResponse (value: string)",
	}
	doTestNaming(t, "Params: just body", rawSpec, expected, nil)
}

func TestNamingRequestJustBodyWithTitle(t *testing.T) {
	rawSpec := utils.Dedent(`{
        "openapi":"3.1.0", "info":{ "title":"Test API", "version":"1.0.0" }, "servers":[ { "url":"https://api.foo.com" } ],
        "paths":{
            "/pet":{
                "get":{
                    "operationId":"getPet",
                    "responses":{
                    "200":{
                        "description":"OK",
                        "content":{
                            "application/json":{
                                "schema":{ "title":"Pet", "type":"object", "properties":{ "value":{ "type":"string" } } }
                            }
                        }
                    }
                    }
                }
            }
        }
        }`)
	expected := []string{
		"Pet (value: string)",
	}
	doTestNaming(t, "Request with title", rawSpec, expected, nil)
}

func TestNamingRequestJustHeaders(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [ { "url": "https://api.foo.com" } ],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"parameters": [
						{ "name": "X-Custom-Header", "in": "header", "required": false, "schema": { "type": "string" } }
					],
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": { "type": "object", "properties": { "value": { "type": "string" } } }
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"GetPetRequest (X-Custom-Header: string)",
		"GetPetResponse (value: string)",
	}
	doTestNaming(t, "Params: just headers", rawSpec, expected, nil)
}

func TestNamingRequestJustQuery(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [ { "url": "https://api.foo.com" } ],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"parameters": [
						{ "name": "id", "in": "query", "required": false, "schema": { "type": "string" } }
					],
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": { "type": "object", "properties": { "value": { "type": "string" } } }
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"GetPetRequest (id: string)",
		"GetPetResponse (value: string)",
	}
	doTestNaming(t, "Params: just query", rawSpec, expected, nil)
}

func TestNamingRequestJustPath(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [ { "url": "https://api.foo.com" } ],
		"paths": {
			"/pet/{id}": {
				"get": {
					"operationId": "getPet",
					"parameters": [
						{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } }
					],
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": { "type": "object", "properties": { "value": { "type": "string" } } }
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"GetPetRequest (id: string)",
		"GetPetResponse (value: string)",
	}
	doTestNaming(t, "Params: just path", rawSpec, expected, nil)
}

func TestNamingRequestHeadersQueryPathRequestBody(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [ { "url": "https://api.foo.com" } ],
		"paths": {
			"/pet/{id}": {
				"post": {
					"operationId": "createPet",
					"parameters": [
						{ "name": "id", "in": "path", "required": true, "schema": { "type": "string" } },
						{ "name": "X-Custom-Header", "in": "header", "required": false, "schema": { "type": "string" } },
						{ "name": "id", "in": "query", "required": false, "schema": { "type": "string" } }
					],
					"requestBody": {
						"content": {
							"application/json": {
								"schema": { "type": "object", "properties": { "payload": { "type": "string" } } }
							}
						}
					},
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": { "type": "object", "properties": { "value": { "type": "string" } } }
								}
							}
						}
					}
				}
			}
		}
	}`)
	expected := []string{
		"CreatePetRequest (idPathParameter: string, X-Custom-Header: string, idQueryParameter: string ...)",
		" CreatePetRequestBody (payload: string)",
		"CreatePetResponse (value: string)",
	}
	doTestNaming(t, "Params: headers,query,path,requestBody", rawSpec, expected, nil)
}

func TestNamingAllOf(t *testing.T) {
	rawSpec := utils.Dedent(`{
		"openapi": "3.1.0",
		"info": { "title": "Test API", "version": "1.0.0" },
		"servers": [{ "url": "https://api.foo.com" }],
		"paths": {
			"/pet": {
				"get": {
					"operationId": "getPet",
					"responses": {
						"200": {
							"description": "OK",
							"content": {
								"application/json": {
									"schema": {
										"allOf": [
											{ "$ref": "#/components/schemas/BasePet" },
											{ "properties": { "nickname": { "type": "string" } } }
										]
									}
								}
							}
						}
					}
				}
			}
		},
		"components": {
			"schemas": {
				"BasePet": {
					"type": "object",
					"properties": {
						"id": { "type": "integer" },
						"name": { "type": "string" }
					}
				}
			}
		}
	}`)
	expected := []string{
		"GetPetResponse (id: integer, name: string, nickname: string)",
	}
	doTestNaming(t, "AllOf", rawSpec, expected, nil)
}

func TestComponentUsedAsRequestResponseAndError(t *testing.T) {
	rawSpec := utils.Dedent(`{
                "openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [ { "url": "https://api.foo.com" } ],
                "paths": {
                    "/pet": {
                        "post": {
                            "operationId": "createPet",
                            "requestBody": { "content": { "application/json": { "schema": { 
                                "$ref": "#/components/schemas/Pet" } } }
                            },
                            "responses": {
                                "200": { "description": "OK", "content": { "application/json": { "schema": {
                                    "$ref": "#/components/schemas/Pet" } } }
                                },
                                "400": { "description": "Error", "content": { "application/json": { "schema": {
                                    "$ref": "#/components/schemas/Pet" } } }
                                }
                            }
                        }
                    }
                },
                "components": {
                    "schemas": {
                        "Pet": {
                            "type": "object",
                            "properties": { "errorCode": { "type": "string", "enum": ["INVALID_ID"] } }
                        }
                    }
                }
            }`)
	expected := []string{
		"Pet (errorCode: enum)",
		" ErrorCode (enum: INVALID_ID)",
		"PetError (error)",
	}
	doTestNaming(t, "Component used as request, response, and error", rawSpec, expected, nil)
}

func TestNamingComponentUsedAsRequestResponseAndErrorWithReadOnlyAndWriteOnlyProperties(t *testing.T) {
	rawSpec := utils.Dedent(`{
        "openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [ { "url": "https://api.foo.com" } ],
        "paths": {
            "/pet": {
                "post": {
                    "operationId": "createPet",
                    "requestBody": {
                        "content": {
                            "application/json": {
                                "schema": {
                                    "$ref": "#/components/schemas/Pet"
                                }
                            }
                        }
                    },
                    "responses": {
                        "200": {
                            "description": "A list of pets",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/Pet"
                                    }
                                }
                            }
                        },
                        "400": {
                            "description": "Invalid ID supplied",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/Pet"
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "components": {
            "schemas": {
                "Pet": {
                    "type": "object",
                    "properties": {
                        "id": {
                            "type": "integer",
                            "readOnly": true
                        },
                        "name": {
                            "type": "string"
                        },
                        "nickname": {
                            "type": "string",
                            "writeOnly": true
                        }
                    }
                }
            }
        }
    }`)
	expected := []string{
		"PetInput (name: string, nickname: string)",
		"PetOutput (id: integer, name: string)",
		"PetError (error)",
	}
	doTestNaming(t, "Component used as request, response, and error with readOnly and writeOnly properties", rawSpec, expected, nil)
}

func TestNamingErrorComponentOnlyUsedAsError(t *testing.T) {
	rawSpec := utils.Dedent(`{
        "openapi": "3.1.0", "info": { "title": "Test API", "version": "1.0.0" }, "servers": [ { "url": "https://api.foo.com" } ],
        "paths": {
            "/pet": {
                "get": {
                    "operationId": "getPet",
                    "responses": {
                        "400": {
                            "description": "Invalid ID supplied",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/PetError"
                                    }
                                }
                            }
                        }
                    }
                },
                "post": {
                    "operationId": "createPet",
                    "responses": {
                        "400": {
                            "description": "Invalid ID supplied",
                            "content": {
                                "application/json": {
                                    "schema": {
                                        "$ref": "#/components/schemas/PetError"
                                    }
                                }
                            }
                        }
                    }
                }
            }
        },
        "components": {
            "schemas": {
                "PetError": {
                    "type": "object",
					"properties": {
						"errorCode": {
							"type": "string",
							"enum": ["INVALID_ID"]
						}
					}
                }
            }
        }
    }`)
	expected := []string{
		"PetError (error)",
		" ErrorCode (enum: INVALID_ID)",
	}
	doTestNaming(t, "Component used as just error", rawSpec, expected, nil)
}

func TestNamingErrorComponentUsedAsResponseAndError(t *testing.T) {
	rawSpec := utils.Dedent(`{
          "openapi": "3.1.0",
          "info": { "title": "Test API", "version": "1.0.0" },
          "servers": [{ "url": "https://api.foo.com" }],
          "paths": {
            "/pet": {
              "get": {
                "operationId": "getPet",
                "responses": {
                  "200": {
                    "description": "OK",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/PetError"
                        }
                      }
                    }
                  },
                  "400": {
                    "description": "Error",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/PetError"
                        }
                      }
                    }
                  }
                }
              },
              "post": {
                "operationId": "createPet",
                "responses": {
                  "400": {
                    "description": "Error",
                    "content": {
                      "application/json": {
                        "schema": {
                          "$ref": "#/components/schemas/PetError"
                        }
                      }
                    }
                  }
                }
              }
            }
          },
          "components": {
            "schemas": {
              "PetError": {
                "type": "object",
                "properties": {
                  "errorCode": {
                    "type": "string",
                    "enum": ["INVALID_ID"]
                  }
                }
              }
            }
          }
        }`)
	expected := []string{
		"PetError (errorCode: enum)",
		" ErrorCode (enum: INVALID_ID)",
		"PetError (error)",
	}
	doTestNaming(t, "Component used as response and error", rawSpec, expected, nil)
}
