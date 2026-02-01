package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/enterprisemodules/gwi/internal/tui"
	"github.com/spf13/cobra"
)

var cdCmd = &cobra.Command{
	Use:   "cd [number|pattern]",
	Short: "Navigate to worktree",
	Long:  `Navigate to a worktree by issue number or pattern. If no argument is provided, opens an interactive selector.`,
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		// Public cd just tells user to use shell integration
		fmt.Println("Use 'gwi cd' with shell integration. Add to your shell config:")
		fmt.Println("  eval \"$(gwi init zsh)\"")
	},
}

var internalCdCmd = &cobra.Command{
	Use:    "_cd [pattern]",
	Hidden: true,
	Args:   cobra.MaximumNArgs(1),
	Run:    runInternalCd,
}

func runInternalCd(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)

	// No pattern - show interactive selector
	if len(args) == 0 {
		issueNumber, err := selectWorktree(repoInfo, cfg)
		if err != nil {
			config.Die("No worktree selected")
		}
		worktreePath := git.FindWorktreeByIssue(base, issueNumber)
		if worktreePath == "" {
			config.Die("No worktree found for issue #%d", issueNumber)
		}
		fmt.Println(worktreePath)
		return
	}

	pattern := args[0]

	// Exact match by issue number
	if num, err := strconv.Atoi(pattern); err == nil {
		worktreePath := git.FindWorktreeByIssue(base, num)
		if worktreePath != "" {
			fmt.Println(worktreePath)
			return
		}
	}

	// Fuzzy match
	worktrees, err := git.ListWorktrees(base)
	if err != nil {
		config.Die("Failed to list worktrees: %v", err)
	}

	var matches []string
	for _, wt := range worktrees {
		name := filepath.Base(wt)
		if strings.Contains(name, pattern) {
			matches = append(matches, wt)
		}
	}

	switch len(matches) {
	case 0:
		config.Die("No worktree found matching: %s", pattern)
	case 1:
		fmt.Println(matches[0])
	default:
		// Multiple matches - use selector
		var options []tui.Option
		for _, match := range matches {
			name := filepath.Base(match)
			options = append(options, tui.Option{
				Label: name,
				Value: match,
			})
		}

		header := fmt.Sprintf("Multiple matches (%s/%s)", repoInfo.Org, repoInfo.Repo)
		selected, err := tui.Select(header, options)
		if err != nil {
			config.Die("No selection made")
		}
		fmt.Println(selected)
	}
}

func selectWorktree(repoInfo *git.RepoInfo, cfg *config.Config) (int, error) {
	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)
	worktrees, err := git.ListWorktrees(base)
	if err != nil {
		return 0, err
	}

	if len(worktrees) == 0 {
		return 0, fmt.Errorf("no worktrees found for %s/%s", repoInfo.Org, repoInfo.Repo)
	}

	var options []tui.Option
	for _, wt := range worktrees {
		name := filepath.Base(wt)
		// Extract issue number from name
		var issueNum string
		for i, c := range name {
			if c == '-' {
				issueNum = name[:i]
				break
			}
		}
		if issueNum != "" {
			options = append(options, tui.Option{
				Label: name,
				Value: issueNum,
			})
		}
	}

	if len(options) == 0 {
		return 0, fmt.Errorf("no valid worktrees found")
	}

	header := fmt.Sprintf("Select worktree (%s/%s)", repoInfo.Org, repoInfo.Repo)
	selected, err := tui.Select(header, options)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(selected)
}
