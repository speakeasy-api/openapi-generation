#!/usr/bin/env bash

set -e

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )
ROOT_DIR=$SCRIPT_DIR/..
PERMS_FILE=$ROOT_DIR/templates/perms.go

is_agent_environment() {
    local agent_env_vars=(
        CLAUDE_CODE
        CURSOR_AGENT
        CODEX
        AIDER
        CLINE
        WINDSURF_AGENT
        GITHUB_COPILOT
        AMAZON_Q
        GEMINI_CODE_ASSIST
        SRC_CODY
        FORCE_AGENT_MODE
    )

    local env_var
    for env_var in "${agent_env_vars[@]}"; do
        if [[ -n "${!env_var:-}" ]]; then
            return 0
        fi
    done

    return 1
}

# Clean untracked and gitignored files
clean_preview=$(git clean -nxfd "$ROOT_DIR/templates")
if [[ "$clean_preview" ]]; then
    echo "Running git clean -xfd $ROOT_DIR/templates"
    echo "$clean_preview"

    if is_agent_environment || [[ ! -t 0 ]] || [[ ! -t 1 ]]; then
        echo "Refusing to prompt during agent or non-interactive execution. Clean the templates tree manually and re-run."
        exit 1
    fi

    read -r -p "Are you sure? [y/N] " response
    case "$response" in
        [yY])
            git clean -xfd "$ROOT_DIR/templates"
            ;;
        *)
            echo "Aborting"
            exit 1
            ;;
    esac
fi

# Update perms.go
go generate $ROOT_DIR/templates/...

# Account for different umask defaults
if [[ "$OSTYPE" == "darwin"* ]]; then
    # macOS
    sed -i '' 's/775/755/g' $PERMS_FILE
    sed -i '' 's/664/644/g' $PERMS_FILE
else
    # Linux
    sed -i 's/775/755/g' $PERMS_FILE
    sed -i 's/664/644/g' $PERMS_FILE
fi

