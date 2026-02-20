import { Link } from "@tanstack/react-router";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";

function SealSVG({ className = "" }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 200 220"
      className={className}
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      {/* Outer ring */}
      <circle cx="100" cy="100" r="88" stroke="var(--color-gold)" strokeWidth="1" opacity="0.25" />
      <circle cx="100" cy="100" r="82" stroke="var(--color-gold)" strokeWidth="0.5" opacity="0.15" />

      {/* Shield body */}
      <path
        d="M100 24 L164 48 L164 106 Q164 158 100 180 Q36 158 36 106 L36 48 Z"
        fill="rgba(152,159,91,0.05)"
        stroke="var(--color-gold)"
        strokeWidth="1.5"
        opacity="0.85"
      />
      {/* Shield inner line */}
      <path
        d="M100 34 L156 56 L156 106 Q156 152 100 170 Q44 152 44 106 L44 56 Z"
        fill="none"
        stroke="var(--color-gold-muted)"
        strokeWidth="0.8"
        opacity="0.35"
      />

      {/* Lock shackle */}
      <path
        d="M82 96 L82 82 Q82 66 100 66 Q118 66 118 82 L118 96"
        fill="none"
        stroke="var(--color-gold)"
        strokeWidth="2"
        strokeLinecap="round"
        opacity="0.9"
      />
      {/* Lock body */}
      <rect
        x="74"
        y="94"
        width="52"
        height="40"
        rx="5"
        fill="rgba(196,169,109,0.08)"
        stroke="var(--color-gold)"
        strokeWidth="1.5"
        opacity="0.9"
      />
      {/* Keyhole circle */}
      <circle cx="100" cy="110" r="7" fill="none" stroke="var(--color-gold)" strokeWidth="1.5" opacity="0.7" />
      {/* Keyhole slot */}
      <path
        d="M96 115 L96 124 L104 124 L104 115"
        stroke="var(--color-gold)"
        strokeWidth="1.5"
        strokeLinejoin="round"
        opacity="0.7"
      />

      {/* Corner ornaments */}
      <text x="52" y="56" fill="var(--color-gold)" fontSize="9" opacity="0.3" textAnchor="middle">✦</text>
      <text x="148" y="56" fill="var(--color-gold)" fontSize="9" opacity="0.3" textAnchor="middle">✦</text>
      <text x="52" y="160" fill="var(--color-gold)" fontSize="9" opacity="0.3" textAnchor="middle">✦</text>
      <text x="148" y="160" fill="var(--color-gold)" fontSize="9" opacity="0.3" textAnchor="middle">✦</text>

      {/* Wax seal below */}
      <ellipse cx="100" cy="196" rx="20" ry="8" fill="rgba(152,159,91,0.12)" stroke="var(--color-gold)" strokeWidth="0.8" opacity="0.5" />
      <line x1="100" y1="180" x2="100" y2="188" stroke="var(--color-gold)" strokeWidth="1" opacity="0.3" />
    </svg>
  );
}

interface Props {
  onSwitchIdentity?: () => void;
}

export function UnauthorizedPage({ onSwitchIdentity }: Props) {
  return (
    <div className="relative flex min-h-dvh flex-col items-center justify-center overflow-hidden bg-dark p-8">
      <div className="dot-pattern opacity-[0.04]" aria-hidden="true" />
      <div className="vignette" aria-hidden="true" />

      {/* Top decorative line */}
      <div className="absolute top-0 left-0 right-0 h-px bg-linear-to-r from-transparent via-gold/40 to-transparent" />

      <div className="relative z-1 max-w-[520px] text-center">
        <SealSVG className="w-[180px] md:w-[220px] h-auto mx-auto mb-8" />

        {/* Ornamental divider */}
        <div className="flex items-center justify-center gap-3 mb-6">
          <div className="h-px w-12 bg-gold/30" />
          <span className="text-gold/40 text-[0.6rem] tracking-[0.3em] font-heading uppercase">Acesso Negado</span>
          <div className="h-px w-12 bg-gold/30" />
        </div>

        <h1 className="font-display text-[1.6rem] md:text-[2rem] font-bold text-ivory mb-4 leading-tight">
          Câmara Restrita
        </h1>

        <p className="text-[1rem] md:text-[1.1rem] text-gold-muted leading-relaxed mb-3">
          Esta câmara é reservada ao casal real e sua guarda de honra.
        </p>
        <p className="text-[0.85rem] text-hint leading-relaxed mb-10">
          Vossa identificação não confere os privilégios necessários para gerenciar a lista de presença.
          Apenas os portadores do brasão do noivo ou da noiva podem adentrar.
        </p>

        <div className="flex flex-col sm:flex-row gap-3 justify-center">
          {onSwitchIdentity && (
            <button
              type="button"
              onClick={onSwitchIdentity}
              className="inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.1em] uppercase text-gold-muted no-underline py-[0.65rem] px-6 border border-[rgba(196,169,109,0.25)] transition-all duration-300 hover:bg-[rgba(196,169,109,0.08)] hover:border-gold/50 hover:text-gold cursor-pointer"
            >
              Trocar Identificação
            </button>
          )}
          <Link
            to="/"
            className="inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.1em] uppercase text-gold no-underline py-[0.65rem] px-6 border border-[rgba(196,169,109,0.35)] transition-all duration-300 hover:bg-[rgba(196,169,109,0.1)] hover:border-gold"
          >
            <ArrowLeft size={15} />
            Voltar ao Início
          </Link>
        </div>
      </div>

      {/* Bottom decorative line */}
      <div className="absolute bottom-0 left-0 right-0 h-px bg-linear-to-r from-transparent via-gold/40 to-transparent" />
    </div>
  );
}
