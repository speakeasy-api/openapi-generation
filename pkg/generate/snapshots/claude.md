# Snapshot Testing Framework

Goal:
* Easy to produce variants of spec and gen.yaml
* Fast iteration
* Easy to automatically update snapshots

## Run Tests

```bash
# Run all (in parallel)
go test -v ./pkg/generate/snapshots/

# Run specific test
go test -run TestSnapPyBase ./pkg/generate/snapshots/
```

## Create New Test

Fork one of the existing `*_base` test files and update:
1. `genYaml` - target configuration (can be partial will be merged with defaults)
2. `expectedSnapshotFiles` file paths (defines which generated files to include in the snapshot)
3. `expectedSnapshot` content (you should never update manually - see `Update Snapshots` section) must end with // end of snapshot

## Fixing the tests
You should never update the snapshot manually. Most of the time you should update code inside
templates ./templates/templates/...


## Update Snapshots

```bash
# Update all snapshots
UPDATE_SNAPS=1 go test ./pkg/generate/snapshots/

# Update specific test
UPDATE_SNAPS=1 go test -run TestSnapPyBase ./pkg/generate/snapshots/
```

## Debug Failed Tests

```bash
# Skip compilation for faster iteration during debugging
SNAPSHOTS_SKIP_COMPILE=1 go test -run TestSnapPyBase ./pkg/generate/snapshots/

# Inspect the generated files for the failed test
cd /tmp/speakeasy-snapshot-tests/{TestName}/

# Review what failed in the snapshot
less /tmp/speakeasy-snapshot-tests/{TestName}/__debug__/snapshot.diff
```

## Watch Templates During Development

For faster iteration when modifying templates, you can watch the templates directory for changes:

```bash
# Set the path to your templates directory and enable watching
export WATCH_TEMPLATES_LOCATION=/path/to/openapi-generation/templates

# Run tests - they will regenerate when templates change
WATCH_TEMPLATES=1 go test -run TestSnapPyBase ./pkg/generate/snapshots/
```

**Important Note**:
* Changes to the spec or go code will not be reloaded
* To exit the watch mode, press `Ctrl+C`.
