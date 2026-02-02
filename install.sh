#!/usr/bin/env bash
#
# gwi installer
# Install with: curl -fsSL https://raw.githubusercontent.com/enterprisemodules/gwi/main/install.sh | bash
#

set -euo pipefail

# Colors
RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[0;33m'
BLUE=$'\033[0;34m'
NC=$'\033[0m'

info() { echo -e "${BLUE}→${NC} $*"; }
success() { echo -e "${GREEN}✓${NC} $*"; }
warn() { echo -e "${YELLOW}!${NC} $*"; }
die() { echo -e "${RED}Error:${NC} $*" >&2; exit 1; }

# Configuration
INSTALL_DIR="${HOME}/.local/bin"
REPO="enterprisemodules/gwi"
VERSION="${GWI_VERSION:-latest}"

# Shell integration line to add
SHELL_INIT='eval "$(gwi init zsh)"'

# Detect OS and architecture
detect_platform() {
  local os arch

  case "$(uname -s)" in
    Darwin) os="darwin" ;;
    Linux)  os="linux" ;;
    *)      die "Unsupported OS: $(uname -s)" ;;
  esac

  case "$(uname -m)" in
    x86_64|amd64)  arch="amd64" ;;
    arm64|aarch64) arch="arm64" ;;
    *)             die "Unsupported architecture: $(uname -m)" ;;
  esac

  echo "${os}_${arch}"
}

# Get download URL for latest release
get_download_url() {
  local platform="$1"
  local url

  if [[ "$VERSION" == "latest" ]]; then
    url="https://github.com/${REPO}/releases/latest/download/gwi_${platform}"
  else
    url="https://github.com/${REPO}/releases/download/${VERSION}/gwi_${platform}"
  fi

  echo "$url"
}

# Check dependencies
check_deps() {
  for cmd in git gh; do
    command -v "$cmd" &>/dev/null || die "Required command not found: $cmd"
  done
  success "Dependencies found: git, gh"

  # Optional dependencies
  if command -v fzf &>/dev/null; then
    info "Optional: fzf found (enhanced selection UI)"
  else
    info "Optional: install fzf for enhanced selection UI"
  fi

  if command -v tmux &>/dev/null; then
    info "Optional: tmux found (gwi up/down/logs)"
  else
    info "Optional: install tmux for dev server management"
  fi

  if command -v direnv &>/dev/null; then
    info "Optional: direnv found (automatic environment loading)"
  else
    info "Optional: install direnv for automatic environment loading"
  fi
}

# Install binary
install_binary() {
  local platform download_url

  platform=$(detect_platform)
  download_url=$(get_download_url "$platform")

  info "Installing gwi to $INSTALL_DIR..."
  info "Platform: $platform"

  mkdir -p "$INSTALL_DIR"

  # Download binary
  if command -v curl &>/dev/null; then
    curl -fsSL "$download_url" -o "$INSTALL_DIR/gwi" || die "Download failed. Check if release exists for $platform"
  elif command -v wget &>/dev/null; then
    wget -q "$download_url" -O "$INSTALL_DIR/gwi" || die "Download failed. Check if release exists for $platform"
  else
    die "Neither curl nor wget found"
  fi

  chmod +x "$INSTALL_DIR/gwi"
  success "Installed gwi binary"
}

# Install from local build
install_local() {
  info "Installing gwi from local build..."

  mkdir -p "$INSTALL_DIR"

  if [[ -f "$(dirname "$0")/gwi" ]]; then
    cp "$(dirname "$0")/gwi" "$INSTALL_DIR/gwi"
  elif command -v go &>/dev/null; then
    info "Building from source..."
    cd "$(dirname "$0")"
    go build -o "$INSTALL_DIR/gwi" .
  else
    die "No local binary found and go not installed"
  fi

  chmod +x "$INSTALL_DIR/gwi"
  success "Installed gwi binary"
}

# Add shell integration
add_shell_integration() {
  local shell_rc shell_type

  # Detect shell
  if [[ -n "${ZSH_VERSION:-}" ]] || [[ "$SHELL" == *"zsh"* ]]; then
    shell_rc="$HOME/.zshrc"
    shell_type="zsh"
  elif [[ -n "${BASH_VERSION:-}" ]] || [[ "$SHELL" == *"bash"* ]]; then
    shell_rc="$HOME/.bashrc"
    shell_type="bash"
  else
    warn "Unknown shell. Add to your shell config:"
    echo '  eval "$(gwi init zsh)"'
    return
  fi

  # Check if already installed
  if grep -q 'gwi init' "$shell_rc" 2>/dev/null; then
    info "Shell integration already present in $shell_rc"
    return
  fi

  info "Adding shell integration to $shell_rc..."
  echo "" >> "$shell_rc"
  echo "# gwi - Git Worktree Issue CLI" >> "$shell_rc"
  echo "eval \"\$(gwi init $shell_type)\"" >> "$shell_rc"
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

  # Check if we're in the repo directory (local install)
  if [[ -f "$(dirname "$0")/go.mod" ]]; then
    install_local
  else
    install_binary
  fi

  add_shell_integration
  check_path

  echo ""
  success "Installation complete!"
  echo ""
  echo "To start using gwi, either:"
  echo "  1. Open a new terminal, or"
  echo "  2. Run: source ~/.zshrc (or ~/.bashrc)"
  echo ""
  echo "Then try: gwi --help"
  echo ""
}

main "$@"
