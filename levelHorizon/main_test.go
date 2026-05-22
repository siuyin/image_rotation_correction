package main

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

func TestParseLine(t *testing.T) {
	tests := []struct {
		name      string
		line      string
		wantAngle float64
		wantOk    bool
	}{
		{"valid line", "0 0.366451 0.229693 0.003736 -0.043064 0", 0.003736 * 180.0 / 3.141592653589793, true},
		{"comment line", "# some comment", 0, false},
		{"empty line", "", 0, false},
		{"insufficient fields", "0 0.366451 0.229693", 0, false},
		{"invalid float", "0 0.366451 0.229693 notafloat -0.043064", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotAngle, gotOk := parseLine(tt.line)
			if gotOk != tt.wantOk {
				t.Errorf("parseLine() ok = %v, want %v", gotOk, tt.wantOk)
			}
			if gotOk && gotAngle != tt.wantAngle {
				t.Errorf("parseLine() angle = %v, want %v", gotAngle, tt.wantAngle)
			}
		})
	}
}

func TestProcessFile(t *testing.T) {
	content := "0 0 0 0.0 0 0\n0 0 0 0.1 0 0\n# comment\n0 0 0 0.2 0 0\n"
	tmpfile, err := os.CreateTemp("", "test_motions")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.Write([]byte(content)); err != nil {
		t.Fatal(err)
	}
	if _, err := tmpfile.Seek(0, 0); err != nil {
		t.Fatal(err)
	}

	var buf bytes.Buffer
	processFile(&buf, tmpfile)
	tmpfile.Close()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")
	if len(lines) != 3 {
		t.Errorf("expected 3 output lines, got %d", len(lines))
	}

	if !strings.Contains(lines[0], "0.0000") {
		t.Errorf("expected frame 0 to have angle 0.0000, got %s", lines[0])
	}
}
