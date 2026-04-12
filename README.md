# agentreel

Turn your apps into launch videos.

https://github.com/user-attachments/assets/ef0f3a97-403f-4534-9b7d-22f2478a5758

> Turn on sound

## Install

```bash
npx agentreel
```

## Usage

```bash
# CLI launch video:
agentreel --cmd "npx my-cli-tool"

# Browser launch video:
agentreel --url http://localhost:3000
```

## How it works

1. You provide a CLI command or URL
2. AI plans and executes the demo (terminal or Playwright browser)
3. AI generates a launch video with text slides, terminal highlights, diagrams, and panels
4. Renders a polished **1080x1080** video with Remotion — ready for Twitter/X, LinkedIn, Reels

## Flags

```
-c, --cmd <command>     CLI command to demo
-u, --url <url>         URL to demo (browser mode)
-t, --title <text>      video title
-s, --subtitle <text>   video subtitle
-o, --output <file>     output file (default: agentreel.mp4)
-a, --auth <file>       Playwright auth state for browser demos
-g, --guidelines <text> highlight generation guidelines
    --music <file>      background music mp3
    --no-share          skip the share prompt
```

## Requirements

- Node.js 18+
- Claude CLI (`claude`) — plans and records demos

## Credits

Default background music: "Boogie Funky" by Petrushka Sound

## License

MIT
