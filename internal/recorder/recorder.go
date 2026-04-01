package recorder

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/creack/pty"
	"golang.org/x/term"

	"github.com/adamgold/agentcast/internal/asciicast"
	"github.com/adamgold/agentcast/internal/storage"
)

type Options struct {
	Title string
	Shell string // override $SHELL
}

type Result struct {
	ID       string
	Path     string
	Duration float64
	Size     int64
}

// Record starts a PTY-based terminal recording session.
// It blocks until the spawned shell exits or a signal is received.
func Record(opts Options) (Result, error) {
	shell := opts.Shell
	if shell == "" {
		shell = os.Getenv("SHELL")
	}
	if shell == "" {
		shell = "/bin/sh"
	}

	id := storage.NewCastID()
	castPath, err := storage.CastPath(id)
	if err != nil {
		return Result{}, fmt.Errorf("storage: %w", err)
	}

	castFile, err := os.Create(castPath)
	if err != nil {
		return Result{}, fmt.Errorf("create cast file: %w", err)
	}
	defer castFile.Close()

	// Get terminal size
	cols, rows, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		cols, rows = 80, 24
	}

	// Auto-title from git branch if no title provided
	title := opts.Title
	if title == "" {
		title = gitAutoTitle()
	}

	// Write asciicast header
	w := asciicast.NewWriter(castFile)
	header := asciicast.Header{
		Width:     cols,
		Height:    rows,
		Timestamp: time.Now().Unix(),
		Title:     title,
		Env: map[string]string{
			"SHELL": shell,
			"TERM":  os.Getenv("TERM"),
		},
	}
	if err := w.WriteHeader(header); err != nil {
		return Result{}, fmt.Errorf("write header: %w", err)
	}

	// Start shell in PTY
	cmd := exec.Command(shell)
	cmd.Env = os.Environ()
	ptmx, err := pty.StartWithSize(cmd, &pty.Winsize{
		Rows: uint16(rows),
		Cols: uint16(cols),
	})
	if err != nil {
		return Result{}, fmt.Errorf("start pty: %w", err)
	}
	defer ptmx.Close()

	// Write lock file
	if err := storage.WriteLock(storage.LockInfo{PID: os.Getpid(), CastID: id}); err != nil {
		return Result{}, fmt.Errorf("write lock: %w", err)
	}
	defer storage.RemoveLock()

	// Put host terminal in raw mode (skip if stdin is not a terminal, e.g. piped input)
	isTTY := term.IsTerminal(int(os.Stdin.Fd()))
	if isTTY {
		oldState, err := term.MakeRaw(int(os.Stdin.Fd()))
		if err != nil {
			return Result{}, fmt.Errorf("raw mode: %w", err)
		}
		defer term.Restore(int(os.Stdin.Fd()), oldState)
	}

	// Handle SIGWINCH (terminal resize)
	sigwinch := make(chan os.Signal, 1)
	signal.Notify(sigwinch, syscall.SIGWINCH)
	defer signal.Stop(sigwinch)

	// Handle SIGTERM/SIGINT for clean shutdown
	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT)
	defer signal.Stop(sigterm)

	startTime := time.Now()
	done := make(chan struct{})

	// Forward stdin -> PTY
	go func() {
		io.Copy(ptmx, os.Stdin)
	}()

	// Capture PTY output -> stdout + cast file
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := ptmx.Read(buf)
			if n > 0 {
				data := buf[:n]
				os.Stdout.Write(data)
				elapsed := time.Since(startTime).Seconds()
				w.WriteEvent(asciicast.Event{
					Time: elapsed,
					Type: asciicast.Output,
					Data: string(data),
				})
			}
			if err != nil {
				close(done)
				return
			}
		}
	}()

	// Handle resize events
	go func() {
		for range sigwinch {
			cols, rows, err := term.GetSize(int(os.Stdin.Fd()))
			if err != nil {
				continue
			}
			pty.Setsize(ptmx, &pty.Winsize{
				Rows: uint16(rows),
				Cols: uint16(cols),
			})
			elapsed := time.Since(startTime).Seconds()
			w.WriteEvent(asciicast.Event{
				Time: elapsed,
				Type: asciicast.Resize,
				Data: fmt.Sprintf("%dx%d", cols, rows),
			})
		}
	}()

	// Wait for shell exit or signal
	select {
	case <-done:
	case <-sigterm:
		cmd.Process.Signal(syscall.SIGHUP)
	}

	cmd.Wait()
	duration := time.Since(startTime).Seconds()

	// Get final file size
	info, _ := castFile.Stat()
	var size int64
	if info != nil {
		size = info.Size()
	}

	return Result{
		ID:       id,
		Path:     castPath,
		Duration: duration,
		Size:     size,
	}, nil
}

// gitAutoTitle generates a title from the current git context.
// Returns empty string if not in a git repo.
func gitAutoTitle() string {
	// Get repo name
	out, err := exec.Command("git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return ""
	}
	repo := filepath.Base(strings.TrimSpace(string(out)))

	// Get branch name
	out, err = exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		return repo
	}
	branch := strings.TrimSpace(string(out))

	// Clean up branch name for display (fix/auth-bug → fix auth bug)
	branch = strings.NewReplacer("/", " ", "-", " ", "_", " ").Replace(branch)

	if branch == "main" || branch == "master" || branch == "HEAD" {
		return repo
	}
	return branch + " in " + repo
}
