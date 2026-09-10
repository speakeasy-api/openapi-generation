#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}=== Changelog Shift Script ===${NC}"

# Step 1: Find all changelog changes between current branch and main
echo -e "\n${YELLOW}Step 1: Finding changelog differences...${NC}"

# Get all added/modified changelogs in current branch vs main
CHANGED_CHANGELOGS=$(git diff main --name-only --diff-filter=AM -- changelogs/ | grep -E 'changelogs/[^/]+/[^/]+\.md$' || true)

if [ -z "$CHANGED_CHANGELOGS" ]; then
    echo -e "${GREEN}No changelog changes found. Nothing to do.${NC}"
    exit 0
fi

echo "Found changelog changes:"
echo "$CHANGED_CHANGELOGS"

# Step 2: Group by target and check for multiple changes per target
echo -e "\n${YELLOW}Step 2: Checking for multiple changelogs per target...${NC}"

# Store target-changelog pairs as newline-separated "target|changelog" strings
target_changelogs=""
seen_targets=""

while IFS= read -r changelog; do
    if [ -z "$changelog" ]; then
        continue
    fi

    # Extract target from path: changelogs/TARGET/feature-version.md
    target=$(echo "$changelog" | cut -d'/' -f2)

    # Check if we've seen this target before
    if echo "$seen_targets" | grep -q "^${target}$"; then
        # Find the existing changelog for this target
        existing=$(echo "$target_changelogs" | grep "^${target}|" | cut -d'|' -f2)
        echo -e "${RED}ERROR: Multiple changelogs found for target '$target':${NC}"
        echo "  - $existing"
        echo "  - $changelog"
        echo -e "${RED}Can't resolve multiple changelogs for the same target with this script${NC}"
        exit 1
    fi

    seen_targets="${seen_targets}${target}"$'\n'
    target_changelogs="${target_changelogs}${target}|${changelog}"$'\n'
done <<< "$CHANGED_CHANGELOGS"

# Step 3 & 4 & 5: Process each changelog
echo -e "\n${YELLOW}Step 3-5: Processing changelogs...${NC}"

# Store new versions as "target:feature|version" strings
new_versions=""

