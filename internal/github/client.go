package github

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
)

// Issue represents a GitHub issue
type Issue struct {
	Number int    `json:"number"`
	Title  string `json:"title"`
	State  string `json:"state"`
}

// PullRequest represents a GitHub pull request
type PullRequest struct {
	Number            int           `json:"number"`
	State             string        `json:"state"`
	Mergeable         string        `json:"mergeable"`
	MergeStateStatus  string        `json:"mergeStateStatus"`
	HeadRefName       string        `json:"headRefName"`
	StatusCheckRollup []CheckStatus `json:"statusCheckRollup"`
}

// CheckStatus represents a CI check status
type CheckStatus struct {
	Name       string `json:"name"`
	Conclusion string `json:"conclusion"`
}

// CheckAuth verifies that gh is authenticated
func CheckAuth() error {
	cmd := exec.Command("gh", "auth", "status")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("GitHub CLI not authenticated. Run: gh auth login")
	}
	return nil
}

// GetIssue fetches an issue by number
func GetIssue(issueNumber int) (*Issue, error) {
	cmd := exec.Command("gh", "issue", "view", strconv.Itoa(issueNumber), "--json", "number,title,state")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("issue #%d not found", issueNumber)
	}

	var issue Issue
	if err := json.Unmarshal(output, &issue); err != nil {
		return nil, err
	}
	return &issue, nil
}

// ListOpenIssues lists open issues for the current repository
func ListOpenIssues(limit int) ([]Issue, error) {
	cmd := exec.Command("gh", "issue", "list", "--state", "open", "--limit", strconv.Itoa(limit), "--json", "number,title")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var issues []Issue
	if err := json.Unmarshal(output, &issues); err != nil {
		return nil, err
	}
	return issues, nil
}

// CreatePR creates a pull request
func CreatePR(path, title, body, branchName string) (string, error) {
	cmd := exec.Command("gh", "pr", "create",
		"--title", title,
		"--body", body,
		"--head", branchName)
	cmd.Dir = path
	output, err := cmd.Output()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			return "", fmt.Errorf("failed to create PR: %s", string(exitErr.Stderr))
		}
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// GetPRForBranch gets the PR number for a branch
func GetPRForBranch(branchName string) (int, error) {
	cmd := exec.Command("gh", "pr", "list", "--head", branchName, "--json", "number", "--jq", ".[0].number")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}

	numStr := strings.TrimSpace(string(output))
	if numStr == "" {
		return 0, fmt.Errorf("no PR found for branch: %s", branchName)
	}

	return strconv.Atoi(numStr)
}

// GetPRStatus gets the status of a PR
func GetPRStatus(prNumber int) (*PullRequest, error) {
	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(prNumber),
		"--json", "mergeable,mergeStateStatus,statusCheckRollup,state,headRefName")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var pr PullRequest
	pr.Number = prNumber
	if err := json.Unmarshal(output, &pr); err != nil {
		return nil, err
	}
	return &pr, nil
}

// GetPRState gets just the state of a PR
func GetPRState(prNumber int) (string, error) {
	cmd := exec.Command("gh", "pr", "view", strconv.Itoa(prNumber), "--json", "state", "--jq", ".state")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

// MergePR merges a pull request
func MergePR(prNumber int, strategy string) error {
	cmd := exec.Command("gh", "pr", "merge", strconv.Itoa(prNumber), "--"+strategy, "--delete-branch")
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// CommentOnIssue adds a comment to an issue
func CommentOnIssue(issueNumber int, body string) error {
	cmd := exec.Command("gh", "issue", "comment", strconv.Itoa(issueNumber), "--body", body)
	return cmd.Run()
}

// ListOpenPRs lists open PRs with branch info
func ListOpenPRs() ([]PullRequest, error) {
	cmd := exec.Command("gh", "pr", "list", "--state", "open", "--json", "number,headRefName")
	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	var prs []PullRequest
	if err := json.Unmarshal(output, &prs); err != nil {
		return nil, err
	}
	return prs, nil
}

// GetFailingChecks returns the names of failing checks for a PR
func GetFailingChecks(pr *PullRequest) []string {
	var failing []string
	for _, check := range pr.StatusCheckRollup {
		if check.Conclusion == "FAILURE" {
			failing = append(failing, check.Name)
		}
	}
	return failing
}

// CloseIssue closes an issue with an optional comment
func CloseIssue(issueNumber int, comment string) error {
	if comment != "" {
		if err := CommentOnIssue(issueNumber, comment); err != nil {
			return fmt.Errorf("failed to add comment: %w", err)
		}
	}
	cmd := exec.Command("gh", "issue", "close", strconv.Itoa(issueNumber))
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to close issue: %s", strings.TrimSpace(string(output)))
	}
	return nil
}

// IsPRMerged checks if a PR has been merged
func IsPRMerged(prNumber int) (bool, error) {
	state, err := GetPRState(prNumber)
	if err != nil {
		return false, err
	}
	return state == "MERGED", nil
}
