"""
CLI demo recorder. Uses `claude` CLI to plan the demo, then records it.

Usage:
    python cli_demo.py <command> <workdir> <output_cast> [context]
"""
import getpass
import json
import os
import pty
import re
import select
import socket
import subprocess
import sys
import time


def find_claude():
    """Find the claude CLI binary."""
    import shutil
    claude = shutil.which("claude")
    if claude:
        return claude
    # Common locations
    for path in [
        os.path.expanduser("~/.local/bin/claude"),
        "/usr/local/bin/claude",
        os.path.expanduser("~/.claude/bin/claude"),
    ]:
        if os.path.isfile(path):
            return path
    return "claude"


def generate_demo_plan(command: str, context: str, guidelines: str = "") -> list[dict]:
    """Use claude CLI to plan a demo sequence."""
    guidelines_block = f"\n\nIMPORTANT guidelines you MUST follow:\n{guidelines}" if guidelines else ""

    prompt = f"""You are planning a terminal demo for a CLI tool. The tool is invoked with: {command}

Context about what this tool does:
{context}{guidelines_block}

Generate a JSON array of demo steps. Each step is an object with:
- "type": "command" (run a shell command)
- "value": the command string
- "delay": seconds to wait after (for dramatic effect)
- "description": one-line description

Make the demo:
- Show the most impressive features
- 5-8 steps max
- Start simple, build to the wow moment
- Use realistic commands

Return ONLY the JSON array, no markdown, no explanation."""

    claude = find_claude()
    print(f"Using claude at: {claude}", file=sys.stderr)

    result = subprocess.run(
        [claude, "-p", prompt, "--output-format", "text"],
        capture_output=True,
        text=True,
        timeout=300,
    )

    if result.returncode != 0:
        print(f"claude stderr: {result.stderr}", file=sys.stderr)
        raise RuntimeError(f"claude failed: {result.stderr}")

    text = result.stdout.strip()
    print(f"Claude output: {text[:200]}...", file=sys.stderr)

    # Strip markdown fences if present
    if "```" in text:
        # Find JSON between fences
        parts = text.split("```")
        for part in parts:
            part = part.strip()
            if part.startswith("json"):
                part = part[4:].strip()
            if part.startswith("["):
                text = part
                break

    return json.loads(text)


def record_demo(steps: list[dict], workdir: str, output_path: str):
    """Execute demo steps in a PTY and record as asciicast v2."""
    cols, rows = 80, 24
    start_time = time.time()

    with open(output_path, "w") as f:
        header = {
            "version": 2,
            "width": cols,
            "height": rows,
            "timestamp": int(start_time),
        }
        f.write(json.dumps(header) + "\n")

        # Build patterns to strip user identity from output
        _user = getpass.getuser()
        _host = socket.gethostname().split(".")[0]
        _home = os.path.expanduser("~")
        _title_seq = re.compile(r'\x1b\][\d;]*[^\x07\x1b]*(?:\x07|\x1b\\)')
        _identity = re.compile(
            r'|'.join(re.escape(s) for s in {_user, _host, _home} if s),
            re.IGNORECASE,
        )

        def _sanitize(text: str) -> str:
            text = _title_seq.sub('', text)
            text = _identity.sub('user', text)
            return text

        def write_event(event_type: str, data: str, sanitize: bool = False):
            if sanitize:
                data = _sanitize(data)
            elapsed = time.time() - start_time
            f.write(json.dumps([round(elapsed, 6), event_type, data]) + "\n")
            f.flush()

        for step in steps:
            if step["type"] != "command":
                continue

            cmd = step["value"]
            desc = step.get("description", "")

            # Show description as dim comment
            if desc:
                write_event("o", f"\x1b[38;5;245m# {desc}\x1b[0m\r\n")
                time.sleep(0.3)

            # Type the command character by character (no $ prompt — the renderer adds one)
            write_event("o", "")
            for char in cmd:
                write_event("o", char)
                time.sleep(0.04)
            write_event("o", "\r\n")
            time.sleep(0.2)

            # Execute in PTY with sanitized env (hide username/hostname)
            pid, fd = pty.fork()
            if pid == 0:
                os.chdir(workdir)
                os.environ["PS1"] = "$ "
                os.environ["PROMPT_COMMAND"] = ""
                os.environ.pop("BASH_COMMAND", None)
                # Suppress terminal title sequences (user@host)
                os.environ["TERM"] = "dumb"
                os.execvp("/bin/sh", ["/bin/sh", "-c", cmd])
            else:
                deadline = time.time() + 15
                while time.time() < deadline:
                    r, _, _ = select.select([fd], [], [], 0.1)
                    if r:
                        try:
                            data = os.read(fd, 4096)
                            if data:
                                write_event("o", data.decode("utf-8", errors="replace"), sanitize=True)
                            else:
                                break
                        except OSError:
                            break
                    pid_result = os.waitpid(pid, os.WNOHANG)
                    if pid_result[0] != 0:
                        # Drain remaining output
                        try:
                            while True:
                                r, _, _ = select.select([fd], [], [], 0.1)
                                if r:
                                    data = os.read(fd, 4096)
                                    if data:
                                        write_event("o", data.decode("utf-8", errors="replace"), sanitize=True)
                                    else:
                                        break
                                else:
                                    break
                        except OSError:
                            pass
                        break

                try:
                    os.close(fd)
                except OSError:
                    pass

            delay = step.get("delay", 1.0)
            time.sleep(delay)
            write_event("o", "\r\n")

    print(f"Saved: {output_path}", file=sys.stderr)


