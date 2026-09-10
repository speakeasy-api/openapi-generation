# Escalation

Escalation configuration with string-encoded integer expiration


## Fields

| Field                                                          | Type                                                           | Required                                                       | Description                                                    | Example                                                        |
| -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- |
| `Enabled`                                                      | `bool`                                                         | :heavy_check_mark:                                             | N/A                                                            | true                                                           |
| `Expiration`                                                   | `int64`                                                        | :heavy_check_mark:                                             | Expiration in seconds, encoded as a string in the API response | 86400                                                          |
| `FallbackApprovers`                                            | []`string`                                                     | :heavy_check_mark:                                             | N/A                                                            | [<br/>"admin-1"<br/>]                                          |