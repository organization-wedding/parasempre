import { useState } from "react";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import ChevronLeft from "lucide-react/dist/esm/icons/chevron-left";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import MessageSquare from "lucide-react/dist/esm/icons/message-square";
import { GiftMessageView } from "../components/GiftMessageView";
import {
  useAdminGiftMessagesQuery,
  useDeleteGiftMessageMutation,
} from "../lib/message-queries";
import type { AdminMessage } from "../schemas/giftMessage";

const PAGE_SIZE = 20;

export function AdminGiftMessagesPage() {
  const [page, setPage] = useState(1);
  const { data, isLoading, error } = useAdminGiftMessagesQuery({ page, limit: PAGE_SIZE });
  const deleteMutation = useDeleteGiftMessageMutation();

  const [deleteTarget, setDeleteTarget] = useState<AdminMessage | null>(null);

  const messages = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  async function handleConfirmDelete() {
    if (!deleteTarget) return;
    try {
      await deleteMutation.mutateAsync(deleteTarget.id);
      setDeleteTarget(null);
    } catch {
      // erro já fica no estado da mutation; mantemos o modal aberto pra o admin ver
    }
  }

  return (
    <div className="mx-auto max-w-[960px]">
        <div className="flex flex-col gap-4 mb-8 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h1 className="font-display text-[1.5rem] md:text-[1.8rem] font-bold text-dark">
              Recados
            </h1>
            <p className="text-[0.85rem] text-hint mt-1">
              Mensagens deixadas por quem comprou um presente. Você pode remover qualquer
              recado inadequado.
            </p>
          </div>
        </div>

        {error && (
          <div className="mb-6 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
            <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
            <span className="text-[0.82rem] text-[#7a2e2b] flex-1">
              {error instanceof Error ? error.message : "Não foi possível carregar os recados."}
            </span>
          </div>
        )}

        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="text-[0.85rem]">Carregando recados...</span>
          </div>
        ) : messages.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <MessageSquare size={48} className="text-gold-muted/40 mb-4" />
            <p className="text-[0.9rem] text-hint">Nenhum recado por enquanto.</p>
          </div>
        ) : (
          <div className="flex flex-col gap-4">
            {messages.map((m) => (
              <article key={m.id} className="bg-ivory border border-gold-muted/40 p-5 flex flex-col gap-3">
                <div className="flex items-start justify-between gap-3">
                  <div className="min-w-0">
                    <p className="font-heading text-[0.78rem] font-semibold text-burgundy truncate">
                      {m.author_name}
                    </p>
                    <p className="text-[0.7rem] text-hint">
                      Presente #{m.gift_id} · transação #{m.gift_transaction_id} ·{" "}
                      {new Date(m.created_at).toLocaleDateString("pt-BR", {
                        day: "2-digit",
                        month: "long",
                        year: "numeric",
                      })}
                    </p>
                  </div>
                  <button
                    type="button"
                    onClick={() => setDeleteTarget(m)}
                    className="inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.45rem] px-3 border border-[#c25550]/40 text-[#c25550] hover:bg-[#c25550] hover:text-ivory transition-all duration-200 cursor-pointer shrink-0"
                  >
                    <Trash2 size={13} />
                    Remover
                  </button>
                </div>
                <GiftMessageView message={m} showAuthor={false} />
              </article>
            ))}

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
          </div>
        )}

      {deleteTarget && (
        <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
          <div
            className="absolute inset-0 bg-dark/50 backdrop-blur-[2px]"
            onClick={() => !deleteMutation.isPending && setDeleteTarget(null)}
          />
          <div className="relative bg-ivory border border-gold-muted/50 shadow-xl max-w-[460px] w-full p-6 anim-fade-in-up">
            <h3 className="font-display text-[1.1rem] font-bold text-dark mb-2">
              Remover recado?
            </h3>
            <p className="text-[0.88rem] text-dark-warm/80 mb-4">
              O recado de <strong className="text-burgundy">{deleteTarget.author_name}</strong>{" "}
              será removido da página do presente. Esta ação não pode ser desfeita.
            </p>
            {deleteMutation.error && (
              <div className="mb-3 flex items-center gap-2 rounded border border-[#c25550]/30 bg-[#fef2f1] px-3 py-2">
                <AlertTriangle size={14} className="text-[#c25550] shrink-0" />
                <span className="text-[0.78rem] text-[#7a2e2b]">
                  {deleteMutation.error instanceof Error
                    ? deleteMutation.error.message
                    : "Falha ao remover."}
                </span>
              </div>
            )}
            <div className="flex justify-end gap-3">
              <button
                type="button"
                onClick={() => setDeleteTarget(null)}
                disabled={deleteMutation.isPending}
                className="font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-2 px-4 text-hint hover:text-burgundy transition-colors cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Cancelar
              </button>
              <button
                type="button"
                onClick={handleConfirmDelete}
                disabled={deleteMutation.isPending}
                className="inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-2 px-4 bg-[#c25550] text-ivory border border-[#c25550] hover:bg-[#a83c37] transition-all duration-200 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <Trash2 size={13} />
                {deleteMutation.isPending ? "Removendo…" : "Remover"}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
