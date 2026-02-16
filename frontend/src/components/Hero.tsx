import ChevronDown from "lucide-react/dist/esm/icons/chevron-down";
import { CoatOfArms } from "./CoatOfArms";
import { OrnamentalDivider } from "./OrnamentalDivider";
import { Countdown } from "./Countdown";

export function Hero() {
  return (
    <section className="relative flex min-h-dvh items-center justify-center overflow-hidden bg-dark">
      <div className="dot-pattern opacity-[0.04]" aria-hidden="true" />
      <div className="noise-texture opacity-15" aria-hidden="true" />
      <div className="vignette z-1" aria-hidden="true" />

      <div className="relative z-2 w-full max-w-[700px] p-6 pt-20 text-center">
        <div className="anim-fade-in-frame relative border border-[rgba(196,169,109,0.25)] p-6 sm:p-10 md:p-14">
          {/* Frame corners */}
          <span className="absolute -top-px -left-px w-7 h-7 border-t-[2.5px] border-l-[2.5px] border-gold pointer-events-none" aria-hidden="true" />
          <span className="absolute -top-px -right-px w-7 h-7 border-t-[2.5px] border-r-[2.5px] border-gold pointer-events-none" aria-hidden="true" />
          <span className="absolute -bottom-px -left-px w-7 h-7 border-b-[2.5px] border-l-[2.5px] border-gold pointer-events-none" aria-hidden="true" />
          <span className="absolute -bottom-px -right-px w-7 h-7 border-b-[2.5px] border-r-[2.5px] border-gold pointer-events-none" aria-hidden="true" />
          <div className="absolute inset-2 border border-[rgba(196,169,109,0.1)] pointer-events-none" aria-hidden="true" />

          <CoatOfArms
            className="anim-fade-in mx-auto mb-5 w-[72px] md:w-[88px] h-auto"
          />

          <p
            className="anim-fade-in font-heading text-[0.75rem] md:text-[0.85rem] font-semibold tracking-[0.35em] uppercase text-gold mb-1"
            style={{ animationDelay: "0.5s" }}
          >
            Save the Date
          </p>

          <OrnamentalDivider />

          <h1
            className="anim-fade-in-slow font-display text-[clamp(1.2rem,5vw,3.2rem)] font-bold text-ivory leading-[1.2] flex flex-col items-center"
            style={{ animationDelay: "0.8s" }}
          >
            <span>
              <span className="text-gold">P</span>edro{" "}
              <span className="text-gold">A</span>rthur
            </span>
            <span className="italic font-light text-[clamp(1rem,3vw,2rem)] text-gold leading-[1.4] not-italic:font-body">
              &
            </span>
            <span>
              <span className="text-gold">R</span>afaella{" "}
              <span className="text-gold">A</span>raujo
            </span>
          </h1>

          <p
            className="anim-fade-in font-heading text-[0.85rem] md:text-[0.95rem] font-semibold tracking-[0.25em] uppercase mt-3"
            style={{ animationDelay: "1.1s" }}
          >
            <span className="text-gold">Para</span>{" "}
            <span className="text-ivory">Sempre</span>
          </p>

          <p
            className="anim-fade-in font-heading text-base md:text-[1.15rem] font-medium tracking-[0.2em] text-gold-muted"
            style={{ animationDelay: "1s" }}
          >
            12 de Outubro de 2026
          </p>

          <OrnamentalDivider />

          <Countdown />
        </div>
      </div>

      <div
        className="anim-fade-in absolute bottom-8 left-1/2 -translate-x-1/2 z-2 text-gold-muted"
        aria-hidden="true"
        style={{ animationDelay: "2s" }}
      >
        <ChevronDown className="anim-bounce" size={28} strokeWidth={1.5} />
      </div>
    </section>
  );
}
