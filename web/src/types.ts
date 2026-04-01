export interface CastProps {
  prompt: string;
  screenshotUrl: string; // local file path or data URL
  duration: string; // e.g. "47 seconds"
  cost: string; // e.g. "$0.08"
  filesChanged: number;
  linesAdded: number;
  linesRemoved: number;
}

export const defaultProps: CastProps = {
  prompt: "Build me a landing page",
  screenshotUrl: "https://placehold.co/800x600/1a1a2e/ffffff?text=Your+App+Here",
  duration: "47 seconds",
  cost: "$0.12",
  filesChanged: 8,
  linesAdded: 342,
  linesRemoved: 12,
};
