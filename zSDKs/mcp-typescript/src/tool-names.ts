// Auto-generated at build time
export const toolNames: Array<{ name: string; description: string }>= [
  {
    "name": "operation-with-leading-and-trailing-underscores",
    "description": ""
  },
  {
    "name": "post-file",
    "description": "Post File\n\nThis is a test endpoint.\nIt has a description."
  },
  {
    "name": "get-polymorphism",
    "description": ""
  },
  {
    "name": "get-union-errors",
    "description": ""
  },
  {
    "name": "get-request-body-flattened-away",
    "description": ""
  },
  {
    "name": "get-fully-flattened-request",
    "description": ""
  },
  {
    "name": "create-with-union",
    "description": "Create with discriminated union request body\n\nTest CLI generation for discriminated unions with dot-notation flags"
  },
  {
    "name": "test-endpoint",
    "description": ""
  },
  {
    "name": "create-user",
    "description": "Create User\n\nCreates a new user in the system. Multiple named examples demonstrate\ndifferent pairing scenarios for documentation generation.\n"
  },
  {
    "name": "get-user",
    "description": "Get User"
  },
  {
    "name": "update-user",
    "description": "Update User"
  },
  {
    "name": "delete-user",
    "description": "Delete User"
  },
  {
    "name": "login",
    "description": "Login"
  },
  {
    "name": "validate",
    "description": "Validate"
  },
  {
    "name": "get-binary-default-response",
    "description": ""
  },
  {
    "name": "test-enum-formats",
    "description": "Test x-speakeasy-enums in different formats\n\nThis endpoint tests the x-speakeasy-enums extension in both array and map formats,\nincluding partial map coverage and both string and integer enum types."
  },
  {
    "name": "binary-and-string-upload",
    "description": ""
  },
  {
    "name": "get-error-in-union",
    "description": ""
  },
  {
    "name": "get-duplicate-export-collision",
    "description": "Tests that a spec-defined error type colliding with a built-in SDK error name does not cause TS2308"
  },
  {
    "name": "get-named-primitive-union",
    "description": "Test named primitive union options using title and x-speakeasy-name-override"
  },
  {
    "name": "get-empty-object-error",
    "description": "Get Empty Object Error\n\nThis endpoint tests the behavior when an error response has an empty object schema."
  },
  {
    "name": "url-validation-stress-test",
    "description": ""
  },
  {
    "name": "parentheses-in-path-allowed",
    "description": "A string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n\n\nA string with {{ double braces }} and { single braces }\nand \\{\\{ escaped curlies \\}\\} and `backticks`.\nand \\`escaped backticks\\` and double slashes\\\\\nand 'single quotes' and \"double quotes\".\nand  \\'escaped single quotes\\' and \\\"escaped double quotes\\\".\n"
  },
  {
    "name": "get-nested-integer-string",
    "description": "Test nested struct with integer:string tag\n\nThis endpoint tests the behavior when a deeply nested struct contains\nan integer field that should be unmarshaled from a string."
  },
  {
    "name": "get-error-only-example",
    "description": "Operation with example only on error response\n\nThis endpoint tests that when an operation has a named example only on\nan error response (not on the success response), we still generate\na default example for the success response."
  },
  {
    "name": "tag1-deprecated1",
    "description": "Deprecated Operation"
  },
  {
    "name": "tag1-list-test1",
    "description": "Get Test1\n\nThis is a {{test}} endpoint.\nIt has a description."
  },
  {
    "name": "tag1-post-file-with-encoding",
    "description": "Post File With Encoding\n\nThis endpoint tests the encoding field with multipart/form-data content type.\nAccording to OpenAPI 3.0.3 spec, the encoding field is valid for both\napplication/x-www-form-urlencoded and multipart/* media types.\n\nThis test includes multiple content types for the file field to verify\nhandling of comma-separated content types in encoding."
  },
  {
    "name": "test-group-tag2-post-test",
    "description": "Post Test2\n\nThis is a test endpoint.\nIt has a description."
  },
  {
    "name": "group-root-group-op",
    "description": "An operation at the group's root level\n\n'group' differs from 'TestGroup' in that it not only contains subgroups,\nbut also an operation.\n"
  },
  {
    "name": "group-sub-group-sub-group-op",
    "description": "An operation at the group's top level"
  },
  {
    "name": "group-sub-group-empty-tail-nested-group-op",
    "description": "An operation at the group's deepest level\n\nNotice that 'group.flattened' has no operations.\n"
  },
  {
    "name": "namespace-tests-conflicts-get-namespace-conflict",
    "description": "Get Namespace Conflict Test\n\nThis endpoint tests the x-speakeasy-model-namespace extension by returning\na model that references two different Pet types from different namespaces.\nThe SDK should properly import and alias both Pet types."
  },
  {
    "name": "namespace-tests-conflicts-put-namespace-conflict",
    "description": "Put Property Name Conflicts Behind\n\nThis endpoint tests property name conflict resolution through\nx-speakeasy-name-override and x-speakeasy-model-namespace extensions."
  },
  {
    "name": "namespace-tests-conflicts-create-namespace-conflict",
    "description": "Create Namespace Conflict Test\n\nThis endpoint tests creating with models from different namespaces.\nUses foo.Pet in the request and bar.Pet in the response."
  },
  {
    "name": "namespace-tests-conflicts-get-triple-namespace-e55",
    "description": "Get Triple Namespace Conflict Test\n\nThis endpoint tests the x-speakeasy-model-namespace extension by returning\na model that references three different Pet types from different namespaces.\nThe SDK should properly import and alias all three Pet types."
  },
  {
    "name": "namespace-tests-conflicts-get-pet-owners",
    "description": "Get Pet Owners\n\nThis endpoint tests using PetOwner models from different namespaces.\nReturns both foo.PetOwner and bar.PetOwner in the response."
  },
  {
    "name": "namespace-tests-single-foo-get-single-namespace-ca2",
    "description": "Get Single Namespace Foo Pet\n\nThis endpoint tests using a single component from the foo namespace.\nNo import aliasing should be needed since there's no conflict within this group."
  },
  {
    "name": "namespace-tests-single-foo-create-single-452",
    "description": "Create Single Namespace Foo Pet\n\nThis endpoint tests creating a component in the foo namespace.\nNo import aliasing should be needed since there's no conflict within this group."
  },
  {
    "name": "namespace-tests-single-bar-get-single-namespace-254",
    "description": "Get Single Namespace Bar Pet\n\nThis endpoint tests using a single component from the bar namespace.\nNo import aliasing should be needed since there's no conflict within this group."
  },
  {
    "name": "namespace-tests-types-get-namespace-types",
    "description": "Get Namespace Types Test\n\nThis endpoint tests x-speakeasy-model-namespace with enums, discriminated unions,\nnon-discriminated unions, and models with nested inline schemas."
  },
  {
    "name": "namespace-tests-types-get-namespace-animal",
    "description": "Get Namespace Animal (Discriminated Union)\n\nThis endpoint tests a discriminated union type in a custom namespace.\nThe response can be either a foo.Dog or foo.Cat."
  },
  {
    "name": "namespace-tests-types-get-namespace-vehicle",
    "description": "Get Namespace Vehicle (Non-Discriminated Union)\n\nThis endpoint tests a non-discriminated union type in a custom namespace.\nThe response can be either a bar.Car or bar.Bike."
  },
  {
    "name": "namespace-tests-types-get-namespace-organization",
    "description": "Get Namespace Organization (Nested Inline Schemas)\n\nThis endpoint tests nested inline object schemas in a custom namespace.\nThe organization model contains nested address and department types that\nshould inherit the foo namespace."
  }
];
