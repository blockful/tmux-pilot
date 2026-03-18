package tui

import (
	"os"
)

// Key represents a keyboard input.
type Key struct {
	Type KeyType
	Rune rune
}

// KeyType identifies the type of key pressed.
type KeyType int

const (
	KeyRune KeyType = iota
	KeyUp
	KeyDown
	KeyEnter
	KeyEscape
	KeyBackspace
	KeyCtrlC
)

// ReadKey reads a single keystroke from stdin.
func ReadKey() (Key, error) {
	buf := make([]byte, 4)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		return Key{}, err
	}

	if n == 0 {
		return Key{}, nil
	}

	// Handle escape sequences
	if buf[0] == '\x1b' && n > 1 {
		if n >= 3 && buf[1] == '[' {
			switch buf[2] {
			case 'A':
				return Key{Type: KeyUp}, nil
			case 'B':
				return Key{Type: KeyDown}, nil
			}
		}
		// Simple escape
		return Key{Type: KeyEscape}, nil
	}

	// Handle special characters
	switch buf[0] {
	case '\r', '\n':
		return Key{Type: KeyEnter}, nil
	case '\x7f', '\b': // DEL or BS
		return Key{Type: KeyBackspace}, nil
	case '\x03': // Ctrl-C
		return Key{Type: KeyCtrlC}, nil
	case '\x1b': // Escape (single byte)
		return Key{Type: KeyEscape}, nil
	default:
		// Regular character
		if buf[0] >= 32 && buf[0] <= 126 { // Printable ASCII
			return Key{Type: KeyRune, Rune: rune(buf[0])}, nil
		}
		// Non-printable characters map to the byte value as rune
		return Key{Type: KeyRune, Rune: rune(buf[0])}, nil
	}
}