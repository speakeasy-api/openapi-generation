#!/usr/bin/env bash

set -euo pipefail
IFS=$'\n\t'

if [[ -t 1 ]] && command -v tput >/dev/null && [[ -n "${TERM:-}" ]]; then
    # Color Constants
    readonly RED=$(tput setaf 1)
    readonly GREEN=$(tput setaf 2)
    readonly YELLOW=$(tput setaf 3)
    readonly BLUE=$(tput setaf 4)
    readonly MAGENTA=$(tput setaf 5)
    readonly CYAN=$(tput setaf 6)
    readonly WHITE=$(tput setaf 7)
    readonly GREY=$(tput setaf 8)
    readonly BOLD=$(tput bold)
    readonly ITALIC=$(tput dim)
    readonly RESET=$(tput sgr0)
    readonly COLORS=$(tput colors)
else
    # Fallback to empty strings if tput is not available or terminal is not interactive
    readonly GREEN=""
    readonly RED=""
    readonly BLUE=""
    readonly MAGENTA=""
    readonly CYAN=""
    readonly YELLOW=""
    readonly GREY=""
    readonly WHITE=""
    readonly BOLD=""
    readonly ITALIC=""
    readonly RESET=""
    readonly COLORS=""

fi

readonly BUILD_ICON=🔨
readonly TEST_ICON=🧪
readonly SNIPPETS_ICON=📝
readonly SKIP_ICON=⏭️
readonly TARGET_ICON=🎯
readonly SUCCESS_ICON=✅
readonly FAIL_ICON=❌
readonly COVERAGE_ICON=📊

# print_formatted accepts a formatted message of the form:
#    "{bold}{color}some text{reset} {color}some other text{reset}"
# replaces the color and reset with the appropriate ANSI escape codes
# and prints it in color if the terminal supports it.
print_formatted() {
    if [[ $# -ne 1 ]]; then
        echo "Usage: print_formatted <message>" >&2
        return 1
    fi

    local message="$1"

    if [[ $COLORS -ge 8 ]]; then
        # Replace color placeholders with actual codes
        message="${message//\{bold\}/$BOLD}"
        message="${message//\{green\}/$GREEN}"
        message="${message//\{red\}/$RED}"
        message="${message//\{grey\}/$GREY}"
        message="${message//\{blue\}/$BLUE}"
        message="${message//\{yellow\}/$YELLOW}"
        message="${message//\{cyan\}/$CYAN}"
        message="${message//\{magenta\}/$MAGENTA}"
        message="${message//\{white\}/$WHITE}"
        message="${message//\{reset\}/$RESET}"
        message="${message//\{italic\}/$ITALIC}"
        printf '%b\n' "$message"
    else
        # Strip color codes if terminal doesn't support colors
        message=$(sed -E 's/\{(bold|green|red|grey|blue|yellow|cyan|magenta|reset|italic|white)\}//g' <<<"$message")
        printf '%s\n' "$message"
    fi
}

format_kwargs() {
    if [[ $# -lt 1 ]]; then
        echo "Usage: format_kwargs <kwargs> where kwargs is a list of | separated key=value pairs" >&2
        return 1
    fi

    local kwargs="$1"
    local formatted_pairs=()
    # split kwargs by | and format each key=value pair
    IFS='|' read -ra pairs <<<"$kwargs"
    for pair in "${pairs[@]}"; do
        local key=$(echo "$pair" | cut -d= -f1)
        local value=$(echo "$pair" | cut -d= -f2)
        formatted_pairs+=("${key}: {bold}{cyan}${value}{reset}")
    done
    local msg=$(printf "%s | " "${formatted_pairs[@]}" | sed 's/ | $//')
    printf '%s' "$msg"
}

print_msg_with_ctx() {
    if [[ $# -lt 2 ]]; then
        echo "Usage: print_msg_with_ctx <msg> <kwargs>" >&2
        return 1
    fi

    local msg="$1"
    local kwargs="$2"
    local ctx=$(format_kwargs "$kwargs")
    print_formatted "${msg} [${ctx}]"
}

print_divider() {
    local text="${1}"
    local divider="━"
    local width=95

    local padding=$(((width - ${#text} - 2) / 2))
    local left_divider=$(printf "%${padding}s" | sed "s/ /$divider/g")
    local right_divider=$(printf "%$((width - padding - ${#text} - 2))s" | sed "s/ /$divider/g")
    print_formatted "{grey}${left_divider}{reset} {bold}${text}{reset} {grey}${right_divider}{reset}"
}

load_local_env() {
    local root env_file line key value file_keys=" "
    root="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.." &>/dev/null && pwd)"
    env_file="${root}/.env"
    if [[ -L "${env_file}" ]]; then
        echo "Ignoring ${env_file}: it is a symlink, and the license election must live in a regular file" >&2
        return 0
    fi
    [[ -f "${env_file}" ]] || return 0

    while IFS= read -r line || [[ -n "${line}" ]]; do
        line="${line#"${line%%[![:space:]]*}"}"
        [[ -z "${line}" || "${line}" == \#* ]] && continue
        line="${line#export }"
        line="${line#"${line%%[![:space:]]*}"}"
        [[ "${line}" == *=* ]] || continue
        key="${line%%=*}"
        key="${key%"${key##*[![:space:]]}"}"
        value="${line#*=}"
        value="${value#"${value%%[![:space:]]*}"}"
        value="${value%"${value##*[![:space:]]}"}"
        if [[ "${value}" == \"*\" && "${value}" == *\" && ${#value} -ge 2 ]]; then
            value="${value:1:${#value}-2}"
        fi
        [[ "${key}" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]] || continue
        if [[ -z "${!key+x}" || "${file_keys}" == *" ${key} "* ]]; then
            export "${key}=${value}"
            file_keys+="${key} "
        fi
    done < "${env_file}"
}

ensure_license_election() {
    load_local_env
    if [[ -n "${SPEAKEASY_LICENSE_TOKEN:-}" || -n "${SPEAKEASY_GENERATED_LICENSE:-}" || "${EXTRA_ARGS:-}" == *--license* ]]; then
        return 0
    fi
    cat >&2 <<'MSG'
No license election for generated output (SPEAKEASY_LICENSE_TOKEN or SPEAKEASY_GENERATED_LICENSE).

  Commercial customers: run `speakeasy auth login`, then `./zero` — it fetches a
                        license token for your workspace into .env.
  Open source / AGPL:   run `./zero --license agpl-3.0-only` (or set
                        SPEAKEASY_GENERATED_LICENSE=agpl-3.0-only in .env, see .env.example).
MSG
    return 1
}

run_cmd() {
    if [[ -n "${LOG_OUTPUT:-}" ]] && [[ -z "${LOG_FILE:-}" ]]; then
        export LOG_FILE=$(mktemp)
        print_formatted "{grey}Logs will be written to: {white}${LOG_FILE}{reset}"
    fi

    if [[ -n "${LOG_FILE:-}" ]]; then
        eval "$@" >>"${LOG_FILE}" 2>&1
    else
        eval "$@"
    fi
}
