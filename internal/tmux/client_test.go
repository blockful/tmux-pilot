package tmux

import "testing"

func TestParseSessions(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    int
		wantErr bool
	}{
		{"multiple", "main\t3\t1\napi\t1\t0\n", 2, false},
		{"single", "dev\t5\t1\n", 1, false},
		{"empty", "", 0, false},
		{"bad number", "x\tnan\t0\n", 0, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseSessions(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("err = %v, wantErr %v", err, tt.wantErr)
			}
			if !tt.wantErr && len(got) != tt.want {
				t.Errorf("got %d sessions, want %d", len(got), tt.want)
			}
		})
	}
}

func TestParseSessions_Fields(t *testing.T) {
	sessions, _ := parseSessions("main\t3\t1\napi\t1\t0\n")
	if sessions[0].Name != "main" || sessions[0].WindowCount != 3 || !sessions[0].Attached {
		t.Errorf("session 0: %+v", sessions[0])
	}
	if sessions[1].Attached {
		t.Error("session 1 should be detached")
	}
}
