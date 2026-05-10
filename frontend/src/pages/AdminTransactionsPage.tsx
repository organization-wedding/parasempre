import { useState } from "react";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import ChevronLeft from "lucide-react/dist/esm/icons/chevron-left";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import CheckCircle2 from "lucide-react/dist/esm/icons/check-circle-2";
import Clock from "lucide-react/dist/esm/icons/clock";
import XCircle from "lucide-react/dist/esm/icons/x-circle";
import RotateCcw from "lucide-react/dist/esm/icons/rotate-ccw";
import X from "lucide-react/dist/esm/icons/x";
import { useAdminTransactionsQuery, useAdminSummaryQuery } from "../lib/transaction-queries";
import { useGiftsQuery } from "../lib/gift-queries";
import { formatBRL } from "../lib/format";
import type { AdminTransaction } from "../types/payment";

const PAGE_SIZE = 20;

const statusUI: Record<string, { icon: React.ComponentType<{ size?: number; className?: string }>; color: string; label: string }> = {
  approved: { icon: CheckCircle2, color: "text-[#3a7a3a]", label: "Pago" },
  pending: { icon: Clock, color: "text-[#a8842c]", label: "Em análise" },
  rejected: { icon: XCircle, color: "text-[#c25550]", label: "Rejeitado" },
  cancelled: { icon: XCircle, color: "text-hint", label: "Cancelado" },
  refunded: { icon: RotateCcw, color: "text-hint", label: "Estornado" },
};

const statusOptions = [
  { value: "", label: "Todos os status" },
  { value: "approved", label: "Pago" },
  { value: "pending", label: "Em análise" },
  { value: "rejected", label: "Rejeitado" },
  { value: "cancelled", label: "Cancelado" },
  { value: "refunded", label: "Estornado" },
];

