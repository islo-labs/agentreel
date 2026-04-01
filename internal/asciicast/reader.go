package asciicast

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
)

// Reader reads asciicast v2 NDJSON from an io.Reader.
type Reader struct {
	scanner *bufio.Scanner
	header  Header
}

func NewReader(r io.Reader) (*Reader, error) {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // up to 1MB lines

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("reading header: %w", err)
		}
		return nil, fmt.Errorf("empty asciicast file")
	}

	var header Header
	if err := json.Unmarshal(scanner.Bytes(), &header); err != nil {
		return nil, fmt.Errorf("parsing header: %w", err)
	}
	if header.Version != 2 {
		return nil, fmt.Errorf("unsupported asciicast version: %d", header.Version)
	}

	return &Reader{scanner: scanner, header: header}, nil
}

// Header returns the parsed header.
func (r *Reader) Header() Header {
	return r.header
}

// NextEvent reads the next event. Returns io.EOF when done.
func (r *Reader) NextEvent() (Event, error) {
	if !r.scanner.Scan() {
		if err := r.scanner.Err(); err != nil {
			return Event{}, err
		}
		return Event{}, io.EOF
	}
	var e Event
	if err := json.Unmarshal(r.scanner.Bytes(), &e); err != nil {
		return Event{}, fmt.Errorf("parsing event: %w", err)
	}
	return e, nil
}
