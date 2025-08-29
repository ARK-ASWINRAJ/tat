#!/usr/bin/env bash
set -euo pipefail

REPO="ARK-ASWINRAJ/tat"
BIN_NAME="tat"
INSTALL_DIR="${HOME}/.local/bin"
TAT_DIR="${HOME}/.tat"

detect_platform() {
  OS="$(uname -s)"
  ARCH="$(uname -m)"
  case "$OS" in
    Linux)   os="linux" ;;
    Darwin)  os="darwin" ;;
    *) echo "Unsupported OS: $OS" >&2; exit 1 ;;
  esac
  case "$ARCH" in
    x86_64|amd64) arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *) echo "Unsupported arch: $ARCH" >&2; exit 1 ;;
  esac
  echo "${os}-${arch}"
}

ensure_path_line() {
  local rc="$1"
  local line='export PATH="$HOME/.local/bin:$PATH"'
  if [ -f "$rc" ]; then
    if ! grep -Fqx "$line" "$rc"; then
      echo "$line" >> "$rc"
      echo "Added PATH to $rc" 
    fi
  else
    echo "$line" >> "$rc"
    echo "Created $rc and added PATH"
  fi
}

install_binary() {
  platform="$(detect_platform)"   # e.g., linux-amd64
  asset="${BIN_NAME}-${platform}"
  url="https://github.com/${REPO}/releases/latest/download/${asset}"

  mkdir -p "$INSTALL_DIR"
  echo "Downloading ${url} ..."
  curl -fsSL -o "${INSTALL_DIR}/${BIN_NAME}" "$url"
  chmod +x "${INSTALL_DIR}/${BIN_NAME}"
  echo "Installed ${BIN_NAME} to ${INSTALL_DIR}/${BIN_NAME}"
}

ensure_path() {
  # Update both rc files safely; user only needs one
  ensure_path_line "${HOME}/.bashrc"
  ensure_path_line "${HOME}/.zshrc"
  export PATH="$HOME/.local/bin:$PATH"
}

install_hooks_if_missing() {
  # If hooks already sourced, skip
  if grep -Fqx 'source ~/.tat/tat.bash' "${HOME}/.bashrc" 2>/dev/null || \
     grep -Fqx 'source ~/.tat/tat.zsh'  "${HOME}/.zshrc"  2>/dev/null; then
    return 0
  fi

  # Run built-in installer if available (your CLI command)
  if command -v tat >/dev/null 2>&1; then
    echo "Installing shell hooks via tat install-shell ..."
    tat install-shell || true
  fi
}

main() {
  install_binary
  ensure_path

  # Initialize and enable
  echo "Initializing TAT ..."
  tat init || true
  # Enable if disabled
  if tat status 2>/dev/null | grep -q 'Enabled: false'; then
    tat enable || true
  fi

  # Install hooks once
  install_hooks_if_missing

  echo ""
  echo "Done. Restart your terminal or run:"
  echo "  source ~/.bashrc   # or source ~/.zshrc"
  echo "Then try: tat search echo"
}

main "$@"
