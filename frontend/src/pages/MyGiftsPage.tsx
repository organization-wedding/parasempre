import { useState } from "react";
import { Link } from "@tanstack/react-router";
import Gift from "lucide-react/dist/esm/icons/gift";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import ChevronLeft from "lucide-react/dist/esm/icons/chevron-left";
import ChevronRight from "lucide-react/dist/esm/icons/chevron-right";
import CheckCircle2 from "lucide-react/dist/esm/icons/check-circle-2";
import Clock from "lucide-react/dist/esm/icons/clock";
import XCircle from "lucide-react/dist/esm/icons/x-circle";
import RotateCcw from "lucide-react/dist/esm/icons/rotate-ccw";
import MessageSquare from "lucide-react/dist/esm/icons/message-square";
import { Header } from "../components/Header";
import { TransactionMessages } from "../components/TransactionMessages";
import { useMyPurchasesQuery, useMyPurchaseQuery } from "../lib/transaction-queries";
import { formatBRL } from "../lib/format";
import type { PublicTransaction } from "../types/payment";

const PAGE_SIZE = 10;

const statusUI: Record<
  string,
  { icon: React.ComponentType<{ size?: number; className?: string }>; color: string; label: string }
> = {
  approved: { icon: CheckCircle2, color: "text-[#3a7a3a]", label: "Pago" },
  pending: { icon: Clock, color: "text-[#a8842c]", label: "Em análise" },
  rejected: { icon: XCircle, color: "text-[#c25550]", label: "Rejeitado" },
  cancelled: { icon: XCircle, color: "text-hint", label: "Cancelado" },
  refunded: { icon: RotateCcw, color: "text-hint", label: "Estornado" },
};

interface Props {
  page: number;
}

export function MyGiftsPage({ page }: Props) {
  const { data, isLoading, error } = useMyPurchasesQuery({ page, limit: PAGE_SIZE });
  const [messageToast, setMessageToast] = useState(false);

  const transactions = data?.data ?? [];
  const total = data?.total ?? 0;
  const totalPages = Math.max(1, Math.ceil(total / PAGE_SIZE));

  function showMessageComingSoon() {
    setMessageToast(true);
    setTimeout(() => setMessageToast(false), 3000);
  }

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />
      <main className="mx-auto max-w-[860px] px-6 pt-24 pb-16">
        <header className="mb-10 text-center">
          <div className="flex items-center justify-center gap-3 mb-4">
            <div className="h-px w-12 bg-gold/40" />
            <span className="font-heading text-[0.68rem] font-semibold tracking-[0.3em] uppercase text-gold-dark">
              Meus Presentes
            </span>
            <div className="h-px w-12 bg-gold/40" />
          </div>
          <h1 className="font-display text-[1.6rem] md:text-[2rem] font-bold text-dark mb-3">
            Presentes que comprei
          </h1>
          <p className="text-[0.95rem] text-dark-warm/70 max-w-[480px] mx-auto leading-relaxed">
            Acompanhe o status dos seus presentes e veja as confirmações de pagamento.
          </p>
        </header>

        {messageToast && (
          <div className="fixed top-20 left-1/2 -translate-x-1/2 z-[200] bg-dark text-gold-light px-5 py-3 shadow-xl text-[0.82rem] font-heading tracking-wide">
            Em breve: funcionalidade de recados será lançada em breve!
          </div>
        )}

        {error && (
          <div className="mb-6 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
            <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
            <span className="text-[0.82rem] text-[#7a2e2b] flex-1">
              {error instanceof Error ? error.message : "Não foi possível carregar seus presentes."}
            </span>
          </div>
        )}

        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="text-[0.85rem]">Carregando presentes...</span>
          </div>
        ) : total === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <Gift size={48} className="text-gold-muted/40 mb-4" />
            <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">
              Você ainda não presenteou ninguém
            </h2>
            <p className="text-[0.88rem] text-hint max-w-[360px] mb-6">
              Escolha um presente da lista e surpreenda os noivos!
            </p>
            <Link
              to="/lista-presentes"
              search={{ page: undefined }}
              className="inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep transition-all duration-200 no-underline"
            >
              Ver lista de presentes
            </Link>
          </div>
        ) : (
          <>
            <div className="flex flex-col gap-4">
              {transactions.map((tx) => (
                <TransactionCard
                  key={tx.id}
                  tx={tx}
                  onMessageClick={showMessageComingSoon}
                />
              ))}
            </div>

            {totalPages > 1 && (
              <div className="mt-10 flex items-center justify-center gap-4">
                <Link
                  to="/meus-presentes"
                  search={{ page: page > 1 ? page - 1 : undefined }}
                  className={`inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.06em] uppercase py-2 px-4 border border-gold-muted/50 bg-ivory text-dark-warm hover:border-burgundy hover:text-burgundy transition-all duration-200 no-underline ${page <= 1 ? "opacity-40 pointer-events-none" : ""}`}
                >
                  <ChevronLeft size={14} />
                  Anterior
                </Link>
                <span className="font-heading text-[0.72rem] tracking-[0.08em] uppercase text-dark-warm/70">
                  Página {page} de {totalPages}
                </span>
                <Link
                  to="/meus-presentes"
                  search={{ page: page + 1 }}
                  className={`inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.06em] uppercase py-2 px-4 border border-gold-muted/50 bg-ivory text-dark-warm hover:border-burgundy hover:text-burgundy transition-all duration-200 no-underline ${page >= totalPages ? "opacity-40 pointer-events-none" : ""}`}
                >
                  Próxima
                  <ChevronRight size={14} />
                </Link>
              </div>
            )}
          </>
        )}
      </main>
    </div>
  );
}

