# Contributing to agentcast

## Setup

```bash
git clone git@github.com:islo-labs/agentcast.git
cd agentcast
make build
cd web && npm install && cd ..
```

## Running locally

```bash
./bin/agentcast --help
./bin/agentcast --cmd "echo hello" -p "test"

# Preview the video template:
cd web && npm run dev
```

## Architecture

| Component | What it does |
|-----------|-------------|
| `internal/cmd/root.go` | CLI orchestrator: session → detect → capture → render |
| `internal/capture/agent.go` | Calls Python scripts for recording + highlights |
| `internal/session/` | Parses Claude Code JSONL, detects CLI vs browser |
| `scripts/cli_demo.py` | Claude plans demo, records in PTY, extracts highlights |
| `scripts/browser_demo.py` | Browser demo via Playwright (coming soon) |
| `web/src/CastVideo.tsx` | Remotion video composition |

## Making changes

- **Video template**: Edit `web/src/CastVideo.tsx`, preview with `cd web && npm run dev`
- **CLI pipeline**: Edit `internal/cmd/root.go`
- **Demo recording**: Edit `scripts/cli_demo.py`
- **Session detection**: Edit `internal/session/detect.go`
