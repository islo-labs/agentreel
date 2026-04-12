# agentreel

Turn your apps into launch videos.

## Run

```bash
npm install
node bin/agentreel.mjs --help
npx remotion studio  # preview video template
```

## Architecture

- `bin/agentreel.mjs` — CLI: PR context, planning, recording, highlight extraction, rendering
- `src/CastVideo.tsx` — Remotion video composition (1080x1080, text slides / panels / diagrams / terminal / browser)
- `src/types.ts` — highlight types (Highlight, DiagramData, PanelData, etc.)
- `src/Root.tsx` — composition config and timing
- `public/music.mp3` — default background track ("Boogie Funky" by Petrushka Sound)
