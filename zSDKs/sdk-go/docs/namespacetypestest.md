# NamespaceTypesTest

A model that references enums and unions from different namespaces


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     | Example                                                         |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `FooSpecies`                                                    | [foo.PetSpecies](./foo/petspecies.md)                           | :heavy_check_mark:                                              | Species of a pet in the foo namespace                           | cat                                                             |
| `BarStatus`                                                     | [bar.PetStatus](./bar/petstatus.md)                             | :heavy_check_mark:                                              | Status of a pet in the bar namespace                            | available                                                       |
| `FooAnimal`                                                     | [foo.Animal](./foo/animal.md)                                   | :heavy_check_mark:                                              | A discriminated union of animal types in the foo namespace      |                                                                 |
| `BarVehicle`                                                    | [bar.Vehicle](./bar/vehicle.md)                                 | :heavy_check_mark:                                              | A non-discriminated union of vehicle types in the bar namespace |                                                                 |
| `FooOrg`                                                        | [foo.Organization](./foo/organization.md)                       | :heavy_check_mark:                                              | An organization with nested inline schemas in the foo namespace |                                                                 |