package tui

import (
	"os"
	"os/signal"
	"syscall"

	"golang.org/x/term"
)

// Terminal manages terminal raw mode and restoration.
type Terminal struct {
	fd            int
	originalState *term.State
	width         int
	height        int
}

// NewTerminal creates a new terminal controller.
func NewTerminal() *Terminal {
	return &Terminal{fd: int(os.Stdin.Fd())}
}

// EnterRawMode switches terminal to raw mode for single-keystroke input.
func (t *Terminal) EnterRawMode() error {
	var err error
	t.originalState, err = term.MakeRaw(t.fd)
	if err != nil {
		return err
	}

	t.width, t.height, err = term.GetSize(t.fd)
	if err != nil {
		t.width, t.height = 80, 24
	}

	return nil
}

// SetupSignalHandlers ensures terminal is restored and UI is cleaned up on
// SIGINT/SIGTERM. The onCleanup function is called before terminal restore.
func (t *Terminal) SetupSignalHandlers(onCleanup func()) {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-c
		if onCleanup != nil {
			onCleanup()
		}
		_ = t.Restore()
		os.Exit(130)
	}()
}

// Restore returns terminal to its original state.
func (t *Terminal) Restore() error {
	if t.originalState != nil {
		return term.Restore(t.fd, t.originalState)
	}
	return nil
}

// Size returns the current terminal dimensions (width, height).
func (t *Terminal) Size() (int, int) {
	return t.width, t.height
}
