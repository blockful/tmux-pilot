#!/usr/bin/env bash
# tmux-pilot installer
#
# Install:     curl -fsSL https://raw.githubusercontent.com/blockful/tmux-pilot/main/install.sh | bash
# Uninstall:   curl -fsSL https://raw.githubusercontent.com/blockful/tmux-pilot/main/install.sh | bash -s -- --uninstall
# Update:      just run the install command again
#
# Environment variables:
#   INSTALL_DIR  — where to put the binary (default: ~/.local/bin)
#   VERSION      — specific version to install (default: latest)

set -euo pipefail

REPO="blockful/tmux-pilot"
BINARY="tmux-pilot"
DEFAULT_DIR="${HOME}/.local/bin"
INSTALL_DIR="${INSTALL_DIR:-$DEFAULT_DIR}"

# --- helpers ---

info()  { printf '\033[1;34m::\033[0m %s\n' "$*"; }
ok()    { printf '\033[1;32m✓\033[0m %s\n' "$*"; }
fail()  { printf '\033[1;31m✗\033[0m %s\n' "$*" >&2; exit 1; }

detect_os() {
  case "$(uname -s)" in
    Linux*)  echo "Linux" ;;
    Darwin*) echo "Darwin" ;;
    *)       fail "Unsupported OS: $(uname -s). Only Linux and macOS are supported." ;;
  esac
}

detect_arch() {
  case "$(uname -m)" in
    x86_64|amd64)  echo "x86_64" ;;
    arm64|aarch64) echo "arm64" ;;
    *)             fail "Unsupported architecture: $(uname -m). Only x86_64 and arm64 are supported." ;;
  esac
}

latest_version() {
  local url="https://api.github.com/repos/${REPO}/releases/latest"
  if command -v curl &>/dev/null; then
    curl -fsSL "$url" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//'
  elif command -v wget &>/dev/null; then
    wget -qO- "$url" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//'
  else
    fail "Neither curl nor wget found. Install one and retry."
  fi
}

download() {
  local url="$1" dest="$2"
  if command -v curl &>/dev/null; then
    curl -fsSL -o "$dest" "$url"
  elif command -v wget &>/dev/null; then
    wget -qO "$dest" "$url"
  fi
}

verify_checksum() {
  local archive="$1" version="$2" expected_file="$3"
  local checksums_url="https://github.com/${REPO}/releases/download/${version}/checksums.txt"
  local tmpcheck
  tmpcheck="$(mktemp)"
  download "$checksums_url" "$tmpcheck" || { rm -f "$tmpcheck"; return 0; }

  local expected
  expected="$(grep "$expected_file" "$tmpcheck" | awk '{print $1}')"
  rm -f "$tmpcheck"

  if [ -z "$expected" ]; then
    return 0
  fi

  local actual
  if command -v sha256sum &>/dev/null; then
    actual="$(sha256sum "$archive" | awk '{print $1}')"
  elif command -v shasum &>/dev/null; then
    actual="$(shasum -a 256 "$archive" | awk '{print $1}')"
  else
    return 0
  fi

  if [ "$actual" != "$expected" ]; then
    fail "Checksum mismatch!\n  Expected: ${expected}\n  Got:      ${actual}"
  fi
}

# --- uninstall ---

uninstall() {
  info "Uninstalling ${BINARY}..."

  local found=false

  if [ -f "${INSTALL_DIR}/${BINARY}" ]; then
    rm -f "${INSTALL_DIR}/${BINARY}"
    ok "Removed ${INSTALL_DIR}/${BINARY}"
    found=true
  fi

  if [ -L "${INSTALL_DIR}/tp" ]; then
    rm -f "${INSTALL_DIR}/tp"
    ok "Removed ${INSTALL_DIR}/tp symlink"
  fi

  # Also check common locations
  for dir in /usr/local/bin "${HOME}/.local/bin" "${HOME}/bin"; do
    if [ "$dir" = "$INSTALL_DIR" ]; then
      continue
    fi
    if [ -f "${dir}/${BINARY}" ]; then
      if rm -f "${dir}/${BINARY}" 2>/dev/null; then
        ok "Removed ${dir}/${BINARY}"
      else
        info "Need sudo to remove ${dir}/${BINARY}"
        sudo rm -f "${dir}/${BINARY}" && ok "Removed ${dir}/${BINARY}"
      fi
      found=true
    fi
    if [ -L "${dir}/tp" ]; then
      if rm -f "${dir}/tp" 2>/dev/null; then
        ok "Removed ${dir}/tp symlink"
      else
        sudo rm -f "${dir}/tp" && ok "Removed ${dir}/tp symlink"
      fi
    fi
  done

  if [ "$found" = true ]; then
    ok "tmux-pilot uninstalled"
  else
    info "tmux-pilot not found in expected locations"
    info "If installed elsewhere, remove the 'tmux-pilot' and 'tp' binaries manually"
  fi
}

