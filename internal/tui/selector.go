package tui

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"

	"github.com/enterprisemodules/gwi/internal/config"
)

// Option represents a selectable option
type Option struct {
	Label      string
	Value      string
	Disabled   bool   // If true, option is shown but not selectable
	Hint       string // Optional hint shown after label (e.g., "already exists")
	InProgress bool   // If true, option is shown in a different color (yellow)
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

// ANSI color codes for fzf
const (
	fzfDim    = "\033[2m"
	fzfReset  = "\033[0m"
	fzfYellow = "\033[33m"
	fzfCyan   = "\033[36m"
)

// selectWithFzf uses fzf for selection
func selectWithFzf(header string, options []Option) (string, error) {
	// Separate enabled and disabled options
	var enabledLabels []string
	labelToValue := make(map[string]string)
	var disabledLabels []string

	for _, opt := range options {
		label := opt.Label

		// Add color for in-progress items
		if opt.InProgress && !opt.Disabled {
			label = fmt.Sprintf("%s%s%s", fzfYellow, opt.Label, fzfReset)
		}

		if opt.Hint != "" {
			if opt.InProgress && !opt.Disabled {
				label = fmt.Sprintf("%s %s(%s)%s", label, fzfCyan, opt.Hint, fzfReset)
			} else {
				label = fmt.Sprintf("%s %s(%s)%s", opt.Label, fzfYellow, opt.Hint, fzfReset)
			}
		}

		if opt.Disabled {
			// Dim the entire line for disabled options
			disabledLabels = append(disabledLabels, fmt.Sprintf("%s%s%s", fzfDim, label, fzfReset))
		} else {
			enabledLabels = append(enabledLabels, label)
			labelToValue[label] = opt.Value
			// Also map the plain label without hint for matching
			labelToValue[opt.Label] = opt.Value
		}
	}

	// Build fzf input: enabled options first, then disabled (shown but not selectable)
	cmd := exec.Command("fzf", "--height=~50%", "--reverse", "--ansi", "--header="+header)
	cmd.Stderr = os.Stderr

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	go func() {
		defer stdin.Close()
		// Write enabled options first
		for _, label := range enabledLabels {
			fmt.Fprintln(stdin, label)
		}
		// Write disabled options (dimmed, shown at bottom)
		for _, label := range disabledLabels {
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

	// Strip ANSI codes for lookup
	cleanSelected := stripAnsi(selected)
	// Try to match with hint first, then without
	if val, ok := labelToValue[selected]; ok {
		return val, nil
	}
	// Try matching the clean version (might have selected a disabled item somehow)
	for label, value := range labelToValue {
		if stripAnsi(label) == cleanSelected || strings.HasPrefix(cleanSelected, stripAnsi(label)+" ") {
			return value, nil
		}
	}

	return "", fmt.Errorf("invalid selection")
}

// stripAnsi removes ANSI escape codes from a string
func stripAnsi(s string) string {
	re := regexp.MustCompile(`\x1b\[[0-9;]*m`)
	return re.ReplaceAllString(s, "")
}

// selectWithNumbered uses a numbered list for selection
func selectWithNumbered(header string, options []Option) (string, error) {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "%s:\n", header)
	fmt.Fprintln(os.Stderr)

	// Build index mapping (only enabled options get numbers)
	enabledIndices := make(map[int]int) // display number -> option index
	displayNum := 1

	for i, opt := range options {
		if opt.Disabled {
			// Show disabled options dimmed without a number
			hint := ""
			if opt.Hint != "" {
				hint = fmt.Sprintf(" (%s)", opt.Hint)
			}
			// Use dim/gray appearance for disabled items
			fmt.Fprintf(os.Stderr, "     \033[2m%s%s\033[0m\n", opt.Label, hint)
		} else {
			hint := ""
			if opt.Hint != "" {
				hint = fmt.Sprintf(" (%s)", opt.Hint)
			}

			// Apply yellow color for in-progress items
			if opt.InProgress {
				fmt.Fprintf(os.Stderr, "  %d) %s%s%s\n", displayNum, config.Yellow(opt.Label), config.Yellow(hint), config.Yellow(""))
			} else {
				fmt.Fprintf(os.Stderr, "  %d) %s%s\n", displayNum, opt.Label, hint)
			}
			enabledIndices[displayNum] = i
			displayNum++
		}
	}

	if len(enabledIndices) == 0 {
		return "", fmt.Errorf("no selectable options available")
	}

	fmt.Fprintln(os.Stderr)
	fmt.Fprintf(os.Stderr, "Select [1-%d]: ", displayNum-1)

	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		return "", err
	}

	input = strings.TrimSpace(input)
	choice, err := strconv.Atoi(input)
	if err != nil || choice < 1 || choice >= displayNum {
		return "", fmt.Errorf("invalid selection")
	}

	optIndex := enabledIndices[choice]
	return options[optIndex].Value, nil
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
