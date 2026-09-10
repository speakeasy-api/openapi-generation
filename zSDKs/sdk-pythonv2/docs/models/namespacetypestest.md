# NamespaceTypesTest

A model that references enums and unions from different namespaces


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     | Example                                                         |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `foo_species`                                                   | [foo.PetSpecies](../models/foo/petspecies.md)                   | :heavy_check_mark:                                              | Species of a pet in the foo namespace                           | cat                                                             |
| `bar_status`                                                    | [bar.PetStatus](../models/bar/petstatus.md)                     | :heavy_check_mark:                                              | Status of a pet in the bar namespace                            | available                                                       |
| `foo_animal`                                                    | [foo.Animal](../models/foo/animal.md)                           | :heavy_check_mark:                                              | A discriminated union of animal types in the foo namespace      |                                                                 |
| `bar_vehicle`                                                   | [bar.Vehicle](../models/bar/vehicle.md)                         | :heavy_check_mark:                                              | A non-discriminated union of vehicle types in the bar namespace |                                                                 |
| `foo_org`                                                       | [foo.Organization](../models/foo/organization.md)               | :heavy_check_mark:                                              | An organization with nested inline schemas in the foo namespace |                                                                 |