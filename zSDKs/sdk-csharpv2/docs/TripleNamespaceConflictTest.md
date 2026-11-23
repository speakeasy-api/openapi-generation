# TripleNamespaceConflictTest

A model that references Pet from foo, bar, and baz namespaces


## Fields

| Field                                                       | Type                                                        | Required                                                    | Description                                                 |
| ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- |
| `FooPet`                                                    | [Speakeasy.OpenAPI.Models.Foo.Pet](./Foo/Pet.md)            | :heavy_check_mark:                                          | A pet in the foo namespace                                  |
| `BarPet`                                                    | [Speakeasy.OpenAPI.Models.Bar.Pet](./Bar/Pet.md)            | :heavy_check_mark:                                          | A pet in the bar namespace                                  |
| `BazPet`                                                    | [Speakeasy.OpenAPI.Models.Baz.Pet](./Baz/Pet.md)            | :heavy_check_mark:                                          | A pet in the baz namespace with completely different schema |