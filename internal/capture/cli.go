package capture

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/creack/pty"

	"github.com/adamgold/agentcast/internal/asciicast"
)

// CLIResult holds the captured CLI session.
type CLIResult struct {
	CastData []byte        // asciicast v2 NDJSON
	Duration time.Duration
}

// RecordCLI runs a command in a PTY and captures the output as asciicast v2.
func RecordCLI(command string, workDir string, timeout time.Duration) (CLIResult, error) {
	if timeout == 0 {
		timeout = 30 * time.Second
	}

	// Parse command
	parts := strings.Fields(command)
	if len(parts) == 0 {
		return CLIResult{}, fmt.Errorf("empty command")
	}

	cmd := exec.Command(parts[0], parts[1:]...)
	cmd.Dir = workDir
	cmd.Env = os.Environ()

	cols, rows := 80, 24

	// Start in PTY
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
	if err != nil {
		return CLIResult{}, fmt.Errorf("start pty: %w", err)
	}
	defer ptmx.Close()

	// Write asciicast to buffer
	var buf bytes.Buffer
	w := asciicast.NewWriter(&buf)
	w.WriteHeader(asciicast.Header{
		Width:     cols,
		Height:    rows,
		Timestamp: time.Now().Unix(),
	})

	startTime := time.Now()
	done := make(chan struct{})

	// Capture output
	go func() {
		readBuf := make([]byte, 32*1024)
		for {
			n, err := ptmx.Read(readBuf)
			if n > 0 {
				elapsed := time.Since(startTime).Seconds()
				w.WriteEvent(asciicast.Event{
					Time: elapsed,
					Type: asciicast.Output,
					Data: string(readBuf[:n]),
				})
			}
			if err != nil {
				close(done)
				return
			}
		}
	}()

	// Wait for command to finish or timeout
	timer := time.NewTimer(timeout)
	select {
	case <-done:
	case <-timer.C:
		cmd.Process.Kill()
	}
	timer.Stop()

	cmd.Wait()
	duration := time.Since(startTime)

	return CLIResult{
		CastData: buf.Bytes(),
		Duration: duration,
	}, nil
}

// RecordCLISequence runs multiple commands in sequence, capturing all output.
func RecordCLISequence(commands []string, workDir string, delayBetween time.Duration) (CLIResult, error) {
	if delayBetween == 0 {
		delayBetween = 500 * time.Millisecond
	}

	var buf bytes.Buffer
	cols, rows := 80, 24
	w := asciicast.NewWriter(&buf)
	w.WriteHeader(asciicast.Header{
		Width:     cols,
		Height:    rows,
		Timestamp: time.Now().Unix(),
	})

	startTime := time.Now()

	for _, command := range commands {
		// Write the prompt + command
		elapsed := time.Since(startTime).Seconds()
		w.WriteEvent(asciicast.Event{
			Time: elapsed,
			Type: asciicast.Output,
			Data: fmt.Sprintf("\x1b[38;5;76m$\x1b[0m %s\r\n", command),
		})

		// Run command and capture output
		parts := strings.Fields(command)
		if len(parts) == 0 {
			continue
		}

		cmd := exec.Command(parts[0], parts[1:]...)
		cmd.Dir = workDir
		cmd.Env = os.Environ()

		ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
			Rows: uint16(rows),
			Cols: uint16(cols),
		})
		if err != nil {
			// Write error and continue
			elapsed = time.Since(startTime).Seconds()
			w.WriteEvent(asciicast.Event{
				Time: elapsed,
				Type: asciicast.Output,
				Data: fmt.Sprintf("\x1b[31mError: %s\x1b[0m\r\n", err),
			})
			continue
		}

		// Capture output
		doneCh := make(chan struct{})
		go func() {
			readBuf := make([]byte, 32*1024)
			for {
				n, readErr := ptmx.Read(readBuf)
				if n > 0 {
					el := time.Since(startTime).Seconds()
					w.WriteEvent(asciicast.Event{
						Time: el,
						Type: asciicast.Output,
						Data: string(readBuf[:n]),
					})
				}
				if readErr != nil {
					close(doneCh)
					return
				}
			}
		}()

		// Wait with timeout
		timer := time.NewTimer(15 * time.Second)
		select {
		case <-doneCh:
		case <-timer.C:
			cmd.Process.Kill()
		}
		timer.Stop()
		cmd.Wait()
		ptmx.Close()

		// Pause between commands
		time.Sleep(delayBetween)
		elapsed = time.Since(startTime).Seconds()
		w.WriteEvent(asciicast.Event{
			Time: elapsed,
			Type: asciicast.Output,
			Data: "\r\n",
		})
	}

	// Drain any remaining reads
	io.Copy(io.Discard, strings.NewReader(""))

	return CLIResult{
		CastData: buf.Bytes(),
		Duration: time.Since(startTime),
	}, nil
}
