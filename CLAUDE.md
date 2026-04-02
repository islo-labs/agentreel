# agentreel

CLI that turns Claude Code sessions into viral demo videos.

## Run

```bash
npm install
node bin/agentreel.mjs --help
npx remotion studio  # preview video template
```

## Architecture

- `bin/agentreel.mjs` — CLI orchestrator (session → detect → capture → render)
- `scripts/cli_demo.py` — Claude plans + records CLI demo + extracts highlights
- `scripts/browser_demo.py` — browser demo via Playwright
- `src/CastVideo.tsx` — Remotion video composition
- `src/types.ts` — highlight types
- `public/music.mp3` — default background track
