# TripleNamespaceConflictTest

A model that references Pet from foo, bar, and baz namespaces


## Fields

| Field                                                       | Type                                                        | Required                                                    | Description                                                 |
| ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- |
| `foo_pet`                                                   | [foo.Pet](../models/foo/pet.md)                             | :heavy_check_mark:                                          | A pet in the foo namespace                                  |
| `bar_pet`                                                   | [bar.Pet](../models/bar/pet.md)                             | :heavy_check_mark:                                          | A pet in the bar namespace                                  |
| `baz_pet`                                                   | [baz.Pet](../models/baz/pet.md)                             | :heavy_check_mark:                                          | A pet in the baz namespace with completely different schema |