package capture

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
)

// findScriptsDir finds the scripts/ directory relative to the binary or cwd.
func findScriptsDir() string {
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	candidates := []string{
		filepath.Join(exeDir, "..", "scripts"),
		filepath.Join(exeDir, "scripts"),
		"scripts",
	}
	for _, c := range candidates {
		if _, err := os.Stat(filepath.Join(c, "cli_demo.py")); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return "scripts"
}

// findVenvPython finds the Python binary in the scripts venv.
func findVenvPython() string {
	scriptsDir := findScriptsDir()
	venvPython := filepath.Join(scriptsDir, ".venv", "bin", "python")
	if _, err := os.Stat(venvPython); err == nil {
		return venvPython
	}
	// Fallback to system python
	return "python3"
}

// AgentRecordCLI uses Claude to plan and record a CLI demo.
// Returns the path to an asciicast file.
func AgentRecordCLI(command string, workDir string, context string) (string, error) {
	python := findVenvPython()
	scriptsDir := findScriptsDir()
	script := filepath.Join(scriptsDir, "cli_demo.py")

	outFile := filepath.Join(os.TempDir(), "agentcast-cli-demo.cast")

	args := []string{script, command, workDir, outFile}
	if context != "" {
		args = append(args, context)
	}

	cmd := exec.Command(python, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	fmt.Fprintf(os.Stderr, "Agent planning CLI demo for: %s\n", command)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("cli demo agent failed: %w", err)
	}

	return outFile, nil
}

// ExtractHighlights asks Claude to pick highlight moments from a recorded session.
// Returns path to a JSON file with the highlights.
func ExtractHighlights(castPath string, context string) (string, error) {
	python := findVenvPython()
	scriptsDir := findScriptsDir()
	script := filepath.Join(scriptsDir, "cli_demo.py")

	outFile := castPath + "-highlights.json"

	args := []string{script, "--highlights", castPath, outFile}
	if context != "" {
		args = append(args, context)
	}

	cmd := exec.Command(python, args...)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("highlight extraction failed: %w", err)
	}

	return outFile, nil
}

// AgentRecordBrowser uses browser-use to demo a web app.
// Returns the path to a video or screenshot file.
func AgentRecordBrowser(url string, task string) (string, error) {
	python := findVenvPython()
	scriptsDir := findScriptsDir()
	script := filepath.Join(scriptsDir, "browser_demo.py")

	outFile := filepath.Join(os.TempDir(), "agentcast-browser-demo.webm")

	cmd := exec.Command(python, script, url, outFile, task)
	cmd.Stdout = os.Stderr
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()

	fmt.Fprintf(os.Stderr, "Agent demoing browser app: %s\n", url)
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("browser demo agent failed: %w", err)
	}

	return outFile, nil
}
