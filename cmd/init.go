package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init [shell]",
	Short: "Output shell integration code",
	Long:  `Output shell integration code for zsh or bash. Add to your shell config with: eval "$(gwi init zsh)"`,
	Args:  cobra.MaximumNArgs(1),
	Run:   runInit,
}

const shellIntegration = `# gwi - Git Worktree Issue CLI shell integration
gwi() {
  if [[ "$1" == "cd" ]]; then
    shift
    local path=$(command gwi _cd "$@")
    if [[ -d "$path" ]]; then
      cd "$path"
      [[ "${GWI_AUTO_ACTIVATE:-0}" == "1" ]] && command gwi activate 2>/dev/null
    else
      echo "Not found" >&2
    fi
  elif [[ "$1" == "main" ]]; then
    local path=$(command gwi _main)
    [[ -d "$path" ]] && cd "$path" || echo "Not found" >&2
  elif [[ "$1" == "list" ]]; then
    local path=$(command gwi _list)
    if [[ -n "$path" && -d "$path" ]]; then
      cd "$path"
      [[ "${GWI_AUTO_ACTIVATE:-0}" == "1" ]] && command gwi activate 2>/dev/null
    fi
  elif [[ "$1" == "start" ]]; then
    local path=$(command gwi _start)
    if [[ -n "$path" && -d "$path" ]]; then
      cd "$path"
      [[ "${GWI_AUTO_ACTIVATE:-0}" == "1" ]] && command gwi activate 2>/dev/null
    fi
  elif [[ "$1" == "create" ]]; then
    local path=$(command gwi _create "${@:2}")
    if [[ -n "$path" && -d "$path" ]]; then
      cd "$path"
      [[ "${GWI_AUTO_ACTIVATE:-0}" == "1" ]] && command gwi activate 2>/dev/null
    fi
  elif [[ "$1" == "rm" ]]; then
    local output=$(command gwi rm "${@:2}")
    echo "$output" | grep -v "^__GWI_CD_TO__:"
    local cd_path=$(echo "$output" | grep "^__GWI_CD_TO__:" | sed 's/^__GWI_CD_TO__://')
    [[ -n "$cd_path" && -d "$cd_path" ]] && cd "$cd_path"
  elif [[ "$1" == "merge" ]]; then
    local output=$(command gwi merge "${@:2}")
    echo "$output" | grep -v "^__GWI_CD_TO__:"
    local cd_path=$(echo "$output" | grep "^__GWI_CD_TO__:" | sed 's/^__GWI_CD_TO__://')
    [[ -n "$cd_path" && -d "$cd_path" ]] && cd "$cd_path"
  else
    command gwi "$@"
  fi
}`

func runInit(cmd *cobra.Command, args []string) {
	// Shell type doesn't matter - we output the same for both zsh and bash
	fmt.Println(shellIntegration)
}
