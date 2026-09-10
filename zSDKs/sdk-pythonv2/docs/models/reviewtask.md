# ReviewTask

A task object with nested policy


## Fields

| Field                                      | Type                                       | Required                                   | Description                                | Example                                    |
| ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ | ------------------------------------------ |
| `id`                                       | *str*                                      | :heavy_check_mark:                         | N/A                                        | task-123                                   |
| `name`                                     | *str*                                      | :heavy_check_mark:                         | N/A                                        | Review Task                                |
| `policy`                                   | [models.MCPPolicy](../models/mcppolicy.md) | :heavy_check_mark:                         | A policy object with steps                 |                                            |