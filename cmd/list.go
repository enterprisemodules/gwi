package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/enterprisemodules/gwi/internal/tui"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Aliases: []string{"ls"},
	Short:   "Interactive worktree selector",
	Long:    `Display and select from available worktrees including main.`,
	Run: func(cmd *cobra.Command, args []string) {
		// Public list shows worktrees without selection
		runListDisplay()
	},
}

var internalListCmd = &cobra.Command{
	Use:    "_list",
	Hidden: true,
	Run:    runInternalList,
}

func runListDisplay() {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)
	mainPath, _ := git.GetMainWorktreePath()

	fmt.Printf("Worktrees for %s%s/%s%s:\n", config.Blue(""), repoInfo.Org, repoInfo.Repo, config.Blue(""))
	fmt.Println()

	// Show main
	if mainPath != "" {
		fmt.Printf("  %smain%s (%s)\n", config.Green(""), config.Green(""), filepath.Base(mainPath))
	}

	// Show issue worktrees
	worktrees, _ := git.ListWorktrees(base)
	for _, wt := range worktrees {
		fmt.Printf("  %s\n", filepath.Base(wt))
	}
}

func runInternalList(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)
	mainPath, _ := git.GetMainWorktreePath()

	var options []tui.Option

	// Add main as first option
	if mainPath != "" {
		options = append(options, tui.Option{
			Label: fmt.Sprintf("main (%s)", filepath.Base(mainPath)),
			Value: mainPath,
		})
	}

	// Add issue worktrees
	worktrees, _ := git.ListWorktrees(base)
	for _, wt := range worktrees {
		name := filepath.Base(wt)
		options = append(options, tui.Option{
			Label: name,
			Value: wt,
		})
	}

	if len(options) == 0 {
		config.Die("No worktrees found")
	}

	header := fmt.Sprintf("Worktrees for %s/%s", repoInfo.Org, repoInfo.Repo)
	selected, err := tui.Select(header, options)
	if err != nil {
		// Silent exit for no selection
		return
	}

	fmt.Println(selected)
}

func selectWorktreeWithMain(repoInfo *git.RepoInfo, cfg *config.Config) (string, error) {
	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)
	mainPath, _ := git.GetMainWorktreePath()

	var options []tui.Option

	// Add main as first option
	if mainPath != "" {
		options = append(options, tui.Option{
			Label: fmt.Sprintf("main (%s)", filepath.Base(mainPath)),
			Value: mainPath,
		})
	}

	// Add issue worktrees
	worktrees, _ := git.ListWorktrees(base)
	for _, wt := range worktrees {
		name := filepath.Base(wt)
		// Extract issue number
		var issueNum string
		for i, c := range name {
			if c == '-' {
				issueNum = name[:i]
				break
			}
		}
		if issueNum != "" {
			if _, err := strconv.Atoi(issueNum); err == nil {
				options = append(options, tui.Option{
					Label: name,
					Value: wt,
				})
			}
		}
	}

	if len(options) == 0 {
		return "", fmt.Errorf("no worktrees found")
	}

	header := fmt.Sprintf("Select worktree (%s/%s)", repoInfo.Org, repoInfo.Repo)
	return tui.Select(header, options)
}
