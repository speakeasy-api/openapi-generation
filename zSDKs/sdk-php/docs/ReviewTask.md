# ReviewTask

A task object with nested policy


## Fields

| Field                       | Type                        | Required                    | Description                 | Example                     |
| --------------------------- | --------------------------- | --------------------------- | --------------------------- | --------------------------- |
| `id`                        | *string*                    | :heavy_check_mark:          | N/A                         | task-123                    |
| `name`                      | *string*                    | :heavy_check_mark:          | N/A                         | Review Task                 |
| `policy`                    | [MCPPolicy](./MCPPolicy.md) | :heavy_check_mark:          | A policy object with steps  |                             |