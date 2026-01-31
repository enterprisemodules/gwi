# gwi - Git Worktree Issue CLI

A portable bash CLI tool that integrates GitHub issues with git worktrees for a streamlined development workflow.

## Installation

### One-command install

```bash
curl -fsSL https://raw.githubusercontent.com/enterprisemodules/gwi/main/install.sh | bash
```

### Manual install

```bash
git clone https://github.com/enterprisemodules/gwi ~/.gwi
~/.gwi/install.sh
```

### Requirements

- `git` - Git version control
- `gh` - GitHub CLI (authenticated)
- `jq` - JSON processor

## Commands

| Command | Description |
|---------|-------------|
| `gwi create <issue-number>` | Create worktree from GitHub issue |
| `gwi pr [issue-number]` | Push, create PR with "Closes #N", remove worktree |
| `gwi merge [issue-number]` | Squash merge PR, delete branch, remove worktree |
| `gwi rm [issue-number]` | Delete worktree (use `--force` for uncommitted changes) |
| `gwi cd <number\|pattern>` | Navigate to worktree (fuzzy match supported) |
| `gwi list` | List all worktrees for current repo |

## Workflow

```bash
# Start working on an issue
gwi create 42
gwi cd 42

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

## Examples

```bash
# Create worktree for issue #37
gwi create 37
# → Creates ~/worktrees/github.com/org/repo/37-issue-title-slugified/

# Navigate to worktree
gwi cd 37                # by issue number
gwi cd auth              # fuzzy match "auth" in title

# Create PR from current worktree (auto-detects issue number)
gwi pr

# Or specify explicitly
gwi pr 37

# Merge the PR
gwi merge 37

# Remove worktree without merging
gwi rm 37
gwi rm --force           # force remove with uncommitted changes
```

## Shell Integration

The installer automatically adds shell integration for the `cd` command. If you need to add it manually:

```bash
# Add to ~/.zshrc or ~/.bashrc
gwi() {
  if [[ "$1" == "cd" ]]; then
    shift
    local path=$(command gwi _cd "$@")
    [[ -d "$path" ]] && cd "$path" || echo "Not found" >&2
  else
    command gwi "$@"
  fi
}
```

## Tab Completion

Zsh completions are installed automatically. They provide:

- `gwi <TAB>` → command names
- `gwi create <TAB>` → open issue numbers from GitHub
- `gwi cd <TAB>` → existing worktree names

## License

MIT
