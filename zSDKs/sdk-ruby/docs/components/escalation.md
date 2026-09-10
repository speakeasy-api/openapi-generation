# Escalation

Escalation configuration with string-encoded integer expiration


## Fields

| Field                                                          | Type                                                           | Required                                                       | Description                                                    | Example                                                        |
| -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- | -------------------------------------------------------------- |
| `enabled`                                                      | *T::Boolean*                                                   | :heavy_check_mark:                                             | N/A                                                            | true                                                           |
| `expiration`                                                   | *::String*                                                     | :heavy_check_mark:                                             | Expiration in seconds, encoded as a string in the API response | 86400                                                          |
| `fallback_approvers`                                           | T::Array<*::String*>                                           | :heavy_check_mark:                                             | N/A                                                            | [<br/>"admin-1"<br/>]                                          |