export function AdminTransactionsPage() {
  const [filterStatus, setFilterStatus] = useState("");
  const [filterGiftId, setFilterGiftId] = useState<number | undefined>();
  const [page, setPage] = useState(1);

  const { data: giftsData } = useGiftsQuery({ limit: 100 });
  const gifts = giftsData?.data ?? [];

  const { data, isLoading, error } = useAdminTransactionsQuery({
    status: filterStatus || undefined,
    gift_id: filterGiftId,
    page,
    limit: PAGE_SIZE,
  });
  const { data: summary } = useAdminSummaryQuery();

  const transactions = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  function clearFilters() {
    setFilterStatus("");
    setFilterGiftId(undefined);
    setPage(1);
  }

  const approvedCount = summary?.by_status.find((b) => b.status === "approved")?.count ?? 0;
  const pendingCount = summary?.by_status.find((b) => b.status === "pending")?.count ?? 0;
  const rejectedCount = summary?.by_status.find((b) => b.status === "rejected")?.count ?? 0;

  return (
    <>
        <div className="flex flex-col gap-4 mb-8 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h1 className="font-display text-[1.5rem] md:text-[1.8rem] font-bold text-dark">
              Pagamentos
            </h1>
            <p className="text-[0.85rem] text-hint mt-1">
              Todas as transações de presentes
            </p>
          </div>
        </div>

        {summary && (
          <div className="grid grid-cols-2 md:grid-cols-4 gap-4 mb-8">
            <SummaryCard
              label="Total recebido"
              value={formatBRL(summary.approved_total_cents)}
              highlight
            />
            <SummaryCard label="Aprovadas" value={String(approvedCount)} />
            <SummaryCard label="Pendentes" value={String(pendingCount)} />
            <SummaryCard label="Rejeitadas" value={String(rejectedCount)} />
          </div>
        )}

        <div className="flex flex-wrap gap-3 mb-6 items-end">
          <div className="flex flex-col gap-1">
            <label className="font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
              Status
            </label>
            <select
              value={filterStatus}
              onChange={(e) => { setFilterStatus(e.target.value); setPage(1); }}
              className="border border-gold-muted/50 bg-ivory text-dark-warm text-[0.82rem] px-3 py-2 outline-none focus:border-burgundy transition-colors"
            >
              {statusOptions.map((o) => (
                <option key={o.value} value={o.value}>{o.label}</option>
              ))}
            </select>
          </div>
          <div className="flex flex-col gap-1">
            <label className="font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
              Presente
            </label>
            <select
              value={filterGiftId ?? ""}
              onChange={(e) => { setFilterGiftId(e.target.value ? Number(e.target.value) : undefined); setPage(1); }}
              className="border border-gold-muted/50 bg-ivory text-dark-warm text-[0.82rem] px-3 py-2 outline-none focus:border-burgundy transition-colors"
            >
              <option value="">Todos os presentes</option>
              {gifts.map((g) => (
                <option key={g.id} value={g.id}>{g.name}</option>
              ))}
            </select>
          </div>
          {(filterStatus || filterGiftId) && (
            <button
              type="button"
              onClick={clearFilters}
              className="inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-2 px-3 border border-gold-muted/50 text-hint hover:border-burgundy hover:text-burgundy transition-all duration-200 cursor-pointer"
            >
              <X size={12} />
              Limpar
            </button>
          )}
        </div>

        {error && (
          <div className="mb-6 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
            <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
            <span className="text-[0.82rem] text-[#7a2e2b] flex-1">
              {error instanceof Error ? error.message : "Não foi possível carregar as transações."}
            </span>
          </div>
        )}

        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="text-[0.85rem]">Carregando transações...</span>
          </div>
        ) : (
          <>
            <div className="hidden md:block overflow-x-auto">
              <table className="w-full min-w-[760px]">
                <thead>
                  <tr className="border-b-2 border-gold-muted/40 bg-dark/[0.04]">
                    {["Data", "Convidado", "Presente", "Método", "Valor", "Status", "MP ID"].map((h) => (
                      <th key={h} className="py-3 px-3 text-left font-heading text-[0.65rem] font-semibold tracking-[0.12em] uppercase text-hint">
                        {h}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody>
                  {transactions.length === 0 ? (
                    <tr>
                      <td colSpan={7} className="py-12 text-center text-[0.85rem] text-hint">
                        Nenhuma transação encontrada.
                      </td>
                    </tr>
                  ) : (
                    transactions.map((tx) => <TransactionRow key={tx.id} tx={tx} />)
                  )}
                </tbody>
              </table>
            </div>

            <div className="md:hidden flex flex-col gap-3">
              {transactions.length === 0 ? (
                <p className="text-center text-[0.85rem] text-hint py-12">
                  Nenhuma transação encontrada.
                </p>
              ) : (
                transactions.map((tx) => <TransactionMobileCard key={tx.id} tx={tx} />)
              )}
            </div>

            {totalPages > 1 && (
              <div className="mt-8 flex items-center justify-center gap-4">
                <button
                  type="button"
                  onClick={() => setPage((p) => p - 1)}
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
    </>
  );
}

function SummaryCard({ label, value, highlight }: { label: string; value: string; highlight?: boolean }) {
  return (
    <div className="bg-ivory border border-gold-muted/40 p-4">
      <p className="font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint mb-1">
        {label}
      </p>
      <p className={`font-display text-[1.4rem] font-bold ${highlight ? "text-burgundy" : "text-dark"}`}>
        {value}
      </p>
    </div>
  );
}

function TransactionRow({ tx }: { tx: AdminTransaction }) {
  const ui = (statusUI[tx.status] ?? statusUI["pending"])!;
  const StatusIcon = ui.icon;
  const date = new Date(tx.created_at).toLocaleDateString("pt-BR");

  return (
    <tr className="border-b border-gold-muted/20 bg-ivory hover:bg-parchment/50 transition-colors">
      <td className="py-3 px-3 text-[0.8rem] text-dark-warm">{date}</td>
      <td className="py-3 px-3 text-[0.8rem] text-dark-warm">
        <span className="font-semibold">{tx.user_uracf}</span>
        {tx.user_phone && <span className="text-hint ml-1.5 text-[0.75rem]">{tx.user_phone}</span>}
      </td>
      <td className="py-3 px-3 text-[0.8rem] text-dark-warm max-w-[160px] truncate">{tx.gift_name}</td>
      <td className="py-3 px-3 text-[0.78rem] text-hint">{tx.payment_method === "pix" ? "PIX" : "Cartão"}</td>
      <td className="py-3 px-3 text-[0.8rem] font-semibold text-dark-warm">{formatBRL(tx.amount_cents)}</td>
      <td className="py-3 px-3">
        <span className={`inline-flex items-center gap-1 ${ui.color}`}>
          <StatusIcon size={13} />
          <span className="font-heading text-[0.65rem] font-semibold tracking-[0.08em] uppercase">{ui.label}</span>
        </span>
      </td>
      <td className="py-3 px-3 text-[0.72rem] text-hint font-mono">
        {tx.mp_payment_id ?? "—"}
      </td>
    </tr>
  );
}

function TransactionMobileCard({ tx }: { tx: AdminTransaction }) {
  const ui = (statusUI[tx.status] ?? statusUI["pending"])!;
  const StatusIcon = ui.icon;
  const date = new Date(tx.created_at).toLocaleDateString("pt-BR");

  return (
    <div className="bg-ivory border border-gold-muted/40 p-4 flex flex-col gap-2">
      <div className="flex items-start justify-between gap-2">
        <div className="min-w-0">
          <p className="font-heading text-[0.78rem] font-semibold text-dark-warm truncate">{tx.gift_name}</p>
          <p className="text-[0.72rem] text-hint">{tx.user_uracf}{tx.user_phone ? ` · ${tx.user_phone}` : ""}</p>
        </div>
        <span className={`flex items-center gap-1 shrink-0 ${ui.color}`}>
          <StatusIcon size={13} />
          <span className="font-heading text-[0.65rem] font-semibold tracking-[0.06em] uppercase">{ui.label}</span>
        </span>
      </div>
      <div className="flex items-center gap-4 flex-wrap">
        <span className="font-display text-[1rem] font-bold text-burgundy">{formatBRL(tx.amount_cents)}</span>
        <span className="text-[0.72rem] text-hint">{date}</span>
        <span className="text-[0.72rem] text-hint">{tx.payment_method === "pix" ? "PIX" : "Cartão"}</span>
      </div>
    </div>
  );
}
