# Escalation

Escalation configuration with string-encoded integer expiration


## Fields

| Field                                                          | Setter Type                                                    | Getter Type                                                    | Required                                                       | Description                                                    | Example                                                        |
| -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- |
| `enabled`                                                      | *boolean*                                                      | *boolean*                                                      | :heavy_check_mark:                                             | N/A                                                            | true                                                           |
| `expiration`                                                   | *String*                                                       | *String*                                                       | :heavy_check_mark:                                             | Expiration in seconds, encoded as a string in the API response | 86400                                                          |
| `fallbackApprovers`                                            | List\<*String*>                                                | List\<*String*>                                                | :heavy_check_mark:                                             | N/A                                                            | [<br/>"admin-1"<br/>]                                          |