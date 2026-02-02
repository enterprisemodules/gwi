# gwi - Git Worktree Issue CLI

A CLI tool that integrates GitHub issues with git worktrees for a streamlined development workflow.

## Installation

```bash
# Using go install
go install github.com/enterprisemodules/gwi@latest

# Or build from source
git clone https://github.com/enterprisemodules/gwi
cd gwi
make install
```

### Shell Integration

Add to your `~/.zshrc` or `~/.bashrc`:

```bash
eval "$(gwi init zsh)"
```

This enables `gwi cd`, `gwi main`, `gwi list`, `gwi create`, and `gwi start` to change your working directory.

### Requirements

- `git` - Git version control
- `gh` - GitHub CLI (authenticated)
- `fzf` - Fuzzy finder (optional, for better selection UI)
- `tmux` - Terminal multiplexer (optional, for `gwi up/down/logs`)
- `direnv` - Directory-specific environments (optional, for automatic env loading)

## Commands

| Command | Description |
|---------|-------------|
| `gwi start` | Select open issue interactively and create worktree |
| `gwi create [issue-number]` | Create worktree from GitHub issue |
| `gwi pr [issue-number]` | Push, create PR with "Closes #N", remove worktree |
| `gwi merge [issue-number]` | Merge PR, delete branch, remove worktree |
| `gwi rm [issue-number]` | Delete worktree (see flags below) |
| `gwi cd [number\|pattern]` | Navigate to worktree (fuzzy match supported) |
| `gwi main` | Navigate back to main repository |
| `gwi list` | Interactive worktree selector (includes main) |
| `gwi status` | Show status of all worktrees with PR info |
| `gwi clean` | Remove orphaned worktrees and branches |
| `gwi activate` | Run setup hook (install deps, etc.) |
| `gwi up` | Start dev server in tmux session |
| `gwi down` | Stop dev server (runs down hook if present) |
| `gwi logs` | Attach to tmux session to view logs |
| `gwi completion [shell]` | Generate shell completions |

### Remove Command Flags

| Flag | Description |
|------|-------------|
| `-f, --force` | Force remove even with uncommitted changes |
| `-y, --yes` | Skip confirmation prompt |
| `-D, --delete-branch` | Also delete the local and remote branch |

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

## Interactive Selection

When using `gwi start` or `gwi create` without arguments, issues that already have worktrees are shown dimmed and cannot be selected. This prevents accidentally trying to create duplicate worktrees.

If you try to create a worktree for an issue that already exists:

```bash
gwi create 42
# Error: Worktree for issue #42 already exists.
#
#   Path: ~/worktrees/github.com/org/repo/42-fix-bug
#
#   Use 'gwi cd 42' to navigate to it, or 'gwi rm 42' to remove it first.
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

| Hook | Description |
|------|-------------|
| `activate` | Setup script (install deps, configure env) |
| `create` | Runs after worktree creation |
| `up` | Command to start dev server (runs in tmux with direnv) |
| `down` | Cleanup script (runs before stopping server) |

### Example Hooks

`.gwi/up`:
```bash
#!/bin/bash
bundle install
bin/dev
```

`.gwi/down`:
```bash
#!/bin/bash
docker-compose down
echo "Cleanup complete"
```

`.gwi/activate`:
```bash
#!/bin/bash
bundle install
yarn install
bin/rails db:migrate
```

## Development Server

The `gwi up` command starts your dev server in a background tmux session with full environment support:

- Loads your shell profile (`.zshrc`/`.bashrc`)
- Activates direnv environment (`.envrc`)
- Loads RVM/rbenv/asdf shims

```bash
gwi up      # Start server in background tmux session
gwi logs    # Attach to view logs (Ctrl+B D to detach)
gwi down    # Run down hook and stop session
```

Each worktree gets its own tmux session named after the directory, so you can run multiple dev servers simultaneously (use different ports via direnv).

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

# Remove worktree (keeps branch)
gwi rm 37

# Remove worktree and delete branch (local + remote)
gwi rm 37 -D

# Force remove with uncommitted changes
gwi rm 37 --force

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
