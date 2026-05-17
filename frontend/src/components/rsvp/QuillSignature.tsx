import { useEffect, useRef, useState } from "react";
import { motion } from "framer-motion";

interface Props {
  firstName: string;
  onComplete?: () => void;
}

export function QuillSignature({ firstName, onComplete }: Props) {
  const textRef = useRef<SVGTextElement>(null);
  const [length, setLength] = useState(0);

  useEffect(() => {
    if (!textRef.current) return;
    try {
      const measured = textRef.current.getComputedTextLength();
      setLength(measured * 1.4);
    } catch {
      setLength(900);
    }
  }, [firstName]);

  const ready = length > 0;
  const traceDuration = 2.2;
  const fillDelay = 1.9;

  return (
    <div
      role="img"
      aria-label={`Assinatura de ${firstName}`}
      className="relative w-full flex flex-col items-center justify-center"
    >
      <p className="font-heading text-[0.6rem] tracking-[0.3em] uppercase text-gold-dark/70 mb-2">
        Selado por
      </p>

      <svg
        viewBox="0 0 600 200"
        preserveAspectRatio="xMidYMid meet"
        className="w-full max-w-[440px] h-auto"
      >
        {/* Hidden measurement text (renders text length to use for stroke animation) */}
        <text
          ref={textRef}
          x="300"
          y="130"
          textAnchor="middle"
          fontFamily="Tangerine, cursive"
          fontWeight={700}
          fontSize={110}
          fill="transparent"
          stroke="transparent"
        >
          {firstName}
        </text>

        {ready && (
          <motion.text
            x="300"
            y="130"
            textAnchor="middle"
            fontFamily="Tangerine, cursive"
            fontWeight={700}
            fontSize={110}
            stroke="var(--color-dark)"
            strokeWidth={1.6}
            strokeLinecap="round"
            strokeLinejoin="round"
            initial={{ strokeDasharray: length, strokeDashoffset: length, fill: "transparent" }}
            animate={{ strokeDashoffset: 0, fill: "var(--color-dark)" }}
            transition={{
              strokeDashoffset: { duration: traceDuration, ease: "easeInOut" },
              fill: { duration: 0.3, delay: fillDelay, ease: "easeIn" },
            }}
            onAnimationComplete={(definition) => {
              const def = definition as { fill?: string } | string;
              if (typeof def !== "string" && def.fill) onComplete?.();
            }}
          >
            {firstName}
          </motion.text>
        )}

        {/* Quill follower */}
        {ready && (
          <motion.g
            initial={{ x: 0, y: 0, rotate: -15, opacity: 0 }}
            animate={{
              x: [60, 160, 280, 400, 500, 540],
              y: [-30, -10, -22, -5, -25, -10],
              rotate: [-18, -14, -12, -10, -8, -5],
              opacity: [0, 1, 1, 1, 1, 0],
            }}
            transition={{
              duration: traceDuration,
              times: [0, 0.18, 0.42, 0.66, 0.85, 1],
              ease: "easeInOut",
            }}
          >
            <path
              d="M 0 0 L -4 -50 Q 0 -56 4 -50 L 0 0 Z"
              fill="var(--color-gold-dark)"
              stroke="var(--color-dark)"
              strokeWidth={0.8}
            />
            <path
              d="M -3 -10 Q -10 -25 -3 -40 M -3 -10 Q -8 -22 -3 -33 M -3 -10 Q -6 -18 -3 -25"
              fill="none"
              stroke="var(--color-dark-warm)"
              strokeWidth={0.5}
              opacity={0.7}
            />
            <circle cx={0} cy={0} r={1.6} fill="var(--color-dark)" />
          </motion.g>
        )}

        {/* Ink dots scattered along the path */}
        {ready &&
          [120, 220, 340, 460].map((cx, i) => (
            <motion.circle
              key={cx}
              cx={cx}
              cy={140 + (i % 2 === 0 ? 4 : -3)}
              r={1.4}
              fill="var(--color-dark)"
              initial={{ opacity: 0 }}
              animate={{ opacity: [0, 0.6, 0.25] }}
              transition={{
                duration: 0.8,
                delay: 0.3 + i * 0.4,
                times: [0, 0.4, 1],
              }}
            />
          ))}
      </svg>

      {/* Ink-wash overlay */}
      <motion.div
        aria-hidden="true"
        className="absolute inset-0 bg-parchment-dark pointer-events-none"
        initial={{ opacity: 0 }}
        animate={{ opacity: [0, 0.15, 0] }}
        transition={{ duration: 0.6, delay: traceDuration - 0.1, ease: "easeOut" }}
      />
    </div>
  );
}
