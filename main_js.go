//go:build js && wasm

// Browser entrypoint. The host page (xterm.js) provides the terminal:
//
//	globalThis.wxtermWrite(Uint8Array)   called with every chunk of terminal output
//	globalThis.wxtermExit()              called when the program quits (optional)
//	globalThis.wxtermSize = {cols, rows} initial terminal size
//
// and this file exports back:
//
//	globalThis.wxtermInput(string)       terminal input (keys, mouse reports)
//	globalThis.wxtermResize(cols, rows)  terminal resized
package main

import (
	"io"
	"syscall/js"

	"wxterm/internal/ui"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

// Set by the build script via ldflags.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

type jsWriter struct{}

func (jsWriter) Write(b []byte) (int, error) {
	arr := js.Global().Get("Uint8Array").New(len(b))
	js.CopyBytesToJS(arr, b)
	js.Global().Call("wxtermWrite", arr)
	return len(b), nil
}

// chanReader turns pushed input chunks into an io.Reader for Bubble Tea.
type chanReader struct {
	ch  chan []byte
	buf []byte
}

func (r *chanReader) Read(p []byte) (int, error) {
	if len(r.buf) == 0 {
		b, ok := <-r.ch
		if !ok {
			return 0, io.EOF
		}
		r.buf = b
	}
	n := copy(p, r.buf)
	r.buf = r.buf[n:]
	return n, nil
}

func main() {
	// Output is not a TTY, so tell lipgloss what the xterm.js terminal supports.
	lipgloss.SetColorProfile(termenv.TrueColor)
	lipgloss.SetHasDarkBackground(true)

	in := &chanReader{ch: make(chan []byte, 256)}
	p := tea.NewProgram(
		ui.NewModel(version),
		tea.WithAltScreen(),
		tea.WithMouseCellMotion(),
		tea.WithInput(in),
		tea.WithOutput(jsWriter{}),
		tea.WithoutSignals(),
	)

	// JS callbacks must never block the event loop: drop input if the buffer is
	// full, and send resizes from a goroutine.
	send := func(msg tea.Msg) { go p.Send(msg) }

	js.Global().Set("wxtermInput", js.FuncOf(func(_ js.Value, args []js.Value) any {
		select {
		case in.ch <- []byte(args[0].String()):
		default:
		}
		return nil
	}))
	js.Global().Set("wxtermResize", js.FuncOf(func(_ js.Value, args []js.Value) any {
		send(tea.WindowSizeMsg{Width: args[0].Int(), Height: args[1].Int()})
		return nil
	}))

	if size := js.Global().Get("wxtermSize"); size.Truthy() {
		send(tea.WindowSizeMsg{Width: size.Get("cols").Int(), Height: size.Get("rows").Int()})
	}

	_, _ = p.Run()

	if exit := js.Global().Get("wxtermExit"); exit.Truthy() {
		exit.Invoke()
	}
}
