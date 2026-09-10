#!/usr/bin/env bash
# Syncs the global `security` array from each base spec into all of its
# associated fragment files. Run this after changing a base spec's security.
#
# Each variant's base spec is read from `x-base-spec` in its overlay.yaml.
# Spec-named fragment directories (e.g. fragments/uber/) are matched directly.

set -euo pipefail
shopt -s nullglob

SPECS_DIR="./tests/specs"
OVERLAYS_DIR="./tests/overlays"
FRAGMENTS_DIR="${SPECS_DIR}/fragments"
resolve_spec() {
	local variant="$1"
	local overlay="${OVERLAYS_DIR}/${variant}/overlay.yaml"

	if [[ ! -f "${overlay}" ]]; then
		echo "warning: no overlay.yaml found for variant '${variant}', skipping" >&2
		return 1
	fi

	local spec
	spec="$(yq -r '.info.x-base-spec // ""' "${overlay}")"
	if [[ -z "${spec}" ]]; then
		echo "warning: missing x-base-spec in ${overlay}, skipping" >&2
		return 1
	fi

	printf '%s\n' "${spec}"
}

sync_dir() {
	local spec="$1"
	local dir="$2"

	if [[ ! -d "${dir}" ]]; then
		return
	fi

	if ! yq -e '.security' "${spec}" >/dev/null 2>&1; then
		return
	fi

	local spec_security
	spec_security="$(yq '.security' "${spec}")"

	for fragment in "${dir}"*.yaml "${dir}"*.yml; do
		local frag_security
		frag_security="$(yq '.security' "${fragment}")"
		if [[ "${frag_security}" != "${spec_security}" ]]; then
			yq -i ".security = load(\"${spec}\").security" "${fragment}"
		fi
	done
}

synced_dirs=()

mark_synced_dir() {
	local dir="$1"
	synced_dirs+=("${dir}")
}

is_synced_dir() {
	local dir="$1"
	local synced_dir
	for synced_dir in "${synced_dirs[@]}"; do
		if [[ "${synced_dir}" == "${dir}" ]]; then
			return 0
		fi
	done

	return 1
}

# primary has no top-level overlay.yaml; its base spec is uber.yaml
sync_dir "${SPECS_DIR}/uber.yaml" "${FRAGMENTS_DIR}/primary/"
mark_synced_dir "${FRAGMENTS_DIR}/primary/"

# Sync variant fragment directories (e.g. fragments/secondary/)
for variant_dir in "${OVERLAYS_DIR}"/*/; do
	variant="$(basename "${variant_dir}")"
	frag_dir="${FRAGMENTS_DIR}/${variant}/"
	is_synced_dir "${frag_dir}" && continue
	spec="$(resolve_spec "${variant}")" || continue

	sync_dir "${spec}" "${frag_dir}"
	mark_synced_dir "${frag_dir}"
done

# Sync spec-named fragment directories (e.g. fragments/uber/)
for spec_file in "${SPECS_DIR}"/*.yaml; do
	spec_name="$(basename "${spec_file}" .yaml)"
	frag_dir="${FRAGMENTS_DIR}/${spec_name}/"
	if ! is_synced_dir "${frag_dir}"; then
		sync_dir "${spec_file}" "${frag_dir}"
		mark_synced_dir "${frag_dir}"
	fi
done

