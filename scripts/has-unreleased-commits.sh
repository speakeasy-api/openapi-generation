set -euo pipefail

latest_tag=$(git describe --tags --abbrev=0 2>/dev/null || true)
if [[ -n "$latest_tag" && "$(git rev-parse HEAD)" == "$(git rev-parse "$latest_tag^{commit}")" ]]; then
  echo 'has_commits=false'
else
  echo 'has_commits=true'
fi
