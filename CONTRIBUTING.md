# Contributing to agentreel

## Setup

```bash
git clone git@github.com:islo-labs/agentreel.git
cd agentreel
npm install

# Python env for demo recording
cd scripts && uv venv .venv && source .venv/bin/activate
uv pip install browser-use playwright anthropic
```

## Running locally

```bash
node bin/agentreel.mjs --help
node bin/agentreel.mjs --cmd "echo hello" -p "test"

# Preview video template:
npx remotion studio
```

## Architecture

| File | What it does |
|------|-------------|
| `bin/agentreel.mjs` | CLI: session parse → detect → capture → render |
| `scripts/cli_demo.py` | Claude plans demo, records in PTY, extracts highlights |
| `scripts/browser_demo.py` | Browser demo via Playwright |
| `src/CastVideo.tsx` | Remotion video composition |
| `src/types.ts` | TypeScript types for highlights |
