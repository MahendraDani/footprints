#!/bin/sh
# footprints installer
#   curl -fsSL https://raw.githubusercontent.com/MahendraDani/footprints/main/install.sh | sh
#
# Env:
#   FOOTPRINTS_VERSION=v0.1.0         install a specific version (default: latest)
#   FOOTPRINTS_INSTALL_DIR=/some/dir  install somewhere specific
#   FOOTPRINTS_DEBUG=1                print every command (for bug reports)
#
# Default install location: ~/.local/bin
# If footprints is already installed, it is replaced in place (so `footprints update` keeps working).
set -eu
[ -n "${FOOTPRINTS_DEBUG:-}" ] && set -x

REPO="MahendraDani/footprints"
BIN="footprints"
VERSION="${FOOTPRINTS_VERSION:-latest}"
INSTALL_DIR=""
PATH_RC=""

say() { printf '%s\n' "$*"; }
step() { printf '\n[%s] %s\n' "$1" "$2"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "'$1' is required but not installed"; }

# fetch URL OUT [progress]
# Never hangs silently: 15s to connect, aborts if the transfer drops below 1 KB/s for 20s,
# then retries once over HTTP/1.1 (works around flaky HTTP/2 connections).
fetch() {
  url="$1"; out="$2"; show="${3:-}"
  for opts in "" "--http1.1"; do
    if [ "$show" = "progress" ] && [ -t 2 ]; then
      # shellcheck disable=SC2086
      curl -fL --progress-bar $opts --connect-timeout 15 --speed-limit 1000 --speed-time 20 \
        "$url" -o "$out" && return 0
    else
      # shellcheck disable=SC2086
      curl -fsSL $opts --connect-timeout 15 --speed-limit 1000 --speed-time 20 \
        "$url" -o "$out" && return 0
    fi
    [ -z "$opts" ] && say "  download failed or stalled, retrying over HTTP/1.1..." >&2
  done
  return 1
}


detect_platform() {
  case "$(uname -s)" in
    Darwin) OS=darwin ;;
    Linux)  OS=linux ;;
    *) die "unsupported OS: $(uname -s). On Windows, use install.ps1 (see the README)." ;;
  esac
  case "$(uname -m)" in
    x86_64|amd64)  ARCH=amd64 ;;
    arm64|aarch64) ARCH=arm64 ;;
    *) die "unsupported architecture: $(uname -m)" ;;
  esac
}

sha256_of() {
  if command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  elif command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | cut -d' ' -f1
  else
    die "need sha256sum or shasum to verify the download"
  fi
}

# Where to install:
#   1. FOOTPRINTS_INSTALL_DIR, if set
#   2. the folder of an existing install, so updates replace the binary in place
#   3. ~/.local/bin
pick_install_dir() {
  if [ -n "${FOOTPRINTS_INSTALL_DIR:-}" ]; then
    INSTALL_DIR="$FOOTPRINTS_INSTALL_DIR"
    return 0
  fi

  existing="$(command -v "$BIN" 2>/dev/null || true)"
  if [ -n "$existing" ] && [ -f "$existing" ] && [ ! -L "$existing" ] && [ -w "$(dirname "$existing")" ]; then
    INSTALL_DIR="$(dirname "$existing")"
    return 0
  fi

  INSTALL_DIR="$HOME/.local/bin"
}

