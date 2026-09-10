echo ""
echo "#######################"
echo "# Merge driver called #"
echo "# $1 #"
echo "# $2 #"
echo "# $3 #"
echo "# $4 #"
echo "#######################"
echo ""

SCRIPT_DIR=$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )

function isSubPath() {
  sd=$(realpath "$1")
  d=$(realpath "$2")
  [ "${sd:0:${#d}}" = "$d" ]
  return $?
}

CONFLICTED_FILE="$( cd "$( dirname "$4" )" && pwd )"
echo "  => CONFLICTED_FILE=$CONFLICTED_FILE"

if isSubPath "$CONFLICTED_FILE" "$SCRIPT_DIR/../zSDKs/sdk-csharp"
then
  (cd "$SCRIPT_DIR/.." ; make build-csharp)
  exit 0
fi

if isSubPath "$CONFLICTED_FILE" "$SCRIPT_DIR/../zSDKs/sdk-go"
then
  (cd "$SCRIPT_DIR/.." ; make build-go)
  exit 0
fi

if isSubPath "$CONFLICTED_FILE" "$SCRIPT_DIR/../zSDKs/sdk-javav2"
then
  (cd "$SCRIPT_DIR/.." ; make build-javav2)
  exit 0
fi

if isSubPath "$CONFLICTED_FILE" "$SCRIPT_DIR/../zSDKs/mcp-typescript"
then
  (cd "$SCRIPT_DIR/.." ; make build-mcp-typescript)
  exit 0
fi

if isSubPath "$CONFLICTED_FILE" "$SCRIPT_DIR/../zSDKs/sdk-php"
then
  (cd "$SCRIPT_DIR/.." ; make build-php)
  exit 0
fi

if isSubPath "$CONFLICTED_FILE" "$SCRIPT_DIR/../zSDKs/sdk-pythonv2"
then
  (cd "$SCRIPT_DIR/.." ; make build-pythonv2)
  exit 0
fi

if isSubPath "$CONFLICTED_FILE" "$SCRIPT_DIR/../zSDKs/sdk-ruby"
then
  (cd "$SCRIPT_DIR/.." ; make build-ruby)
  exit 0
fi

if isSubPath "$CONFLICTED_FILE" "$SCRIPT_DIR/../zSDKs/sdk-terraform"
then
  (cd "$SCRIPT_DIR/.." ; make build-terraform)
  exit 0
fi

if isSubPath "$CONFLICTED_FILE" "$SCRIPT_DIR/../zSDKs/sdk-typescriptv2"
then
  (cd "$SCRIPT_DIR/.." ; make build-typescriptv2)
  exit 0
fi

echo "Error: merge driver called with unexpected file: $4"
exit 1