while IFS='|' read -r target changelog; do
    if [ -z "$target" ] || [ -z "$changelog" ]; then
        continue
    fi

    echo -e "\n${GREEN}Processing $target: $changelog${NC}"

    # Extract filename without path and extension
    filename=$(basename "$changelog" .md)

    # Extract feature and version from filename: feature-major.minor.patch
    if [[ ! "$filename" =~ ^(.+)-([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        echo -e "${RED}ERROR: Invalid changelog filename format: $filename${NC}"
        exit 1
    fi

    feature="${BASH_REMATCH[1]}"
    major="${BASH_REMATCH[2]}"
    minor="${BASH_REMATCH[3]}"
    patch="${BASH_REMATCH[4]}"

    echo "  Feature: $feature"
    echo "  Original version: $major.$minor.$patch"

    # Step 3: Determine uptick type
    uptick_type=""
    if [ "$patch" != "0" ]; then
        uptick_type="patch"
    elif [ "$minor" != "0" ]; then
        uptick_type="minor"
    elif [ "$major" != "0" ]; then
        uptick_type="major"
    else
        echo -e "${RED}ERROR: Version $major.$minor.$patch has all zeros${NC}"
        exit 1
    fi

    echo "  Uptick type: $uptick_type"

    # Step 4: Find latest version in features.ts
    features_file="templates/templates/$target/features.ts"

    if [ ! -f "$features_file" ]; then
        echo -e "${RED}ERROR: features.ts not found: $features_file${NC}"
        exit 1
    fi

    # Extract current version from features.ts
    # Looking for lines like: methodSecurity: "0.1.1",
    current_version=$(grep -E "^\s*${feature}:\s*\"[0-9]+\.[0-9]+\.[0-9]+\"" "$features_file" | sed -E 's/.*"([0-9]+\.[0-9]+\.[0-9]+)".*/\1/' || true)

    if [ -z "$current_version" ]; then
        echo -e "${RED}ERROR: Could not find feature '$feature' in $features_file${NC}"
        exit 1
    fi

    echo "  Current version in features.ts: $current_version"

    # Parse current version
    if [[ ! "$current_version" =~ ^([0-9]+)\.([0-9]+)\.([0-9]+)$ ]]; then
        echo -e "${RED}ERROR: Invalid version format: $current_version${NC}"
        exit 1
    fi

    curr_major="${BASH_REMATCH[1]}"
    curr_minor="${BASH_REMATCH[2]}"
    curr_patch="${BASH_REMATCH[3]}"

    # Step 5: Calculate new version based on uptick type
    new_major=$curr_major
    new_minor=$curr_minor
    new_patch=$curr_patch

    case "$uptick_type" in
        "major")
            new_major=$((curr_major + 1))
            new_minor=0
            new_patch=0
            ;;
        "minor")
            new_minor=$((curr_minor + 1))
            new_patch=0
            ;;
        "patch")
            new_patch=$((curr_patch + 1))
            ;;
    esac

    new_version="$new_major.$new_minor.$new_patch"
    echo "  New version: $new_version"

    # Create new changelog file
    new_changelog="changelogs/$target/${feature}-${new_version}.md"

    if [ -f "$new_changelog" ]; then
        echo "  ✓ Changelog already exists at correct version: $new_changelog"
    else
        echo "  Creating new changelog: $new_changelog"

        # Extract OUR version of the changelog (in case of conflict, :2: is our version)
        if git show ":2:$changelog" > "$new_changelog" 2>/dev/null; then
            echo "  ✓ Extracted OUR version from conflict"
        elif [ -f "$changelog" ] && ! grep -q "^<<<<<<< " "$changelog"; then
            # File exists and has no conflict markers - safe to copy
            cp "$changelog" "$new_changelog"
            echo "  ✓ Copied from existing file (no conflict)"
        else
            echo -e "${RED}ERROR: Could not extract changelog content from $changelog${NC}"
            exit 1
        fi

        # Update the version string in the first line of the changelog
        # Format: ## feature: X.Y.Z - YYYY-MM-DD
        sed -i.bak -E "s/^(## ${feature}: )[0-9]+\.[0-9]+\.[0-9]+/\1${new_version}/" "$new_changelog"
        rm -f "${new_changelog}.bak"
        echo "  ✓ Updated version string in changelog header to $new_version"

        git add "$new_changelog"
    fi

    # Store new version for features.ts update
    new_versions="${new_versions}${target}:${feature}|${new_version}"$'\n'
done <<< "$target_changelogs"

# Step 6: Resolve conflicts
echo -e "\n${YELLOW}Step 6: Resolving conflicts...${NC}"

# Get list of conflicted files
CONFLICTED_FILES=$(git diff --name-only --diff-filter=U || true)

if [ -n "$CONFLICTED_FILES" ]; then
    echo "Found conflicted files:"
    echo "$CONFLICTED_FILES"

    while IFS= read -r conflicted_file; do
        if [ -z "$conflicted_file" ]; then
            continue
        fi

        echo "Processing conflict: $conflicted_file"

        # For old changelog files, accept theirs
        if [[ "$conflicted_file" =~ ^changelogs/.*\.md$ ]]; then
            echo "  Accepting 'theirs' for old changelog: $conflicted_file"
            git checkout --theirs "$conflicted_file"
            git add "$conflicted_file"
        # For features.ts files, accept theirs (will be updated in next step)
        elif [[ "$conflicted_file" =~ features\.ts$ ]]; then
            echo "  Accepting 'theirs' for features.ts: $conflicted_file"
            git checkout --theirs "$conflicted_file"
            git add "$conflicted_file"
        # For auto-generated files, accept theirs
        elif [[ "$conflicted_file" =~ ^zSDKs/ ]] || [[ "$conflicted_file" =~ \.gen\.lock$ ]] || [[ "$conflicted_file" =~ ^tests/config/ ]]; then
            echo "  Accepting 'theirs' for auto-generated file: $conflicted_file"
            git checkout --theirs "$conflicted_file"
            git add "$conflicted_file"
        else
            echo -e "${YELLOW}  WARNING: Skipping non-changelog conflict (needs manual resolution): $conflicted_file${NC}"
        fi
    done <<< "$CONFLICTED_FILES"
fi

# Step 7: Update all features.ts files
echo -e "\n${YELLOW}Step 7: Updating features.ts files...${NC}"

while IFS='|' read -r target_feature new_version; do
    if [ -z "$target_feature" ] || [ -z "$new_version" ]; then
        continue
    fi

    target="${target_feature%%:*}"
    feature="${target_feature#*:}"

    features_file="templates/templates/$target/features.ts"
    echo "Updating $features_file: $feature -> $new_version"

    # Update the version in features.ts
    sed -i.bak -E "s/(${feature}: \")[0-9]+\.[0-9]+\.[0-9]+(\")/\1${new_version}\2/" "$features_file"
    rm -f "${features_file}.bak"

    git add "$features_file"
done <<< "$new_versions"

echo -e "\n${GREEN}=== Changelog shift complete! ===${NC}"
echo -e "${YELLOW}Next steps:${NC}"
echo "1. Review the changes: git status"
echo "2. Run: make build-review-sdks-without-mprocs"
echo "3. Complete the merge: git merge --continue (or git rebase --continue)"
