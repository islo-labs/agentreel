import React from "react";
import {
  AbsoluteFill,
  Audio,
  interpolate,
  spring,
  staticFile,
  useCurrentFrame,
  useVideoConfig,
  Sequence,
} from "remotion";
import { CastProps, Highlight, TermLine } from "./types";

const ACCENT = "#50fa7b";
const DIM = "#6272a4";
const WHITE = "#f8f8f2";
const TERM_BG = "#282a36";
const TITLE_BAR = "#1e1f29";
const CURSOR_COLOR = "#f8f8f2";

// Timing constants (in seconds)
const TITLE_DUR = 2.5;
const HIGHLIGHT_DUR = 4;
const TRANSITION_DUR = 0.4;
const END_DUR = 2.5;

export const CastVideo: React.FC<CastProps> = ({
  title,
  subtitle,
  highlights,
  endText,
  gradient,
}) => {
  const { fps } = useVideoConfig();
  const g = gradient || ["#0f0f1a", "#1a0f2e"];

  const titleFrames = Math.round(TITLE_DUR * fps);
  const highlightFrames = Math.round(HIGHLIGHT_DUR * fps);
  const endFrames = Math.round(END_DUR * fps);

  return (
    <AbsoluteFill
      style={{
        background: `linear-gradient(135deg, ${g[0]}, ${g[1]})`,
      }}
    >
      {/* Background music */}
      <MusicTrack />

      {/* Title card */}
      <Sequence durationInFrames={titleFrames}>
        <TitleCard title={title} subtitle={subtitle} />
      </Sequence>

      {/* Highlights */}
      {highlights.map((h, i) => (
        <Sequence
          key={i}
          from={titleFrames + i * highlightFrames}
          durationInFrames={highlightFrames}
        >
          <HighlightClip highlight={h} index={i} total={highlights.length} />
        </Sequence>
      ))}

      {/* End card */}
      <Sequence
        from={titleFrames + highlights.length * highlightFrames}
        durationInFrames={endFrames}
      >
        <EndCard text={endText || title} />
      </Sequence>
    </AbsoluteFill>
  );
};

// ─── Music ────────────────────────────────────────────────

