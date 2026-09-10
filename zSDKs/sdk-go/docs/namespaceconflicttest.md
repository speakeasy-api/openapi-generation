# NamespaceConflictTest

A model that references both foo.Pet and bar.Pet to test import aliasing


## Fields

| Field                      | Type                       | Required                   | Description                |
| -------------------------- | -------------------------- | -------------------------- | -------------------------- |
| `FooPet`                   | [foo.Pet](./foo/pet.md)    | :heavy_check_mark:         | A pet in the foo namespace |
| `BarPet`                   | [bar.Pet](./bar/pet.md)    | :heavy_check_mark:         | A pet in the bar namespace |