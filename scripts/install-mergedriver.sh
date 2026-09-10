#!/usr/bin/env bash

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

function containedWithin() {
  SUBSET="$1"
  SUPERSET="$2"
  CHECK=$(comm -23 <(sort $SUBSET | uniq ) <(sort $SUPERSET | uniq ) | head -1)
  if [[ ! -z $CHECK ]]; then
    return 1
  fi
  return 0
}

if containedWithin "$SCRIPT_DIR/.gitconfig" "$SCRIPT_DIR/../.git/config"; then
  printf "merge driver already installed\n"
  exit 0
else
  printf "installing merge driver\n"
  cat "$SCRIPT_DIR/.gitconfig" >> "$SCRIPT_DIR/../.git/config"
fi

