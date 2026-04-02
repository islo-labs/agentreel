# Contributing to agentcast

## Setup

```bash
git clone git@github.com:islo-labs/agentcast.git
cd agentcast
npm install

# Python env for demo recording
cd scripts && uv venv .venv && source .venv/bin/activate
uv pip install browser-use playwright anthropic
```

## Running locally

```bash
node bin/agentcast.mjs --help
node bin/agentcast.mjs --cmd "echo hello" -p "test"

# Preview video template in browser:
npx remotion studio
```

## Architecture

| File | What it does |
|------|-------------|
| `bin/agentcast.mjs` | CLI: session parse → detect → capture → render |
| `scripts/cli_demo.py` | Claude plans demo, records in PTY, extracts highlights |
| `scripts/browser_demo.py` | Browser demo via Playwright (coming soon) |
| `src/CastVideo.tsx` | Remotion video composition |
| `src/types.ts` | TypeScript types for highlights |
