package session

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Action represents one thing the agent did.
type Action struct {
	Time     time.Time
	Type     ActionType
	Tool     string // Read, Write, Edit, Bash, Grep, etc.
	FilePath string // for file operations
	Command  string // for Bash
	Detail   string // short description
	Size     int    // bytes written/edited
}

type ActionType int

const (
	ActionRead ActionType = iota
	ActionWrite
	ActionEdit
	ActionBash
	ActionSearch
	ActionAgent
	ActionOther
)

// Session is a parsed Claude Code session.
type Session struct {
	ID        string
	Project   string
	Title     string // custom title if set
	Prompt    string // first user message
	StartTime time.Time
	EndTime   time.Time
	Actions   []Action
	Duration  time.Duration
}

// Stats returns aggregate stats.
func (s Session) Stats() Stats {
	st := Stats{Duration: s.Duration}
	files := map[string]bool{}
	for _, a := range s.Actions {
		switch a.Type {
		case ActionRead:
			st.Reads++
		case ActionWrite:
			st.Writes++
			files[a.FilePath] = true
			st.BytesWritten += a.Size
		case ActionEdit:
			st.Edits++
			files[a.FilePath] = true
		case ActionBash:
			st.Commands++
		case ActionSearch:
			st.Searches++
		case ActionAgent:
			st.Agents++
		}
	}
	st.FilesChanged = len(files)
	st.TotalActions = len(s.Actions)
	return st
}

type Stats struct {
	Duration     time.Duration
	TotalActions int
	Reads        int
	Writes       int
	Edits        int
	Commands     int
	Searches     int
	Agents       int
	FilesChanged int
	BytesWritten int
}

// FindLatestSession finds the most recent Claude Code session for the current project.
func FindLatestSession() (string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Claude Code encodes the path with - instead of /
	projectKey := strings.ReplaceAll(cwd, "/", "-")

	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	projectDir := filepath.Join(home, ".claude", "projects", projectKey)

	entries, err := os.ReadDir(projectDir)
	if err != nil {
		return "", fmt.Errorf("no Claude Code sessions found for this project")
	}

	// Find the newest .jsonl file (not in subagents/)
	var newest string
	var newestTime time.Time
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".jsonl") {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		if info.ModTime().After(newestTime) {
			newestTime = info.ModTime()
			newest = filepath.Join(projectDir, e.Name())
		}
	}

	if newest == "" {
		return "", fmt.Errorf("no session files found in %s", projectDir)
	}
	return newest, nil
}

// Parse reads a Claude Code JSONL session file and extracts the action timeline.
func Parse(path string) (Session, error) {
	f, err := os.Open(path)
	if err != nil {
		return Session{}, err
	}
	defer f.Close()

	s := Session{
		ID: strings.TrimSuffix(filepath.Base(path), ".jsonl"),
	}

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 10*1024*1024) // 10MB max line

	for scanner.Scan() {
		var raw map[string]json.RawMessage
		if err := json.Unmarshal(scanner.Bytes(), &raw); err != nil {
			continue
		}

		var msgType string
		if v, ok := raw["type"]; ok {
			json.Unmarshal(v, &msgType)
		}

		var timestamp string
		if v, ok := raw["timestamp"]; ok {
			json.Unmarshal(v, &timestamp)
		}
		ts, _ := time.Parse(time.RFC3339Nano, timestamp)

		if !ts.IsZero() {
			if s.StartTime.IsZero() || ts.Before(s.StartTime) {
				s.StartTime = ts
			}
			if ts.After(s.EndTime) {
				s.EndTime = ts
			}
		}

		if msgType == "assistant" {
			parseAssistantMessage(raw, ts, &s)
		}
		if msgType == "user" && s.Prompt == "" {
			s.Prompt = extractPrompt(raw)
		}
		if msgType == "custom-title" {
			var ct struct{ CustomTitle string `json:"customTitle"` }
			data, _ := json.Marshal(raw)
			json.Unmarshal(data, &ct)
			if ct.CustomTitle != "" {
				s.Title = ct.CustomTitle
			}
		}
	}

	if !s.StartTime.IsZero() && !s.EndTime.IsZero() {
		s.Duration = s.EndTime.Sub(s.StartTime)
	}

	sort.Slice(s.Actions, func(i, j int) bool {
		return s.Actions[i].Time.Before(s.Actions[j].Time)
	})

	return s, nil
}

