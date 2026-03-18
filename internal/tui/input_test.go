package tui

import (
	"os"
	"testing"
)

func TestReadKey(t *testing.T) {
	tests := []struct {
		name     string
		input    []byte
		expected Key
	}{
		{"regular char", []byte{'a'}, Key{Type: KeyRune, Rune: 'a'}},
		{"space", []byte{' '}, Key{Type: KeyRune, Rune: ' '}},
		{"enter", []byte{'\n'}, Key{Type: KeyEnter}},
		{"carriage return", []byte{'\r'}, Key{Type: KeyEnter}},
		{"backspace", []byte{'\x7f'}, Key{Type: KeyBackspace}},
		{"ctrl-c", []byte{'\x03'}, Key{Type: KeyCtrlC}},
		{"escape", []byte{'\x1b'}, Key{Type: KeyEscape}},
		{"up arrow", []byte{'\x1b', '[', 'A'}, Key{Type: KeyUp}},
		{"down arrow", []byte{'\x1b', '[', 'B'}, Key{Type: KeyDown}},
		{"alt+x escape sequence", []byte{'\x1b', 'x'}, Key{Type: KeyEscape}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a temporary pipe to simulate stdin
			r, w, err := os.Pipe()
			if err != nil {
				t.Fatal(err)
			}
			defer r.Close()
			defer w.Close()

			// Replace stdin
			oldStdin := os.Stdin
			os.Stdin = r
			defer func() { os.Stdin = oldStdin }()

			// Write test input
			go func() {
				w.Write(tt.input)
				w.Close()
			}()

			// Read the key
			got, err := ReadKey()
			if err != nil {
				t.Fatalf("ReadKey() error = %v", err)
			}

			if got.Type != tt.expected.Type {
				t.Errorf("ReadKey() Type = %v, want %v", got.Type, tt.expected.Type)
			}
			if got.Rune != tt.expected.Rune {
				t.Errorf("ReadKey() Rune = %v, want %v", got.Rune, tt.expected.Rune)
			}
		})
	}
}

func TestKeyTypes(t *testing.T) {
	// Test that all key types are distinct
	types := []KeyType{KeyRune, KeyUp, KeyDown, KeyEnter, KeyEscape, KeyBackspace, KeyCtrlC}
	
	for i, typ1 := range types {
		for j, typ2 := range types {
			if i != j && typ1 == typ2 {
				t.Errorf("KeyType %v and %v have the same value", typ1, typ2)
			}
		}
	}
}