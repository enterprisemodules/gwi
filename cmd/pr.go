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
	"github.com/spf13/cobra"
)

var prCmd = &cobra.Command{
	Use:   "pr [issue-number]",
	Short: "Push and create PR",
	Long:  `Push branch, create a pull request with "Closes #N", then remove the worktree.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runPR,
}

func runPR(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	var issueNumber int
	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)

	if len(args) > 0 {
		issueNumber, err = strconv.Atoi(args[0])
		if err != nil {
			config.Die("Invalid issue number: %s", args[0])
		}
	} else {
		// Try to detect from current directory
		if num, ok := git.DetectIssueNumber(base); ok {
			issueNumber = num
		} else {
			// Interactive selection
			issueNumber, err = selectWorktree(repoInfo, cfg)
			if err != nil {
				config.Die("No worktree selected")
			}
		}
	}

	worktreePath := git.FindWorktreeByIssue(base, issueNumber)
	if worktreePath == "" {
		config.Die("No worktree found for issue #%d", issueNumber)
	}

	// Check for uncommitted changes
	if git.HasUncommittedChanges(worktreePath) {
		config.Warn("Worktree has uncommitted changes")
		status, _ := git.GetStatusShort(worktreePath)
		lines := strings.Split(status, "\n")
		for i, line := range lines {
			if i >= 5 {
				break
			}
			fmt.Printf("  %s%s%s\n", config.Yellow(""), line, config.Yellow(""))
		}
		if !confirmPrompt("Continue anyway?") {
			config.Die("Aborted. Commit your changes first.")
		}
	}

	branchName := filepath.Base(worktreePath)

	config.Info("Fetching issue #%d...", issueNumber)
	issue, err := github.GetIssue(issueNumber)
	if err != nil {
		config.Die("%v", err)
	}

	config.Info("Pushing branch: %s", branchName)
	if err := git.Push(worktreePath, branchName); err != nil {
		config.Die("Failed to push: %v", err)
	}

	config.Info("Creating pull request...")
	prURL, err := github.CreatePR(worktreePath, issue.Title, fmt.Sprintf("Closes #%d", issueNumber), branchName)
	if err != nil {
		config.Die("Failed to create PR: %v", err)
	}

	config.Success("Pull request created: %s", prURL)

	// Update GitHub Project status to "In Review"
	if cfg.GitHub.ProjectsEnabled {
		if err := github.UpdateIssueStatus(issueNumber, cfg.GitHub.InReviewValue, cfg); err != nil {
			if cfg.Verbose {
				config.Warn("Failed to update project status: %v", err)
			}
		} else {
			config.Info("Updated issue #%d to '%s' in GitHub Projects", issueNumber, cfg.GitHub.InReviewValue)
		}
	}

	config.Info("Removing worktree...")
	if err := git.RemoveWorktree(worktreePath, false); err != nil {
		config.Warn("Failed to remove worktree: %v", err)
	}

	config.Success("Done! PR is ready for review.")
}

func confirmPrompt(prompt string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", prompt)
	var response string
	fmt.Scanln(&response)
	response = strings.ToLower(strings.TrimSpace(response))
	return response == "y" || response == "yes"
}
