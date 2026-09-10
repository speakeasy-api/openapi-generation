# Escalation

Escalation configuration with string-encoded integer expiration


## Fields

| Field                                                          | Type                                                           | Required                                                       | Description                                                    | Example                                                        |
| -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- |
| `enabled`                                                      | *bool*                                                         | :heavy_check_mark:                                             | N/A                                                            | true                                                           |
| `expiration`                                                   | *string*                                                       | :heavy_check_mark:                                             | Expiration in seconds, encoded as a string in the API response | 86400                                                          |
| `fallbackApprovers`                                            | array<*string*>                                                | :heavy_check_mark:                                             | N/A                                                            | [<br/>"admin-1"<br/>]                                          |