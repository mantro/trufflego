#!/bin/bash
set -euo pipefail

# Find path of the script (and follow symlink if needed)
SRC="${BASH_SOURCE[0]}"
REPO=$(cd -- "$(dirname -- "$([ -L "$SRC" ] && readlink -f "$SRC" || echo "$SRC")")" &>/dev/null && pwd)

set -euo pipefail

DIR="$(pwd)"

cd "$REPO"
go build

cd "$DIR"
"$REPO/trufflego" $@
