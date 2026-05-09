import { useState } from "react";
import { Link } from "@tanstack/react-router";
import Gift from "lucide-react/dist/esm/icons/gift";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";
import ChevronLeft from "lucide-react/dist/esm/icons/chevron-left";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import ExternalLink from "lucide-react/dist/esm/icons/external-link";
import MessageSquare from "lucide-react/dist/esm/icons/message-square";
import { Header } from "../components/Header";
import { GiftMessageView } from "../components/GiftMessageView";
import { isNotFoundError } from "../lib/api";
import { useGiftQuery } from "../lib/gift-queries";
import { useGiftMessagesQuery } from "../lib/message-queries";
import { formatBRL } from "../lib/format";

interface Props {
  giftId: number;
}

const MESSAGES_PAGE_SIZE = 5;

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
          <>
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

            <GiftMessagesSection giftId={gift.id} />
          </>
        ) : null}
      </main>
    </div>
  );
}

function GiftMessagesSection({ giftId }: { giftId: number }) {
  const [page, setPage] = useState(1);
  const { data, isLoading, error } = useGiftMessagesQuery(giftId, page);

  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / MESSAGES_PAGE_SIZE));
  const messages = data?.data ?? [];

  return (
    <section className="mt-16 border-t border-gold-muted/40 pt-10">
      <header className="mb-6 flex items-center gap-3">
        <MessageSquare size={18} className="text-burgundy" />
        <h2 className="font-display text-[1.2rem] font-bold text-dark">Recados</h2>
        {total > 0 && (
          <span className="font-heading text-[0.7rem] tracking-[0.08em] uppercase text-hint">
            {total}
          </span>
        )}
      </header>

      {isLoading ? (
        <div className="flex items-center gap-2 text-hint text-[0.85rem]">
          <div className="w-4 h-4 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin" />
          Carregando recados…
        </div>
      ) : error ? (
        <div className="flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
          <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
          <span className="text-[0.82rem] text-[#7a2e2b] flex-1">
            {error instanceof Error ? error.message : "Não foi possível carregar os recados."}
          </span>
        </div>
      ) : messages.length === 0 ? (
        <p className="text-[0.88rem] text-hint">
          Ainda não há recados para este presente. Seja o primeiro a presentear e deixar uma mensagem!
        </p>
      ) : (
        <>
          <div className="flex flex-col gap-6">
            {messages.map((m) => (
              <article key={m.id} className="bg-ivory border border-gold-muted/30 p-5">
                <GiftMessageView message={m} />
              </article>
            ))}
          </div>

          {totalPages > 1 && (
            <div className="mt-8 flex items-center justify-center gap-4">
              <button
                type="button"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={page <= 1}
                className="inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.06em] uppercase py-2 px-4 border border-gold-muted/50 bg-ivory text-dark-warm hover:border-burgundy hover:text-burgundy transition-all duration-200 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
              >
                <ChevronLeft size={14} />
                Anterior
              </button>
              <span className="font-heading text-[0.72rem] tracking-[0.08em] uppercase text-dark-warm/70">
                Página {page} de {totalPages}
              </span>
              <button
                type="button"
                onClick={() => setPage((p) => p + 1)}
                disabled={page >= totalPages}
                className="inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.06em] uppercase py-2 px-4 border border-gold-muted/50 bg-ivory text-dark-warm hover:border-burgundy hover:text-burgundy transition-all duration-200 cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Próxima
                <ChevronRight size={14} />
              </button>
            </div>
          )}
        </>
      )}
    </section>
  );
}
