# TripleNamespaceConflictTest

A model that references Pet from foo, bar, and baz namespaces


## Fields

| Field                                                       | Type                                                        | Required                                                    | Description                                                 |
| ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- | ----------------------------------------------------------- |
| `foo_pet`                                                   | [Components::Foo::Pet](../models/shared/pet.md)             | :heavy_check_mark:                                          | A pet in the foo namespace                                  |
| `bar_pet`                                                   | [Components::Bar::Pet](../models/shared/pet.md)             | :heavy_check_mark:                                          | A pet in the bar namespace                                  |
| `baz_pet`                                                   | [Components::Baz::Pet](../models/shared/pet.md)             | :heavy_check_mark:                                          | A pet in the baz namespace with completely different schema |