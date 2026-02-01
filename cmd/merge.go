package cmd

import (
	"fmt"
	"path/filepath"
	"strconv"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/enterprisemodules/gwi/internal/github"
	"github.com/spf13/cobra"
)

var mergeCmd = &cobra.Command{
	Use:   "merge [issue-number]",
	Short: "Merge PR and cleanup",
	Long:  `Merge the pull request, delete the branch, and remove the worktree.`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runMerge,
}

func runMerge(cmd *cobra.Command, args []string) {
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
	var branchName string

	if worktreePath != "" {
		branchName = filepath.Base(worktreePath)
	} else {
		// Try to find the branch name from a PR
		prs, err := github.ListOpenPRs()
		if err != nil {
			config.Die("Failed to list PRs: %v", err)
		}
		for _, pr := range prs {
			if pr.Number == issueNumber {
				branchName = pr.HeadRefName
				break
			}
		}
		if branchName == "" {
			config.Die("No worktree or PR found for issue #%d", issueNumber)
		}
	}

	// Find the PR number for this branch
	prNumber, err := github.GetPRForBranch(branchName)
	if err != nil {
		config.Die("No PR found for branch: %s", branchName)
	}

	// Check PR status
	pr, err := github.GetPRStatus(prNumber)
	if err != nil {
		config.Die("Failed to get PR status: %v", err)
	}

	if pr.Mergeable == "CONFLICTING" {
		config.Die("PR #%d has merge conflicts. Resolve them first.", prNumber)
	}

	if pr.MergeStateStatus == "BLOCKED" {
		config.Warn("PR #%d is blocked (required checks or reviews pending)", prNumber)
		if !confirmPrompt("Continue anyway?") {
			config.Die("Aborted")
		}
	}

	// Check for failing CI
	failingChecks := github.GetFailingChecks(pr)
	if len(failingChecks) > 0 {
		config.Warn("PR #%d has failing checks:", prNumber)
		for _, check := range failingChecks {
			if len(failingChecks) > 3 {
				break
			}
			fmt.Printf("  - %s\n", check)
		}
		if !confirmPrompt("Continue anyway?") {
			config.Die("Aborted")
		}
	}

	// Get the last commit message to post as issue comment
	var lastCommitMsg string
	if worktreePath != "" {
		lastCommitMsg, _ = git.GetLastCommitMessage("")
	} else {
		lastCommitMsg, _ = git.GetLastCommitMessage("origin/" + branchName)
	}

	// Post summary comment to the issue
	if lastCommitMsg != "" {
		config.Info("Adding summary to issue #%d...", issueNumber)
		body := fmt.Sprintf("**Merged in PR #%d**\n\n%s", prNumber, lastCommitMsg)
		github.CommentOnIssue(issueNumber, body)
	}

	config.Info("Merging PR #%d (%s)...", prNumber, cfg.MergeStrategy)
	if err := github.MergePR(prNumber, cfg.MergeStrategy); err != nil {
		config.Die("Failed to merge PR: %v", err)
	}

	// Remove worktree if it exists
	if worktreePath != "" {
		config.Info("Removing worktree...")
		if err := git.RemoveWorktree(worktreePath, false); err != nil {
			// Try force remove
			git.RemoveWorktree(worktreePath, true)
		}
	}

	// Clean up local branch
	git.DeleteBranch(branchName)

	config.Success("PR merged and cleaned up!")
}
