#!/usr/bin/env bash
# Exercise version selection without downloading a tool or changing PATH binaries.
set -euo pipefail
cd -- "$(dirname -- "$0")/.."
mkdir -p temp
fixture=$(mktemp -d "$PWD/temp/lint-version-test.XXXXXX")
trap 'rm -rf -- "$fixture"' EXIT
mkdir -p "$fixture/scripts" "$fixture/bin"
cp scripts/lint.sh "$fixture/scripts/"
export LINT_TEST_LOG="$fixture/calls"
cat > "$fixture/bin/golangci-lint" <<'TOOL'
#!/usr/bin/env bash
if [ "$1" = version ] && [ "${2:-}" = --json ]; then
    echo '{"version":"0.0.0"}'
    exit 0
fi
echo 'wrong PATH runner executed' >&2
exit 1
TOOL
cat > "$fixture/bin/go" <<'GO'
#!/usr/bin/env bash
set -euo pipefail
printf 'go %s\n' "$*" >> "$LINT_TEST_LOG"
cat > "$GOBIN/golangci-lint" <<'PINNED'
#!/usr/bin/env bash
if [ "$1" = version ]; then
    if [ "${2:-}" = --json ]; then
        echo '{"version":"2.15.0"}'
    else
        echo 'golangci-lint has version 2.15.0'
    fi
else
    printf 'lint %s\n' "$*" >> "$LINT_TEST_LOG"
fi
PINNED
chmod +x "$GOBIN/golangci-lint"
GO
chmod +x "$fixture/bin/go" "$fixture/bin/golangci-lint"
PATH="$fixture/bin:$PATH" bash "$fixture/scripts/lint.sh" v2.15.0
PATH="$fixture/bin:$PATH" bash "$fixture/scripts/lint.sh" v2.15.0 --version
python3 - "$LINT_TEST_LOG" <<'PY'
from pathlib import Path
import sys
calls = Path(sys.argv[1]).read_text().splitlines()
assert calls == [
    'go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@v2.15.0',
    'lint run ./...',
    'lint run ./examples/...',
], calls
print('lint-version-test: wrong PATH version replaced; exact cached version reused')
PY
