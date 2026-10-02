//go:build !js

package ui

import (
	"os/exec"
	"runtime"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func openBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		// Only ever called with a fixed https URL; guard keeps it that way
		// so the shell/opener below never sees anything else.
		if !strings.HasPrefix(url, "https://") {
			return nil
		}
		var cmd *exec.Cmd
		switch runtime.GOOS {
		case "windows":
			cmd = exec.Command("cmd", "/c", "start", url) // #nosec G204 -- fixed opener, url restricted to https above
		case "darwin":
			cmd = exec.Command("open", url) // #nosec G204 -- fixed opener, url restricted to https above
		default:
			cmd = exec.Command("xdg-open", url) // #nosec G204 -- fixed opener, url restricted to https above
		}
		_ = cmd.Run()
		return nil
	}
}
