#!/usr/bin/env bash
# Compares the speed of the working tree with a base commit, for make perf and
# the performance job of CI. It builds BenchmarkRenderClasses on both sides,
# the base getting the head's benchmark and inputs so that both do the same
# work, runs them in turns COUNT times (default 10), and compares them with
# benchstat. It fails when a benchmark is slower than the base by more than
# THRESHOLD percent (default 20) and benchstat finds the difference
# significant. With fewer than 4 rounds no difference can be significant,
# and benchstat gives the spread from 6 rounds on.
#
# Usage: tools/perf/compare.sh <base>
set -euo pipefail

BENCHSTAT="golang.org/x/perf/cmd/benchstat@v0.0.0-20261009192801-be2c69fb417e"
COUNT="${COUNT:-10}"
THRESHOLD="${THRESHOLD:-20}"

if [[ $# -ne 1 ]]; then
	echo "usage: tools/perf/compare.sh <base>" >&2
	exit 2
fi
if [[ ! "${COUNT}" =~ ^[1-9][0-9]*$ || ! "${THRESHOLD}" =~ ^[0-9]+$ ]]; then
	echo "compare.sh: COUNT and THRESHOLD must be whole numbers" >&2
	exit 2
fi

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
base="$(git -C "${ROOT}" rev-parse --verify "$1^{commit}")"
head="$(git -C "${ROOT}" rev-parse HEAD)"

tmp="$(mktemp -d)"
cleanup() {
	git -C "${ROOT}" worktree remove --force "${tmp}/base" 2>/dev/null || true
	git -C "${ROOT}" worktree prune
	rm -rf "${tmp}"
}
trap cleanup EXIT

git -C "${ROOT}" worktree add --quiet --detach "${tmp}/base" "${base}"
cp "${ROOT}/bench_test.go" "${tmp}/base/"
rm -rf "${tmp}/base/testdata/bench"
mkdir -p "${tmp}/base/testdata"
cp -R "${ROOT}/testdata/bench" "${tmp}/base/testdata/"

echo "Base: ${base}"
echo "Head: the working tree at ${head}"
echo

(cd "${ROOT}" && go test -c -o "${tmp}/head.test" .)
if ! (cd "${tmp}/base" && go test -c -o "${tmp}/base.test" .) 2>"${tmp}/base-build.txt"; then
	echo "No comparison: the base does not build the benchmark."
	sed 's/^/  /' "${tmp}/base-build.txt"
	exit 0
fi

args=(-test.run '^$' -test.bench BenchmarkRenderClasses -test.count 1)
for ((i = 1; i <= COUNT; i++)); do
	echo "round ${i}/${COUNT}" >&2
	(cd "${tmp}/base" && "${tmp}/base.test" "${args[@]}") >>"${tmp}/base.txt"
	(cd "${ROOT}" && "${tmp}/head.test" "${args[@]}") >>"${tmp}/head.txt"
done

cd "${tmp}"
go run "${BENCHSTAT}" base.txt head.txt
# In the csv output, a table starts with a line of labels and a line of
# units; the delta of each benchmark is in the sixth column, ~ when the
# difference is not significant.
slower="$(go run "${BENCHSTAT}" -format csv base.txt head.txt 2>/dev/null | awk -F, -v limit="${THRESHOLD}" '
	$1 == "" && $2 == "sec/op" { sec = 1; next }
	$1 == "" { sec = 0; next }
	sec && $1 != "geomean" && $6 ~ /^\+/ {
		d = $6
		sub(/^\+/, "", d)
		sub(/%$/, "", d)
		if (d + 0 > limit) print "  " $1 " " $6
	}
')"

echo
if [[ -n "${slower}" ]]; then
	echo "Slower than the base by more than ${THRESHOLD}%, significantly:"
	echo "${slower}"
	exit 1
fi
echo "No benchmark is significantly slower than the base by more than ${THRESHOLD}%."
