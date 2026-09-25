//go:build windows
// +build windows

package zulu

import (
	"fmt"
	"os"
	"time"

	"github.com/inconshreveable/mousetrap"
)

func runMouseTrap(command *Command) {
	if MousetrapHelpText != "" && mousetrap.StartedByExplorer() {
		command.Print(MousetrapHelpText)
		if MousetrapDisplayDuration > 0 {
			time.Sleep(MousetrapDisplayDuration)
		} else {
			command.Println("Press return to continue...")
			fmt.Scanln()
		}
		os.Exit(1)
	}
}
