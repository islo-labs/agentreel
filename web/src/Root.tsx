import { Composition } from "remotion";
import { CastVideo } from "./CastVideo";
import { defaultProps, CastProps } from "./types";

export const RemotionRoot: React.FC = () => {
  return (
    <Composition
      id="CastVideo"
      component={CastVideo}
      durationInFrames={300}
      fps={30}
      width={1080}
      height={1080}
      defaultProps={defaultProps}
      calculateMetadata={({ props }) => {
        // 3 acts: prompt (2s) + result (4s) + stats (2s) + transitions
        const fps = 30;
        const promptFrames = fps * 2;
        const resultFrames = fps * 4;
        const statsFrames = fps * 3;
        const transitionFrames = fps * 0.5 * 2;
        return {
          durationInFrames:
            promptFrames + resultFrames + statsFrames + transitionFrames,
        };
      }}
    />
  );
};
