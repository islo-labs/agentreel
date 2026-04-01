package player

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/adamgold/agentcast/internal/asciicast"
)

type Options struct {
	Speed float64 // playback speed multiplier (default 1.0)
}

// Play replays an asciicast file to the terminal with timing.
func Play(f *os.File, opts Options) error {
	if opts.Speed <= 0 {
		opts.Speed = 1.0
	}

	r, err := asciicast.NewReader(f)
	if err != nil {
		return fmt.Errorf("parse cast file: %w", err)
	}

	h := r.Header()
	fmt.Fprintf(os.Stderr, "\x1b[2m# Playing %dx%d recording (%.0fx speed, q to quit)\x1b[0m\r\n", h.Width, h.Height, opts.Speed)

	var prevTime float64

	for {
		e, err := r.NextEvent()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read event: %w", err)
		}

		if e.Type != asciicast.Output {
			continue
		}

		// Sleep for the time delta, adjusted by speed
		delta := e.Time - prevTime
		if delta > 0 {
			// Cap idle time at 2 seconds to avoid long waits
			if delta > 2.0 {
				delta = 2.0
			}
			time.Sleep(time.Duration(float64(time.Second) * delta / opts.Speed))
		}
		prevTime = e.Time

		os.Stdout.WriteString(e.Data)
	}

	fmt.Fprintf(os.Stderr, "\r\n\x1b[2m# Playback complete\x1b[0m\r\n")
	return nil
}
