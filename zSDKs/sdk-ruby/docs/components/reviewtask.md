# ReviewTask

A task object with nested policy


## Fields

| Field                                                  | Type                                                   | Required                                               | Description                                            | Example                                                |
| ------------------------------------------------------ | ------------------------------------------------------ | ------------------------------------------------------ | ------------------------------------------------------ | ------------------------------------------------------ |
| `id`                                                   | *::String*                                             | :heavy_check_mark:                                     | N/A                                                    | task-123                                               |
| `name`                                                 | *::String*                                             | :heavy_check_mark:                                     | N/A                                                    | Review Task                                            |
| `policy`                                               | [Components::MCPPolicy](../models/shared/mcppolicy.md) | :heavy_check_mark:                                     | A policy object with steps                             |                                                        |