# ReviewTask

A task object with nested policy


## Fields

| Field                                         | Setter Type                                   | Getter Type                                   | Required                                      | Description                                   | Example                                       |
| --------------------------------------------- | --------------------------------------------- | --------------------------------------------- | --------------------------------------------- | --------------------------------------------- | --------------------------------------------- |
| `id`                                          | *String*                                      | *String*                                      | :heavy_check_mark:                            | N/A                                           | task-123                                      |
| `name`                                        | *String*                                      | *String*                                      | :heavy_check_mark:                            | N/A                                           | Review Task                                   |
| `policy`                                      | [MCPPolicy](../../models/shared/MCPPolicy.md) | [MCPPolicy](../../models/shared/MCPPolicy.md) | :heavy_check_mark:                            | A policy object with steps                    |                                               |