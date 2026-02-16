export function CoatOfArms({ className = "" }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 120 155"
      className={className}
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-label="Brasão Para Sempre"
    >
      {/* Crown */}
      <path
        d="M32 42 L40 14 L50 30 L60 4 L70 30 L80 14 L88 42Z"
        fill="var(--color-gold)"
        stroke="var(--color-gold-dark)"
        strokeWidth="1.2"
      />
      <rect
        x="32"
        y="40"
        width="56"
        height="7"
        rx="1"
        fill="var(--color-gold)"
        stroke="var(--color-gold-dark)"
        strokeWidth="0.8"
      />
      {/* Crown jewels */}
      <circle cx="60" cy="14" r="3" fill="var(--color-burgundy)" />
      <circle cx="44" cy="24" r="2" fill="var(--color-burgundy)" />
      <circle cx="76" cy="24" r="2" fill="var(--color-burgundy)" />
      {/* Crown band details */}
      <circle cx="42" cy="43" r="1.5" fill="var(--color-burgundy)" opacity="0.8" />
      <circle cx="52" cy="43" r="1.5" fill="var(--color-burgundy)" opacity="0.8" />
      <circle cx="60" cy="43" r="1.5" fill="var(--color-gold-dark)" />
      <circle cx="68" cy="43" r="1.5" fill="var(--color-burgundy)" opacity="0.8" />
      <circle cx="78" cy="43" r="1.5" fill="var(--color-burgundy)" opacity="0.8" />
      {/* Shield */}
      <path
        d="M18 50 L18 105 Q18 140 60 152 Q102 140 102 105 L102 50Z"
        fill="var(--color-burgundy)"
        stroke="var(--color-gold)"
        strokeWidth="2.5"
      />
      {/* Inner shield border */}
      <path
        d="M26 57 L26 103 Q26 133 60 144 Q94 133 94 103 L94 57Z"
        fill="none"
        stroke="var(--color-gold)"
        strokeWidth="1"
        opacity="0.4"
      />
      {/* Shield cross pattern */}
      <line x1="60" y1="54" x2="60" y2="148" stroke="var(--color-gold)" strokeWidth="0.8" opacity="0.15" />
      <line x1="22" y1="90" x2="98" y2="90" stroke="var(--color-gold)" strokeWidth="0.8" opacity="0.15" />
      {/* Small cross at top of shield */}
      <path d="M60 62 L60 74 M54 68 L66 68" stroke="var(--color-gold)" strokeWidth="1.8" opacity="0.5" strokeLinecap="round" />
      {/* Initials */}
      <text x="60" y="93" textAnchor="middle" fill="var(--color-gold)" fontFamily="'Cinzel Decorative', serif" fontSize="18" fontWeight="700">
        PR
      </text>
      {/* Intertwined wedding rings */}
      <circle cx="50" cy="118" r="11" fill="none" stroke="var(--color-gold)" strokeWidth="1.8" />
      <circle cx="70" cy="118" r="11" fill="none" stroke="var(--color-gold)" strokeWidth="1.8" />
      {/* Small fleur-de-lis at bottom of shield */}
      <path d="M60 135 Q58 131 56 133 Q58 130 60 128 Q62 130 64 133 Q62 131 60 135Z" fill="var(--color-gold)" opacity="0.6" />
    </svg>
  );
}
