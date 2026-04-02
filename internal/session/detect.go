package session

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
)

type ResultType int

const (
	ResultUnknown ResultType = iota
	ResultCLI
	ResultBrowser
)

// DetectedResult describes what the agent built and how to demo it.
type DetectedResult struct {
	Type    ResultType
	Command string // CLI command to run for demo, or URL for browser
	WorkDir string // where to run it
}

// DetectResult analyzes a session to figure out what was built.
func DetectResult(sess Session) DetectedResult {
	r := DetectedResult{WorkDir: findWorkDir(sess)}

	// Check for web app signals
	for _, a := range sess.Actions {
		if a.Type == ActionBash {
			cmd := strings.ToLower(a.Command)
			if containsAny(cmd, "npm run dev", "npm start", "npx next", "npx vite", "yarn dev", "pnpm dev", "flask run", "uvicorn", "python -m http", "localhost:") {
				r.Type = ResultBrowser
				r.Command = extractURL(a.Command)
				if r.Command == "" {
					r.Command = "http://localhost:3000"
				}
				return r
			}
		}
	}

	// Check for CLI signals — look at what was built
	// 1. Check package.json for bin field
	if cmd := detectNodeCLI(r.WorkDir, sess); cmd != "" {
		r.Type = ResultCLI
		r.Command = cmd
		return r
	}

	// 2. Check for Go binary
	if cmd := detectGoCLI(r.WorkDir, sess); cmd != "" {
		r.Type = ResultCLI
		r.Command = cmd
		return r
	}

	// 3. Look at the last few bash commands for clues
	if cmd := detectFromBashHistory(sess); cmd != "" {
		r.Type = ResultCLI
		r.Command = cmd
		return r
	}

	return r
}

func detectNodeCLI(workDir string, sess Session) string {
	// Check if a package.json with bin was written
	for _, a := range sess.Actions {
		if (a.Type == ActionWrite || a.Type == ActionEdit) && strings.HasSuffix(a.FilePath, "package.json") {
			data, err := os.ReadFile(a.FilePath)
			if err != nil {
				continue
			}
			var pkg struct {
				Name string            `json:"name"`
				Bin  json.RawMessage   `json:"bin"`
			}
			if json.Unmarshal(data, &pkg) == nil && len(pkg.Bin) > 0 {
				// Has a bin field — it's a CLI
				if pkg.Name != "" {
					return "npx " + pkg.Name + " --help"
				}
			}
		}
	}
	return ""
}

func detectGoCLI(workDir string, sess Session) string {
	for _, a := range sess.Actions {
		if a.Type == ActionBash && strings.Contains(a.Command, "go build") {
			// Extract the binary name from go build -o
			parts := strings.Fields(a.Command)
			for i, p := range parts {
				if p == "-o" && i+1 < len(parts) {
					binary := parts[i+1]
					return binary + " --help"
				}
			}
		}
	}
	return ""
}

func detectFromBashHistory(sess Session) string {
	// Walk backwards through bash actions to find the last "run" command
	// that looks like running the built thing
	var candidates []string
	for i := len(sess.Actions) - 1; i >= 0; i-- {
		a := sess.Actions[i]
		if a.Type != ActionBash {
			continue
		}
		cmd := strings.TrimSpace(a.Command)

		// Skip build/test/install commands
		if containsAny(cmd, "go build", "go test", "npm install", "npm test",
			"git ", "mkdir", "ls ", "cat ", "echo ", "cd ") {
			continue
		}

		// Look for commands that run the built thing
		if containsAny(cmd, "npx ", "./bin/", "./dist/", "go run", "python ", "node ") {
			candidates = append(candidates, cmd)
			if len(candidates) >= 3 {
				break
			}
		}
	}

	if len(candidates) > 0 {
		return candidates[0]
	}
	return ""
}

func findWorkDir(sess Session) string {
	// Look for cwd from session actions
	for _, a := range sess.Actions {
		if a.FilePath != "" {
			dir := filepath.Dir(a.FilePath)
			if dir != "" && dir != "." {
				return dir
			}
		}
	}
	cwd, _ := os.Getwd()
	return cwd
}

func extractURL(cmd string) string {
	// Find localhost:PORT in command
	for _, part := range strings.Fields(cmd) {
		if strings.Contains(part, "localhost:") {
			if !strings.HasPrefix(part, "http") {
				return "http://" + part
			}
			return part
		}
	}
	return ""
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
