#!/usr/bin/env bash

# check-csharpv2-pack.sh
#
# Packs a generated csharpv2 SDK from its solution and fails unless:
#   - the solution packs into exactly one .nupkg, the SDK package
#   - lib/ has one folder per SDK target framework, holding only <assembly>.dll and <assembly>.xml
#   - README.md and every file under docs/ are bundled
#   - no source generator assembly or analyzers/ folder is bundled
#   - each nuspec dependency group lists exactly the package references restored for that target
#     framework, whatever their PrivateAssets (build-only packages excepted)
#   - System.Text.Json and Microsoft.Bcl.AsyncInterfaces are dependencies of the netstandard2.0
#     group only
#   - projects restoring the package from a local feed build for every package target framework,
#     and for net8.0 when the package has netstandard2.0 but no net8.0 asset; executables (every
#     framework but netstandard2.0) run. SDKs with Server-Sent Events enumerate a stream through
#     the public IAsyncEnumerable surface.
#
# Usage:
#   ./check-csharpv2-pack.sh <sdk_dir>

set -euo pipefail
IFS=$'\n\t'

readonly SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" &>/dev/null && pwd)"
source "${SCRIPT_DIR}/utils.sh"

readonly BUILD_ONLY_PACKAGES="Microsoft.SourceLink.GitHub"
readonly WARNINGS_NOT_AS_ERRORS="NETSDK1138;NU1901;NU1902;NU1903;NU1904"

