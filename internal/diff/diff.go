package diff

import (
	"os/exec"
	"strings"
	"time"
)

// Summary holds everything we need to animate a diff.
type Summary struct {
	Title    string
	Repo     string
	Branch   string
	Files    []FileDiff
	Stats    Stats
	Duration time.Duration // from commit timestamps
}

type FileDiff struct {
	Path      string
	Added     int
	Removed   int
	Lines     []DiffLine // actual diff lines (trimmed to interesting parts)
}

type DiffLine struct {
	Text string // raw text (without +/- prefix)
	Type LineType
}

type LineType int

const (
	LineContext LineType = iota
	LineAdd
	LineRemove
	LineHeader // @@ ... @@
)

type Stats struct {
	FilesChanged int
	Additions    int
	Deletions    int
}

// FromGit reads the current git state and returns a Summary.
// It compares staged + unstaged changes against HEAD.
func FromGit() (Summary, error) {
	s := Summary{}

	// Repo name
	out, err := run("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return s, err
	}
	parts := strings.Split(strings.TrimSpace(out), "/")
	s.Repo = parts[len(parts)-1]

	// Branch
	out, err = run("git", "rev-parse", "--abbrev-ref", "HEAD")
	if err == nil {
		s.Branch = strings.TrimSpace(out)
	}

	// Title from branch
	s.Title = branchToTitle(s.Branch, s.Repo)

	// Get diff (staged + unstaged vs HEAD)
	out, err = run("git", "diff", "HEAD")
	if err != nil {
		// Maybe no commits yet, try just diff
		out, _ = run("git", "diff")
	}
	if strings.TrimSpace(out) == "" {
		// Try diff of last commit
		out, _ = run("git", "diff", "HEAD~1", "HEAD")
	}

	s.Files = parseDiff(out)

	for _, f := range s.Files {
		s.Stats.FilesChanged++
		s.Stats.Additions += f.Added
		s.Stats.Deletions += f.Removed
	}

	// Duration from recent commits
	s.Duration = estimateDuration()

	return s, nil
}

// FromRef compares two git refs and returns a Summary.
func FromRef(base string) (Summary, error) {
	s := Summary{}

	out, _ := run("git", "rev-parse", "--show-toplevel")
	parts := strings.Split(strings.TrimSpace(out), "/")
	s.Repo = parts[len(parts)-1]

	out, _ = run("git", "rev-parse", "--abbrev-ref", "HEAD")
	s.Branch = strings.TrimSpace(out)
	s.Title = branchToTitle(s.Branch, s.Repo)

	out, err := run("git", "diff", base+"...HEAD")
	if err != nil {
		out, _ = run("git", "diff", base, "HEAD")
	}

	s.Files = parseDiff(out)
	for _, f := range s.Files {
		s.Stats.FilesChanged++
		s.Stats.Additions += f.Added
		s.Stats.Deletions += f.Removed
	}

	s.Duration = estimateDuration()
	return s, nil
}

func parseDiff(raw string) []FileDiff {
	var files []FileDiff
	var current *FileDiff

	for _, line := range strings.Split(raw, "\n") {
		if strings.HasPrefix(line, "diff --git") {
			if current != nil {
				files = append(files, *current)
			}
			current = &FileDiff{}
			continue
		}
		if current == nil {
			continue
		}
		if strings.HasPrefix(line, "+++ b/") {
			current.Path = strings.TrimPrefix(line, "+++ b/")
			continue
		}
		if strings.HasPrefix(line, "--- ") || strings.HasPrefix(line, "+++ ") || strings.HasPrefix(line, "index ") {
			continue
		}
		if strings.HasPrefix(line, "@@") {
			current.Lines = append(current.Lines, DiffLine{Text: line, Type: LineHeader})
			continue
		}
		if strings.HasPrefix(line, "+") {
			current.Added++
			current.Lines = append(current.Lines, DiffLine{Text: line[1:], Type: LineAdd})
		} else if strings.HasPrefix(line, "-") {
			current.Removed++
			current.Lines = append(current.Lines, DiffLine{Text: line[1:], Type: LineRemove})
		} else if strings.HasPrefix(line, " ") {
			current.Lines = append(current.Lines, DiffLine{Text: line[1:], Type: LineContext})
		}
	}
	if current != nil && current.Path != "" {
		files = append(files, *current)
	}

	// Trim to most interesting lines per file (keep ±context around changes)
	for i := range files {
		files[i].Lines = trimLines(files[i].Lines, 20)
	}

	return files
}

// trimLines keeps up to maxLines, prioritizing add/remove lines with 1 line of context.
func trimLines(lines []DiffLine, maxLines int) []DiffLine {
	if len(lines) <= maxLines {
		return lines
	}

	// Mark lines near changes as important
	important := make([]bool, len(lines))
	for i, l := range lines {
		if l.Type == LineAdd || l.Type == LineRemove || l.Type == LineHeader {
			for j := max(0, i-1); j <= min(len(lines)-1, i+1); j++ {
				important[j] = true
			}
		}
	}

	var result []DiffLine
	for i, l := range lines {
		if important[i] {
			result = append(result, l)
			if len(result) >= maxLines {
				break
			}
		}
	}
	return result
}

func estimateDuration() time.Duration {
	// Get timestamps of last 2 commits
	out, err := run("git", "log", "--format=%ct", "-2")
	if err != nil {
		return 0
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	if len(lines) < 2 {
		return 0
	}
	var t1, t2 int64
	fmt_sscanf(lines[0], &t1)
	fmt_sscanf(lines[1], &t2)
	if t1 > t2 {
		return time.Duration(t1-t2) * time.Second
	}
	return 0
}

func fmt_sscanf(s string, v *int64) {
	for _, c := range s {
		if c >= '0' && c <= '9' {
			*v = *v*10 + int64(c-'0')
		}
	}
}

func branchToTitle(branch, repo string) string {
	if branch == "" || branch == "main" || branch == "master" || branch == "HEAD" {
		return repo
	}
	clean := strings.NewReplacer("/", " ", "-", " ", "_", " ").Replace(branch)
	return clean
}

func run(name string, args ...string) (string, error) {
	out, err := exec.Command(name, args...).Output()
	return string(out), err
}
