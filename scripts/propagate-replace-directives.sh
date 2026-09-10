#!/bin/bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

# This script propagates replace directives from a source go.mod file to another module's go.mod file.

set -euo pipefail

rel_path() {
    perl -e 'use File::Spec; print File::Spec->abs2rel(@ARGV) . "\n"' "$1" "$2" 2>/dev/null || echo "$1"
}


# Function to check if a file exists and is readable
check_file() {
    if [[ ! -f "$1" || ! -r "$1" ]]; then
        echo "Error: $1 does not exist or is not readable" >&2
        exit 1
    fi
}

# Function to propagate replace directives
propagate_replaces() {
    local source_mod="$1"
    local target_mod="$2"
    local clean="$3"
    local source_dir=$(dirname "$source_mod")
    local target_dir=$(dirname "$target_mod")

    # If there are replace directives in the source mod, return
    if grep -q 'replace ' "$source_mod"; then
        # Get existing replace directives from target mod
        local target_replaces=$(go mod edit -json "$target_mod" | jq -r '.Replace[]? | "\(.Old.Path)"')

        if [[ "$clean" != "noclean" ]]; then
          # Drop existing replace directives from target mod
          while IFS= read -r old_path; do
              if [[ -n "$old_path" ]]; then
                  go mod edit -dropreplace="$old_path" "$target_mod"
              fi
          done <<< "$target_replaces"
        fi

        # Apply replace directives to target mod
        go mod edit -json "$source_mod" | jq -r '
            .Replace[] |
            .Old.Path as $old_path |
            .New.Path as $new_path |
            .New.Version as $version |
            [
                $old_path,
                $new_path,
                ($version // "")
            ] |
            @tsv
        ' | while IFS=$'\t' read -r old_path new_path version; do
            # Adjust local replacement paths relative to the target module.
            if [[ "$new_path" = /* ]]; then
                new_path=$(rel_path "$new_path" "$target_dir")
            elif [[ "$new_path" = ./* || "$new_path" = ../* ]]; then
                if ! (cd "$source_dir" && [[ -d "$new_path" ]]); then
                    echo "local replacement target is unavailable: $new_path" >&2
                    exit 1
                fi
                new_path=$(rel_path "$source_dir/$new_path" "$target_dir")
            fi

            if [[ -n "$version" ]]; then
                new_path="${new_path}@${version}"
            fi

            echo "  $ go mod edit -replace=\"$old_path=$new_path\" \"$target_mod\""
            go mod edit -replace="$old_path=$new_path" "$target_mod"
        done
    fi

    # Get the module name of the source go.mod
    local source_module=$(cd "$(dirname "$source_mod")" && GOWORK=off go list -m)

    local relative_source_path=$(rel_path "$source_dir" "$target_dir")

    echo "  $ go mod edit -replace=\"$source_module=$relative_source_path\" \"$target_mod\""
    go mod edit -replace="$source_module=$relative_source_path" "$target_mod"
}

if [[ $# -ne 2 ]]; then
    echo "Usage: $0 <path_to_source_go.mod> <path_to_target_go.mod>" >&2
    exit 1
fi

SOURCE_MOD="$1"
TARGET_MOD="$2"

# if TARGET_MOD has wasm in it, we should not clean it
SHOULD_CLEAN=""
if [[ $TARGET_MOD == *"wasm"* ]]; then
  SHOULD_CLEAN="noclean"
fi

# Check if the source go.mod file exists and is readable
check_file "$SOURCE_MOD"
check_file "$TARGET_MOD"

# Propagate replace directives
propagate_replaces "$SOURCE_MOD" "$TARGET_MOD" "$SHOULD_CLEAN"

(cd "$(dirname "$TARGET_MOD")" && GOWORK=off go mod tidy)

RELATIVE_TARGET_MOD=$(rel_path "$TARGET_MOD" "$(pwd)")

header="==== $RELATIVE_TARGET_MOD ===="
separator=$(printf '=%.0s' $(seq 1 ${#header}))
echo "$header"
cat "$TARGET_MOD"
echo "$separator"

echo "Replace directives have been propagated from $SOURCE_MOD to $TARGET_MOD"
