# NamespaceConflictTest

A model that references both foo.Pet and bar.Pet to test import aliasing


## Fields

| Field                                     | Type                                      | Required                                  | Description                               |
| ----------------------------------------- | ----------------------------------------- | ----------------------------------------- | ----------------------------------------- |
| `fooPet`                                  | [\OpenAPI\OpenAPI\Foo\Pet](../foo/Pet.md) | :heavy_check_mark:                        | A pet in the foo namespace                |
| `barPet`                                  | [\OpenAPI\OpenAPI\Bar\Pet](../bar/Pet.md) | :heavy_check_mark:                        | A pet in the bar namespace                |