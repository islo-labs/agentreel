# agentcast

Turn Claude Code sessions into viral demo videos.

<p align="center">
  <img src="demo.gif" alt="agentcast demo" width="600" />
</p>

## How it works

1. Reads your Claude Code session to understand what was built
2. Claude plans and executes a demo of your CLI tool (or screenshots your web app)
3. Claude picks the highlights — the 3-4 best moments
4. Renders a Screen Studio-quality video with music, cursor, and transitions

## Install

```bash
git clone git@github.com:islo-labs/agentcast.git
cd agentcast
make build
cd web && npm install && cd ..

# Set up Python env for demo recording
cd scripts && uv venv .venv && source .venv/bin/activate
uv pip install browser-use playwright anthropic
playwright install chromium
```

## Usage

```bash
# Auto-detect from your last Claude session:
agentcast

# Manual CLI demo:
agentcast --cmd "npx @islo-labs/overtime" -p "Cron for AI agents"

# Manual browser demo:
agentcast --url http://localhost:3000 -p "My web app"

# With options:
agentcast --cmd "npx my-tool" \
  --title "my-tool" \
  --prompt "What the tool does" \
  --music ~/music/track.mp3 \
  --output demo.mp4
```

## What you get

A 15-20 second video with:
- **Title card** with your tool name
- **3-4 highlight clips** — terminal window on gradient background, text typing with cursor, smooth transitions
- **End card** with install command
- **Background music** with fade in/out

1080x1080, ready for Twitter/X, LinkedIn, Reels.

## Project structure

```
agentcast/
├── cmd/cast/main.go           # entry point
├── internal/
│   ├── cmd/root.go            # orchestrator
│   ├── capture/agent.go       # calls Python scripts
│   └── session/               # Claude Code session parser + detection
├── scripts/
│   ├── cli_demo.py            # Claude plans + records CLI demo + highlights
│   └── browser_demo.py        # browser demo (coming soon)
├── web/
│   ├── src/                   # Remotion video components
│   └── public/music.mp3       # default background track
└── Makefile
```

## Requirements

- Go 1.21+
- Node.js 18+
- Python 3.10+ with uv
- Claude CLI (`claude`)

## Credits

Default background music: ["Go Create"](https://uppbeat.io/track/all-good-folks/go-create) by All Good Folks (via [Uppbeat](https://uppbeat.io))

## License

MIT
