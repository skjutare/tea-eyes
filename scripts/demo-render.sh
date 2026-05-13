#!/usr/bin/env bash
# Render a TUI binary to a PNG via tea-eyes' embedded VHS tape generator.
# Usage: scripts/demo-render.sh <binary> <out.png>
set -euo pipefail

if [ "$#" -ne 2 ]; then
  echo "usage: $0 <binary> <out.png>" >&2
  exit 2
fi

BIN="$1"
OUT="$2"
WORK="$(mktemp -d)"
trap 'rm -rf "$WORK"' EXIT

cat > "$WORK/demo.tape" <<EOF
Output $OUT.gif
Set Shell "bash"
Set FontFamily "JetBrains Mono"
Set FontSize 14
Set Theme "Dracula"
Set Padding 20
Set Width 800
Set Height 480
Hide
Type "exec $BIN"
Enter
Show
Sleep 600ms
Tab
Sleep 500ms
Screenshot $OUT
EOF

vhs "$WORK/demo.tape"
rm -f "$OUT.gif"
echo "wrote $OUT"