func parseAssistantMessage(raw map[string]json.RawMessage, ts time.Time, s *Session) {
	var msg struct {
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	}
	data, _ := json.Marshal(raw)
	if err := json.Unmarshal(data, &msg); err != nil {
		return
	}

	var blocks []json.RawMessage
	if err := json.Unmarshal(msg.Message.Content, &blocks); err != nil {
		return
	}

	for _, block := range blocks {
		var b struct {
			Type  string          `json:"type"`
			Name  string          `json:"name"`
			Input json.RawMessage `json:"input"`
		}
		if err := json.Unmarshal(block, &b); err != nil {
			continue
		}
		if b.Type != "tool_use" {
			continue
		}

		action := parseToolUse(b.Name, b.Input, ts)
		if action != nil {
			s.Actions = append(s.Actions, *action)
		}
	}
}

func parseToolUse(name string, inputRaw json.RawMessage, ts time.Time) *Action {
	var input map[string]json.RawMessage
	json.Unmarshal(inputRaw, &input)

	getString := func(key string) string {
		if v, ok := input[key]; ok {
			var s string
			json.Unmarshal(v, &s)
			return s
		}
		return ""
	}

	switch name {
	case "Read":
		fp := getString("file_path")
		return &Action{
			Time: ts, Type: ActionRead, Tool: "Read",
			FilePath: fp,
			Detail:   fmt.Sprintf("Read %s", shortPath(fp)),
		}
	case "Write":
		fp := getString("file_path")
		content := getString("content")
		return &Action{
			Time: ts, Type: ActionWrite, Tool: "Write",
			FilePath: fp, Size: len(content),
			Detail: fmt.Sprintf("Write %s (%d bytes)", shortPath(fp), len(content)),
		}
	case "Edit":
		fp := getString("file_path")
		return &Action{
			Time: ts, Type: ActionEdit, Tool: "Edit",
			FilePath: fp,
			Detail: fmt.Sprintf("Edit %s", shortPath(fp)),
		}
	case "Bash":
		cmd := getString("command")
		short := cmd
		if len(short) > 50 {
			short = short[:50] + "..."
		}
		return &Action{
			Time: ts, Type: ActionBash, Tool: "Bash",
			Command: cmd,
			Detail:  fmt.Sprintf("$ %s", short),
		}
	case "Grep", "Glob":
		pattern := getString("pattern")
		return &Action{
			Time: ts, Type: ActionSearch, Tool: name,
			Detail: fmt.Sprintf("%s %s", name, pattern),
		}
	case "Agent":
		desc := getString("description")
		return &Action{
			Time: ts, Type: ActionAgent, Tool: "Agent",
			Detail: fmt.Sprintf("Agent: %s", desc),
		}
	default:
		return nil // skip non-visual tools
	}
}

func shortPath(p string) string {
	parts := strings.Split(p, "/")
	if len(parts) <= 2 {
		return p
	}
	return filepath.Join(parts[len(parts)-2], parts[len(parts)-1])
}

func extractPrompt(raw map[string]json.RawMessage) string {
	var msg struct {
		Message struct {
			Content json.RawMessage `json:"content"`
		} `json:"message"`
	}
	data, _ := json.Marshal(raw)
	json.Unmarshal(data, &msg)

	// Content can be a string or array of blocks
	var text string
	if err := json.Unmarshal(msg.Message.Content, &text); err == nil {
		return cleanPrompt(text)
	}

	var blocks []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	}
	if err := json.Unmarshal(msg.Message.Content, &blocks); err == nil {
		for _, b := range blocks {
			if b.Type == "text" && b.Text != "" {
				return cleanPrompt(b.Text)
			}
		}
	}
	return ""
}

// cleanPrompt extracts a clean, short prompt from potentially long user messages.
func cleanPrompt(s string) string {
	// Take the first meaningful line
	for _, line := range strings.Split(s, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "│") || strings.HasPrefix(line, "├") ||
			strings.HasPrefix(line, "└") || strings.HasPrefix(line, "─") ||
			strings.HasPrefix(line, "┌") || strings.HasPrefix(line, "┐") {
			continue
		}
		if len(line) > 200 {
			line = line[:200]
		}
		return line
	}
	return s
}
