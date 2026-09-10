# Test Coverage Overview

## Global Security

| Test file        | multiple schemes | multiple options | hoisted | flattened |
|:----------------:|:----------------:|:----------------:|:-------:|:---------:|
| review           | yes              | yes              | no      | n.a.      |
| primary          | yes              | no               | no      | n.a.      |
| simple-security  | no               | no               | no      | yes       |
| hoisted-security | no               | no               | yes     | no        |
| no-security      | no               | no               | no      | n.a.      |


## Server selection

### Global

| Test file        | Named | Variables |
|:----------------:|:-----:|:---------:|
| primary          | no    | yes       |
| simple-security  | yes   | yes       |
| hoisted-security | yes   | no        |
| no-security      | no    | no        |


### Operation

| Test file   | operation               | Named | Variables |
|:-----------:|-------------------------|:-----:|:---------:|
| primary     | selectServerWithID      | yes   | no        |
| primary     | serverWithTemplates     | no    | yes       |
| primary     | serverByIDWithTemplates | yes   | yes       |
