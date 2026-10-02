//go:build js

package tea

import (
	"errors"
	"os"
)

func (p *Program) initInput() error { return nil }

func openInputTTY() (*os.File, error) { return nil, errors.New("no tty in browser") }

const suspendSupported = false

func suspendProcess() {}

func (p *Program) listenForResize(done chan struct{}) { close(done) }
