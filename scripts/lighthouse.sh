#!/usr/bin/env bash
# Audit the deployed app with the Lighthouse CLI and write HTML + JSON reports.
# Same task whether it is run by the nightly workflow, by `gh workflow run`, or by hand.
#
#   ./scripts/lighthouse.sh https://example.com
#   LIGHTHOUSE_URL=https://example.com ./scripts/lighthouse.sh
set -euo pipefail

LIGHTHOUSE_VERSION="${LIGHTHOUSE_VERSION:-12}"
WARMUP_ATTEMPTS="${WARMUP_ATTEMPTS:-60}"
WARMUP_DELAY_SECONDS="${WARMUP_DELAY_SECONDS:-5}"

url="${1:-${LIGHTHOUSE_URL:-}}"
if [ -z "$url" ]; then
  echo "error: no target URL." >&2
  echo "Pass one as an argument or set LIGHTHOUSE_URL." >&2
  echo "In CI it comes from the LIGHTHOUSE_URL repository variable:" >&2
  echo "  gh variable set LIGHTHOUSE_URL --body 'https://<your-render-host>'" >&2
  exit 1
fi

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
run_dir="$repo_root/.lighthouse/$(date -u +%Y-%m-%dT%H-%M-%SZ)"
mkdir -p "$run_dir"

# Render's free plan spins the service down when idle. Without this the cold start
# is measured as page load time and the performance score is meaningless.
echo "Waking $url ..."
for attempt in $(seq 1 "$WARMUP_ATTEMPTS"); do
  if curl --fail --silent --show-error --output /dev/null --max-time 30 "$url"; then
    echo "Target is up after ${attempt} attempt(s)."
    break
  fi
  if [ "$attempt" -eq "$WARMUP_ATTEMPTS" ]; then
    echo "error: $url did not respond after ${WARMUP_ATTEMPTS} attempts." >&2
    exit 1
  fi
  sleep "$WARMUP_DELAY_SECONDS"
done

for preset in mobile desktop; do
  echo "Running Lighthouse (${preset}) ..."

  # Mobile is Lighthouse's default form factor, so it has no preset flag of its own.
  args=(
    "$url"
    --output=html
    --output=json
    --output-path="$run_dir/$preset"
    --chrome-flags="--headless=new --no-sandbox --disable-dev-shm-usage"
    --quiet
  )
  if [ "$preset" = "desktop" ]; then
    args+=(--preset=desktop)
  fi

  npx --yes "lighthouse@${LIGHTHOUSE_VERSION}" "${args[@]}"
done

# Print the scores as a markdown table, and append the same table to the GitHub job
# summary when running in Actions, so the numbers are readable without downloading
# the HTML report.
summary="$(node "$repo_root/scripts/lighthouse-summary.mjs" "$run_dir" "$url")"
echo "$summary"
if [ -n "${GITHUB_STEP_SUMMARY:-}" ]; then
  echo "$summary" >>"$GITHUB_STEP_SUMMARY"
fi

echo "Reports written to $run_dir"
