# agentcast

CLI tool that turns Claude Code sessions into viral demo videos.

## Build & Run

```bash
make build           # build to ./bin/agentcast
go run ./cmd/cast    # run directly
```

## Architecture

- `internal/cmd/root.go` — CLI orchestrator
- `internal/capture/agent.go` — calls Python scripts for recording + highlights
- `internal/session/` — Claude Code JSONL parser + auto-detection
- `scripts/cli_demo.py` — Claude-powered demo recording
- `web/src/` — Remotion video components
