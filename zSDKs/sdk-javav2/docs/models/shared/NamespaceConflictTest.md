# NamespaceConflictTest

A model that references both foo.Pet and bar.Pet to test import aliasing


## Fields

| Field                             | Setter Type                       | Getter Type                       | Required                          | Description                       |
| --------------------------------- | --------------------------------- | --------------------------------- | --------------------------------- | --------------------------------- |
| `fooPet`                          | [Pet](../../models/shared/Pet.md) | [Pet](../../models/shared/Pet.md) | :heavy_check_mark:                | A pet in the foo namespace        |
| `barPet`                          | [Pet](../../models/shared/Pet.md) | [Pet](../../models/shared/Pet.md) | :heavy_check_mark:                | A pet in the bar namespace        |