# PetOwner

A pet owner in the bar namespace


## Fields

| Field                                        | Setter Type                                  | Getter Type                                  | Required                                     | Description                                  | Example                                      |
| -------------------------------------------- | -------------------------------------------- | -------------------------------------------- | -------------------------------------------- | -------------------------------------------- | -------------------------------------------- |
| `ownerId`                                    | *long*                                       | *long*                                       | :heavy_check_mark:                           | N/A                                          | 456                                          |
| `fullName`                                   | *String*                                     | *String*                                     | :heavy_check_mark:                           | N/A                                          | Bob Smith                                    |
| `pet`                                        | @Nullable [Pet](../../models/shared/Pet.md)  | Optional\<[Pet](../../models/shared/Pet.md)> | :heavy_minus_sign:                           | A pet in the bar namespace                   |                                              |