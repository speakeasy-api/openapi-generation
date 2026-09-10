# NamespaceConflictTest

A model that references both foo.Pet and bar.Pet to test import aliasing


## Fields

| Field                                      | Type                                       | Required                                   | Description                                |
| ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| `FooPet`                                   | [Speakeasy.OpenAPI.Foo.Pet](../Foo/Pet.md) | :heavy_check_mark:                         | A pet in the foo namespace                 |
| `BarPet`                                   | [Speakeasy.OpenAPI.Bar.Pet](../Bar/Pet.md) | :heavy_check_mark:                         | A pet in the bar namespace                 |