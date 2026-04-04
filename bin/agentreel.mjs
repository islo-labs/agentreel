#!/usr/bin/env node

import { execFileSync } from "node:child_process";
import { readFileSync, statSync, existsSync, mkdirSync, copyFileSync } from "node:fs";
import { join, dirname, resolve } from "node:path";
import { tmpdir } from "node:os";
import { fileURLToPath } from "node:url";
import { createInterface } from "node:readline";

const __dirname = dirname(fileURLToPath(import.meta.url));
const ROOT = resolve(__dirname, "..");

// ── CLI flags ───────────────────────────────────────────────

function parseArgs() {
  const args = process.argv.slice(2);
  const flags = {};
  for (let i = 0; i < args.length; i++) {
    const arg = args[i];
    if (arg === "--help" || arg === "-h") { printUsage(); process.exit(0); }
    if (arg === "--version" || arg === "-v") {
      const pkg = JSON.parse(readFileSync(join(ROOT, "package.json"), "utf-8"));
      console.log(pkg.version);
      process.exit(0);
    }
    if (arg === "--cmd" || arg === "-c") flags.cmd = args[++i];
    else if (arg === "--url" || arg === "-u") flags.url = args[++i];
    else if (arg === "--prompt" || arg === "-p") flags.prompt = args[++i];
    else if (arg === "--title" || arg === "-t") flags.title = args[++i];
    else if (arg === "--output" || arg === "-o") flags.output = args[++i];
    else if (arg === "--music") flags.music = args[++i];
    else if (arg === "--auth" || arg === "-a") flags.auth = args[++i];
    else if (arg === "--guidelines" || arg === "-g") flags.guidelines = args[++i];
    else if (arg === "--no-share") flags.noShare = true;
  }
  return flags;
}

function printUsage() {
  console.log(`agentreel — Turn your web apps and CLIs into viral clips

Usage:
  agentreel --cmd "npx my-cli-tool"              # CLI demo
  agentreel --url http://localhost:3000           # browser demo

Flags:
  -c, --cmd <command>     CLI command to demo
  -u, --url <url>         URL to demo (browser mode)
  -p, --prompt <text>     description of what the tool does
  -t, --title <text>      video title
  -o, --output <file>     output file (default: agentreel.mp4)
  -a, --auth <file>       Playwright storage state (cookies/auth) for browser demos
  -g, --guidelines <text>  guidelines for highlight generation (e.g. "focus on speed")
      --music <file>      path to background music mp3
      --no-share          skip the share prompt
  -h, --help              show help
  -v, --version           show version`);
}

// ── Recording + Highlights ──────────────────────────────────

function findPython() {
  const venvPython = join(ROOT, "scripts", ".venv", "bin", "python");
  if (existsSync(venvPython)) return venvPython;
  return "python3";
}

function ensureBrowserDeps() {
  const venvDir = join(ROOT, "scripts", ".venv");
  const venvPython = join(venvDir, "bin", "python");
  const browsersDir = join(venvDir, "playwright-browsers");

  if (existsSync(venvPython)) {
    try {
      execFileSync(venvPython, ["-c", "import playwright"], {
        stdio: "ignore",
      });
      return; // all good
    } catch {
      // playwright missing, install below
    }
  } else {
    console.error("  Setting up Python environment...");
    execFileSync("python3", ["-m", "venv", venvDir], {
      stdio: ["ignore", "inherit", "inherit"],
    });
  }

  const pip = join(venvDir, "bin", "pip");
  console.error("  Installing playwright...");
  execFileSync(pip, ["install", "-q", "playwright"], {
    stdio: ["ignore", "inherit", "inherit"],
  });

  console.error("  Installing Chromium (one-time, ~150MB)...");
  execFileSync(venvPython, ["-m", "playwright", "install", "chromium"], {
    stdio: ["ignore", "inherit", "inherit"],
    env: { ...process.env, PLAYWRIGHT_BROWSERS_PATH: browsersDir },
    cwd: tmpdir(),
  });
}

function recordCLI(command, workDir, context) {
  const python = findPython();
  const script = join(ROOT, "scripts", "cli_demo.py");
  const outFile = join(tmpdir(), "agentreel-cli-demo.cast");

  const args = [script, command, workDir, outFile];
  if (context) args.push(context);

  console.error(`Agent planning CLI demo for: ${command}`);
  execFileSync(python, args, { stdio: ["ignore", "inherit", "inherit"], env: process.env });
  return outFile;
}

