#!/bin/bash
set -euo pipefail

# Find path of the script (and follow symlink if needed)
REPO=$(cd -- "$(dirname -- "$([ -L "$SRC" ] && readlink -f "$SRC" || echo "$SRC")")" &>/dev/null && pwd)

set -euo pipefail

DIR="$(pwd)"

cd "$REPO"
go build

cd "$DIR"
"$REPO/trufflego" $@