function TransactionCard({
  tx,
  onMessageClick,
}: {
  tx: PublicTransaction;
  onMessageClick: () => void;
}) {
  const [showQR, setShowQR] = useState(false);
  const { data: detail, isLoading: qrLoading } = useMyPurchaseQuery(tx.id, showQR);

  const ui = (statusUI[tx.status] ?? statusUI["pending"])!;
  const StatusIcon = ui.icon;

  const createdAt = new Date(tx.created_at).toLocaleDateString("pt-BR", {
    day: "2-digit",
    month: "long",
    year: "numeric",
  });

  return (
    <div className="bg-ivory border border-gold-muted/40 p-5 flex flex-col gap-4">
      <div className="flex items-start justify-between gap-4">
        <div className="flex flex-col gap-1 min-w-0">
          <h3 className="font-heading text-[0.95rem] font-semibold text-dark-warm leading-snug truncate">
            {tx.gift_name}
          </h3>
          <div className="flex items-center gap-3 flex-wrap">
            <span className="font-display text-[1.05rem] font-bold text-burgundy">
              {formatBRL(tx.amount_cents)}
            </span>
            <span className="text-[0.75rem] text-hint">{createdAt}</span>
            <span className="text-[0.75rem] text-hint capitalize">
              {tx.payment_method === "pix" ? "PIX" : "Cartão de crédito"}
            </span>
          </div>
        </div>
        <div className={`flex items-center gap-1.5 shrink-0 ${ui.color}`}>
          <StatusIcon size={16} />
          <span className="font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase">
            {ui.label}
          </span>
        </div>
      </div>

      <div className="flex flex-wrap gap-2">
        {tx.status === "approved" && (
          <span className="text-[0.78rem] text-[#3a7a3a]">
            Pago em {createdAt}
          </span>
        )}

        {tx.status === "pending" && tx.payment_method === "pix" && (
          <button
            type="button"
            onClick={() => setShowQR((v) => !v)}
            className="inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.45rem] px-3 border border-burgundy text-burgundy hover:bg-burgundy hover:text-gold-light transition-all duration-200 cursor-pointer"
          >
            {showQR ? "Ocultar QR Code" : "Ver QR Code"}
          </button>
        )}

        {tx.status === "pending" && tx.payment_method === "credit_card" && (
          <span className="text-[0.78rem] text-[#a8842c]">
            Pagamento em análise. Você receberá uma confirmação em breve.
          </span>
        )}

        {tx.status === "rejected" && (
          <Link
            to="/lista-presentes/$giftId/comprar"
            params={{ giftId: String(tx.gift_id) }}
            className="inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.45rem] px-3 bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep transition-all duration-200 no-underline"
          >
            Tentar novamente
          </Link>
        )}

        <button
          type="button"
          onClick={onMessageClick}
          className="inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.45rem] px-3 border border-gold-muted/60 text-hint hover:border-burgundy hover:text-burgundy transition-all duration-200 cursor-pointer"
        >
          <MessageSquare size={13} />
          Deixar recado
        </button>
      </div>

      {showQR && (
        <div className="border-t border-gold-muted/30 pt-4">
          {qrLoading ? (
            <div className="flex items-center gap-2 text-hint text-[0.82rem]">
              <div className="w-4 h-4 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin" />
              Carregando QR Code...
            </div>
          ) : detail?.pix ? (
            <div className="flex flex-col items-center gap-4">
              {detail.pix.qr_code_base64 && (
                <img
                  src={`data:image/png;base64,${detail.pix.qr_code_base64}`}
                  alt="QR Code PIX"
                  className="w-[180px] h-[180px] border border-gold-muted/40"
                />
              )}
              {detail.pix.qr_code && (
                <div className="w-full">
                  <p className="font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-hint mb-1.5">
                    Código copia-e-cola
                  </p>
                  <textarea
                    readOnly
                    value={detail.pix.qr_code}
                    onClick={(e) => e.currentTarget.select()}
                    className="w-full text-[0.75rem] font-mono p-3 border border-gold-muted/40 bg-parchment-dark resize-none"
                    rows={3}
                  />
                </div>
              )}
            </div>
          ) : (
            <p className="text-[0.82rem] text-hint">
              QR Code indisponível. O pagamento pode ter expirado.
            </p>
          )}
        </div>
      )}

      <TransactionMessages transactionId={tx.id} />
    </div>
  );
}
