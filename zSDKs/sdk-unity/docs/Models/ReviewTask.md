# ReviewTask

A task object with nested policy


## Fields

| Field                               | Type                                | Required                            | Description                         | Example                             |
| ----------------------------------- | ----------------------------------- | ----------------------------------- | ----------------------------------- | ----------------------------------- |
| `Id`                                | *string*                            | :heavy_check_mark:                  | N/A                                 | task-123                            |
| `Name`                              | *string*                            | :heavy_check_mark:                  | N/A                                 | Review Task                         |
| `Policy`                            | [MCPPolicy](../Models/MCPPolicy.md) | :heavy_check_mark:                  | A policy object with steps          |                                     |