[[ $# -eq 1 && -d "$1" ]] || { echo "Usage: $0 <sdk_dir>" >&2; exit 1; }
readonly SDK_DIR="$(cd -- "$1" && pwd)"
readonly SDK_NAME="$(basename "${SDK_DIR}")"

WORK_DIR="$(mktemp -d)"
readonly WORK_DIR
trap 'rm -rf "${WORK_DIR}"' EXIT
readonly FEED="${WORK_DIR}/feed"

function fail() {
    {
        print_msg_with_ctx "${FAIL_ICON} {red}{bold}Package check failed{reset}" "sdk=${SDK_NAME}"
        printf '  - %s\n' "$@"
    } >&2
    exit 1
}

function find_one() {
    local -r what="$1"
    shift
    local matches
    matches="$(find "$@" -not -path "${SDK_DIR}/Tests/*" -not -path '*/bin/*' -not -path '*/obj/*' | sort)"
    [[ -n "${matches}" ]] || fail "${what} not found"
    [[ "$(wc -l <<<"${matches}" | tr -d ' ')" -eq 1 ]] || fail "more than one ${what}: ${matches//$'\n'/ }"
    printf '%s' "${matches}"
}

SOLUTION="$(find_one solution "${SDK_DIR}" -maxdepth 1 \( -name '*.sln' -o -name '*.slnx' \))"
SDK_PROJECT="$(find_one "SDK project" "${SDK_DIR}" -name '*.csproj' -not -name '*.Generators.csproj')"
readonly SOLUTION SDK_PROJECT

dotnet msbuild "${SDK_PROJECT}" -getProperty:PackageId -getProperty:PackageVersion -getProperty:AssemblyName \
    -getProperty:TargetFramework -getProperty:TargetFrameworks >"${WORK_DIR}/properties.json"
target_frameworks="$(python3 -c '
import json, sys
p = json.load(open(sys.argv[1]))["Properties"]
print("\n".join(t.strip() for t in (p["TargetFrameworks"] or p["TargetFramework"]).split(";") if t.strip()))
' "${WORK_DIR}/properties.json")"

print_msg_with_ctx "${BUILD_ICON} {blue}{bold}Packing SDK{reset}" "sdk=${SDK_NAME}|solution=$(basename "${SOLUTION}")"
# Test SDK projects treat warnings as errors; production projects do not, so NuGet pack warnings
# such as NU5123 (long docs/ paths) stay warnings here too.
(cd "${SDK_DIR}" && dotnet pack "${SOLUTION}" -c Release -o "${FEED}" -m:1 -nodeReuse:false -p:TreatWarningsAsErrors=false) \
    >"${WORK_DIR}/pack.log" 2>&1 || { cat "${WORK_DIR}/pack.log"; fail "dotnet pack failed"; }

python3 - "${FEED}" "${WORK_DIR}" "${SDK_DIR}" "$(dirname "${SDK_PROJECT}")/obj/project.assets.json" \
    "${BUILD_ONLY_PACKAGES}" <<'PY' || fail "package contents differ from the SDK project"
import json
import os
import sys
import xml.etree.ElementTree as ET
import zipfile

feed, work_dir, sdk_dir, assets_file = sys.argv[1:5]
build_only = set(sys.argv[5].split(";"))
failures = []

properties = json.load(open(os.path.join(work_dir, "properties.json")))["Properties"]
package_id = properties["PackageId"]
assembly = properties["AssemblyName"]
tfms = [t.strip() for t in (properties["TargetFrameworks"] or properties["TargetFramework"]).split(";") if t.strip()]
sdk_nupkg = f"{package_id}.{properties['PackageVersion']}.nupkg"

nupkgs = sorted(f for f in os.listdir(feed) if f.endswith(".nupkg"))
if nupkgs != [sdk_nupkg]:
    failures.append(f"expected only {sdk_nupkg}, found {' '.join(nupkgs) or 'none'}")
if sdk_nupkg not in nupkgs:
    print("\n".join(f"  - {f}" for f in failures))
    sys.exit(1)

package = zipfile.ZipFile(os.path.join(feed, sdk_nupkg))
entries = [e for e in package.namelist() if not e.endswith("/")]
nuspec_names = [e for e in entries if "/" not in e and e.endswith(".nuspec")]
nuspec = ET.fromstring(package.read(nuspec_names[0]))
ns = {"n": nuspec.tag[1:].split("}")[0]}
metadata = nuspec.find("n:metadata", ns)

if metadata.findtext("n:id", namespaces=ns) != package_id:
    failures.append(f"package id {metadata.findtext('n:id', namespaces=ns)}, expected {package_id}")

lib = {}
for entry in entries:
    parts = entry.split("/")
    if parts[0] == "lib":
        lib.setdefault(parts[1], set()).add("/".join(parts[2:]))
if sorted(lib) != sorted(tfms):
    failures.append(f"lib/ target frameworks {sorted(lib)}, expected {sorted(tfms)}")
for tfm in tfms:
    expected_files = {f"{assembly}.dll", f"{assembly}.xml"}
    files = lib.get(tfm, set())
    if files != expected_files:
        failures.append(f"lib/{tfm}/ holds {sorted(files)}, expected {sorted(expected_files)}")

readme = metadata.findtext("n:readme", namespaces=ns)
if not readme or readme not in entries:
    failures.append(f"readme {readme!r} not bundled")
docs_dir = os.path.join(sdk_dir, "docs")
for root, _, files in os.walk(docs_dir):
    for name in files:
        path = "docs/" + os.path.relpath(os.path.join(root, name), docs_dir).replace(os.sep, "/")
        if path not in entries:
            failures.append(f"{path} not bundled")

for entry in entries:
    if entry.startswith("analyzers/") or ".Generators." in os.path.basename(entry):
        failures.append(f"source generator output bundled: {entry}")


def group_name(tfm):
    return ".NETStandard" + tfm[len("netstandard"):] if tfm.startswith("netstandard") else tfm


groups = {}
for group in metadata.findall("n:dependencies/n:group", ns):
    groups[group.get("targetFramework")] = {
        d.get("id"): d.get("version") for d in group.findall("n:dependency", ns)
    }
if sorted(groups) != sorted(group_name(t) for t in tfms):
    failures.append(f"nuspec dependency groups {sorted(groups)}, expected {sorted(group_name(t) for t in tfms)}")


def minimum_version(version_range):
    return version_range[1:].split(",")[0].strip() if version_range.startswith("[") and version_range.endswith(", )") else version_range


restored = json.load(open(assets_file))["project"]["frameworks"]
for tfm in tfms:
    expected = {
        name: minimum_version(reference.get("version", ""))
        for name, reference in restored.get(tfm, {}).get("dependencies", {}).items()
        if reference.get("target") == "Package" and not reference.get("autoReferenced") and name not in build_only
    }
    actual = groups.get(group_name(tfm), {})
    for dependency, version in sorted(expected.items()):
        if dependency not in actual:
            failures.append(f"{group_name(tfm)}: package reference {dependency} {version} missing from the nuspec")
        elif actual[dependency] != version:
            failures.append(f"{group_name(tfm)}: {dependency} version {actual[dependency]}, project has {version}")
    for dependency in sorted(set(actual) - set(expected)):
        failures.append(f"{group_name(tfm)}: nuspec dependency {dependency} is not a package reference of the SDK project")

for dependency in ("System.Text.Json", "Microsoft.Bcl.AsyncInterfaces"):
    for name, deps in sorted(groups.items()):
        if (dependency in deps) != (name == ".NETStandard2.0"):
            state = "missing from" if name == ".NETStandard2.0" else "listed in"
            failures.append(f"{dependency} {state} the {name} dependency group")

if failures:
    for failure in failures:
        print(f"  - {failure}")
    sys.exit(1)

with open(os.path.join(work_dir, "package.txt"), "w") as out:
    out.write(f"{package_id}\n{metadata.findtext('n:version', namespaces=ns)}\n")
PY

package_id="$(sed -n 1p "${WORK_DIR}/package.txt")"
package_version="$(sed -n 2p "${WORK_DIR}/package.txt")"
print_msg_with_ctx "${SUCCESS_ICON} {green}{bold}Package contents match the SDK project{reset}" \
    "package=${package_id}|version=${package_version}|frameworks=${target_frameworks//$'\n'/;}"

UTILS_FILE="$(find_one Utils/Utilities.cs "${SDK_DIR}" -path '*/Utils/Utilities.cs')"
UTILS_NAMESPACE="$(sed -nE 's/^namespace ([A-Za-z0-9_.]+).*/\1/p' "${UTILS_FILE}" | head -1)"
readonly UTILS_FILE UTILS_NAMESPACE
[[ -n "${UTILS_NAMESPACE}" ]] || fail "no namespace in ${UTILS_FILE}"
if [[ -f "$(dirname "${UTILS_FILE}")/Sse/SseStreamExtensions.cs" ]]; then
    consumer_code="using System.Text;
using System.Text.Json.Serialization.Metadata;
using ${UTILS_NAMESPACE}.Sse;
public static class Consumer {
  public static async System.Threading.Tasks.Task<int> Run() {
    var response = new System.Net.Http.HttpResponseMessage { Content = new System.Net.Http.StringContent(\"data: \\\"a\\\"\\n\\n\", Encoding.UTF8, \"text/event-stream\") };
    var typeInfo = (JsonTypeInfo<string>)System.Text.Json.JsonSerializerOptions.Default.GetTypeInfo(typeof(string));
    var events = 0;
    await foreach (var data in response.AsAsyncEnumerable(typeInfo)) { System.Console.WriteLine(\"event: \" + data); events++; }
    return events == 1 ? 0 : 1;
  }
}"
else
    consumer_code="public static class Consumer {
  public static System.Threading.Tasks.Task<int> Run() {
    System.Console.WriteLine(typeof(${UTILS_NAMESPACE}.Utilities).Assembly.FullName);
    return System.Threading.Tasks.Task.FromResult(0);
  }
}"
fi

consumer_frameworks="${target_frameworks}"
if grep -qx netstandard2.0 <<<"${target_frameworks}" && ! grep -qx net8.0 <<<"${target_frameworks}"; then
    consumer_frameworks+=$'\nnet8.0'
fi

for tfm in ${consumer_frameworks}; do
    consumer_dir="${WORK_DIR}/consumer-${tfm}"
    mkdir -p "${consumer_dir}"
    output_type=Exe
    [[ "${tfm}" == netstandard* ]] && output_type=Library
    cat >"${consumer_dir}/Consumer.csproj" <<EOF
<Project Sdk="Microsoft.NET.Sdk">
  <PropertyGroup>
    <TargetFramework>${tfm}</TargetFramework>
    <OutputType>${output_type}</OutputType>
    <LangVersion>latest</LangVersion>
    <Nullable>enable</Nullable>
    <TreatWarningsAsErrors>true</TreatWarningsAsErrors>
    <WarningsNotAsErrors>${WARNINGS_NOT_AS_ERRORS}</WarningsNotAsErrors>
    <RestoreSources>${FEED};https://api.nuget.org/v3/index.json</RestoreSources>
    <RestorePackagesPath>${WORK_DIR}/packages</RestorePackagesPath>
  </PropertyGroup>
  <ItemGroup>
    <PackageReference Include="${package_id}" Version="${package_version}" />
  </ItemGroup>
</Project>
EOF
    printf '%s\n' "${consumer_code}" >"${consumer_dir}/Consumer.cs"
    [[ "${output_type}" == Exe ]] && printf 'return await Consumer.Run();\n' >"${consumer_dir}/Program.cs"

    print_msg_with_ctx "${BUILD_ICON} {blue}{bold}Building package consumer{reset}" "sdk=${SDK_NAME}|framework=${tfm}|output=${output_type}"
    (cd "${consumer_dir}" && dotnet build -c Release -nodeReuse:false) >"${consumer_dir}/build.log" 2>&1 \
        || { cat "${consumer_dir}/build.log"; fail "consumer ${tfm}: build failed"; }
    if [[ "${output_type}" == Exe ]]; then
        (cd "${consumer_dir}" && dotnet run -c Release --no-build) || fail "consumer ${tfm}: run failed"
    fi
done

print_msg_with_ctx "${SUCCESS_ICON} {green}{bold}Package consumers build and run{reset}" "sdk=${SDK_NAME}|frameworks=${consumer_frameworks//$'\n'/;}"