# If INSTALL_DIR isn't on PATH, add it to the user's shell config.
# Sets PATH_RC to the file we touched so main() can tell the user what to source.
add_to_path() {
  case ":$PATH:" in *":$INSTALL_DIR:"*) return 0 ;; esac

  # Write $HOME instead of the expanded path, so the config line stays portable
  case "$INSTALL_DIR" in
    "$HOME"/*) dir_expr="\$HOME${INSTALL_DIR#"$HOME"}" ;;
    *)         dir_expr="$INSTALL_DIR" ;;
  esac

  line="export PATH=\"$dir_expr:\$PATH\""
  case "$(basename "${SHELL:-}")" in
    zsh)  rc="${ZDOTDIR:-$HOME}/.zshrc" ;;
    bash)
      if [ "$OS" = "darwin" ]; then rc="$HOME/.bash_profile"
      elif [ -f "$HOME/.bashrc" ]; then rc="$HOME/.bashrc"
      else rc="$HOME/.bash_profile"; fi ;;
    fish)
      rc="$HOME/.config/fish/config.fish"
      line="fish_add_path \"$INSTALL_DIR\""
      mkdir -p "$(dirname "$rc")" ;;
    *) rc="" ;;
  esac

  if [ -z "$rc" ]; then
    say ""
    say "Add this line to your shell profile, then restart your terminal:"
    say "    $line"
    return 0
  fi

  if ! grep -qsF "$dir_expr" "$rc" && ! grep -qsF "$INSTALL_DIR" "$rc"; then
    printf '\n# footprints\n%s\n' "$line" >> "$rc"
  fi
  PATH_RC="$rc"
}

main() {
  say "footprints installer"

  step "1/5" "Checking your system"
  need curl
  need tar
  detect_platform
  pick_install_dir
  say "  platform: $OS/$ARCH"
  say "  install to: $INSTALL_DIR"

  asset="${BIN}_${OS}_${ARCH}.tar.gz"
  if [ "$VERSION" = "latest" ]; then
    base="https://github.com/$REPO/releases/latest/download"
  else
    base="https://github.com/$REPO/releases/download/$VERSION"
  fi

  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  step "2/5" "Downloading $asset ($VERSION)"
  say "  $base/$asset"
  fetch "$base/$asset" "$tmp/$asset" progress || die "download failed. Check your connection and try again, or set FOOTPRINTS_DEBUG=1 and re-run to see details."

  step "3/5" "Verifying checksum"
  say "  fetching checksums.txt..."
  fetch "$base/checksums.txt" "$tmp/checksums.txt" || die "could not download checksums.txt"
  say "  computing SHA-256..."
  want="$(grep " $asset\$" "$tmp/checksums.txt" | cut -d' ' -f1)"
  [ -n "$want" ] || die "no checksum listed for $asset"
  got="$(sha256_of "$tmp/$asset")"
  [ "$want" = "$got" ] || die "checksum mismatch for $asset (expected $want, got $got)"
  say "  ok"

  step "4/5" "Installing"
  tar -xzf "$tmp/$asset" -C "$tmp" "$BIN"
  mkdir -p "$INSTALL_DIR" || die "could not create $INSTALL_DIR"
  mv "$tmp/$BIN" "$INSTALL_DIR/$BIN.new"
  chmod +x "$INSTALL_DIR/$BIN.new"
  mv -f "$INSTALL_DIR/$BIN.new" "$INSTALL_DIR/$BIN" # atomic replace, safe even while the old binary runs

  # Check the binary runs. stdin is closed (it would otherwise be this script's pipe),
  # and network features are off so this can't block or print a first-run notice.
  ver="$(FOOTPRINTS_NO_UPDATE_CHECK=1 FOOTPRINTS_NO_TELEMETRY=1 "$INSTALL_DIR/$BIN" --version </dev/null 2>&1)" \
    || die "the installed binary failed to run: $ver"
  say "  $ver"

  step "5/5" "Setting up your PATH"
  add_to_path
  if [ -z "$PATH_RC" ]; then
    say "  $INSTALL_DIR is already on your PATH"
  else
    say "  added $INSTALL_DIR to $PATH_RC"
  fi

  # Warn if some other copy earlier on PATH will win
  resolved="$(command -v "$BIN" 2>/dev/null || true)"
  if [ -n "$resolved" ] && [ "$resolved" != "$INSTALL_DIR/$BIN" ]; then
    say ""
    say "Note: your shell currently runs $resolved instead of $INSTALL_DIR/$BIN."
    say "Remove the old one to use this version."
  fi

  say ""
  say "Done! Installed $ver"
  if [ -n "$PATH_RC" ]; then
    say ""
    say "Last step: footprints isn't on this terminal's PATH yet."
    say "Open a new terminal tab, or run:"
    say ""
    say "    source $PATH_RC"
    say ""
    say "Then try:  $BIN graph"
  else
    say "Try:  $BIN graph"
  fi
}

main "$@"