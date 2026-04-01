package vt

import (
	"fmt"
	"testing"
)

func TestBasicText(t *testing.T) {
	s := NewScreen(60, 20)
	s.Write([]byte("$ echo hello world\r\n"))
	s.Write([]byte("hello world\r\n"))
	s.Write([]byte("$ ls -la\r\n"))

	for y := 0; y < 5; y++ {
		line := ""
		for x := 0; x < 30; x++ {
			ch := s.Cells[y][x].Char
			if ch == 0 {
				ch = ' '
			}
			line += string(ch)
		}
		fmt.Printf("Row %d: '%s'\n", y, line)
	}

	// Check first row has content
	if s.Cells[0][0].Char != '$' {
		t.Errorf("Row 0 col 0: got %q, want '$'", s.Cells[0][0].Char)
	}
	if s.Cells[0][2].Char != 'e' {
		t.Errorf("Row 0 col 2: got %q, want 'e'", s.Cells[0][2].Char)
	}
	if s.Cells[1][0].Char != 'h' {
		t.Errorf("Row 1 col 0: got %q, want 'h'", s.Cells[1][0].Char)
	}
}

func TestANSIColors(t *testing.T) {
	s := NewScreen(60, 20)
	// Red text: ESC[31m
	s.Write([]byte("\x1b[31mred text\x1b[0m normal"))

	// Check red text cell
	if s.Cells[0][0].Char != 'r' {
		t.Errorf("got char %q, want 'r'", s.Cells[0][0].Char)
	}
	if s.Cells[0][0].FG == DefaultFG {
		t.Error("expected non-default FG color for 'r'")
	}
	// Check normal text after reset
	if s.Cells[0][9].FG != DefaultFG {
		t.Error("expected default FG after reset")
	}
}
