# agentcast

CLI that turns Claude Code sessions into viral demo videos.

## Run

```bash
npm install
node bin/agentcast.mjs --help
npx remotion studio  # preview video template
```

## Architecture

- `bin/agentcast.mjs` — CLI orchestrator (session → detect → capture → render)
- `scripts/cli_demo.py` — Claude plans + records CLI demo + extracts highlights
- `scripts/browser_demo.py` — browser demo via Playwright (future)
- `src/CastVideo.tsx` — Remotion video composition
- `src/types.ts` — highlight types
- `public/music.mp3` — default background track
