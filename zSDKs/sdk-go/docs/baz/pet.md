# Pet

A pet in the baz namespace with completely different schema


## Fields

| Field                                      | Type                                       | Required                                   | Description                                | Example                                    |
| ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| `PetID`                                    | `int64`                                    | :heavy_check_mark:                         | N/A                                        | 789                                        |
| `PetName`                                  | `string`                                   | :heavy_check_mark:                         | N/A                                        | Rex                                        |
| `AdoptedAt`                                | [*time.Time](https://pkg.go.dev/time#Time) | :heavy_minus_sign:                         | N/A                                        | 2024-01-15T10:30:00Z                       |