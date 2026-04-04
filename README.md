# agentreel

Turn your web apps and CLIs into viral clips.

https://github.com/user-attachments/assets/474fd85d-3b35-48f4-82b8-1b337840fb51

> 🔊 Turn on sound

## Install

```bash
npx agentreel
```

## Usage

```bash
# CLI demo:
agentreel --cmd "npx my-cli-tool"

# Browser demo:
agentreel --url http://localhost:3000

# With context for smarter demo planning:
agentreel --cmd "npx my-tool" --prompt "A CLI that manages cron jobs"
```

## How it works

1. You provide a CLI command or URL
2. AI plans and executes a demo (terminal or browser)
3. AI picks the 3-4 best highlight moments
4. Renders a polished video with music, transitions, and overlays
5. Prompts you to share on Twitter

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
- Claude CLI (`claude`) — used to plan demo sequences

## Credits

Default background music: ["Go Create"](https://uppbeat.io/track/all-good-folks/go-create) by All Good Folks (via [Uppbeat](https://uppbeat.io))

## License

MIT
