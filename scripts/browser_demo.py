"""
Browser demo recorder. Uses claude CLI to generate a Playwright script,
then runs it with video recording.

Usage:
    python browser_demo.py <url> <output> [task]
"""
import asyncio
import os
import subprocess
import sys
import tempfile


def generate_playwright_script(url: str, task: str) -> str:
    """Use claude CLI to generate a Playwright demo script."""
    prompt = (
        f"Generate a Playwright Python async function that demos a web app at {url}. "
        f"Task: {task}. "
        f"The function signature is: async def demo(page). "
        f"It should navigate, wait for load, click key elements, scroll, "
        f"and take about 15 seconds. Return ONLY the function code, no imports."
    )

    result = subprocess.run(
        ["claude", "-p", prompt, "--output-format", "text"],
        capture_output=True,
        text=True,
        timeout=60,
    )

    text = result.stdout.strip()
    if text.startswith("```"):
        text = text.split("\n", 1)[1].rsplit("```", 1)[0].strip()
    return text


async def record_browser_demo(url: str, task: str, output_path: str):
    """Generate and run a Playwright demo with video recording."""
    from playwright.async_api import async_playwright

    print(f"Generating demo script for {url}...", file=sys.stderr)
    script_code = generate_playwright_script(url, task)
    print(f"Script ready", file=sys.stderr)

    video_dir = tempfile.mkdtemp()

    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        context = await browser.new_context(
            viewport={"width": 1280, "height": 800},
            record_video_dir=video_dir,
            record_video_size={"width": 1280, "height": 800},
        )
        page = await context.new_page()

        try:
            # Build and run the demo function
            local_ns = {}
            full_code = f"import asyncio\n{script_code}"
            compiled = compile(full_code, "<demo>", "exec")
            exec(compiled, local_ns)  # noqa: S102 - generated demo code

            if "demo" in local_ns:
                await local_ns["demo"](page)
            else:
                await page.goto(url, wait_until="networkidle")
                await page.wait_for_timeout(5000)
        except Exception as e:
            print(f"Demo error: {e}, falling back to simple load", file=sys.stderr)
            await page.goto(url, wait_until="networkidle")
            await page.wait_for_timeout(3000)

        await page.close()
        await context.close()
        await browser.close()

    for f in os.listdir(video_dir):
        if f.endswith(".webm"):
            src = os.path.join(video_dir, f)
            os.rename(src, output_path)
            print(f"Saved: {output_path}", file=sys.stderr)
            return

    print("No video, taking screenshot fallback", file=sys.stderr)
    async with async_playwright() as p:
        browser = await p.chromium.launch(headless=True)
        pg = await browser.new_page(viewport={"width": 1280, "height": 800})
        await pg.goto(url, wait_until="networkidle")
        await pg.screenshot(path=output_path)
        await browser.close()


if __name__ == "__main__":
    if len(sys.argv) < 3:
        print("Usage: python browser_demo.py <url> <output> [task]", file=sys.stderr)
        sys.exit(1)

    asyncio.run(record_browser_demo(sys.argv[1], sys.argv[2],
                                     sys.argv[3] if len(sys.argv) > 3 else "Explore the main features"))
