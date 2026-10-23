#!/usr/bin/env bash
# Example internal-import gate (AGENTS.md "Public API freeze until Page wrap").
#
# Scans examples/**/*.go for imports of
# github.com/chinmay-sawant/ownframe/internal/ and compares them with
# scripts/example-internal-allowlist.txt. Existing pairs pass. A new file
# that imports internal/, a new import path, a missing file, or a stale
# pair fails. Login uses the public clipboard API and must stay off the list.
#
# Usage:
#   bash scripts/check-example-internal.sh
#
# Wired as `make internal-import-check`, which `make lint` calls.
set -euo pipefail

if [ "${BASH_VERSINFO[0]:-0}" -lt 4 ]; then
	echo "internal-import-check: bash 4+ required" >&2
	exit 1
fi

script_dir=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
repo_root=$(CDPATH= cd -- "$script_dir/.." && pwd)
cd "$repo_root"

prefix='github.com/chinmay-sawant/ownframe/internal/'
allowlist=scripts/example-internal-allowlist.txt

if [ ! -f "$allowlist" ]; then
	echo "internal-import-check: missing $allowlist" >&2
	exit 1
fi

prune_dirs=(
	./examples/.git
	./examples/temp
	./examples/node_modules
	./examples/testdata
)

find_expr=(-path "${prune_dirs[0]}")
for dir in "${prune_dirs[@]:1}"; do
	find_expr+=(-o -path "$dir")
done

mapfile -t scanned_files < <(
	find ./examples \( "${find_expr[@]}" \) -prune -o -type f -name '*.go' -print | LC_ALL=C sort
)

declare -A scanned=()
declare -A actual=()
actual_pairs=()
for file in "${scanned_files[@]}"; do
	file=${file#./}
	scanned["$file"]=1
	# Quoted import paths only; comments that mention the prefix still
	# match, which is the conservative side of a freeze.
	while IFS= read -r spec; do
		spec=${spec#\"}
		spec=${spec%\"}
		key="$file"$'\t'"$spec"
		if [ -z "${actual[$key]+set}" ]; then
			actual["$key"]=1
			actual_pairs+=("$key")
		fi
	done < <(grep -oE "\"${prefix}[^\"]+\"" "$file" 2>/dev/null || true)
done

problems=()

allowed_pairs=()
declare -A allowed=()
lineno=0
while IFS= read -r line || [ -n "$line" ]; do
	lineno=$((lineno + 1))
	line=${line%$'\r'}
	case $line in
		'' | '#'*) continue ;;
	esac
	path=${line%%$'\t'*}
	spec=${line#*$'\t'}
	if [ "$spec" = "$line" ]; then
		problems+=("$allowlist:$lineno: expected 'path<TAB>import', got: $line")
		continue
	fi
	path=${path#./}
	if [ -z "$path" ] || [ -z "$spec" ]; then
		problems+=("$allowlist:$lineno: empty path or import")
		continue
	fi
	case $spec in
		"$prefix"*) ;;
		*)
			problems+=("$allowlist:$lineno: import must start with $prefix, got '$spec'")
			continue
			;;
	esac
	key="$path"$'\t'"$spec"
	if [ -n "${allowed[$key]+set}" ]; then
		problems+=("$allowlist:$lineno: duplicate entry for $path -> $spec")
		continue
	fi
	allowed["$key"]=1
	allowed_pairs+=("$key")
done < "$allowlist"

for key in "${actual_pairs[@]}"; do
	if [ -z "${allowed[$key]+set}" ]; then
		path=${key%%$'\t'*}
		spec=${key#*$'\t'}
		problems+=("$path imports $spec and is not in $allowlist; use a public API or add an allowlist entry only for a recorded freeze exception")
	fi
done

for key in "${allowed_pairs[@]}"; do
	path=${key%%$'\t'*}
	spec=${key#*$'\t'}
	if [ ! -e "$path" ]; then
		problems+=("$path is allowlisted for $spec but does not exist; remove the stale entry")
		continue
	fi
	if [ -z "${scanned[$path]+set}" ]; then
		problems+=("$path is allowlisted but was not scanned; remove the entry")
		continue
	fi
	if [ -z "${actual[$key]+set}" ]; then
		problems+=("$path no longer imports $spec (recorded in $allowlist); remove the stale entry")
	fi
done

if [ ${#problems[@]} -gt 0 ]; then
	for problem in "${problems[@]}"; do
		echo "internal-import-check: $problem" >&2
	done
	exit 1
fi

case ${#allowed_pairs[@]} in
	0) echo "internal-import-check: clean (0 allowlisted pairs)." ;;
	1) echo "internal-import-check: clean (1 allowlisted pair)." ;;
	*) echo "internal-import-check: clean (${#allowed_pairs[@]} allowlisted pairs)." ;;
esac
