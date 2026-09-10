# ReviewTask

A task object with nested policy


## Fields

| Field                       | Type                        | Required                    | Description                 | Example                     |
| --------------------------- | --------------------------- | --------------------------- | --------------------------- | --------------------------- |
| `ID`                        | `string`                    | :heavy_check_mark:          | N/A                         | task-123                    |
| `Name`                      | `string`                    | :heavy_check_mark:          | N/A                         | Review Task                 |
| `Policy`                    | [MCPPolicy](./mcppolicy.md) | :heavy_check_mark:          | A policy object with steps  |                             |