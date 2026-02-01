package cmd

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/enterprisemodules/gwi/internal/config"
	"github.com/enterprisemodules/gwi/internal/git"
	"github.com/enterprisemodules/gwi/internal/github"
	"github.com/spf13/cobra"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show status of all worktrees",
	Long:  `Display all worktrees with their git status, push/pull state, and PR status.`,
	Run:   runStatus,
}

func runStatus(cmd *cobra.Command, args []string) {
	cfg := config.Load()
	repoInfo, err := git.GetRepoInfo()
	if err != nil {
		config.Die("%v", err)
	}

	base := cfg.WorktreeBasePath(repoInfo.Org, repoInfo.Repo)

	fmt.Printf("%sgwi status%s for %s%s/%s%s\n", config.Green(""), config.Green(""), config.Blue(""), repoInfo.Org, repoInfo.Repo, config.Blue(""))
	fmt.Println()

	worktrees, err := git.ListWorktrees(base)
	if err != nil || len(worktrees) == 0 {
		fmt.Println("No worktrees found.")
		return
	}

	re := regexp.MustCompile(`^(\d+)-`)

	for _, dir := range worktrees {
		name := filepath.Base(dir)
		branchName := name

		// Extract issue number
		var issueNumber int
		matches := re.FindStringSubmatch(name)
		if matches != nil {
			issueNumber, _ = strconv.Atoi(matches[1])
		}

		// Check git status
		var statusIcon string
		var changes string
		if git.HasUncommittedChanges(dir) {
			statusIcon = config.Yellow("●")
			count := git.GetUncommittedCount(dir)
			changes = fmt.Sprintf(" (%d changes)", count)
		} else {
			statusIcon = config.Green("●")
			changes = ""
		}

		// Check if branch is pushed
		var pushStatus string
		ahead, behind, err := git.GetAheadBehind(dir, branchName)
		if err == nil {
			if ahead > 0 {
				pushStatus = fmt.Sprintf(" ↑%d", ahead)
			}
			if behind > 0 {
				pushStatus += fmt.Sprintf(" ↓%d", behind)
			}
		}

		// Check PR status
		var prStatus string
		if issueNumber > 0 {
			prNumber, err := github.GetPRForBranch(branchName)
			if err == nil && prNumber > 0 {
				state, err := github.GetPRState(prNumber)
				if err == nil {
					switch state {
					case "OPEN":
						prStatus = fmt.Sprintf(" %sPR #%d%s", config.Blue(""), prNumber, config.Blue(""))
					case "MERGED":
						prStatus = fmt.Sprintf(" %sPR #%d merged%s", config.Green(""), prNumber, config.Green(""))
					case "CLOSED":
						prStatus = fmt.Sprintf(" %sPR #%d closed%s", config.Red(""), prNumber, config.Red(""))
					}
				}
			} else {
				prStatus = fmt.Sprintf(" %sno PR%s", config.Yellow(""), config.Yellow(""))
			}
		}

		fmt.Printf("  %s %s%s%s%s\n", statusIcon, name, changes, pushStatus, prStatus)
	}
}
