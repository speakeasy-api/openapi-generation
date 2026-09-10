# PetOwner

A pet owner in the bar namespace


## Fields

| Field                                                         | Type                                                          | Required                                                      | Description                                                   | Example                                                       |
| ------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------- | ------------------------------------------------------------- |
| `owner_id`                                                    | *::Integer*                                                   | :heavy_check_mark:                                            | N/A                                                           | 456                                                           |
| `full_name`                                                   | *::String*                                                    | :heavy_check_mark:                                            | N/A                                                           | Bob Smith                                                     |
| `pet`                                                         | [T.nilable(Components::Bar::Pet)](../../models/shared/pet.md) | :heavy_minus_sign:                                            | A pet in the bar namespace                                    |                                                               |