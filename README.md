# gwi - Git Worktree Issue CLI

A CLI tool that integrates GitHub issues with git worktrees for a streamlined development workflow.

Available in two versions:
- **Go** (recommended): Single binary, no dependencies, cross-platform
- **Bash**: Portable script, requires `jq` and `gh`

## Installation

### Go Version (Recommended)

```bash
# Using go install
go install github.com/enterprisemodules/gwi@latest

# Or build from source
git clone https://github.com/enterprisemodules/gwi
cd gwi
make install
```

### Bash Version

```bash
curl -fsSL https://raw.githubusercontent.com/enterprisemodules/gwi/main/install.sh | bash
```

### Shell Integration

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(gwi init zsh)"
```

This enables `gwi cd`, `gwi main`, `gwi list`, and `gwi start` to change your working directory.

### Requirements

**Go version:** None (single binary)

**Bash version:**
- `git` - Git version control
- `gh` - GitHub CLI (authenticated)
- `jq` - JSON processor
- `fzf` - Fuzzy finder (optional, for better selection UI)

## Commands

| Command | Description |
|---------|-------------|
| `gwi start` | Select open issue interactively and create worktree |
| `gwi create [issue-number]` | Create worktree from GitHub issue |
| `gwi pr [issue-number]` | Push, create PR with "Closes #N", remove worktree |
| `gwi merge [issue-number]` | Merge PR, delete branch, remove worktree |
| `gwi rm [issue-number]` | Delete worktree (`--force` for uncommitted changes) |
| `gwi cd [number\|pattern]` | Navigate to worktree (fuzzy match supported) |
| `gwi main` | Navigate back to main repository |
| `gwi list` | Interactive worktree selector (includes main) |
| `gwi status` | Show status of all worktrees with PR info |
| `gwi clean` | Remove orphaned worktrees and branches |
| `gwi activate` | Run setup hook (install deps, etc.) |
| `gwi up` | Start dev server in tmux session |
| `gwi down` | Stop dev server |
| `gwi logs` | Attach to tmux session to view logs |
| `gwi completion [shell]` | Generate shell completions |

## Workflow

```bash
# Start working on an issue (interactive selection)
gwi start

# Or create worktree for a specific issue
gwi create 42

# ... make your changes ...

# Create PR and clean up worktree
gwi pr

# After review, merge and clean up
gwi merge 42
```

## Directory Structure

Worktrees are organized by GitHub org and repo:

```
~/worktrees/
  github.com/
    <org>/
      <repo>/
        37-final-creation-of-product-fails/
        42-add-user-authentication/
```

## Configuration

| Variable | Description | Default |
|----------|-------------|---------|
| `GWI_WORKTREE_BASE` | Base directory for worktrees | `~/worktrees` |
| `GWI_MERGE_STRATEGY` | Merge strategy: squash, merge, rebase | `squash` |
| `GWI_AUTO_ACTIVATE` | Auto-run activate hook on cd/start | `0` |
| `GWI_HOOK_DIR` | Global hooks directory | `~/.config/gwi/hooks` |

## Hooks

Hooks are executable scripts searched in order:

1. `<worktree>/.gwi/<hook>` - Per-worktree override
2. `<main-repo>/.gwi/<hook>` - Per-project hook (recommended)
3. `~/.config/gwi/hooks/<org>/<repo>/<hook>` - Global fallback

Hook types:
- `activate` - Setup script (install deps, configure env)
- `up` - Command to start dev server (runs in tmux)

## Examples

```bash
# Interactive issue selection
gwi start                # Select issue, create worktree, cd into it

# Create worktree for issue #37
gwi create 37
# → Creates ~/worktrees/github.com/org/repo/37-issue-title-slugified/

# Navigate to worktree
gwi cd 37                # by issue number
gwi cd auth              # fuzzy match "auth" in title
gwi list                 # interactive selector

# Check status of all worktrees
gwi status
#   ● 37-fix-bug (2 changes) ↑1 PR #42
#   ● 38-add-feature no PR

# Create PR from current worktree (auto-detects issue number)
gwi pr

# Merge the PR
gwi merge 37

# Remove worktree without merging
gwi rm 37
gwi rm --force           # force remove with uncommitted changes

# Development server (requires tmux)
gwi up                   # Start server in background
gwi logs                 # View logs (Ctrl+B D to detach)
gwi down                 # Stop server
```

## Tab Completion

```bash
# Generate completions
gwi completion zsh > "${fpath[1]}/_gwi"
gwi completion bash > /etc/bash_completion.d/gwi
gwi completion fish > ~/.config/fish/completions/gwi.fish
```

## Building from Source

```bash
# Build
make build

# Install to ~/.local/bin
make install

# Build for all platforms
make build-all

# Run tests
make test
```

## License

MIT
