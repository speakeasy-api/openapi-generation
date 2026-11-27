#!/usr/bin/env bash
#
# Exercises the generated Unix installer offline. curl, uname and tar are
# replaced by PATH shims: curl serves a locally built archive, checksums.txt
# and the latest-release lookup, and tar records each extraction before running
# the real tar. Each case checks the exit status, what was downloaded and
# extracted, and what was installed.
#
# Usage: scripts/test-unix-installer.sh [path/to/install.sh]

# check evaluates its single-quoted condition after each run, so expansion is
# deliberately deferred and status is read there.
# shellcheck disable=SC2016,SC2034

set -euo pipefail
unset CLI_VERSION CLI_INSTALL_DIR

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
installer=${1:-"$script_dir/../zSDKs/sdk-cli/scripts/install.sh"}
installer=$(cd "$(dirname "$installer")" && pwd)/$(basename "$installer")

real_tar=$(command -v tar)
if command -v sha256sum >/dev/null 2>&1; then
    sha256() { sha256sum "$1" | awk '{ print $1 }'; }
else
    sha256() { shasum -a 256 "$1" | awk '{ print $1 }'; }
fi

work=$(mktemp -d)
trap 'rm -rf -- "$work"' EXIT

shims="$work/shims"
mkdir -p "$shims" "$work/payload"
printf 'stub binary %s\n' "$$" >"$work/payload/cli"
chmod +x "$work/payload/cli"
archive="$work/cli.tar.gz"
"$real_tar" -czf "$archive" -C "$work/payload" cli

cat >"$shims/curl" <<'EOF'
#!/usr/bin/env bash
url="" out=""
while [ $# -gt 0 ]; do
    case "$1" in
        -o) out="$2"; shift 2 ;;
        -*) shift ;;
        *) url="$1"; shift ;;
    esac
done
echo "curl $url" >>"$STUB_LOG"
case "$url" in
    */releases/latest)
        [ -n "$STUB_LATEST" ] || exit 22
        printf '{\n  "tag_name": "%s",\n  "name": "release"\n}\n' "$STUB_LATEST"
        ;;
    */checksums.txt)
        [ "$STUB_CHECKSUMS" = "absent" ] && exit 22
        printf '%s\n' "$STUB_CHECKSUMS_BODY" >"$out"
        ;;
    *)
        cp "$STUB_ARCHIVE" "$out"
        ;;
esac
EOF

cat >"$shims/uname" <<'EOF'
#!/usr/bin/env bash
case "$1" in
    -s) echo "$STUB_OS" ;;
    -m) echo "$STUB_ARCH" ;;
    *) echo "$STUB_OS" ;;
esac
EOF

cat >"$shims/tar" <<EOF
#!/usr/bin/env bash
echo "tar \$*" >>"\$STUB_LOG"
exec "$real_tar" "\$@"
EOF
chmod +x "$shims/curl" "$shims/uname" "$shims/tar"

failures=0
case_dir=""

fail() {
    echo "FAIL: $1" >&2
    failures=$((failures + 1))
}

ok() {
    echo "ok: $1"
}

check() {
    if eval "$1"; then ok "$2"; else fail "$2"; fi
}

# run_installer <checksums body> [checksums mode]
run_installer() {
    case_dir=$(mktemp -d "$work/case.XXXXXX")
    mkdir -p "$case_dir/install" "$case_dir/home"
    : >"$case_dir/log"
    set +e
    env PATH="$shims:$PATH" HOME="$case_dir/home" \
        STUB_LOG="$case_dir/log" STUB_ARCHIVE="$archive" \
        STUB_OS="${STUB_OS:-Linux}" STUB_ARCH="${STUB_ARCH:-x86_64}" \
        STUB_LATEST="${STUB_LATEST:-}" STUB_CHECKSUMS="${2:-present}" STUB_CHECKSUMS_BODY="$1" \
        CLI_INSTALL_DIR="$case_dir/install" \
        bash "$installer" >"$case_dir/stdout" 2>"$case_dir/stderr"
    status=$?
    set -e
}

installed() { [ -f "$case_dir/install/cli" ] && cmp -s "$case_dir/install/cli" "$work/payload/cli"; }
nothing_installed() { [ ! -e "$case_dir/install/cli" ]; }
extracted() { grep -q '^tar ' "$case_dir/log"; }
logged() { grep -q -- "$1" "$case_dir/log"; }
stderr_has() { grep -q -- "$1" "$case_dir/stderr"; }

digest=$(sha256 "$archive")
archive_name="cli_Linux_x86_64.tar.gz"

echo "--- pinned version installs after checksum verification"
CLI_VERSION=v0.0.0-test run_installer "$digest  $archive_name"
check '[ "$status" -eq 0 ]' "exits 0"
check installed "binary is installed"
check 'logged "/releases/download/v0.0.0-test/$archive_name\$"' "downloads the pinned archive"
check 'logged "/releases/download/v0.0.0-test/checksums.txt\$"' "downloads checksums from the same release"
check '! logged "/releases/latest"' "does not look up the latest release"
check 'stderr_has "Checksum verified"' "reports the verified checksum"
check '[ ! -s "$case_dir/stdout" ]' "logs go to stderr"

echo "--- binary-mode checksum entry is accepted"
CLI_VERSION=v0.0.0-test run_installer "$digest *$archive_name"
check '[ "$status" -eq 0 ] && installed' "installs with a *-prefixed entry"

echo "--- latest version comes from the release lookup"
STUB_LATEST=v1.2.3 run_installer "$digest  $archive_name"
check '[ "$status" -eq 0 ] && installed' "installs the latest release"
check 'logged "/releases/download/v1.2.3/$archive_name\$"' "downloads the looked-up tag"

echo "--- checksum mismatch stops before extraction"
CLI_VERSION=v0.0.0-test run_installer "$(printf '0%.0s' $(seq 64))  $archive_name"
check '[ "$status" -ne 0 ]' "exits non-zero"
check '! extracted' "nothing is extracted"
check nothing_installed "nothing is installed"
check 'stderr_has "Checksum mismatch"' "reports the mismatch"

echo "--- missing checksum entry stops before extraction"
CLI_VERSION=v0.0.0-test run_installer "$digest  some_other_archive.tar.gz"
check '[ "$status" -ne 0 ] && ! extracted && nothing_installed' "exits non-zero without extracting"
check 'stderr_has "no entry for $archive_name"' "names the missing entry"

echo "--- missing checksums.txt stops before extraction"
CLI_VERSION=v0.0.0-test run_installer "" absent
check '[ "$status" -ne 0 ] && ! extracted && nothing_installed' "exits non-zero without extracting"
check 'stderr_has "Failed to download checksums"' "reports the failed download"

echo "--- failed latest-release lookup exits non-zero"
STUB_LATEST="" run_installer "$digest  $archive_name"
check '[ "$status" -ne 0 ] && nothing_installed' "exits non-zero"
check '! logged "/releases/download/"' "downloads no archive"
check 'stderr_has "Could not determine the latest release"' "explains the failed lookup"

echo "--- unsupported operating system exits non-zero"
STUB_OS=FreeBSD CLI_VERSION=v0.0.0-test run_installer "$digest  $archive_name"
check '[ "$status" -ne 0 ] && nothing_installed' "exits non-zero"
check '[ ! -s "$case_dir/stdout" ]' "prints nothing on stdout"
check '! logged "/releases/download/"' "downloads nothing"
check 'stderr_has "Unsupported operating system: FreeBSD"' "names the operating system"

if [ "$failures" -ne 0 ]; then
    echo "$failures check(s) failed" >&2
    exit 1
fi
echo "all installer checks passed"
