# NamespaceConflictTest

A model that references both foo.Pet and bar.Pet to test import aliasing


## Fields

| Field                                           | Type                                            | Required                                        | Description                                     |
| ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- | ----------------------------------------------- |
| `foo_pet`                                       | [Components::Foo::Pet](../models/shared/pet.md) | :heavy_check_mark:                              | A pet in the foo namespace                      |
| `bar_pet`                                       | [Components::Bar::Pet](../models/shared/pet.md) | :heavy_check_mark:                              | A pet in the bar namespace                      |