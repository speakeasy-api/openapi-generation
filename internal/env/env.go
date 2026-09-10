package env

import (
	"fmt"
	"os"
	"strconv"
)

func IsDebug() bool {
	return os.Getenv("SPEAKEASY_DEBUG") == "true"
}

func IsForceGeneration() bool {
	return os.Getenv("SPEAKEASY_FORCE_GENERATION") == "true"
}

func IsGithubAction() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

func DebugOperationFilter() string {
	return os.Getenv("SPEAKEASY_DEBUG_OPERATION_FILTER")
}

func DebugCycles() bool {
	return os.Getenv("SPEAKEASY_DEBUG_CYCLES") != "" && os.Getenv("SPEAKEASY_DEBUG_CYCLES") != "false"
}

func DebugConcurrency() int {
	debugConcurrency := os.Getenv("SPEAKEASY_DEBUG_CONCURRENCY")
	if debugConcurrency == "" {
		return 0
	}
	concurrency, err := strconv.Atoi(debugConcurrency)
	if err != nil {
		fmt.Println("Error parsing SPEAKEASY_DEBUG_CONCURRENCY:", err)
		return 0
	}
	return concurrency
}

func DebugWriteFiles() bool {
	return os.Getenv("SPEAKEASY_DEBUG_WRITE_FILES") != "" && os.Getenv("SPEAKEASY_DEBUG_WRITE_FILES") != "false"
}

func UpdateSnapshots() bool {
	return os.Getenv("UPDATE_SNAPS") != "" && os.Getenv("UPDATE_SNAPS") != "false"
}

func SnapshotSkipCompile() bool {
	return os.Getenv("SNAPSHOTS_SKIP_COMPILE") != "" && os.Getenv("SNAPSHOTS_SKIP_COMPILE") != "false"
}

func SnapshotsSkipCleanDir() bool {
	return os.Getenv("SNAPSHOTS_SKIP_CLEAN_DIR") != "" && os.Getenv("SNAPSHOTS_SKIP_CLEAN_DIR") != "false"
}

func DebugDisableMockServer() bool {
	return os.Getenv("SPEAKEASY_DEBUG_DISABLE_MOCK_SERVER") != "" && os.Getenv("SPEAKEASY_DEBUG_DISABLE_MOCK_SERVER") != "false"
}

// Usage:
// export WATCH_TEMPLATES_LOCATION=/path/to/openapi-generation/templates
// go run cmd/generate/main.go --watch-templates
func WatchTemplatesLocation() string {
	location := os.Getenv("WATCH_TEMPLATES_LOCATION")
	if location == "" {
		location = os.Getenv("SPEAKEASY_WATCH_TEMPLATES_LOCATION")
	}
	return location
}

// Usage:
// export WATCH_TEMPLATES_LOCATION=/path/to/openapi-generation/templates
// WATCH_TEMPLATES=1 go run cmd/generate/main.go
func WatchTemplatesEnabled() bool {
	return os.Getenv("WATCH_TEMPLATES") != "" && os.Getenv("WATCH_TEMPLATES") != "false"
}

func DebugInferDiscriminators() bool {
	return os.Getenv("SPEAKEASY_DEBUG_INFER_DISCRIMINATORS") != "" && os.Getenv("SPEAKEASY_DEBUG_INFER_DISCRIMINATORS") != "false"
}

func DebugPreApplyUnionDiscriminators() bool {
	return os.Getenv("SPEAKEASY_DEBUG_PRE_APPLY_UNION_DISCRIMINATORS") != "" && os.Getenv("SPEAKEASY_DEBUG_PRE_APPLY_UNION_DISCRIMINATORS") != "false"
}

// DebugGojaPort returns the DAP debug port for the goja JS runtime.
// Set SPEAKEASY_DEBUG_GOJA=true for the default port (4711), or a number for a custom port.
// Returns 0 when debugging is disabled.
func DebugGojaPort() int {
	val := os.Getenv("SPEAKEASY_DEBUG_GOJA")
	if val == "" || val == "false" {
		return 0
	}
	if val == "true" {
		return 4711
	}
	port, err := strconv.Atoi(val)
	if err != nil {
		fmt.Println("Error parsing SPEAKEASY_DEBUG_GOJA:", err)
		return 0
	}
	return port
}