function extractHighlightsFromCast(castPath, context, guidelines) {
  const python = findPython();
  const script = join(ROOT, "scripts", "cli_demo.py");
  const outFile = castPath + "-highlights.json";

  const args = [script, "--highlights", castPath, outFile];
  if (context) args.push(context);
  if (guidelines) args.push(guidelines);

  execFileSync(python, args, { stdio: ["ignore", "inherit", "inherit"], env: process.env });
  return outFile;
}

// ── Browser Recording ───────────────────────────────────────

function browserEnv() {
  const browsersDir = join(ROOT, "scripts", ".venv", "playwright-browsers");
  return { ...process.env, PLAYWRIGHT_BROWSERS_PATH: browsersDir };
}

function recordBrowser(url, task, authState) {
  const python = findPython();
  const script = join(ROOT, "scripts", "browser_demo.py");
  const outFile = join(tmpdir(), "agentreel-browser-demo.mp4");

  console.error(`Agent demoing browser app: ${url}`);
  const args = [script, url, outFile, task];
  if (authState) args.push("--auth", authState);
  execFileSync(python, args, {
    stdio: ["ignore", "inherit", "inherit"],
    env: browserEnv(),
    timeout: 300000,
  });
  return outFile;
}

function extractBrowserHighlights(videoPath, task) {
  const python = findPython();
  const script = join(ROOT, "scripts", "browser_demo.py");
  const outFile = videoPath + "-highlights.json";

  execFileSync(python, [script, "--highlights", videoPath, outFile, task], {
    stdio: ["ignore", "inherit", "inherit"],
    env: browserEnv(),
  });
  return outFile;
}

// ── Browser Highlight Builder ───────────────────────────────

function buildBrowserHighlights(clicks, videoPath, task, guidelines) {
  const CLIP_DUR = 7;
  const MIN_HIGHLIGHTS = 3;
  const MAX_HIGHLIGHTS = 4;
  // Ask Claude to generate labels/overlays based on the task
  let labels, overlays;
  try {
    const guidelinesLine = guidelines ? `\nGuidelines: ${guidelines}` : "";
    const genPrompt = `Generate exactly 4 highlight labels and overlay captions for a short browser demo video.
Task: ${task}${guidelinesLine}

Return a JSON object: {"labels": ["word1", "word2", "word3", "word4"], "overlays": ["**caption1**", "**caption2**", "**caption3**", "**caption4**"]}
Labels: 1-2 words each, specific to this app (not generic). Overlays: short punchy captions with **markdown bold** for emphasis. Return ONLY JSON.`;
    const result = execFileSync("claude", ["-p", genPrompt, "--output-format", "text"], {
      encoding: "utf-8", timeout: 30000, stdio: ["ignore", "pipe", "ignore"],
    }).trim();
    const parsed = JSON.parse(result.replace(/```json?\n?/g, "").replace(/```/g, "").trim());
    labels = parsed.labels?.length >= 4 ? parsed.labels : null;
    overlays = parsed.overlays?.length >= 4 ? parsed.overlays : null;
  } catch { /* fall through */ }
  if (!labels) labels = ["Overview", "Interact", "Navigate", "Result"];
  if (!overlays) overlays = ["**First look**", "**Key action**", "**Exploring**", "**The result**"];

  // Estimate video duration from last click or default to 25s
  const lastClickTime = clicks.length > 0 ? clicks[clicks.length - 1].timeSec : 0;
  const videoDur = Math.max(25, lastClickTime + 5);

  // Build click-based highlights
  const clickHighlights = [];
  if (clicks.length >= 1) {
    // Group clicks that are within 3s of each other
    const clusters = [];
    let cluster = [clicks[0]];

    for (let i = 1; i < clicks.length; i++) {
      if (clicks[i].timeSec - cluster[cluster.length - 1].timeSec < 3) {
        cluster.push(clicks[i]);
      } else {
        clusters.push(cluster);
        cluster = [clicks[i]];
      }
    }
    clusters.push(cluster);

    // Take top clusters by density, sorted by time
    const ranked = clusters
      .map((c) => ({ cluster: c }))
      .sort((a, b) => b.cluster.length - a.cluster.length)
      .slice(0, MAX_HIGHLIGHTS)
      .sort((a, b) => a.cluster[0].timeSec - b.cluster[0].timeSec);

    for (const r of ranked) {
      const first = r.cluster[0];
      const last = r.cluster[r.cluster.length - 1];
      const center = (first.timeSec + last.timeSec) / 2;
      const startSec = Math.max(0, center - CLIP_DUR / 2);
      const endSec = startSec + CLIP_DUR;

      const hlClicks = r.cluster.map(c => ({
        x: Math.max(0, Math.min(1280, c.x)),
        y: Math.max(0, Math.min(800, c.y)),
        timeSec: c.timeSec - startSec,
      }));

      const focusX = hlClicks.reduce((s, c) => s + c.x, 0) / hlClicks.length / 1280;
      const focusY = hlClicks.reduce((s, c) => s + c.y, 0) / hlClicks.length / 800;

      clickHighlights.push({
        videoSrc: "browser-demo.mp4",
        videoStartSec: Math.round(startSec * 10) / 10,
        videoEndSec: Math.round(endSec * 10) / 10,
        focusX,
        focusY,
        clicks: hlClicks,
      });
    }
  }

  // Pad to MIN_HIGHLIGHTS with evenly-spaced filler clips
  const highlights = [...clickHighlights];
  if (highlights.length < MIN_HIGHLIGHTS) {
    // Find time gaps not covered by existing highlights
    const covered = highlights.map(h => [h.videoStartSec, h.videoEndSec]);
    const fillerCount = MIN_HIGHLIGHTS - highlights.length;

    // Divide full video into slots, pick uncovered ones
    const slotDur = videoDur / (fillerCount + covered.length + 1);
    for (let i = 0; i < fillerCount; i++) {
      const candidate = slotDur * (i + 1);
      // Skip if overlaps with existing highlight
      const overlaps = covered.some(([s, e]) => candidate >= s && candidate <= e);
      const startSec = overlaps
        ? Math.max(0, videoDur - CLIP_DUR * (fillerCount - i))
        : Math.max(0, candidate - CLIP_DUR / 2);
      highlights.push({
        videoSrc: "browser-demo.mp4",
        videoStartSec: Math.round(startSec * 10) / 10,
        videoEndSec: Math.round((startSec + CLIP_DUR) * 10) / 10,
      });
    }
  }

  // Sort by start time and assign labels
  highlights.sort((a, b) => a.videoStartSec - b.videoStartSec);
  for (let i = 0; i < highlights.length; i++) {
    highlights[i].label = labels[i % labels.length];
    highlights[i].overlay = overlays[i % overlays.length];
  }

  console.error(`  ${highlights.length} highlights (${clickHighlights.length} from clicks, ${highlights.length - clickHighlights.length} filler)`);
  return highlights;
}

