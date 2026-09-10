# TripleNamespaceConflictTest

A model that references Pet from foo, bar, and baz namespaces


## Fields

| Field                                                       | Type                                                        | Required                                                    | Description                                                 |
| ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- |
| `FooPet`                                                    | [foo.Pet](./foo/pet.md)                                     | :heavy_check_mark:                                          | A pet in the foo namespace                                  |
| `BarPet`                                                    | [bar.Pet](./bar/pet.md)                                     | :heavy_check_mark:                                          | A pet in the bar namespace                                  |
| `BazPet`                                                    | [baz.Pet](./baz/pet.md)                                     | :heavy_check_mark:                                          | A pet in the baz namespace with completely different schema |