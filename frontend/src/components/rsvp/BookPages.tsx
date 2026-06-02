import { useEffect } from "react";
import { motion, useReducedMotion, type Variants } from "framer-motion";
import { playPageFlip } from "../../lib/page-flip-sound";

interface Props {
  pageCount: number;
  direction: "forward" | "backward";
  onComplete?: () => void;
}

const pageVariants: Variants = {
  closed: { rotateY: 0 },
  flipped: { rotateY: -178 },
};

export function BookPages({ pageCount, direction, onComplete }: Props) {
  const pages = Array.from({ length: pageCount }, (_, i) => i);
  const flipping = direction === "forward";
  const reducedMotion = useReducedMotion();
  const stagger = direction === "forward" ? 0.12 : 0.09;
  const delayChildren = 0.05;

  useEffect(() => {
    if (reducedMotion) return;
    const volume = direction === "forward" ? 0.18 : 0.13;
    const timers = pages.map((_, i) =>
      window.setTimeout(
        () => playPageFlip(volume),
        Math.round((delayChildren + i * stagger) * 1000),
      ),
    );
    return () => {
      for (const id of timers) window.clearTimeout(id);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [direction, reducedMotion, pageCount]);

  return (
    <motion.div
      aria-hidden="true"
      className="absolute inset-0 z-20 pointer-events-none"
      style={{ transformStyle: "preserve-3d" }}
      variants={{
        forward: {
          transition: { staggerChildren: stagger, delayChildren },
        },
        backward: {
          transition: { staggerChildren: stagger, staggerDirection: -1, delayChildren },
        },
      }}
      animate={direction}
      onAnimationComplete={onComplete}
    >
      {pages.map((i) => (
        <motion.div
          key={i}
          className="book-face absolute inset-0"
          style={{
            transformOrigin: "left center",
            zIndex: pageCount - i,
          }}
          variants={pageVariants}
          initial={flipping ? "closed" : "flipped"}
          animate={flipping ? "flipped" : "closed"}
          transition={{ duration: 0.42, ease: [0.32, 0.72, 0, 1] }}
        >
          <div className="absolute inset-0 book-parchment border-r border-gold-muted/30 shadow-[2px_0_8px_rgba(0,0,0,0.08)]" />
          <div className="absolute inset-0 book-spine-shadow pointer-events-none" />
          <div className="absolute inset-4 border border-gold/15 pointer-events-none" />
        </motion.div>
      ))}
    </motion.div>
  );
}
