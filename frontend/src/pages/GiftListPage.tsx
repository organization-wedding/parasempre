import { useState } from "react";
import { Link, useNavigate } from "@tanstack/react-router";
import Gift from "lucide-react/dist/esm/icons/gift";
import Package from "lucide-react/dist/esm/icons/package";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import ChevronLeft from "lucide-react/dist/esm/icons/chevron-left";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import { Header } from "../components/Header";
import { GiftFilters } from "../components/GiftFilters";
import { useGiftsQuery, type GiftSort } from "../lib/gift-queries";
import { useDebounce } from "../lib/useDebounce";
import { formatBRL, reaisToCents } from "../lib/format";
import type { PublicGift } from "../types/gift";

const PAGE_SIZE = 20;

interface Props {
  page: number;
}

export function GiftListPage({ page }: Props) {
  const navigate = useNavigate();

  const [search, setSearch] = useState("");
  const [priceMin, setPriceMin] = useState("");
  const [priceMax, setPriceMax] = useState("");
  const [sort, setSort] = useState<"" | GiftSort>("");

  const debouncedSearch = useDebounce(search, 300);
  const debouncedMin = useDebounce(priceMin, 300);
  const debouncedMax = useDebounce(priceMax, 300);

  const { data, isLoading, error } = useGiftsQuery({
    page,
    limit: PAGE_SIZE,
    search: debouncedSearch.trim() || undefined,
    price_min: reaisToCents(debouncedMin),
    price_max: reaisToCents(debouncedMax),
    sort: sort || undefined,
  });

  const hasActiveFilter =
    debouncedSearch.trim() !== "" ||
    reaisToCents(debouncedMin) !== undefined ||
    reaisToCents(debouncedMax) !== undefined ||
    sort !== "";

  const gifts = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const goToPage = (n: number) => {
    void navigate({
      to: "/lista-presentes",
      search: { page: n > 1 ? n : undefined },
    });
  };

  // Ao alterar qualquer filtro, volta para a primeira página.
  const resetToFirstPage = () => {
    if (page > 1) goToPage(1);
  };

  const clearFilters = () => {
    setSearch("");
    setPriceMin("");
    setPriceMax("");
    setSort("");
    resetToFirstPage();
  };

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />

      <main className="mx-auto max-w-[1280px] px-6 pt-24 pb-16">
        <div className="relative mb-10">
          <header className="text-center">
            <div className="flex items-center justify-center gap-3 mb-4">
              <div className="h-px w-12 bg-gold/40" />
              <span className="font-heading text-[0.68rem] font-semibold tracking-[0.3em] uppercase text-gold-dark">
                Lista de Presentes
              </span>
              <div className="h-px w-12 bg-gold/40" />
            </div>
            <h1 className="font-display text-[1.6rem] md:text-[2rem] font-bold text-dark mb-3">
              Nossa Lista de Presentes
            </h1>
            <p className="text-[0.95rem] text-dark-warm/70 max-w-[560px] mx-auto leading-relaxed">
              Escolha um presente para celebrar conosco este novo capítulo. Sua
              presença e carinho são o maior presente que poderíamos receber.
            </p>
          </header>
          <div className="mt-4 flex justify-end md:absolute md:top-0 md:right-0 md:mt-0">
            <Link
              to="/meus-presentes"
              search={{ page: undefined }}
              className="inline-flex items-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.5rem] px-4 border border-burgundy text-burgundy hover:bg-burgundy hover:text-gold-light transition-all duration-200 no-underline"
            >
              <Package size={14} />
              Presentes que comprei
            </Link>
          </div>
        </div>

        {error && (
          <div className="mb-6 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
            <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
            <span className="text-[0.82rem] text-[#7a2e2b] flex-1">
              {error instanceof Error ? error.message : "Não foi possível carregar a lista."}
            </span>
          </div>
        )}

        {!isLoading && (total > 0 || hasActiveFilter) && (
          <div className="mb-8">
            <GiftFilters
              search={search}
              priceMin={priceMin}
              priceMax={priceMax}
              sort={sort}
              onSearchChange={(v) => { setSearch(v); resetToFirstPage(); }}
              onPriceMinChange={(v) => { setPriceMin(v); resetToFirstPage(); }}
              onPriceMaxChange={(v) => { setPriceMax(v); resetToFirstPage(); }}
              onSortChange={(v) => { setSort(v); resetToFirstPage(); }}
              onClear={clearFilters}
            />
          </div>
        )}

        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div
              className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4"
              role="status"
              aria-label="Carregando"
            />
            <span className="text-[0.85rem]">Carregando presentes...</span>
          </div>
        ) : total === 0 && hasActiveFilter ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <Gift size={48} className="text-gold-muted/40 mb-4" />
            <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">
              Nenhum presente encontrado
            </h2>
            <p className="text-[0.88rem] text-hint max-w-[360px] mb-6">
              Tente ajustar a busca ou a faixa de preço.
            </p>
            <button
              type="button"
              onClick={clearFilters}
              className="inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 border border-burgundy text-burgundy hover:bg-burgundy hover:text-gold-light transition-all duration-200 cursor-pointer"
            >
              Limpar filtros
            </button>
          </div>
        ) : total === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <Gift size={48} className="text-gold-muted/40 mb-4" />
            <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">
              Lista em preparação
            </h2>
            <p className="text-[0.88rem] text-hint max-w-[360px]">
              Em breve os presentes estarão disponíveis aqui. Volte mais tarde.
            </p>
          </div>
        ) : gifts.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <Gift size={48} className="text-gold-muted/40 mb-4" />
            <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">
              Página sem resultados
            </h2>
            <p className="text-[0.88rem] text-hint max-w-[360px] mb-6">
              Não há presentes nesta página. Volte para o início da lista.
            </p>
            <button
              type="button"
              onClick={() => goToPage(1)}
              className="inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 border border-burgundy text-burgundy hover:bg-burgundy hover:text-gold-light transition-all duration-200 cursor-pointer"
            >
              Voltar ao início
            </button>
          </div>
        ) : (
          <>
            <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 gap-6">
              {gifts.map((gift) => (
                <GiftCard key={gift.id} gift={gift} />
              ))}
            </div>

            {totalPages > 1 && (
              <div className="mt-10 flex items-center justify-center gap-4">
                <button
                  type="button"
                  onClick={() => goToPage(page - 1)}
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
                  onClick={() => goToPage(page + 1)}
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
      </main>
    </div>
  );
}

function GiftCard({ gift }: { gift: PublicGift }) {
  return (
    <Link
      to="/lista-presentes/$giftId"
      params={{ giftId: String(gift.id) }}
      className="group flex flex-col bg-ivory border border-gold-muted/40 overflow-hidden no-underline transition-all duration-300 hover:border-burgundy hover:shadow-[0_6px_20px_rgba(28,20,16,0.08)] hover:-translate-y-0.5"
    >
      <div className="relative aspect-[4/3] bg-parchment-dark overflow-hidden">
        <div className="absolute inset-0 flex items-center justify-center">
          <Gift size={56} className="text-gold-muted/40" />
        </div>
        {gift.image_url ? (
          <img
            src={gift.image_url}
            alt={gift.name}
            loading="lazy"
            onError={(e) => {
              e.currentTarget.style.display = "none";
            }}
            className="relative w-full h-full object-cover transition-transform duration-500 group-hover:scale-[1.03]"
          />
        ) : null}
      </div>
      <div className="flex flex-col gap-2 p-5">
        <h3 className="font-heading text-[0.95rem] font-semibold text-dark-warm leading-snug line-clamp-2">
          {gift.name}
        </h3>
        <span className="font-display text-[1.1rem] font-bold text-burgundy">
          {formatBRL(gift.price_cents)}
        </span>
      </div>
    </Link>
  );
}
