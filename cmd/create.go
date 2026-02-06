package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/enterprisemodules/gwi/internal/github"
	"github.com/enterprisemodules/gwi/internal/hooks"
	"github.com/enterprisemodules/gwi/internal/tui"
	"github.com/spf13/cobra"
)

var (
	includeInProgress bool
)

var createCmd = &cobra.Command{
	Use:   "create [issue-number]",
	Short: "Create worktree from GitHub issue",
	Long:  `Create a new git worktree for a GitHub issue. If no issue number is provided, opens an interactive selector.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runCreate,
}

var internalCreateCmd = &cobra.Command{
	Use:    "_create [issue-number]",
	Hidden: true,
	Args:   cobra.MaximumNArgs(1),
	Run:    runInternalCreate,
}

func runCreate(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	var issueNumber int
	if len(args) > 0 {
		issueNumber, err = strconv.Atoi(args[0])
		if err != nil {
			config.Die("Invalid issue number: %s", args[0])
		}
	} else {
		issueNumber, err = selectIssue(repoInfo)
		if err != nil {
			config.Die("No issue selected")
		}
	}

	createWorktree(cfg, repoInfo, issueNumber, false)
}

func runInternalCreate(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	var issueNumber int
	if len(args) > 0 {
		issueNumber, err = strconv.Atoi(args[0])
		if err != nil {
			config.Die("Invalid issue number: %s", args[0])
		}
	} else {
		issueNumber, err = selectIssue(repoInfo)
		if err != nil {
			config.Die("No issue selected")
		}
	}

	// Create worktree silently and output just the path
	worktreePath := createWorktree(cfg, repoInfo, issueNumber, true)
	fmt.Println(worktreePath)
}

func selectIssue(repoInfo *git.RepoInfo) (int, error) {
	cfg := config.Load()

	if err := github.CheckAuth(); err != nil {
		return 0, err
	}

	// Get issues with their project status
	issues, err := github.ListOpenIssuesWithStatus(50, cfg.GitHub.StatusFieldName)
	if err != nil {
		return 0, err
	}

	if len(issues) == 0 {
		return 0, fmt.Errorf("no open issues found")
	}

	// Get existing worktrees to mark them as disabled
	existingIssues := getExistingWorktreeIssues(cfg, repoInfo)

	var options []tui.Option
	for _, issue := range issues {
		_, exists := existingIssues[issue.Number]

		// Check if issue is in progress
		isInProgress := issue.ProjectStatus == cfg.GitHub.InProgressValue

		// Determine if option should be disabled
		disabled := exists || (isInProgress && !includeInProgress)

		// Determine hint based on status
		hint := ""
		if exists {
			hint = "already exists"
		} else if isInProgress {
			hint = "in progress"
		}

		options = append(options, tui.Option{
			Label:      fmt.Sprintf("#%d %s", issue.Number, issue.Title),
			Value:      strconv.Itoa(issue.Number),
			Disabled:   disabled,
			Hint:       hint,
			InProgress: isInProgress && !exists, // Mark as in-progress only if not already existing
		})
	}

	header := fmt.Sprintf("Select issue (%s/%s)", repoInfo.Org, repoInfo.Repo)
	selected, err := tui.Select(header, options)
	if err != nil {
		return 0, err
	}

	return strconv.Atoi(selected)
}

// getExistingWorktreeIssues returns a set of issue numbers that have existing worktrees
func getExistingWorktreeIssues(cfg *config.Config, repoInfo *git.RepoInfo) map[int]bool {
	result := make(map[int]bool)
	basePath := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)

	worktrees, err := git.ListWorktrees(basePath)
	if err != nil {
		return result
	}

	for _, wt := range worktrees {
		dir := filepath.Base(wt)
		// Extract issue number from directory name (e.g., "42-fix-bug")
		if idx := strings.Index(dir, "-"); idx > 0 {
			if num, err := strconv.Atoi(dir[:idx]); err == nil {
				result[num] = true
			}
		}
	}

	return result
}

func createWorktree(cfg *config.Config, repoInfo *git.RepoInfo, issueNumber int, silent bool) string {
	if err := github.CheckAuth(); err != nil {
		config.Die("%v", err)
	}

	if !silent {
		config.Info("Fetching issue #%d...", issueNumber)
	}

	issue, err := github.GetIssue(issueNumber)
	if err != nil {
		config.Die("%v", err)
	}

	if issue.State == "CLOSED" {
		config.Warn("Issue #%d is closed", issueNumber)
	}

	slug := git.Slugify(issue.Title)
	branchName := fmt.Sprintf("%d-%s", issueNumber, slug)
	worktreePath := filepath.Join(cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo), branchName)

	// Check if worktree already exists
	if _, err := os.Stat(worktreePath); err == nil {
		if !silent {
			config.Die("Worktree for issue #%d already exists.\n\n  Path: %s\n\n  Use 'gwi cd %d' to navigate to it, or 'gwi rm %d' to remove it first.", issueNumber, worktreePath, issueNumber, issueNumber)
		}
		// In silent mode (shell integration), just return the path to cd to it
		fmt.Println(worktreePath)
		return worktreePath
	}

	if !silent {
		config.Info("Fetching from origin...")
	}
	if err := git.Fetch(); err != nil {
		config.Die("Failed to fetch: %v", err)
	}

	// Create worktree directory structure
	if err := os.MkdirAll(filepath.Dir(worktreePath), 0755); err != nil {
		config.Die("Failed to create directory: %v", err)
	}

	// Check if branch already exists (local or remote)
	if git.BranchExists(branchName) {
		if !silent {
			config.Info("Using existing local branch: %s", branchName)
		}
		if err := git.CreateWorktreeFromBranch(worktreePath, branchName); err != nil {
			config.Die("Failed to create worktree: %v", err)
		}
	} else if git.RemoteBranchExists(branchName) {
		if !silent {
			config.Info("Using existing remote branch: %s", branchName)
		}
		if err := git.CreateWorktreeFromRemote(worktreePath, branchName, "origin/"+branchName); err != nil {
			config.Die("Failed to create worktree: %v", err)
		}
	} else {
		if !silent {
			config.Info("Creating worktree: %s", branchName)
		}
		if err := git.CreateWorktree(worktreePath, branchName, "origin/main"); err != nil {
			config.Die("Failed to create worktree: %v", err)
		}
	}

	if !silent {
		config.Success("Worktree created at: %s", worktreePath)
	}

	// Run create hook if it exists
	hooks.RunHook("create", worktreePath, cfg, repoInfo)

	// Update GitHub Project status to "In Progress"
	// This happens even in silent mode, messages go to stderr so they don't break shell integration
	if cfg.Verbose {
		config.Info("GitHub Projects enabled: %v", cfg.GitHub.ProjectsEnabled)
	}
	if cfg.GitHub.ProjectsEnabled {
		branchName := filepath.Base(worktreePath)
		if cfg.Verbose {
			config.Info("Branch name: %s", branchName)
		}
		if issueNum, ok := github.ParseIssueFromBranch(branchName); ok {
			if cfg.Verbose {
				config.Info("Parsed issue number: %d", issueNum)
				config.Info("Attempting to update to: %s", cfg.GitHub.InProgressValue)
			}
			if err := github.UpdateIssueStatus(issueNum, cfg.GitHub.InProgressValue, cfg); err != nil {
				if cfg.Verbose {
					config.Warn("Failed to update project status: %v", err)
				}
			} else {
				config.Info("Updated issue #%d to '%s' in GitHub Projects", issueNum, cfg.GitHub.InProgressValue)
			}
		} else if cfg.Verbose {
			config.Warn("Could not parse issue number from branch: %s", branchName)
		}
	}

	// Output cd instruction for shell wrapper (only in interactive mode)
	if !silent {
		fmt.Printf("__GWI_CD_TO__:%s\n", worktreePath)
	}

	return worktreePath
}

func init() {
	createCmd.Flags().BoolVar(&includeInProgress, "include-in-progress", false, "Allow selecting issues that are already in progress")
}