def extract_highlights(cast_path: str, context: str, guidelines: str = "") -> list[dict]:
    """Ask Claude to pick highlight moments from the recorded session."""
    # Read the asciicast and strip to just the text content
    lines_output = []
    with open(cast_path) as f:
        for i, line in enumerate(f):
            if i == 0:
                continue  # skip header
            try:
                event = json.loads(line)
                if event[1] == "o":
                    lines_output.append(event[2])
            except (json.JSONDecodeError, IndexError):
                continue

    raw_output = "".join(lines_output)
    # Clean ANSI escape codes
    clean = re.sub(r'\x1b\[[0-9;]*[a-zA-Z]', '', raw_output)
    # Strip other common escape sequences (title sets, OSC, etc.)
    clean = re.sub(r'\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)', '', clean)
    # Collapse carriage-return overwrites (spinners, progress bars).
    # \r means "go back to line start" — keep only the final version of each line.
    collapsed_lines = []
    for line in clean.split('\n'):
        parts = line.split('\r')
        final = parts[-1].strip() if parts else ""
        if final:
            # Deduplicate consecutive identical lines (spinner frames)
            if not collapsed_lines or final != collapsed_lines[-1]:
                collapsed_lines.append(final)
    clean = '\n'.join(collapsed_lines)

    guidelines_block = f"\n\nAdditional guidelines: {guidelines}" if guidelines else ""

    # Demo mode: more chapters, more lines, show full flows
    is_demo = "demo" in guidelines.lower() if guidelines else False

    # If terminal output is empty (e.g. tmux sessions, interactive tools),
    # tell Claude to generate representative highlights from context alone
    output_block = f"\nTerminal output:\n---\n{clean[:6000 if is_demo else 3000]}\n---" if clean.strip() else "\n(No terminal output captured — generate representative highlights from the context below.)"

    if is_demo:
        prompt = f"""You are creating chapter-based highlights for a demo walkthrough video.
{output_block}

Context: {context}{guidelines_block}

Create 4-6 chapters that walk through the demo. For each chapter, return:
- "label": chapter name (1-3 words)
- "lines": array of objects with "text" (string), optionally "color" (hex), "bold" (bool), "dim" (bool), "isPrompt" (bool)

Each chapter: 12-20 lines. Include the prompt (isPrompt: true) and realistic output.
Colors: green="#50fa7b", yellow="#f1fa8c", purple="#bd93f9", red="#ff5555", dim="#6272a4", white="#f8f8f2"

Return ONLY a JSON array. No markdown fences."""
    else:
        prompt = f"""You are creating a highlights reel for a CLI demo video.
{output_block}

Context: {context}{guidelines_block}

Pick 3-4 highlight moments. For each, return:
- "label": 1-2 word label
- "lines": array of objects with "text", optionally "color" (hex), "bold", "dim", "isPrompt"
- "zoomLine": (optional) index of the most impressive line

Each highlight: 4-8 lines max.
Colors: green="#50fa7b", yellow="#f1fa8c", purple="#bd93f9", red="#ff5555", dim="#6272a4", white="#f8f8f2"

Return ONLY a JSON array. No markdown fences."""

    claude = find_claude()
    result = subprocess.run(
        [claude, "-p", prompt, "--output-format", "text"],
        capture_output=True,
        text=True,
        timeout=300,
    )

    text = result.stdout.strip()
    if "```" in text:
        parts = text.split("```")
        for part in parts:
            part = part.strip()
            if part.startswith("json"):
                part = part[4:].strip()
            if part.startswith("["):
                text = part
                break

    try:
        return json.loads(text)
    except (json.JSONDecodeError, ValueError):
        print(f"Could not parse highlights, using defaults", file=sys.stderr)
        return [
            {"label": "Run", "lines": [
                {"text": "Running...", "isPrompt": True},
                {"text": "  Done.", "color": "#50fa7b"},
            ]},
        ]


if __name__ == "__main__":
    if len(sys.argv) < 4:
        print("Usage: python cli_demo.py <command> <workdir> <output> [context]", file=sys.stderr)
        print("       python cli_demo.py --highlights <cast_file> <output_json> [context]", file=sys.stderr)
        sys.exit(1)

    # Highlights mode: extract highlights from existing recording
    if sys.argv[1] == "--highlights":
        cast_file = sys.argv[2]
        output = sys.argv[3]
        context = sys.argv[4] if len(sys.argv) > 4 else ""
        guidelines = sys.argv[5] if len(sys.argv) > 5 else ""
        print(f"Extracting highlights from: {cast_file}", file=sys.stderr)
        highlights = extract_highlights(cast_file, context, guidelines)
        with open(output, "w") as f:
            json.dump(highlights, f, indent=2)
        print(f"Saved {len(highlights)} highlights to: {output}", file=sys.stderr)
        sys.exit(0)

    # Record mode: plan + record + extract highlights
    command = sys.argv[1]
    workdir = sys.argv[2]
    output = sys.argv[3]
    context = sys.argv[4] if len(sys.argv) > 4 else ""
    guidelines = sys.argv[5] if len(sys.argv) > 5 else ""

    print(f"Planning demo for: {command}", file=sys.stderr)
    steps = generate_demo_plan(command, context, guidelines)
    print(f"Generated {len(steps)} steps:", file=sys.stderr)
    for s in steps:
        print(f"  $ {s['value']}  — {s.get('description', '')}", file=sys.stderr)

    print("Recording...", file=sys.stderr)
    record_demo(steps, workdir, output)

    # Extract highlights from the recording
    highlights_path = output.replace(".cast", "-highlights.json")
    print("Extracting highlights...", file=sys.stderr)
    highlights = extract_highlights(output, context, guidelines)
    with open(highlights_path, "w") as f:
        json.dump(highlights, f, indent=2)
    print(f"Saved {len(highlights)} highlights to: {highlights_path}", file=sys.stderr)
