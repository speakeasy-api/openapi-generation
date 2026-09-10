# NamespaceTypesTest

A model that references enums and unions from different namespaces


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     | Example                                                         |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `FooSpecies`                                                    | [PetSpecies](../Foo/PetSpecies.md)                              | :heavy_check_mark:                                              | Species of a pet in the foo namespace                           | cat                                                             |
| `BarStatus`                                                     | [PetStatus](../Bar/PetStatus.md)                                | :heavy_check_mark:                                              | Status of a pet in the bar namespace                            | available                                                       |
| `FooAnimal`                                                     | [Animal](../Foo/Animal.md)                                      | :heavy_check_mark:                                              | A discriminated union of animal types in the foo namespace      |                                                                 |
| `BarVehicle`                                                    | [Vehicle](../Bar/Vehicle.md)                                    | :heavy_check_mark:                                              | A non-discriminated union of vehicle types in the bar namespace |                                                                 |
| `FooOrg`                                                        | [Organization](../Foo/Organization.md)                          | :heavy_check_mark:                                              | An organization with nested inline schemas in the foo namespace |                                                                 |