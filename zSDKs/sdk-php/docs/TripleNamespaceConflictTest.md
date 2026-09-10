# TripleNamespaceConflictTest

A model that references Pet from foo, bar, and baz namespaces


## Fields

| Field                                                       | Type                                                        | Required                                                    | Description                                                 |
| ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- |
| `fooPet`                                                    | [\OpenAPI\OpenAPI\Foo\Pet](../foo/Pet.md)                   | :heavy_check_mark:                                          | A pet in the foo namespace                                  |
| `barPet`                                                    | [\OpenAPI\OpenAPI\Bar\Pet](../bar/Pet.md)                   | :heavy_check_mark:                                          | A pet in the bar namespace                                  |
| `bazPet`                                                    | [\OpenAPI\OpenAPI\Baz\Pet](../baz/Pet.md)                   | :heavy_check_mark:                                          | A pet in the baz namespace with completely different schema |