#!/bin/sh
# footprints installer
#   curl -fsSL https://raw.githubusercontent.com/MahendraDani/footprints/main/install.sh | sh
#
# Env:
#   FOOTPRINTS_VERSION=v0.1.0         install a specific version (default: latest)
#   FOOTPRINTS_INSTALL_DIR=/some/dir  install somewhere specific
#
# Default install location: ~/.local/bin
# If footprints is already installed, it is replaced in place (so `footprints update` keeps working).
set -eu

REPO="MahendraDani/footprints"
BIN="footprints"
VERSION="${FOOTPRINTS_VERSION:-latest}"
INSTALL_DIR=""
PATH_RC=""

say() { printf '%s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "'$1' is required but not installed"; }

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
  need curl
  need tar
  detect_platform
  pick_install_dir

  asset="${BIN}_${OS}_${ARCH}.tar.gz"
  if [ "$VERSION" = "latest" ]; then
    base="https://github.com/$REPO/releases/latest/download"
  else
    base="https://github.com/$REPO/releases/download/$VERSION"
  fi

  tmp="$(mktemp -d)"
  trap 'rm -rf "$tmp"' EXIT

  say "Downloading $asset ($VERSION)..."
  curl -fsSL "$base/$asset" -o "$tmp/$asset" || die "download failed: $base/$asset"
  curl -fsSL "$base/checksums.txt" -o "$tmp/checksums.txt" || die "could not download checksums.txt"

  want="$(grep " $asset\$" "$tmp/checksums.txt" | cut -d' ' -f1)"
  [ -n "$want" ] || die "no checksum listed for $asset"
  got="$(sha256_of "$tmp/$asset")"
  [ "$want" = "$got" ] || die "checksum mismatch for $asset"

  tar -xzf "$tmp/$asset" -C "$tmp" "$BIN"
  mkdir -p "$INSTALL_DIR" || die "could not create $INSTALL_DIR"
  mv "$tmp/$BIN" "$INSTALL_DIR/$BIN.new"
  chmod +x "$INSTALL_DIR/$BIN.new"
  mv -f "$INSTALL_DIR/$BIN.new" "$INSTALL_DIR/$BIN" # atomic replace, safe even while the old binary runs

  say "Installed $("$INSTALL_DIR/$BIN" --version) to $INSTALL_DIR"
  add_to_path

  # Warn if some other copy earlier on PATH will win
  resolved="$(command -v "$BIN" 2>/dev/null || true)"
  if [ -n "$resolved" ] && [ "$resolved" != "$INSTALL_DIR/$BIN" ]; then
    say ""
    say "Note: your shell currently runs $resolved instead of $INSTALL_DIR/$BIN."
    say "Remove the old one to use this version."
  fi

  say ""
  if [ -n "$PATH_RC" ]; then
    say "Last step: footprints isn't on this terminal's PATH yet."
    say "Open a new terminal tab, or run:"
    say ""
    say "    source $PATH_RC"
    say ""
    say "Then try:  $BIN graph"
  else
    say "Ready. Try:  $BIN graph"
  fi
}

main "$@"