package tui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/enterprisemodules/gwi/internal/config"
)

// Option represents a selectable option
type Option struct {
	Label string
	Value string
}

// hasFzf checks if fzf is available
func hasFzf() bool {
	_, err := exec.LookPath("fzf")
	return err == nil
}

// Select presents an interactive selection UI and returns the selected value
func Select(header string, options []Option) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options to select from")
	}

	if hasFzf() {
		return selectWithFzf(header, options)
	}
	return selectWithNumbered(header, options)
}

// selectWithFzf uses fzf for selection
func selectWithFzf(header string, options []Option) (string, error) {
	// Build input for fzf
	var labels []string
	labelToValue := make(map[string]string)
	for _, opt := range options {
		labels = append(labels, opt.Label)
		labelToValue[opt.Label] = opt.Value
	}

	cmd := exec.Command("fzf", "--height=~50%", "--reverse", "--header="+header)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		for _, label := range labels {
			fmt.Fprintln(stdin, label)
		}
	}()

	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("no selection made")
	}

	selected := strings.TrimSpace(string(output))
	if selected == "" {
		return "", fmt.Errorf("no selection made")
	}

	return labelToValue[selected], nil
}

// selectWithNumbered uses a numbered list for selection
func selectWithNumbered(header string, options []Option) (string, error) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s:\n", header)
	fmt.Fprintln(os.Stderr)

	for i, opt := range options {
		fmt.Fprintf(os.Stderr, "  %s%d)%s %s\n", config.Yellow(""), i+1, config.Yellow(""), opt.Label)
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Select [1-%d]: ", len(options))

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice > len(options) {
		return "", fmt.Errorf("invalid selection")
	}

	return options[choice-1].Value, nil
}

// Confirm asks for user confirmation
func Confirm(prompt string) bool {
	fmt.Fprintf(os.Stderr, "%s [y/N]: ", prompt)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return false
	}

	input = strings.TrimSpace(strings.ToLower(input))
	return input == "y" || input == "yes"
}