// ── Render ──────────────────────────────────────────────────

async function renderVideo(props, output, musicPath) {
  const publicDir = join(ROOT, "public");
  if (!existsSync(publicDir)) mkdirSync(publicDir, { recursive: true });
  if (musicPath && existsSync(musicPath)) {
    copyFileSync(musicPath, join(publicDir, "music.mp3"));
  }

  const absOutput = resolve(output);
  const propsJSON = JSON.stringify(props);

  // Render using Remotion's Node.js API — no CLI binary needed
  const { bundle } = await import("@remotion/bundler");
  const { renderMedia, selectComposition } = await import("@remotion/renderer");

  const entryPoint = join(ROOT, "src", "index.ts");

  console.error("  Bundling...");
  const serveUrl = await bundle({
    entryPoint,
    webpackOverride: (config) => config,
  });

  console.error("  Preparing renderer...");
  const composition = await selectComposition({
    serveUrl,
    id: "CastVideo",
    inputProps: props,
    onBrowserDownload: () => {
      console.error("  Downloading renderer (one-time, ~90MB)...");
      return { onProgress: () => {} };
    },
  });

  console.error("  Rendering...");
  await renderMedia({
    composition,
    serveUrl,
    codec: "h264",
    outputLocation: absOutput,
    inputProps: props,
  });

  const size = statSync(absOutput).size;
  console.error(`\nDone: ${output} (${Math.round(size / 1024)} KB)`);
}

// ── Upload + Share ──────────────────────────────────────────

// Video upload placeholder — will add agentreel.dev hosting later
async function uploadVideo(_filePath) {
  return null;
}

function openShareURL(videoURL, text) {
  const tweetText = encodeURIComponent(text);
  const encodedURL = encodeURIComponent(videoURL);
  const intentURL = `https://twitter.com/intent/tweet?text=${tweetText}&url=${encodedURL}`;

  console.error(`\n  Share: ${videoURL}`);
  console.error(`  Tweet: ${intentURL}\n`);

  // Open in browser
  const cmd = process.platform === "darwin" ? "open" : "xdg-open";
  try {
    execFileSync(cmd, [intentURL], { stdio: "ignore" });
  } catch {
    console.error("  (Could not open browser — copy the link above)");
  }
}

