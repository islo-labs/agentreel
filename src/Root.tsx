import { Composition } from "remotion";
import { CastVideo } from "./CastVideo";
import { defaultProps, CastProps } from "./types";

// Duration constants per mode
const REEL = { title: 2.5, termHighlight: 4.5, browserHighlight: 7.0, end: 3.5 };
const DEMO = { title: 2.0, termHighlight: 12.0, browserHighlight: 10.0, end: 3.0 };

export const RemotionRoot: React.FC = () => {
  return (
    <Composition
      id="CastVideo"
      component={CastVideo as unknown as React.FC<Record<string, unknown>>}
      durationInFrames={450}
      fps={30}
      width={1080}
      height={1080}
      defaultProps={defaultProps as unknown as Record<string, unknown>}
      calculateMetadata={({ props }) => {
        const p = props as unknown as CastProps;
        const fps = 30;
        const isDemo = p.mode === "demo";
        const timing = isDemo ? DEMO : REEL;

        const titleFrames = Math.round(timing.title * fps);
        const highlightFrames = p.highlights.reduce((sum, h) => {
          const dur = h.videoSrc ? timing.browserHighlight : timing.termHighlight;
          return sum + Math.round(dur * fps);
        }, 0);
        const endFrames = Math.round(timing.end * fps);

        return {
          durationInFrames: titleFrames + highlightFrames + endFrames,
          width: isDemo ? 1920 : 1080,
          height: isDemo ? 1080 : 1080,
        };
      }}
    />
  );
};
