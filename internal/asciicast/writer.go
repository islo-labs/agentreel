package asciicast

import (
	"encoding/json"
	"io"
	"sync"
)

// Writer writes asciicast v2 NDJSON to an io.Writer.
// It is safe for concurrent use.
type Writer struct {
	w  io.Writer
	mu sync.Mutex
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w: w}
}

// WriteHeader writes the asciicast header as the first line.
func (w *Writer) WriteHeader(h Header) error {
	h.Version = 2
	return w.writeLine(h)
}

// WriteEvent appends an event line.
func (w *Writer) WriteEvent(e Event) error {
	return w.writeLine(e)
}

func (w *Writer) writeLine(v interface{}) error {
	w.mu.Lock()
	defer w.mu.Unlock()
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	data = append(data, '\n')
	_, err = w.w.Write(data)
	return err
}
