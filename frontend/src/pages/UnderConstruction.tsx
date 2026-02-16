import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";

function HourglassSVG({ className = "" }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 160 240"
      className={className}
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      {/* Top plate */}
      <rect x="28" y="16" width="104" height="10" rx="3" fill="var(--color-gold)" opacity="0.9" />
      <rect x="24" y="12" width="112" height="6" rx="2" fill="var(--color-gold)" />
      {/* Decorative knobs */}
      <circle cx="24" cy="15" r="6" fill="var(--color-gold)" />
      <circle cx="136" cy="15" r="6" fill="var(--color-gold)" />
      <circle cx="24" cy="15" r="3" fill="var(--color-burgundy)" />
      <circle cx="136" cy="15" r="3" fill="var(--color-burgundy)" />
      {/* Top glass bulb */}
      <path
        d="M36 26 L36 85 Q36 120 80 130 Q124 120 124 85 L124 26"
        fill="rgba(196, 169, 109, 0.06)"
        stroke="var(--color-gold)"
        strokeWidth="1.5"
        opacity="0.7"
      />
      {/* Sand in top */}
      <path
        d="M50 40 L50 70 Q50 95 80 105 Q110 95 110 70 L110 40Z"
        fill="var(--color-gold-muted)"
        opacity="0.15"
      />
      {/* Sand stream falling */}
      <line x1="80" y1="125" x2="80" y2="155" stroke="var(--color-gold)" strokeWidth="1.5" strokeDasharray="3 5" opacity="0.4">
        <animate attributeName="stroke-dashoffset" from="0" to="-16" dur="1.5s" repeatCount="indefinite" />
      </line>
      {/* Bottom glass bulb */}
      <path
        d="M36 214 L36 155 Q36 120 80 110 Q124 120 124 155 L124 214"
        fill="rgba(196, 169, 109, 0.06)"
        stroke="var(--color-gold)"
        strokeWidth="1.5"
        opacity="0.7"
      />
      {/* Sand accumulated at bottom */}
      <path
        d="M44 214 L44 190 Q44 165 80 155 Q116 165 116 190 L116 214Z"
        fill="var(--color-gold-muted)"
        opacity="0.2"
      />
      {/* Bottom plate */}
      <rect x="28" y="214" width="104" height="10" rx="3" fill="var(--color-gold)" opacity="0.9" />
      <rect x="24" y="222" width="112" height="6" rx="2" fill="var(--color-gold)" />
      {/* Bottom knobs */}
      <circle cx="24" cy="225" r="6" fill="var(--color-gold)" />
      <circle cx="136" cy="225" r="6" fill="var(--color-gold)" />
      <circle cx="24" cy="225" r="3" fill="var(--color-burgundy)" />
      <circle cx="136" cy="225" r="3" fill="var(--color-burgundy)" />
      {/* Center pinch ring */}
      <ellipse cx="80" cy="120" rx="10" ry="4" fill="none" stroke="var(--color-gold)" strokeWidth="1.5" opacity="0.5" />
      {/* Side support bars */}
      <path d="M36 26 Q20 120 36 214" fill="none" stroke="var(--color-gold)" strokeWidth="1" opacity="0.25" />
      <path d="M124 26 Q140 120 124 214" fill="none" stroke="var(--color-gold)" strokeWidth="1" opacity="0.25" />
      {/* Decorative stars */}
      <text x="14" y="75" fill="var(--color-gold)" fontSize="8" opacity="0.3">&#10022;</text>
      <text x="142" y="170" fill="var(--color-gold)" fontSize="8" opacity="0.3">&#10022;</text>
      <text x="8" y="180" fill="var(--color-gold)" fontSize="6" opacity="0.2">&#10022;</text>
      <text x="148" y="65" fill="var(--color-gold)" fontSize="6" opacity="0.2">&#10022;</text>
    </svg>
  );
}

export function UnderConstruction() {
  return (
    <div className="relative flex min-h-dvh flex-col items-center justify-center overflow-hidden bg-dark p-8">
      <div className="dot-pattern opacity-[0.04]" aria-hidden="true" />
      <div className="vignette" aria-hidden="true" />

      <div className="relative z-1 max-w-[480px] text-center">
        <HourglassSVG className="w-[160px] md:w-[200px] h-auto mx-auto mb-8" />

        <h1 className="font-display text-[1.6rem] md:text-[2rem] font-bold text-ivory mb-3">
          Em Construção
        </h1>
        <p className="text-[1.1rem] md:text-[1.25rem] text-gold-muted leading-relaxed mb-8">
          Os artesãos reais estão preparando esta página com todo o cuidado
          que um evento desta grandeza merece. Em breve estará pronta.
        </p>

        <a
          href="/"
          className="inline-flex items-center gap-2 font-heading text-[0.75rem] font-semibold tracking-[0.1em] uppercase text-gold no-underline py-[0.65rem] px-6 border border-[rgba(196,169,109,0.35)] transition-all duration-300 hover:bg-[rgba(196,169,109,0.1)] hover:border-gold"
        >
          <ArrowLeft size={16} />
          Voltar ao Início
        </a>
      </div>
    </div>
  );
}
