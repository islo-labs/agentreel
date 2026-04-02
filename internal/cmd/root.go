package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/adamgold/agentcast/internal/capture"
)

var version = "dev"

var rootCmd = &cobra.Command{
	Use:   "agentcast",
	Short: "Generate a viral demo video for your CLI tool",
	Long: `Records your CLI tool, picks the highlights, and renders
a Screen Studio-quality video. Just tell it what to run.

  agentcast --cmd "npx @islo-labs/overtime" -p "Cron for AI agents"`,
	RunE: runAgentcast,
}

func runAgentcast(cmd *cobra.Command, args []string) error {
	cmdFlag, _ := cmd.Flags().GetString("cmd")
	promptFlag, _ := cmd.Flags().GetString("prompt")
	output, _ := cmd.Flags().GetString("output")
	title, _ := cmd.Flags().GetString("title")
	musicFlag, _ := cmd.Flags().GetString("music")

	if cmdFlag == "" {
		return fmt.Errorf("--cmd is required. Example: agentcast --cmd \"npx my-tool\"")
	}
	if output == "" {
		output = "agentcast.mp4"
	}

	workDir, _ := os.Getwd()
	context := promptFlag

	// 1. Record the CLI demo (Claude plans + executes commands)
	fmt.Fprintf(os.Stderr, "Step 1/3: Recording CLI demo...\n")
	castPath, err := capture.AgentRecordCLI(cmdFlag, workDir, context)
	if err != nil {
		return fmt.Errorf("record: %w", err)
	}

	// 2. Extract highlights from the recording
	fmt.Fprintf(os.Stderr, "Step 2/3: Extracting highlights...\n")
	highlightsPath, err := capture.ExtractHighlights(castPath, context)
	if err != nil {
		return fmt.Errorf("highlights: %w", err)
	}

	// Read highlights JSON
	highlightsData, err := os.ReadFile(highlightsPath)
	if err != nil {
		return fmt.Errorf("read highlights: %w", err)
	}

	var highlights []interface{}
	if err := json.Unmarshal(highlightsData, &highlights); err != nil {
		return fmt.Errorf("parse highlights: %w", err)
	}

	fmt.Fprintf(os.Stderr, "  %d highlights extracted\n", len(highlights))

	// 3. Render via Remotion
	fmt.Fprintf(os.Stderr, "Step 3/3: Rendering video...\n")

	// Copy custom music if provided
	webDir := findWebDir()
	if musicFlag != "" {
		publicDir := filepath.Join(webDir, "public")
		os.MkdirAll(publicDir, 0o755)
		if data, err := os.ReadFile(musicFlag); err == nil {
			os.WriteFile(filepath.Join(publicDir, "music.mp3"), data, 0o644)
			fmt.Fprintf(os.Stderr, "  Using music: %s\n", musicFlag)
		}
	}

	videoTitle := title
	if videoTitle == "" {
		videoTitle = cmdFlag
	}

	props := map[string]interface{}{
		"title":      videoTitle,
		"subtitle":   promptFlag,
		"highlights": highlights,
		"endText":    cmdFlag,
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
	rootCmd.Flags().StringP("cmd", "c", "", "CLI command to demo (required)")
	rootCmd.Flags().StringP("prompt", "p", "", "description of what the tool does")
	rootCmd.Flags().StringP("title", "t", "", "video title (defaults to command)")
	rootCmd.Flags().StringP("output", "o", "", "output file (default: agentcast.mp4)")
	rootCmd.Flags().String("music", "", "path to background music mp3")
}