function askYesNo(question) {
  return new Promise((resolve) => {
    const rl = createInterface({ input: process.stdin, output: process.stderr });
    rl.question(question, (answer) => {
      rl.close();
      resolve(answer.trim().toLowerCase() !== "n");
    });
  });
}

async function shareFlow(outputPath, title, prompt) {
  const shouldShare = await askYesNo("Share to Twitter? [Y/n] ");
  if (!shouldShare) return;

  const name = title || "this";
  const desc = prompt || "";
  const text = desc
    ? `Introducing ${name} — ${desc}\n\nMade with https://github.com/islo-labs/agentreel`
    : `Introducing ${name}\n\nMade with https://github.com/islo-labs/agentreel`;
  const tweetText = encodeURIComponent(text);
  const intentURL = `https://twitter.com/intent/tweet?text=${tweetText}`;

  console.error(`\n  Opening Twitter — attach your video to the tweet.`);
  console.error(`  Video: ${resolve(outputPath)}\n`);

  const openCmd = process.platform === "darwin" ? "open" : "xdg-open";
  try {
    execFileSync(openCmd, [intentURL], { stdio: "ignore" });
  } catch {
    console.error(`  Link: ${intentURL}`);
  }
}

// ── Auto-describe ──────────────────────────────────────────

function autoDescribe(cmd, url) {
  const target = cmd || url;
  try {
    const result = execFileSync("claude", [
      "-p",
      `Describe what this tool/app does in one short sentence (under 10 words). No quotes, no period. Just the description.\n\n${target}`,
      "--output-format", "text",
    ], { encoding: "utf-8", timeout: 30000, stdio: ["ignore", "pipe", "ignore"] });
    const desc = result.trim();
    if (desc && desc.length < 100) return desc;
  } catch { /* fall through */ }
  return cmd ? cmd.split(/\s+/).pop() : "Web app demo";
}

// ── Main ────────────────────────────────────────────────────

async function main() {
  const flags = parseArgs();
  const output = flags.output || "agentreel.mp4";
  const noShare = flags.noShare;

  let demoCmd = flags.cmd;
  let demoURL = flags.url;
  let prompt = flags.prompt;

  // Auto-generate description if not provided
  if (!prompt) {
    console.error("Generating description...");
    prompt = autoDescribe(demoCmd, demoURL);
    console.error(`  "${prompt}"`);
  }

  if (!demoCmd && !demoURL) {
    console.error("Please provide --cmd or --url.\n");
    printUsage();
    process.exit(1);
  }

  let videoTitle = flags.title || demoCmd || demoURL;

  if (demoCmd) {
    console.error("Step 1/3: Recording CLI demo...");
    const castPath = recordCLI(demoCmd, process.cwd(), prompt);

    console.error("Step 2/3: Extracting highlights...");
    const highlightsPath = extractHighlightsFromCast(castPath, prompt, flags.guidelines);
    const highlights = JSON.parse(readFileSync(highlightsPath, "utf-8"));
    console.error(`  ${highlights.length} highlights extracted`);

    console.error("Step 3/3: Rendering video...");
    await renderVideo({
      title: videoTitle,
      subtitle: prompt,
      highlights,
      endText: demoCmd,
    }, output, flags.music);

    if (!noShare) {
      await shareFlow(resolve(output), videoTitle, prompt);
    }
    return;
  }

  if (demoURL) {
    const task = prompt || "Explore the main features of this app";

    ensureBrowserDeps();
    console.error("Step 1/3: Recording browser demo...");
    const videoPath = recordBrowser(demoURL, task, flags.auth);

    // Copy video to Remotion public dir so it can be served
    const publicDir = join(ROOT, "public");
    if (!existsSync(publicDir)) mkdirSync(publicDir, { recursive: true });
    copyFileSync(videoPath, join(publicDir, "browser-demo.mp4"));

    console.error("Step 2/3: Building highlights...");

    // Read click data — this is the primary signal for highlights
    const clicksPath = videoPath.replace(".mp4", "-clicks.json");
    let allClicks = [];
    if (existsSync(clicksPath)) {
      allClicks = JSON.parse(readFileSync(clicksPath, "utf-8"));
      console.error(`  ${allClicks.length} clicks captured`);
    }

    const highlights = buildBrowserHighlights(allClicks, videoPath, task, flags.guidelines);

    console.error("Step 3/3: Rendering video...");
    await renderVideo({
      title: videoTitle,
      subtitle: prompt,
      highlights,
      endText: demoURL,
      endUrl: demoURL,
    }, output, flags.music);

    if (!noShare) {
      await shareFlow(resolve(output), videoTitle, prompt);
    }
    return;
  }
}

main();
