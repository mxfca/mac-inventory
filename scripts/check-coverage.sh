#!/usr/bin/env bash
# Fails if the total coverage in coverage.out (written by the go-test hook) is below MIN_COVERAGE.
set -euo pipefail

min="${MIN_COVERAGE:-80}"
profile="${1:-coverage.out}"

total=$(go tool cover -func="$profile" | awk '/^total:/ { sub(/%/, "", $3); print $3 }')
echo "coverage: ${total}% (min ${min}%)"
awk -v t="$total" -v m="$min" 'BEGIN { exit (t + 0 < m) }'
