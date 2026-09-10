# ApprovalConfig

Approval configuration with escalation


## Fields

| Field                                                           | Type                                                            | Required                                                        | Description                                                     | Example                                                         |
| --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- | --------------------------------------------------------------- |
| `approvers`                                                     | List[*str*]                                                     | :heavy_check_mark:                                              | N/A                                                             | [<br/>"user-1",<br/>"user-2"<br/>]                              |
| `escalation`                                                    | [models.Escalation](../models/escalation.md)                    | :heavy_check_mark:                                              | Escalation configuration with string-encoded integer expiration |                                                                 |