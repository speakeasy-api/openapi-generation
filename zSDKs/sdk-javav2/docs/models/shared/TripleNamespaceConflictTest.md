# TripleNamespaceConflictTest

A model that references Pet from foo, bar, and baz namespaces


## Fields

| Field                                                       | Setter Type                                                 | Getter Type                                                 | Required                                                    | Description                                                 |
| ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- |
| `fooPet`                                                    | [Pet](../../models/shared/Pet.md)                           | [Pet](../../models/shared/Pet.md)                           | :heavy_check_mark:                                          | A pet in the foo namespace                                  |
| `barPet`                                                    | [Pet](../../models/shared/Pet.md)                           | [Pet](../../models/shared/Pet.md)                           | :heavy_check_mark:                                          | A pet in the bar namespace                                  |
| `bazPet`                                                    | [Pet](../../models/shared/Pet.md)                           | [Pet](../../models/shared/Pet.md)                           | :heavy_check_mark:                                          | A pet in the baz namespace with completely different schema |