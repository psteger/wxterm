//go:build js

package ui

import (
	"strings"
	"syscall/js"

	tea "github.com/charmbracelet/bubbletea"
)

func openBrowser(url string) tea.Cmd {
	return func() tea.Msg {
		if !strings.HasPrefix(url, "https://") {
			return nil
		}
		js.Global().Call("open", url, "_blank", "noopener")
		return nil
	}
}