# --- install ---

install() {
  info "Installing ${BINARY}..."

  local os arch version archive_name url tmpdir

  os="$(detect_os)"
  arch="$(detect_arch)"

  if [ -n "${VERSION:-}" ]; then
    version="${VERSION}"
    [[ "$version" == v* ]] || version="v${version}"
  else
    info "Fetching latest version..."
    version="$(latest_version)"
  fi

  [ -z "$version" ] && fail "Could not determine latest version."

  # Check if already installed at this version
  if command -v "$BINARY" &>/dev/null; then
    local current
    current="$("$BINARY" --version 2>/dev/null | awk '{print $2}')" || true
    if [ "v${current}" = "$version" ] || [ "$current" = "$version" ]; then
      ok "Already up to date (${version})"
      return 0
    fi
    info "Updating from ${current} → ${version}"
  else
    info "Version: ${version}"
  fi

  archive_name="${BINARY}_${os}_${arch}.tar.gz"
  url="https://github.com/${REPO}/releases/download/${version}/${archive_name}"

  tmpdir="$(mktemp -d)"
  trap 'rm -rf "$tmpdir"' EXIT

  info "Downloading ${archive_name}..."
  download "$url" "${tmpdir}/${archive_name}" || fail "Download failed. Check that ${version} exists at ${url}"

  info "Verifying checksum..."
  verify_checksum "${tmpdir}/${archive_name}" "$version" "$archive_name"
  ok "Checksum verified"

  info "Extracting..."
  tar -xzf "${tmpdir}/${archive_name}" -C "$tmpdir"

  [ -f "${tmpdir}/${BINARY}" ] || fail "Binary not found in archive."

  mkdir -p "$INSTALL_DIR"
  mv "${tmpdir}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
  chmod +x "${INSTALL_DIR}/${BINARY}"

  # Create 'tp' shortcut symlink
  ln -sf "${INSTALL_DIR}/${BINARY}" "${INSTALL_DIR}/tp"

  ok "Installed ${BINARY} ${version} to ${INSTALL_DIR}/${BINARY}"
  ok "Shortcut: 'tp' is ready to use"

  # Check PATH
  if ! echo "$PATH" | tr ':' '\n' | grep -qx "$INSTALL_DIR"; then
    echo ""
    info "Add ${INSTALL_DIR} to your PATH:"
    echo ""
    echo "  echo 'export PATH=\"${INSTALL_DIR}:\$PATH\"' >> ~/.bashrc"
    echo "  source ~/.bashrc"
    echo ""
  fi

  "${INSTALL_DIR}/${BINARY}" --version 2>/dev/null && ok "Ready to use!" || true
}

# --- entry ---

case "${1:-}" in
  --uninstall|-u)
    uninstall
    ;;
  --help|-h)
    echo "tmux-pilot installer"
    echo ""
    echo "Install:     curl -fsSL https://raw.githubusercontent.com/blockful/tmux-pilot/main/install.sh | bash"
    echo "Update:      same command (detects current version)"
    echo "Uninstall:   curl -fsSL ... | bash -s -- --uninstall"
    echo ""
    echo "Options:"
    echo "  --uninstall, -u    Remove tmux-pilot and tp symlink"
    echo "  --help, -h         Show this help"
    echo ""
    echo "Environment:"
    echo "  INSTALL_DIR        Install location (default: ~/.local/bin)"
    echo "  VERSION            Specific version to install (default: latest)"
    ;;
  *)
    install
    ;;
esac
