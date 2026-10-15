#!/bin/sh
# footprints uninstaller
#   curl -fsSL https://raw.githubusercontent.com/MahendraDani/footprints/main/uninstall.sh | sh
#
# By default this removes:
#   1. the footprints binary
#   2. the PATH line the installer added to your shell config
#   3. the ~/.footprints workspace (cached Chrome history copy, generated graphs, settings)
#
# To keep the workspace:
#   curl -fsSL https://raw.githubusercontent.com/MahendraDani/footprints/main/uninstall.sh | sh -s -- --keep-data
#   (or set FOOTPRINTS_KEEP_DATA=1)
set -eu

BIN="footprints"
DATA_DIR="$HOME/.footprints"
KEEP_DATA="${FOOTPRINTS_KEEP_DATA:-}"

say() { printf '%s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }

for arg in "$@"; do
  case "$arg" in
    --keep-data) KEEP_DATA=1 ;;
    --purge) ;;  # old flag, now the default; accepted so existing instructions keep working
    -h|--help)
      say "usage: uninstall.sh [--keep-data]"
      say "  --keep-data  keep $DATA_DIR (cached history copy, graphs, settings)"
      say "  default      remove the binary, the PATH entry, and $DATA_DIR"
      exit 0 ;;
    *) die "unknown option: $arg" ;;
  esac
done

remove_binaries() {
  found=0
  for f in "$(command -v "$BIN" 2>/dev/null || true)" \
           "${FOOTPRINTS_INSTALL_DIR:+$FOOTPRINTS_INSTALL_DIR/$BIN}" \
           "$HOME/.local/bin/$BIN" \
           "$HOME/.footprints/bin/$BIN" \
           "/usr/local/bin/$BIN"; do
    [ -n "$f" ] && [ -f "$f" ] || continue
    if rm -f "$f" 2>/dev/null; then
      say "  removed $f"
      found=1
    else
      say "  could not remove $f (try: sudo rm $f)"
    fi
  done
  [ "$found" = 1 ] || say "  no footprints binary found"
}

# Remove the "# footprints" comment and the PATH line right after it (the pair install.sh appends).
clean_rc() {
  rc="$1"
  [ -f "$rc" ] || return 0
  grep -qx '# footprints' "$rc" || return 0
  t="$(mktemp)"
  awk '
    held { held = 0; if ($0 ~ /PATH|fish_add_path/) next; print "# footprints" }
    $0 == "# footprints" { held = 1; next }
    { print }
    END { if (held) print "# footprints" }
  ' "$rc" > "$t"
  cat "$t" > "$rc"   # cat keeps the file's permissions and symlinks intact
  rm -f "$t"
  say "  cleaned $rc"
}

remove_workspace() {
  if [ ! -d "$DATA_DIR" ]; then
    say "  nothing to delete ($DATA_DIR doesn't exist)"
    return 0
  fi
  # Safety: only ever delete a real directory named .footprints directly under $HOME
  [ -n "$HOME" ] && [ "$HOME" != "/" ] && [ ! -L "$DATA_DIR" ] || die "refusing to delete $DATA_DIR"
  rm -rf "$DATA_DIR"
  say "  deleted $DATA_DIR"
}

say "[1/3] Removing the footprints binary"
remove_binaries

say ""
say "[2/3] Cleaning shell config"
for rc in "${ZDOTDIR:-$HOME}/.zshrc" "$HOME/.zshrc" "$HOME/.bashrc" "$HOME/.bash_profile" "$HOME/.config/fish/config.fish"; do
  clean_rc "$rc"
done

say ""
say "[3/3] Workspace"
if [ -n "$KEEP_DATA" ]; then
  if [ -d "$DATA_DIR" ]; then
    say "  kept $DATA_DIR (a cached copy of your Chrome history and generated graphs)"
    say "  to delete it later:  rm -rf $DATA_DIR"
  else
    say "  nothing to keep ($DATA_DIR doesn't exist)"
  fi
else
  remove_workspace
fi

say ""
say "Done."