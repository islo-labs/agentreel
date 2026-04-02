# agentreel

Turn your Claude Code sessions into viral demo videos.

https://github.com/user-attachments/assets/070ee610-298c-4989-8d7e-369ca495469e

## Install

```bash
npx agentreel
```

## Usage

```bash
# After Claude builds something, just run:
agentreel

# It reads your session, detects what was built, records a demo,
# picks the highlights, and renders a video. One command.

# Manual mode:
agentreel --cmd "npx my-cli-tool"          # CLI demo
agentreel --url http://localhost:3000       # browser demo
```

## How it works

1. Reads your Claude Code session log
2. Detects what was built — CLI tool or web app
3. Claude plans and executes a demo (terminal or browser)
4. Claude picks the 3-4 best highlight moments
5. Renders a Screen Studio-quality video with music, transitions, and overlays
6. Prompts you to share on Twitter

## What you get

A 15-20 second 1080x1080 video with:
- **Title card** with your project name
- **Highlight clips** — terminal or browser window on animated gradient
- **Text overlays** — bold captions that work on mute
- **Cursor + typing** — looks like someone's actually using it
- **Background music** with fade in/out
- **End CTA** — install command + URL

Ready for Twitter/X, LinkedIn, Reels.

## Supports

- **CLI demos** — records your tool in a terminal, shows the highlights
- **Browser demos** — records your web app via Playwright, shows the key moments

## Requirements

- Node.js 18+
- Python 3.10+
- Claude CLI (`claude`)

## Credits

Default background music: ["Go Create"](https://uppbeat.io/track/all-good-folks/go-create) by All Good Folks (via [Uppbeat](https://uppbeat.io))

## License

MIT
