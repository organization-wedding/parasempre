import { useState } from "react";
import { Link } from "@tanstack/react-router";
import Gift from "lucide-react/dist/esm/icons/gift";
import Plus from "lucide-react/dist/esm/icons/plus";
import Upload from "lucide-react/dist/esm/icons/upload";
import Pencil from "lucide-react/dist/esm/icons/pencil";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import X from "lucide-react/dist/esm/icons/x";
import ChevronLeft from "lucide-react/dist/esm/icons/chevron-left";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import { Header } from "../components/Header";
import { DashboardTabs } from "../components/DashboardTabs";
import {
  useDeleteGiftMutation,
  useGiftsQuery,
} from "../lib/gift-queries";
import { useUserMeQuery } from "../lib/user-queries";
import { UnauthorizedPage } from "./UnauthorizedPage";
import { formatBRL } from "../lib/format";
import type { PublicGift } from "../types/gift";

const PAGE_SIZE = 20;

export function GiftAdminPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading, error } = useGiftsQuery({ page, limit: PAGE_SIZE });
  const gifts = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  const deleteMutation = useDeleteGiftMutation();
  const [deleteTarget, setDeleteTarget] = useState<PublicGift | null>(null);
  const [uiError, setUiError] = useState<string | null>(null);

  const { data: userMe, isLoading: roleLoading } = useUserMeQuery();
  const isAuthorized = userMe?.role === "groom" || userMe?.role === "bride";

  const effectiveError = uiError ?? (error instanceof Error ? error.message : null);

  async function handleDelete() {
    if (!deleteTarget) return;
    try {
      await deleteMutation.mutateAsync(deleteTarget.id);
      setDeleteTarget(null);
    } catch (mutationError) {
      setUiError(mutationError instanceof Error ? mutationError.message : "Erro ao excluir presente");
      setDeleteTarget(null);
    }
  }

  if (!roleLoading && userMe && !isAuthorized) {
    return <UnauthorizedPage />;
  }

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />

      <main className="mx-auto max-w-[1280px] px-6 pt-24 pb-16">
        {roleLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="text-[0.85rem]">Verificando permissões...</span>
          </div>
        ) : (
          <>
            <DashboardTabs active="presentes" />
            <div className="flex flex-col gap-4 mb-8 sm:flex-row sm:items-start sm:justify-between">
              <div>
                <h1 className="font-display text-[1.5rem] md:text-[1.8rem] font-bold text-dark">
                  Gerenciar Presentes
                </h1>
                <span className="inline-flex items-center gap-1.5 text-[0.82rem] text-dark-warm/70 mt-2.5">
                  <Gift size={14} />
                  {total} {total === 1 ? "presente" : "presentes"}
                </span>
              </div>

              <div className="flex gap-2.5 shrink-0">
                <Link
                  to="/dashboard/presentes/importar"
                  className="inline-flex items-center gap-[0.4rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.1rem] cursor-pointer no-underline transition-all duration-300 whitespace-nowrap bg-transparent text-burgundy border border-burgundy hover:bg-burgundy hover:text-gold-light"
                >
                  <Upload size={14} />
                  Importar CSV
                </Link>
                <Link
                  to="/dashboard/presentes/novo"
                  className="inline-flex items-center gap-[0.4rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.1rem] cursor-pointer no-underline transition-all duration-300 whitespace-nowrap bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] hover:-translate-y-px"
                >
                  <Plus size={14} />
                  Novo Presente
                </Link>
              </div>
            </div>

            {effectiveError && (
              <div className="mb-6 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
                <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
                <span className="text-[0.82rem] text-[#7a2e2b] flex-1">{effectiveError}</span>
                <button
                  type="button"
                  onClick={() => setUiError(null)}
                  className="text-[#c25550]/60 hover:text-[#c25550] cursor-pointer"
                >
                  <X size={14} />
                </button>
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
            ) : total === 0 ? (
              <div className="flex flex-col items-center justify-center py-20 text-center">
                <Gift size={48} className="text-gold-muted/40 mb-4" />
                <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">
                  Nenhum presente cadastrado
                </h2>
                <p className="text-[0.88rem] text-hint max-w-[360px] mb-6">
                  Comece adicionando o primeiro presente da sua lista.
                </p>
                <Link
                  to="/dashboard/presentes/novo"
                  className="inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-burgundy text-gold-light border border-burgundy transition-all duration-300 hover:bg-burgundy-deep no-underline"
                >
                  <Plus size={14} />
                  Adicionar Primeiro Presente
                </Link>
              </div>
            ) : (
              <>
                <div className="overflow-x-auto rounded border border-gold-muted/50 shadow-[0_2px_8px_rgba(28,20,16,0.06)]">
                  <table className="w-full min-w-[600px]">
                    <thead>
                      <tr className="border-b-2 border-gold-muted/40 bg-dark/[0.04]">
                        <th className="py-3 px-3 w-20"></th>
                        <th className="py-3 px-3 text-left font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-hint">
                          Nome
                        </th>
                        <th className="py-3 px-3 text-left font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-hint">
                          Preço
                        </th>
                        <th className="py-3 px-3 w-24"></th>
                      </tr>
                    </thead>
                    <tbody>
                      {gifts.map((gift) => (
                        <tr
                          key={gift.id}
                          className="border-b border-gold-muted/20 last:border-b-0 bg-ivory hover:bg-parchment/50 transition-colors"
                        >
                          <td className="py-2 px-3">
                            <div className="relative w-14 h-14 bg-parchment-dark overflow-hidden rounded">
                              <div className="absolute inset-0 flex items-center justify-center">
                                <Gift size={20} className="text-gold-muted/40" />
                              </div>
                              {gift.image_url ? (
                                <img
                                  src={gift.image_url}
                                  alt={gift.name}
                                  loading="lazy"
                                  onError={(e) => {
                                    e.currentTarget.style.display = "none";
                                  }}
                                  className="relative w-full h-full object-cover"
                                />
                              ) : null}
                            </div>
                          </td>
                          <td className="py-3 px-3">
                            <span className="text-[0.9rem] text-dark-warm font-medium">
                              {gift.name}
                            </span>
                          </td>
                          <td className="py-3 px-3">
                            <span className="font-display text-[0.95rem] font-bold text-burgundy">
                              {formatBRL(gift.price_cents)}
                            </span>
                          </td>
                          <td className="py-3 px-3">
                            <div className="flex gap-1 justify-end">
                              <Link
                                to="/dashboard/presentes/$giftId"
                                params={{ giftId: String(gift.id) }}
                                className="p-1.5 text-dark-warm/40 hover:text-burgundy transition-colors"
                                title="Editar"
                              >
                                <Pencil size={15} />
                              </Link>
                              <button
                                type="button"
                                onClick={() => setDeleteTarget(gift)}
                                className="p-1.5 text-dark-warm/40 hover:text-[#c25550] transition-colors cursor-pointer"
                                title="Excluir"
                              >
                                <Trash2 size={15} />
                              </button>
                            </div>
                          </td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>

                {totalPages > 1 && (
                  <div className="mt-6 flex items-center justify-center gap-4">
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
                      onClick={() => setPage((p) => Math.min(totalPages, p + 1))}
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
          </>
        )}
      </main>

      {deleteTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div
            className="absolute inset-0 bg-dark/50 backdrop-blur-[2px]"
            onClick={() => !deleteMutation.isPending && setDeleteTarget(null)}
          />
          <div className="relative bg-ivory border border-gold-muted/50 shadow-xl max-w-[420px] w-full p-6 anim-fade-in-up">
            <h3 className="font-display text-[1.15rem] font-bold text-dark mb-3">
              Excluir presente?
            </h3>
            <p className="text-[0.88rem] text-dark-warm/70 mb-6 leading-relaxed">
              O presente <strong className="text-dark-warm">{deleteTarget.name}</strong> será removido da lista. Esta ação não pode ser desfeita.
            </p>
            <div className="flex gap-3 w-full">
              <button
                type="button"
                onClick={() => setDeleteTarget(null)}
                disabled={deleteMutation.isPending}
                className="flex-1 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-4 border border-gold-muted/60 text-hint bg-transparent hover:border-burgundy hover:text-burgundy cursor-pointer disabled:opacity-50 transition-colors"
              >
                Cancelar
              </button>
              <button
                type="button"
                onClick={() => void handleDelete()}
                disabled={deleteMutation.isPending}
                className="flex-1 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-4 border border-[#c25550] bg-[#c25550] text-white hover:bg-[#a83f3b] cursor-pointer disabled:opacity-50 transition-colors"
              >
                {deleteMutation.isPending ? "Excluindo..." : "Excluir"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
