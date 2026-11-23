#!/usr/bin/env bash

# build-csharpv2-netstandard2.sh
#
# Generates csharpv2 variants with `targetFramework: netstandard2.0` from their
# test spec, config and hooks, then builds the SDK project for every target
# framework with warnings as errors, except CS0618 and NU1901-NU1904. Fails when
# the CS0618 warnings differ between the target frameworks. Tests are neither
# built nor run. Overwrites ./testSDKs/sdk-csharpv2-<variant> and removes
# ./testusages/csharpv2/<variant> and ./testprojects/csharpv2/sdk-csharpv2-<variant>.
#
# Usage:
#   ./build-csharpv2-netstandard2.sh <variant...>

set -euo pipefail
IFS=$'\n\t'

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"
ensure_license_election

readonly WARNINGS_NOT_AS_ERRORS="CS0618;NU1901;NU1902;NU1903;NU1904"
readonly BUILD_LOG="build-netstandard2.log"

function fail() {
    print_msg_with_ctx "${FAIL_ICON} {red}{bold}$1{reset}" "variant=$2"
    exit 1
}

function cs0618_pattern() {
    printf '%s' "warning CS0618: .*::TargetFramework=${1//./\\.}\\]\$"
}

function cs0618_count() {
    grep -cE "$(cs0618_pattern "$2")" "$1" || true
}

function cs0618_for() {
    local -r log="$1"
    local -r tfm="$2"
    grep -E "$(cs0618_pattern "${tfm}")" "${log}" \
        | sed -E -e 's/ \[[^]]*\]$//' -e "s#/obj/([^/]+)/${tfm}/#/obj/\1/#" \
        | sort -u || true
}

function check_cs0618_parity() {
    local -r variant="$1"
    local -r log="$2"
    local -r net_tfm="$3"

    [[ -f "${log}" ]] || fail "${BUILD_LOG} not written" "${variant}"

    local -r total="$(grep -c 'warning CS0618' "${log}" || true)"
    local -r netstandard_lines="$(cs0618_count "${log}" netstandard2.0)"
    local -r net_lines="$(cs0618_count "${log}" "${net_tfm}")"
    if ((total != netstandard_lines + net_lines)); then
        fail "CS0618 lines without a known target framework in ${BUILD_LOG}" "${variant}"
    fi

    if ! diff <(cs0618_for "${log}" netstandard2.0) <(cs0618_for "${log}" "${net_tfm}"); then
        fail "CS0618 differs between netstandard2.0 and ${net_tfm}" "${variant}"
    fi

    local -r unique_count="$(cs0618_for "${log}" netstandard2.0 | wc -l | tr -d ' ')"
    print_msg_with_ctx "${SUCCESS_ICON} {green}{bold}CS0618 identical across target frameworks{reset}" \
        "variant=${variant}|netstandard2.0=${unique_count}|${net_tfm}=${unique_count}"
}

function build_variant() {
    local -r variant="$1"
    local -r sdk_dir="./testSDKs/sdk-csharpv2-${variant}"

    EXTRA_ARGS="--skip-compile" "${SCRIPT_DIR}/build-target.sh" csharpv2 "${variant}"
    run_cmd go run ./cmd/regen -l csharpv2 --skip-compile \
        --set csharp.targetFramework=netstandard2.0 "${sdk_dir}"

    local sdk_project
    sdk_project="$(cd "${sdk_dir}" && find src -mindepth 2 -maxdepth 2 -name '*.csproj' \
        ! -name '*.Models.csproj' ! -name '*.Generators.csproj')"
    [[ -n "${sdk_project}" ]] || fail "SDK project not found" "${variant}"

    local net_tfm
    net_tfm="$(sed -nE 's#.*<TargetFrameworks>netstandard2\.0;([^<]+)</TargetFrameworks>.*#\1#p' "${sdk_dir}/${sdk_project}")"
    [[ -n "${net_tfm}" ]] || fail "${sdk_project} does not target netstandard2.0" "${variant}"

    print_msg_with_ctx "${BUILD_ICON} {green}{bold}Building netstandard2.0 SDK{reset}" "variant=${variant}|project=${sdk_project}"
    (cd "${sdk_dir}" && dotnet build "${sdk_project}" -c Debug -warnaserror "-warnNotAsError:${WARNINGS_NOT_AS_ERRORS}" \
        -p:BuildInParallel=false -nodeReuse:false "-flp:logfile=${BUILD_LOG};warningsonly;verbosity=minimal")

    check_cs0618_parity "${variant}" "${sdk_dir}/${BUILD_LOG}" "${net_tfm}"
}

for variant in "$@"; do
    if [[ "${variant}" == "review" || ! -d "./tests/config/${variant}/csharpv2" ]]; then
        print_msg_with_ctx "${SKIP_ICON} {yellow}{bold}Skipping netstandard2.0 build{reset}" "variant=${variant}"
        continue
    fi
    build_variant "${variant}"
done
