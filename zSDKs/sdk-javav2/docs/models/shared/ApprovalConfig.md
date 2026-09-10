# ApprovalConfig

Approval configuration with escalation


## Fields

| Field                                                           | Setter Type                                                     | Getter Type                                                     | Required                                                        | Description                                                     | Example                                                         |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `approvers`                                                     | List\<*String*>                                                 | List\<*String*>                                                 | :heavy_check_mark:                                              | N/A                                                             | [<br/>"user-1",<br/>"user-2"<br/>]                              |
| `escalation`                                                    | [Escalation](../../models/shared/Escalation.md)                 | [Escalation](../../models/shared/Escalation.md)                 | :heavy_check_mark:                                              | Escalation configuration with string-encoded integer expiration |                                                                 |