const MusicTrack: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps, durationInFrames } = useVideoConfig();

  // Fade in over 1s, fade out over 1.5s
  const fadeIn = interpolate(frame, [0, fps], [0, 0.3], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const fadeOut = interpolate(
    frame,
    [durationInFrames - fps * 1.5, durationInFrames],
    [0.3, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );
  const volume = Math.min(fadeIn, fadeOut);

  try {
    return <Audio src={staticFile("music.mp3")} volume={volume} />;
  } catch {
    return null; // No music file — that's fine
  }
};

// ─── Cursor Component ─────────────────────────────────────

const Cursor: React.FC<{ visible: boolean; blink?: boolean }> = ({
  visible,
  blink = true,
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  if (!visible) return null;

  // Blink every 500ms
  const blinkOn = blink ? frame % Math.round(fps / 2) < fps / 4 : true;

  return (
    <span
      style={{
        display: "inline-block",
        width: 9,
        height: 18,
        backgroundColor: blinkOn ? CURSOR_COLOR : "transparent",
        marginLeft: 1,
        verticalAlign: "text-bottom",
        borderRadius: 1,
      }}
    />
  );
};

// ─── Title Card ───────────────────────────────────────────

const TitleCard: React.FC<{ title: string; subtitle?: string }> = ({
  title,
  subtitle,
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const titleSpring = spring({ fps, frame, config: { damping: 14 } });
  const subSpring = spring({
    fps,
    frame: Math.max(0, frame - 6),
    config: { damping: 14 },
  });

  const fadeOut = interpolate(
    frame,
    [fps * (TITLE_DUR - TRANSITION_DUR), fps * TITLE_DUR],
    [1, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );

  return (
    <AbsoluteFill
      style={{
        opacity: fadeOut,
        justifyContent: "center",
        alignItems: "center",
      }}
    >
      <div
        style={{
          transform: `scale(${titleSpring}) translateY(${interpolate(titleSpring, [0, 1], [30, 0])}px)`,
          textAlign: "center",
        }}
      >
        <div
          style={{
            fontFamily:
              '-apple-system, BlinkMacSystemFont, "SF Pro Display", sans-serif',
            fontSize: 72,
            fontWeight: 800,
            color: WHITE,
            letterSpacing: -2,
          }}
        >
          {title}
        </div>
        {subtitle && (
          <div
            style={{
              transform: `translateY(${interpolate(subSpring, [0, 1], [20, 0])}px)`,
              opacity: subSpring,
              fontFamily:
                '-apple-system, BlinkMacSystemFont, "SF Pro Display", sans-serif',
              fontSize: 28,
              color: DIM,
              marginTop: 16,
              letterSpacing: 1,
            }}
          >
            {subtitle}
          </div>
        )}
      </div>
    </AbsoluteFill>
  );
};

// ─── Highlight Clip (Screen Studio style) ─────────────────

const HighlightClip: React.FC<{
  highlight: Highlight;
  index: number;
  total: number;
}> = ({ highlight, index, total }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  // Window entrance
  const enterSpring = spring({
    fps,
    frame,
    config: { damping: 18, stiffness: 80 },
  });
  const windowScale = interpolate(enterSpring, [0, 1], [0.85, 1]);
  const windowY = interpolate(enterSpring, [0, 1], [60, 0]);

  // Fade transitions
  const fadeIn = interpolate(frame, [0, fps * TRANSITION_DUR], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const fadeOut = interpolate(
    frame,
    [fps * (HIGHLIGHT_DUR - TRANSITION_DUR), fps * HIGHLIGHT_DUR],
    [1, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );
  const opacity = Math.min(fadeIn, fadeOut);

  // Subtle zoom during clip
  const zoomProgress = interpolate(
    frame,
    [fps * 0.5, fps * HIGHLIGHT_DUR],
    [1, 1.04],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
  );

  // Line timing
  const lineDelay = fps * 0.15;
  const firstLineFrame = fps * 0.3;

  // Figure out which line the cursor is on and where it is
  const lastVisibleLineIdx = highlight.lines.findIndex((_, i) => {
    const lf = firstLineFrame + (i + 1) * lineDelay;
    return frame < lf;
  });
  const cursorLineIdx =
    lastVisibleLineIdx === -1
      ? highlight.lines.length - 1
      : Math.max(0, lastVisibleLineIdx - 1);

  return (
    <AbsoluteFill style={{ opacity }}>
      {/* Label + progress */}
      <div
        style={{
          position: "absolute",
          top: 50,
          left: 0,
          width: "100%",
          textAlign: "center",
          zIndex: 10,
        }}
      >
        <span
          style={{
            fontFamily: "monospace",
            fontSize: 14,
            color: ACCENT,
            letterSpacing: 3,
            textTransform: "uppercase",
            opacity: interpolate(enterSpring, [0, 1], [0, 0.7]),
          }}
        >
          {highlight.label}
        </span>
        <div
          style={{
            marginTop: 12,
            display: "flex",
            justifyContent: "center",
            gap: 8,
          }}
        >
          {Array.from({ length: total }).map((_, i) => (
            <div
              key={i}
              style={{
                width: i === index ? 24 : 8,
                height: 8,
                borderRadius: 4,
                backgroundColor:
                  i === index ? ACCENT : "rgba(255,255,255,0.15)",
              }}
            />
          ))}
        </div>
      </div>

      {/* Terminal window */}
      <AbsoluteFill
        style={{
          justifyContent: "center",
          alignItems: "center",
          padding: 40,
          paddingTop: 110,
        }}
      >
        <div
          style={{
            transform: `scale(${windowScale * zoomProgress}) translateY(${windowY}px)`,
            width: 820,
            borderRadius: 12,
            overflow: "hidden",
            boxShadow:
              "0 30px 100px rgba(0,0,0,0.5), 0 0 0 1px rgba(255,255,255,0.05)",
          }}
        >
          {/* macOS title bar */}
          <div
            style={{
              backgroundColor: TITLE_BAR,
              padding: "12px 16px",
              display: "flex",
              alignItems: "center",
              gap: 8,
            }}
          >
            <div
              style={{
                width: 12,
                height: 12,
                borderRadius: 6,
                backgroundColor: "#ff5555",
              }}
            />
            <div
              style={{
                width: 12,
                height: 12,
                borderRadius: 6,
                backgroundColor: "#f1fa8c",
              }}
            />
            <div
              style={{
                width: 12,
                height: 12,
                borderRadius: 6,
                backgroundColor: "#50fa7b",
              }}
            />
            <div
              style={{
                flex: 1,
                textAlign: "center",
                fontFamily: "monospace",
                fontSize: 13,
                color: "rgba(255,255,255,0.3)",
              }}
            >
              Terminal
            </div>
          </div>

          {/* Terminal body */}
          <div
            style={{
              backgroundColor: TERM_BG,
              padding: "20px 24px",
              minHeight: 280,
            }}
          >
            {highlight.lines.map((line, lineIdx) => {
              const lineFrame = firstLineFrame + lineIdx * lineDelay;
              const lineSpring = spring({
                fps,
                frame: Math.max(0, frame - lineFrame),
                config: { damping: 20, stiffness: 120 },
              });
              const lineOpacity = interpolate(lineSpring, [0, 1], [0, 1]);
              const lineX = interpolate(lineSpring, [0, 1], [10, 0]);

              // Typing effect for prompt lines
              let displayText = line.text;
              let isTyping = false;
              if (line.isPrompt) {
                const typingStart = lineFrame;
                const typingEnd = lineFrame + fps * 0.6;
                if (frame < typingEnd) {
                  const typingProgress = interpolate(
                    frame,
                    [typingStart, typingEnd],
                    [0, 1],
                    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
                  );
                  const chars = Math.floor(typingProgress * line.text.length);
                  displayText = line.text.slice(0, chars);
                  isTyping = chars < line.text.length;
                }
              }

              // Zoom highlight on specific line
              const isZoomed = highlight.zoomLine === lineIdx;
              const zoomScale = isZoomed
                ? interpolate(
                    frame,
                    [lineFrame + fps * 0.5, lineFrame + fps * 1],
                    [1, 1.06],
                    { extrapolateLeft: "clamp", extrapolateRight: "clamp" }
                  )
                : 1;

              // Is this the line where the cursor should be?
              const showCursor = lineIdx === cursorLineIdx;

              return (
                <div
                  key={lineIdx}
                  style={{
                    opacity: lineOpacity,
                    transform: `translateX(${lineX}px) scale(${zoomScale})`,
                    transformOrigin: "left center",
                    fontFamily:
                      '"SF Mono", "Fira Code", "Cascadia Code", monospace',
                    fontSize: 16,
                    lineHeight: 1.7,
                    color: line.dim ? DIM : line.color || WHITE,
                    fontWeight: line.bold ? 700 : 400,
                    whiteSpace: "pre",
                    display: "flex",
                    alignItems: "center",
                  }}
                >
                  {line.isPrompt && (
                    <span style={{ color: ACCENT, marginRight: 8 }}>$</span>
                  )}
                  <span>{displayText}</span>
                  {showCursor && (
                    <Cursor visible={true} blink={!isTyping} />
                  )}
                </div>
              );
            })}
          </div>
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};

// ─── End Card ─────────────────────────────────────────────

const EndCard: React.FC<{ text: string }> = ({ text }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const s = spring({ fps, frame, config: { damping: 14 } });

  return (
    <AbsoluteFill style={{ justifyContent: "center", alignItems: "center" }}>
      <div style={{ transform: `scale(${s})`, textAlign: "center" }}>
        <div
          style={{
            fontFamily: "monospace",
            fontSize: 32,
            color: WHITE,
            padding: "16px 32px",
            borderRadius: 12,
            backgroundColor: "rgba(255,255,255,0.05)",
            border: "1px solid rgba(255,255,255,0.1)",
          }}
        >
          <span style={{ color: DIM }}>$ </span>
          {text}
          <Cursor visible={true} blink={true} />
        </div>
        <div
          style={{
            marginTop: 20,
            fontFamily: "monospace",
            fontSize: 14,
            color: DIM,
            letterSpacing: 2,
          }}
        >
          agentcast.dev
        </div>
      </div>
    </AbsoluteFill>
  );
};
