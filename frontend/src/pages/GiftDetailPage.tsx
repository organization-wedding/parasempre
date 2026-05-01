import { Link } from "@tanstack/react-router";
import Gift from "lucide-react/dist/esm/icons/gift";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";
import ExternalLink from "lucide-react/dist/esm/icons/external-link";
import { Header } from "../components/Header";
import { isNotFoundError } from "../lib/api";
import { useGiftQuery } from "../lib/gift-queries";
import { formatBRL } from "../lib/format";

interface Props {
  giftId: number;
}

export function GiftDetailPage({ giftId }: Props) {
  const validId = Number.isFinite(giftId) && giftId > 0;
  const { data: gift, isLoading, error } = useGiftQuery(giftId, validId);

  const notFound = !validId || isNotFoundError(error);

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />

      <main className="mx-auto max-w-[1100px] px-6 pt-24 pb-16">
        <Link
          to="/lista-presentes"
          search={{ page: undefined }}
          className="inline-flex items-center gap-1.5 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase text-hint hover:text-burgundy transition-colors no-underline mb-6"
        >
          <ArrowLeft size={14} />
          Lista de presentes
        </Link>

        {isLoading && !notFound ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div
              className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4"
              role="status"
              aria-label="Carregando"
            />
            <span className="text-[0.85rem]">Carregando presente...</span>
          </div>
        ) : notFound ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <Gift size={48} className="text-gold-muted/40 mb-4" />
            <h2 className="font-heading text-[1.05rem] font-semibold text-dark-warm mb-2">
              Presente não disponível
            </h2>
            <p className="text-[0.88rem] text-hint max-w-[380px] mb-6">
              Este presente pode ter sido removido ou está temporariamente indisponível.
            </p>
            <Link
              to="/lista-presentes"
              search={{ page: undefined }}
              className="inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 border border-burgundy text-burgundy hover:bg-burgundy hover:text-gold-light transition-all duration-200 cursor-pointer no-underline"
            >
              <ArrowLeft size={14} />
              Voltar à lista
            </Link>
          </div>
        ) : error ? (
          <div className="mb-6 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
            <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
            <span className="text-[0.82rem] text-[#7a2e2b] flex-1">
              {error instanceof Error ? error.message : "Não foi possível carregar o presente."}
            </span>
          </div>
        ) : gift ? (
          <div className="grid grid-cols-1 md:grid-cols-2 gap-10 md:gap-14">
            <div className="relative aspect-[4/3] bg-parchment-dark border border-gold-muted/40 overflow-hidden">
              <div className="absolute inset-0 flex items-center justify-center">
                <Gift size={72} className="text-gold-muted/40" />
              </div>
              {gift.image_url ? (
                <img
                  src={gift.image_url}
                  alt={gift.name}
                  onError={(e) => {
                    e.currentTarget.style.display = "none";
                  }}
                  className="relative w-full h-full object-cover"
                />
              ) : null}
            </div>

            <div className="flex flex-col gap-5">
              <h1 className="font-display text-[1.6rem] md:text-[1.9rem] font-bold text-dark leading-tight">
                {gift.name}
              </h1>
              <span className="font-display text-[1.6rem] font-bold text-burgundy">
                {formatBRL(gift.price_cents)}
              </span>

              {gift.description && (
                <p className="text-[0.95rem] text-dark-warm/80 leading-relaxed whitespace-pre-line">
                  {gift.description}
                </p>
              )}

              <div className="flex flex-col gap-3 mt-2">
                <Link
                  to="/lista-presentes/$giftId/comprar"
                  params={{ giftId: String(gift.id) }}
                  className="inline-flex items-center justify-center gap-2 font-heading text-[0.75rem] font-semibold tracking-[0.1em] uppercase py-3 px-6 bg-burgundy text-gold-light border border-burgundy transition-all duration-300 hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] cursor-pointer no-underline rounded-[2px]"
                >
                  Quero presentear
                </Link>

                {gift.store_url && (
                  <a
                    href={gift.store_url}
                    target="_blank"
                    rel="noopener noreferrer"
                    className="inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.7rem] px-5 border border-gold-muted/60 text-dark-warm hover:border-burgundy hover:text-burgundy transition-all duration-200 no-underline"
                  >
                    Ver na loja
                    <ExternalLink size={13} />
                  </a>
                )}
              </div>
            </div>
          </div>
        ) : null}
      </main>
    </div>
  );
}
