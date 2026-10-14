#!/bin/sh
# footprints installer
#   curl -fsSL https://raw.githubusercontent.com/MahendraDani/footprints/main/install.sh | sh
# Env: FOOTPRINTS_VERSION=v0.1.0 to pin a version, FOOTPRINTS_INSTALL_DIR to change the target.
set -eu

REPO="MahendraDani/footprints"
BIN="footprints"
INSTALL_DIR="${FOOTPRINTS_INSTALL_DIR:-$HOME/.footprints/bin}"
VERSION="${FOOTPRINTS_VERSION:-latest}"

say() { printf '%s\n' "$*"; }
die() { printf 'error: %s\n' "$*" >&2; exit 1; }
need() { command -v "$1" >/dev/null 2>&1 || die "'$1' is required but not installed"; }

detect_platform() {
  case "$(uname -s)" in
    Darwin) OS=darwin ;;
    Linux)  OS=linux ;;
    *) die "unsupported OS: $(uname -s). Download a build from https://github.com/$REPO/releases" ;;
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

add_to_path() {
  case ":$PATH:" in *":$INSTALL_DIR:"*) return 0 ;; esac
  case "$(basename "${SHELL:-}")" in
    zsh)  rc="$HOME/.zshrc" ;;
    bash) if [ -f "$HOME/.bashrc" ]; then rc="$HOME/.bashrc"; else rc="$HOME/.bash_profile"; fi ;;
    *)    rc="" ;;
  esac
  line="export PATH=\"$INSTALL_DIR:\$PATH\""
  if [ -n "$rc" ]; then
    if ! grep -qsF "$INSTALL_DIR" "$rc"; then
      printf '\n# footprints\n%s\n' "$line" >> "$rc"
    fi
    say "Added $INSTALL_DIR to your PATH in $rc. Restart your shell or run: source $rc"
  else
    say "Add this to your shell profile:  $line"
  fi
}

main() {
  need curl
  need tar
  detect_platform

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
  mkdir -p "$INSTALL_DIR"
  mv "$tmp/$BIN" "$INSTALL_DIR/$BIN.new"
  chmod +x "$INSTALL_DIR/$BIN.new"
  mv -f "$INSTALL_DIR/$BIN.new" "$INSTALL_DIR/$BIN" # atomic replace, safe even while the old binary runs

  say "Installed: $("$INSTALL_DIR/$BIN" --version)"
  add_to_path
  say "Try it:  $BIN graph"
}

main "$@"