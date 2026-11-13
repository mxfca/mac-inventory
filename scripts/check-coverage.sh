#!/usr/bin/env bash
# Fails if coverage in coverage.out (written by the go-test hook) is below MIN_COVERAGE,
# both in total and on the lines changed since BASE_REF (SonarCloud's quality gate looks at new code).
set -euo pipefail

min="${MIN_COVERAGE:-80}"
profile="${1:-coverage.out}"
base="${BASE_REF:-origin/main}"

total=$(go tool cover -func="$profile" | awk '/^total:/ { sub(/%/, "", $3); print $3 }')
echo "coverage: ${total}% (min ${min}%)"
awk -v t="$total" -v m="$min" 'BEGIN { exit (t + 0 < m) }'

if ! git rev-parse --verify --quiet "$base" >/dev/null; then
	echo "new code coverage: skipped ($base not found)"
	exit 0
fi

module=$(go list -m)
# Pass 1 reads the changed lines of non-test Go files ("file:line"); pass 2 reads the profile
# blocks ("file:start.col,end.col statements count"). A changed line counts once, and is
# covered when any block holding it ran.
git diff -U0 "$base" -- '*.go' ':!*_test.go' | awk -v module="$module" -v min="$min" '
	/^\+\+\+ b\// { file = substr($0, 7); next }
	/^@@/ {
		split($3, r, ","); start = substr(r[1], 2) + 0; n = (r[2] == "") ? 1 : r[2] + 0
		for (i = 0; i < n; i++) changed[file ":" (start + i)] = 1
	}
	END {
		while ((getline line < "'"$profile"'") > 0) {
			if (line ~ /^mode:/) continue
			split(line, f, " "); split(f[1], loc, ":"); split(loc[2], span, ",")
			split(span[1], s, "."); split(span[2], e, ".")
			file = loc[1]; sub("^" module "/", "", file)
			for (l = s[1]; l <= e[1]; l++) {
				key = file ":" l
				if (!(key in changed)) continue
				seen[key] = 1
				if (f[3] > 0) hit[key] = 1
			}
		}
		for (k in seen) { toCover++; if (k in hit) covered++ }
		if (toCover == 0) { print "new code coverage: no changed lines to cover"; exit 0 }
		pct = 100 * covered / toCover
		printf "new code coverage: %.1f%% of %d lines (min %s%%)\n", pct, toCover, min
		exit (pct < min)
	}'
