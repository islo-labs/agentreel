import { Composition } from "remotion";
import { CastVideo } from "./CastVideo";
import { defaultProps, CastProps } from "./types";

export const RemotionRoot: React.FC = () => {
  return (
    <Composition
      id="CastVideo"
      component={CastVideo}
      durationInFrames={450}
      fps={30}
      width={1080}
      height={1080}
      defaultProps={defaultProps}
      calculateMetadata={({ props }: { props: CastProps }) => {
        const fps = 30;
        const titleFrames = Math.round(2.5 * fps);
        const highlightFrames = Math.round(4 * fps) * props.highlights.length;
        const endFrames = Math.round(2.5 * fps);
        return {
          durationInFrames: titleFrames + highlightFrames + endFrames,
        };
      }}
    />
  );
};
