# ApprovalConfig

Approval configuration with escalation


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     | Example                                                         |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `Approvers`                                                     | []`string`                                                      | :heavy_check_mark:                                              | N/A                                                             | [<br/>"user-1",<br/>"user-2"<br/>]                              |
| `Escalation`                                                    | [Escalation](./escalation.md)                                   | :heavy_check_mark:                                              | Escalation configuration with string-encoded integer expiration |                                                                 |