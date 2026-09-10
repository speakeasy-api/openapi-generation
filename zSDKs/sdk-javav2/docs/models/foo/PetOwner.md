# PetOwner

A pet owner in the foo namespace


## Fields

| Field                                               | Setter Type                                         | Getter Type                                         | Required                                            | Description                                         | Example                                             |
| --------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- | --------------------------------------------------- |
| `id`                                                | *String*                                            | *String*                                            | :heavy_check_mark:                                  | N/A                                                 | owner-123                                           |
| `name`                                              | *String*                                            | *String*                                            | :heavy_check_mark:                                  | N/A                                                 | Jane Doe                                            |
| `pets`                                              | @Nullable List\<[Pet](../../models/shared/Pet.md)>  | Optional\<List\<[Pet](../../models/shared/Pet.md)>> | :heavy_minus_sign:                                  | N/A                                                 |                                                     |