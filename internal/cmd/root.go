package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/adamgold/agentcast/internal/capture"
	"github.com/adamgold/agentcast/internal/session"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "agentcast",
	Short: "Turn Claude Code sessions into viral demo videos",
	Long: `Run after Claude builds something. Reads the session, figures out
what was built, records a demo, picks the highlights, renders a video.

  agentcast                                          # auto-detect from session
  agentcast --cmd "npx @islo-labs/overtime"           # manual CLI demo
  agentcast --url http://localhost:3000               # manual browser demo`,
	RunE: runAgentcast,
}

func runAgentcast(cmd *cobra.Command, args []string) error {
	cmdFlag, _ := cmd.Flags().GetString("cmd")
	urlFlag, _ := cmd.Flags().GetString("url")
	promptFlag, _ := cmd.Flags().GetString("prompt")
	output, _ := cmd.Flags().GetString("output")
	title, _ := cmd.Flags().GetString("title")
	musicFlag, _ := cmd.Flags().GetString("music")
	sessionFlag, _ := cmd.Flags().GetString("session")

	if output == "" {
		output = "agentcast.mp4"
	}

	workDir, _ := os.Getwd()

	// --- Resolve what to demo ---
	demoCmd := cmdFlag
	demoURL := urlFlag
	prompt := promptFlag

	// If no manual flags, read the Claude session to auto-detect
	if demoCmd == "" && demoURL == "" {
		var sessionPath string
		if sessionFlag != "" {
			sessionPath = sessionFlag
		} else {
			var err error
			sessionPath, err = session.FindLatestSession()
			if err != nil {
				return fmt.Errorf("no session found and no --cmd or --url provided.\n\nUsage:\n  agentcast --cmd \"npx my-tool\" -p \"what it does\"\n  agentcast --url http://localhost:3000")
			}
		}

		fmt.Fprintf(os.Stderr, "Reading session: %s\n", filepath.Base(sessionPath))
		sess, err := session.Parse(sessionPath)
		if err != nil {
			return fmt.Errorf("parse session: %w", err)
		}

		if prompt == "" {
			prompt = sess.Prompt
		}

		detected := session.DetectResult(sess)
		switch detected.Type {
		case session.ResultCLI:
			demoCmd = detected.Command
			if detected.WorkDir != "" {
				workDir = detected.WorkDir
			}
			fmt.Fprintf(os.Stderr, "Detected CLI: %s\n", demoCmd)
		case session.ResultBrowser:
			demoURL = detected.Command
			fmt.Fprintf(os.Stderr, "Detected browser: %s\n", demoURL)
		default:
			return fmt.Errorf("couldn't detect what was built. Use --cmd or --url to specify")
		}
	}

	// --- Record the demo ---
	if demoCmd != "" {
		return runCLIDemo(demoCmd, workDir, prompt, title, output, musicFlag)
	}
	if demoURL != "" {
		return runBrowserDemo(demoURL, prompt, title, output, musicFlag)
	}

	return fmt.Errorf("nothing to demo. Use --cmd or --url")
}

func runCLIDemo(demoCmd, workDir, prompt, title, output, musicFlag string) error {
	// 1. Record
	fmt.Fprintf(os.Stderr, "Step 1/3: Recording CLI demo...\n")
	castPath, err := capture.AgentRecordCLI(demoCmd, workDir, prompt)
	if err != nil {
		return fmt.Errorf("record: %w", err)
	}

	// 2. Highlights
	fmt.Fprintf(os.Stderr, "Step 2/3: Extracting highlights...\n")
	highlightsPath, err := capture.ExtractHighlights(castPath, prompt)
	if err != nil {
		return fmt.Errorf("highlights: %w", err)
	}
	highlights, err := readHighlights(highlightsPath)
	if err != nil {
		return err
	}
	fmt.Fprintf(os.Stderr, "  %d highlights extracted\n", len(highlights))

	// 3. Render
	return renderVideo(title, demoCmd, prompt, highlights, output, musicFlag)
}

func runBrowserDemo(demoURL, prompt, title, output, musicFlag string) error {
	// 1. Record browser
	fmt.Fprintf(os.Stderr, "Step 1/3: Recording browser demo...\n")
	task := prompt
	if task == "" {
		task = "Explore the main features of this app"
	}
	_, err := capture.AgentRecordBrowser(demoURL, task)
	if err != nil {
		return fmt.Errorf("browser demo: %w", err)
	}

	// TODO: extract highlights from browser recording + render
	fmt.Fprintf(os.Stderr, "Browser demo recorded. Video rendering coming soon.\n")
	return nil
}

func readHighlights(path string) ([]interface{}, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read highlights: %w", err)
	}
	var highlights []interface{}
	if err := json.Unmarshal(data, &highlights); err != nil {
		return nil, fmt.Errorf("parse highlights: %w", err)
	}
	return highlights, nil
}

func renderVideo(title, endText, subtitle string, highlights []interface{}, output, musicFlag string) error {
	fmt.Fprintf(os.Stderr, "Step 3/3: Rendering video...\n")

	webDir := findWebDir()

	// Copy custom music if provided
	if musicFlag != "" {
		publicDir := filepath.Join(webDir, "public")
		os.MkdirAll(publicDir, 0o755)
		if data, err := os.ReadFile(musicFlag); err == nil {
			os.WriteFile(filepath.Join(publicDir, "music.mp3"), data, 0o644)
		}
	}

	videoTitle := title
	if videoTitle == "" {
		videoTitle = endText
	}

	props := map[string]interface{}{
		"title":      videoTitle,
		"subtitle":   subtitle,
		"highlights": highlights,
		"endText":    endText,
	}
	propsJSON, _ := json.Marshal(props)
	absOutput, _ := filepath.Abs(output)

	renderCmd := exec.Command(
		filepath.Join(webDir, "node_modules", ".bin", "remotion"),
		"render", "CastVideo", absOutput,
		"--props", string(propsJSON),
	)
	renderCmd.Dir = webDir
	renderCmd.Stdout = os.Stderr
	renderCmd.Stderr = os.Stderr

	if err := renderCmd.Run(); err != nil {
		return fmt.Errorf("render: %w", err)
	}

	info, _ := os.Stat(absOutput)
	if info != nil {
		fmt.Fprintf(os.Stderr, "\nDone: %s (%.0f KB)\n", output, float64(info.Size())/1024)
	}
	return nil
}

func findWebDir() string {
	exe, _ := os.Executable()
	exeDir := filepath.Dir(exe)
	for _, c := range []string{
		filepath.Join(exeDir, "..", "web"),
		filepath.Join(exeDir, "web"),
		"web",
	} {
		if _, err := os.Stat(filepath.Join(c, "package.json")); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}
	return "web"
}

func Execute() int {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}
	return 0
}

func init() {
	rootCmd.Version = version
	rootCmd.CompletionOptions.HiddenDefaultCmd = true
	rootCmd.Flags().StringP("cmd", "c", "", "CLI command to demo")
	rootCmd.Flags().StringP("url", "u", "", "URL to demo (browser mode)")
	rootCmd.Flags().StringP("prompt", "p", "", "description of what the tool does")
	rootCmd.Flags().StringP("title", "t", "", "video title")
	rootCmd.Flags().StringP("output", "o", "", "output file (default: agentcast.mp4)")
	rootCmd.Flags().String("music", "", "path to background music mp3")
	rootCmd.Flags().String("session", "", "path to Claude Code session .jsonl")
}
