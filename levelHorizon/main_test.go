package main

import (
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
