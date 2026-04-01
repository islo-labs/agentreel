package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/adamgold/agentcast/internal/asciicast"
)

// Recording represents a local .cast file with parsed metadata.
type Recording struct {
	ID       string
	Path     string
	Header   asciicast.Header
	Size     int64
	Duration float64
}

// LockInfo is written to .recording while a session is active.
type LockInfo struct {
	PID    int    `json:"pid"`
	CastID string `json:"cast_id"`
}

// DataDir returns the storage directory, creating it if needed.
func DataDir() (string, error) {
	dir := os.Getenv("XDG_DATA_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".local", "share")
	}
	dir = filepath.Join(dir, "cast")
	return dir, os.MkdirAll(dir, 0o755)
}

// NewCastID generates a timestamp-based recording ID.
func NewCastID() string {
	return time.Now().Format("2006-01-02T15-04-05")
}

// CastPath returns the full path for a cast ID.
func CastPath(id string) (string, error) {
	dir, err := DataDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, id+".cast"), nil
}

// WriteLock writes the active recording lock file.
func WriteLock(info LockInfo) error {
	dir, err := DataDir()
	if err != nil {
		return err
	}
	data, err := json.Marshal(info)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, ".recording"), data, 0o644)
}

// ReadLock reads the active recording lock file.
func ReadLock() (LockInfo, error) {
	dir, err := DataDir()
	if err != nil {
		return LockInfo{}, err
	}
	data, err := os.ReadFile(filepath.Join(dir, ".recording"))
	if err != nil {
		return LockInfo{}, err
	}
	var info LockInfo
	return info, json.Unmarshal(data, &info)
}

// RemoveLock removes the active recording lock file.
func RemoveLock() error {
	dir, err := DataDir()
	if err != nil {
		return err
	}
	return os.Remove(filepath.Join(dir, ".recording"))
}

// List returns all local recordings, newest first.
func List() ([]Recording, error) {
	dir, err := DataDir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var recs []Recording
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".cast") {
			continue
		}
		path := filepath.Join(dir, entry.Name())
		rec, err := readRecording(path)
		if err != nil {
			continue // skip unreadable files
		}
		recs = append(recs, rec)
	}

	sort.Slice(recs, func(i, j int) bool {
		return recs[i].Header.Timestamp > recs[j].Header.Timestamp
	})
	return recs, nil
}

// MostRecent returns the most recent recording.
func MostRecent() (Recording, error) {
	recs, err := List()
	if err != nil {
		return Recording{}, err
	}
	if len(recs) == 0 {
		return Recording{}, fmt.Errorf("no recordings found")
	}
	return recs[0], nil
}

// Resolve returns a recording by ID, or the most recent if id is empty.
func Resolve(id string) (Recording, error) {
	if id == "" {
		return MostRecent()
	}
	path, err := CastPath(id)
	if err != nil {
		return Recording{}, err
	}
	return readRecording(path)
}

func readRecording(path string) (Recording, error) {
	f, err := os.Open(path)
	if err != nil {
		return Recording{}, err
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		return Recording{}, err
	}

	r, err := asciicast.NewReader(f)
	if err != nil {
		return Recording{}, err
	}

	h := r.Header()
	id := strings.TrimSuffix(filepath.Base(path), ".cast")

	// Scan to find duration from last event
	var lastTime float64
	for {
		e, err := r.NextEvent()
		if err != nil {
			break
		}
		lastTime = e.Time
	}

	return Recording{
		ID:       id,
		Path:     path,
		Header:   h,
		Size:     info.Size(),
		Duration: lastTime,
	}, nil
}

// FormatDuration formats seconds as "Xm Ys".
func FormatDuration(secs float64) string {
	d := time.Duration(secs * float64(time.Second))
	m := int(d.Minutes())
	s := int(d.Seconds()) % 60
	if m > 0 {
		return fmt.Sprintf("%dm %ds", m, s)
	}
	return fmt.Sprintf("%ds", s)
}

// FormatSize formats bytes as human-readable.
func FormatSize(bytes int64) string {
	switch {
	case bytes >= 1024*1024:
		return fmt.Sprintf("%.1f MB", float64(bytes)/(1024*1024))
	case bytes >= 1024:
		return fmt.Sprintf("%.1f KB", float64(bytes)/1024)
	default:
		return strconv.FormatInt(bytes, 10) + " B"
	}
}
