package asciicast

import (
	"encoding/json"
	"fmt"
)

// Event types in asciicast v2.
const (
	Output = "o" // terminal output
	Input  = "i" // keyboard input
	Resize = "r" // terminal resize (data = "COLSxROWS")
	Marker = "m" // navigation marker
)

// Event is a single asciicast v2 event: [time, type, data].
type Event struct {
	Time float64 // seconds since recording start
	Type string  // one of Output, Input, Resize, Marker
	Data string
}

func (e Event) MarshalJSON() ([]byte, error) {
	return json.Marshal([]interface{}{e.Time, e.Type, e.Data})
}

func (e *Event) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return fmt.Errorf("event: expected JSON array: %w", err)
	}
	if len(raw) != 3 {
		return fmt.Errorf("event: expected 3 elements, got %d", len(raw))
	}
	if err := json.Unmarshal(raw[0], &e.Time); err != nil {
		return fmt.Errorf("event time: %w", err)
	}
	if err := json.Unmarshal(raw[1], &e.Type); err != nil {
		return fmt.Errorf("event type: %w", err)
	}
	if err := json.Unmarshal(raw[2], &e.Data); err != nil {
		return fmt.Errorf("event data: %w", err)
	}
	return nil
}
