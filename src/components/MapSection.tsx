import { useState } from "react";
import MapPin from "lucide-react/dist/esm/icons/map-pin";
import { OrnamentalDivider } from "./OrnamentalDivider";
import { VENUE } from "../config";

export function MapSection() {
  const [mapFlipped, setMapFlipped] = useState(false);

  return (
    <section className="map-section px-6 py-16 md:py-24" id="local">
      <div className="noise-texture opacity-[0.08]" aria-hidden="true" />

      <div className="relative z-1 mx-auto max-w-[900px]">
        <div className="mb-10 text-center md:mb-12">
          <p className="font-heading text-[0.7rem] font-semibold tracking-[0.3em] uppercase text-gold-dark mb-2">
            Celebração
          </p>
          <h2 className="font-display text-[1.75rem] md:text-[2.25rem] font-bold text-burgundy mb-2">
            Local da Cerimônia
          </h2>
          <OrnamentalDivider variant="muted" />
        </div>

        <div className="map-card-container">
          <div className={`map-card${mapFlipped ? " map-card--flipped" : ""}`}>
            {/* Front face — Venue info */}
            <div className="map-card__face flex flex-col items-center justify-center gap-5 text-center p-8">
              <MapPin size={32} color="var(--color-gold-dark)" strokeWidth={1.5} />
              <p className="font-heading text-[0.85rem] font-semibold tracking-[0.15em] uppercase text-gold-dark">
                {VENUE.city}
              </p>
              <h3 className="font-display text-2xl md:text-[1.85rem] font-bold text-burgundy">
                {VENUE.name}
              </h3>
              <p className="text-base text-hint leading-relaxed max-w-[400px]">
                {VENUE.address}
              </p>
              <button
                className="inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.1em] uppercase text-gold-light bg-burgundy border border-burgundy py-[0.65rem] px-6 cursor-pointer transition-all duration-300 mt-2 hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(107,29,29,0.35)] hover:-translate-y-px"
                onClick={() => setMapFlipped(true)}
              >
                <MapPin size={15} />
                Ver no Mapa
              </button>
            </div>

            {/* Back face — Google Maps iframe */}
            <div className="map-card__face map-card__back">
              <button
                className="absolute top-3.5 right-3.5 z-2 inline-flex items-center gap-1.5 font-heading text-[0.65rem] font-semibold tracking-[0.08em] uppercase text-dark-warm bg-[rgba(250,246,239,0.92)] backdrop-blur-[4px] border border-gold-muted py-1.5 px-3.5 cursor-pointer transition-all duration-300 hover:bg-ivory hover:border-gold"
                onClick={() => setMapFlipped(false)}
              >
                Voltar
              </button>
              {mapFlipped && (
                <iframe
                  src={VENUE.mapEmbed}
                  className="block w-full h-full border-0"
                  allowFullScreen
                  loading="lazy"
                  referrerPolicy="no-referrer-when-downgrade"
                  title="Localização do casamento"
                />
              )}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
