import { useMemo } from "react";
import { motion, useReducedMotion } from "framer-motion";

interface Props {
  active: boolean;
  count?: number;
}

interface ParticleSeed {
  left: number;
  delay: number;
  duration: number;
  drift: number;
  lift: number;
  size: number;
  repeatDelay: number;
}

export function MagicParticles({ active, count = 16 }: Props) {
  const reducedMotion = useReducedMotion();

  const seeds = useMemo<ParticleSeed[]>(
    () =>
      Array.from({ length: count }, () => ({
        left: Math.random() * 100,
        delay: Math.random() * 2.5,
        duration: 2.4 + Math.random() * 1.8,
        drift: (Math.random() - 0.5) * 60,
        lift: 180 + Math.random() * 120,
        size: 0.6 + Math.random() * 0.8,
        repeatDelay: Math.random() * 1.5,
      })),
    [count],
  );

  if (!active || reducedMotion) return null;

  return (
    <div
      aria-hidden="true"
      className="pointer-events-none absolute inset-0 z-40 overflow-visible"
    >
      {seeds.map((p, i) => (
        <motion.span
          key={i}
          className="absolute text-gold/80 select-none"
          style={{
            left: `${p.left}%`,
            bottom: "0%",
            fontSize: `${p.size}rem`,
            textShadow: "0 0 6px rgba(196,169,109,0.45)",
          }}
          initial={{ opacity: 0, y: 0, x: 0, scale: 0.4 }}
          animate={{
            opacity: [0, 0.9, 0],
            y: [-10, -p.lift],
            x: [0, p.drift],
            scale: [0.4, 1.1, 0.6],
          }}
          transition={{
            duration: p.duration,
            delay: p.delay,
            ease: "easeOut",
            repeat: Infinity,
            repeatDelay: p.repeatDelay,
          }}
        >
          ✦
        </motion.span>
      ))}
    </div>
  );
}
