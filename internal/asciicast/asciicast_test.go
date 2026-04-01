package asciicast

import (
	"bytes"
	"io"
	"testing"
)

func TestRoundTrip(t *testing.T) {
	var buf bytes.Buffer
	w := NewWriter(&buf)

	header := Header{
		Width:     120,
		Height:    30,
		Timestamp: 1711990000,
		Title:     "test recording",
		Env:       map[string]string{"SHELL": "/bin/zsh", "TERM": "xterm-256color"},
	}
	if err := w.WriteHeader(header); err != nil {
		t.Fatalf("WriteHeader: %v", err)
	}

	events := []Event{
		{Time: 0.1, Type: Output, Data: "$ "},
		{Time: 1.2, Type: Output, Data: "ls\r\n"},
		{Time: 1.5, Type: Output, Data: "file1.txt  file2.txt\r\n"},
		{Time: 5.0, Type: Resize, Data: "80x24"},
		{Time: 10.0, Type: Marker, Data: "tool:edit src/main.go"},
	}
	for _, e := range events {
		if err := w.WriteEvent(e); err != nil {
			t.Fatalf("WriteEvent: %v", err)
		}
	}

	r, err := NewReader(&buf)
	if err != nil {
		t.Fatalf("NewReader: %v", err)
	}

	h := r.Header()
	if h.Version != 2 {
		t.Errorf("version = %d, want 2", h.Version)
	}
	if h.Width != 120 || h.Height != 30 {
		t.Errorf("dimensions = %dx%d, want 120x30", h.Width, h.Height)
	}
	if h.Title != "test recording" {
		t.Errorf("title = %q, want %q", h.Title, "test recording")
	}

	for i, want := range events {
		got, err := r.NextEvent()
		if err != nil {
			t.Fatalf("event %d: %v", i, err)
		}
		if got.Time != want.Time || got.Type != want.Type || got.Data != want.Data {
			t.Errorf("event %d = %+v, want %+v", i, got, want)
		}
	}

	_, err = r.NextEvent()
	if err != io.EOF {
		t.Errorf("expected EOF, got %v", err)
	}
}

func TestEventJSON(t *testing.T) {
	e := Event{Time: 1.234, Type: Output, Data: "hello\r\n"}
	data, err := e.MarshalJSON()
	if err != nil {
		t.Fatal(err)
	}
	want := `[1.234,"o","hello\r\n"]`
	if string(data) != want {
		t.Errorf("got %s, want %s", data, want)
	}

	var decoded Event
	if err := decoded.UnmarshalJSON(data); err != nil {
		t.Fatal(err)
	}
	if decoded != e {
		t.Errorf("decoded = %+v, want %+v", decoded, e)
	}
}
