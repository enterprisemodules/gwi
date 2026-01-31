#!/usr/bin/env bash
#
# gwi installer
# Install with: curl -fsSL https://raw.githubusercontent.com/enterprisemodules/gwi/main/install.sh | bash
#

set -euo pipefail

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info() { echo -e "${BLUE}→${NC} $*"; }
success() { echo -e "${GREEN}✓${NC} $*"; }
warn() { echo -e "${YELLOW}!${NC} $*"; }
die() { echo -e "${RED}Error:${NC} $*" >&2; exit 1; }

# Configuration
INSTALL_DIR="${HOME}/.local/bin"
GWI_URL="https://raw.githubusercontent.com/enterprisemodules/gwi/main/gwi"
COMPLETION_URL="https://raw.githubusercontent.com/enterprisemodules/gwi/main/completions/_gwi"

# Shell function to add
SHELL_FUNCTION='
# gwi - Git Worktree Issue CLI shell integration
gwi() {
  if [[ "$1" == "cd" ]]; then
    shift
    local path=$(command gwi _cd "$@")
    [[ -d "$path" ]] && cd "$path" || echo "Not found" >&2
  elif [[ "$1" == "main" ]]; then
    local path=$(command gwi _main)
    [[ -d "$path" ]] && cd "$path" || echo "Not found" >&2
  elif [[ "$1" == "list" ]]; then
    local path=$(command gwi _list)
    [[ -n "$path" && -d "$path" ]] && cd "$path"
  else
    command gwi "$@"
  fi
}'

# Detect if running from local clone or remote
is_local_install() {
  [[ -f "$(dirname "$0")/gwi" ]]
}

# Check dependencies
check_deps() {
  for cmd in git gh jq; do
    command -v "$cmd" &>/dev/null || die "Required command not found: $cmd"
  done
  success "Dependencies found: git, gh, jq"
}

# Install main script
install_script() {
  info "Installing gwi to $INSTALL_DIR..."

  mkdir -p "$INSTALL_DIR"

  if is_local_install; then
    # Local install from repo clone
    cp "$(dirname "$0")/gwi" "$INSTALL_DIR/gwi"
  else
    # Remote install
    curl -fsSL "$GWI_URL" -o "$INSTALL_DIR/gwi"
  fi

  chmod +x "$INSTALL_DIR/gwi"
  success "Installed gwi script"
}

# Install completions
install_completions() {
  local zsh_completions="${HOME}/.zsh/completions"

  if [[ -n "${ZSH_VERSION:-}" ]] || [[ -d "$HOME/.oh-my-zsh" ]] || [[ -f "$HOME/.zshrc" ]]; then
    info "Installing Zsh completions..."
    mkdir -p "$zsh_completions"

    if is_local_install; then
      cp "$(dirname "$0")/completions/_gwi" "$zsh_completions/_gwi" 2>/dev/null || true
    else
      curl -fsSL "$COMPLETION_URL" -o "$zsh_completions/_gwi" 2>/dev/null || true
    fi

    # Add completions directory to fpath if not already there
    if ! grep -q 'fpath.*\.zsh/completions' "$HOME/.zshrc" 2>/dev/null; then
      echo 'fpath=(~/.zsh/completions $fpath)' >> "$HOME/.zshrc"
    fi

    success "Installed Zsh completions"
  fi
}

# Add shell integration
add_shell_integration() {
  local shell_rc

  # Detect shell
  if [[ -n "${ZSH_VERSION:-}" ]] || [[ "$SHELL" == *"zsh"* ]]; then
    shell_rc="$HOME/.zshrc"
  elif [[ -n "${BASH_VERSION:-}" ]] || [[ "$SHELL" == *"bash"* ]]; then
    shell_rc="$HOME/.bashrc"
  else
    warn "Unknown shell. Add the following to your shell config manually:"
    echo "$SHELL_FUNCTION"
    return
  fi

  # Check if already installed
  if grep -q "gwi.*Git Worktree Issue" "$shell_rc" 2>/dev/null; then
    info "Shell integration already present in $shell_rc"
    return
  fi

  info "Adding shell integration to $shell_rc..."
  echo "$SHELL_FUNCTION" >> "$shell_rc"
  success "Added shell integration"
}

# Ensure ~/.local/bin is in PATH
check_path() {
  if [[ ":$PATH:" != *":$INSTALL_DIR:"* ]]; then
    warn "$INSTALL_DIR is not in your PATH"

    local shell_rc
    if [[ "$SHELL" == *"zsh"* ]]; then
      shell_rc="$HOME/.zshrc"
    else
      shell_rc="$HOME/.bashrc"
    fi

    if ! grep -q "PATH.*\.local/bin" "$shell_rc" 2>/dev/null; then
      echo 'export PATH="$HOME/.local/bin:$PATH"' >> "$shell_rc"
      success "Added $INSTALL_DIR to PATH in $shell_rc"
    fi
  fi
}

# Main installation
main() {
  echo ""
  echo -e "${GREEN}gwi${NC} - Git Worktree Issue CLI Installer"
  echo ""

  check_deps
  install_script
  install_completions
  add_shell_integration
  check_path

  echo ""
  success "Installation complete!"
  echo ""
  echo "To start using gwi, either:"
  echo "  1. Open a new terminal, or"
  echo "  2. Run: source ~/.zshrc (or ~/.bashrc)"
  echo ""
  echo "Then try: gwi help"
  echo ""
}

main "$@"
