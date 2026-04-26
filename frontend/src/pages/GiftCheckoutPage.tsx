import { useMemo, useState } from "react";
import { Link } from "@tanstack/react-router";
import Gift from "lucide-react/dist/esm/icons/gift";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import CheckCircle2 from "lucide-react/dist/esm/icons/check-circle-2";
import MessageSquare from "lucide-react/dist/esm/icons/message-square";
import Package from "lucide-react/dist/esm/icons/package";
import { Header } from "../components/Header";
import { PaymentBrick } from "../components/PaymentBrick";
import { useGiftQuery } from "../lib/gift-queries";
import { useCreatePurchaseMutation } from "../lib/payment-queries";
import { isNotFoundError } from "../lib/api";
import { formatBRL } from "../lib/format";
import { MERCADO_PAGO_PUBLIC_KEY } from "../config";
import type { PaymentBrickFormData } from "../lib/mercado-pago";
import type { PurchaseResponse } from "../schemas/payment";

interface Props {
  giftId: number;
}

type CheckoutOutcome =
  | { kind: "approved"; resp: PurchaseResponse }
  | { kind: "pending"; resp: PurchaseResponse }
  | { kind: "pix"; resp: PurchaseResponse };

export function GiftCheckoutPage({ giftId }: Props) {
  const validId = Number.isFinite(giftId) && giftId > 0;
  const { data: gift, isLoading, error } = useGiftQuery(giftId, validId);
  const purchaseMutation = useCreatePurchaseMutation(giftId);

  const [outcome, setOutcome] = useState<CheckoutOutcome | null>(null);
  const [submitError, setSubmitError] = useState<string | null>(null);

  const idempotencyKey = useMemo(() => crypto.randomUUID(), []);

  const notFound = !validId || isNotFoundError(error);
  const mpConfigured = MERCADO_PAGO_PUBLIC_KEY.length > 0;

  async function handleSubmit(formData: PaymentBrickFormData) {
    setSubmitError(null);
    try {
      const resp = await purchaseMutation.mutateAsync({
        payment_method_id: formData.payment_method_id,
        token: formData.token,
        issuer_id: formData.issuer_id,
        installments: formData.installments,
        payer: {
          email: formData.payer.email,
          identification: {
            type: "CPF",
            number: formData.payer.identification.number,
          },
        },
        idempotency_key: idempotencyKey,
      });

      if (resp.status === "approved") {
        setOutcome({ kind: "approved", resp });
      } else if (resp.payment_method === "pix") {
        setOutcome({ kind: "pix", resp });
      } else {
        setOutcome({ kind: "pending", resp });
      }
    } catch (err) {
      const message = err instanceof Error ? err.message : "Falha no pagamento. Tente novamente.";
      setSubmitError(message);
      throw err;
    }
  }

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />
      <main className="mx-auto max-w-[760px] px-6 pt-24 pb-16">
        <Link
          to="/lista-presentes/$giftId"
          params={{ giftId: String(giftId) }}
          className="inline-flex items-center gap-1.5 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase text-hint hover:text-burgundy transition-colors no-underline mb-6"
        >
          <ArrowLeft size={14} />
          Voltar ao presente
        </Link>

        {!mpConfigured ? (
          <div className="rounded border border-gold-muted/40 bg-parchment-dark p-6">
            <div className="flex items-start gap-3">
              <AlertTriangle size={18} className="text-burgundy mt-0.5 shrink-0" />
              <div>
                <h2 className="font-heading text-[0.95rem] font-semibold text-dark-warm mb-1">
                  Pagamentos indisponíveis
                </h2>
                <p className="text-[0.85rem] text-hint">
                  O checkout ainda não foi configurado neste ambiente. Entre em contato com os
                  noivos para presentear de outra forma.
                </p>
              </div>
            </div>
          </div>
        ) : isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="text-[0.85rem]">Carregando presente...</span>
          </div>
        ) : notFound || !gift ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <Gift size={48} className="text-gold-muted/40 mb-4" />
            <h2 className="font-heading text-[1.05rem] font-semibold text-dark-warm mb-2">
              Presente não disponível
            </h2>
            <Link
              to="/lista-presentes"
              search={{ page: undefined }}
              className="mt-6 inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 border border-burgundy text-burgundy hover:bg-burgundy hover:text-gold-light transition-all duration-200 cursor-pointer no-underline"
            >
              <ArrowLeft size={14} />
              Voltar à lista
            </Link>
          </div>
        ) : outcome ? (
          <CheckoutResult outcome={outcome} giftName={gift.name} />
        ) : (
          <div className="flex flex-col gap-6">
            <header className="flex items-baseline justify-between border-b border-gold-muted/40 pb-4">
              <div>
                <p className="font-heading text-[0.7rem] font-semibold tracking-[0.1em] uppercase text-hint">
                  Presenteando
                </p>
                <h1 className="font-display text-[1.4rem] font-bold text-dark mt-1">
                  {gift.name}
                </h1>
              </div>
              <span className="font-display text-[1.4rem] font-bold text-burgundy">
                {formatBRL(gift.price_cents)}
              </span>
            </header>

            {submitError && (
              <div className="flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
                <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
                <span className="text-[0.82rem] text-[#7a2e2b] flex-1">{submitError}</span>
              </div>
            )}

            <PaymentBrick amount={gift.price_cents / 100} onSubmit={handleSubmit} />

            <p className="text-[0.72rem] text-hint">
              O pagamento é processado de forma segura pelo Mercado Pago. Nenhum dado
              de cartão passa pelos nossos servidores.
            </p>
          </div>
        )}
      </main>
    </div>
  );
}

function PostPurchaseCTAs({ onMessageClick }: { onMessageClick: () => void }) {
  return (
    <div className="flex flex-col sm:flex-row gap-3 justify-center mt-6">
      <button
        type="button"
        onClick={onMessageClick}
        className="inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 border border-gold-muted/60 text-hint cursor-pointer transition-all duration-200 hover:border-burgundy hover:text-burgundy"
      >
        <MessageSquare size={14} />
        Deixar mensagem para os noivos
      </button>
      <Link
        to="/meus-presentes"
        search={{ page: undefined }}
        className="inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep transition-all duration-200 no-underline"
      >
        <Package size={14} />
        Ver meus presentes
      </Link>
    </div>
  );
}

function CheckoutResult({ outcome, giftName }: { outcome: CheckoutOutcome; giftName: string }) {
  const [messageToast, setMessageToast] = useState(false);

  function showMessageComingSoon() {
    setMessageToast(true);
    setTimeout(() => setMessageToast(false), 3000);
  }

  if (outcome.kind === "approved") {
    return (
      <div className="flex flex-col items-center text-center py-10">
        {messageToast && (
          <div className="fixed top-20 left-1/2 -translate-x-1/2 z-[200] bg-dark text-gold-light px-5 py-3 shadow-xl text-[0.82rem] font-heading tracking-wide">
            Em breve: funcionalidade de mensagens será lançada em breve!
          </div>
        )}
        <CheckCircle2 size={56} className="text-[#3a7a3a] mb-4" />
        <h2 className="font-display text-[1.5rem] font-bold text-dark mb-2">
          Pagamento aprovado!
        </h2>
        <p className="text-[0.95rem] text-dark-warm/80 max-w-[420px]">
          Obrigado por presentear com{" "}
          <strong className="text-burgundy">{giftName}</strong>. Em breve os noivos receberão
          a notificação.
        </p>
        <PostPurchaseCTAs onMessageClick={showMessageComingSoon} />
        <Link
          to="/lista-presentes"
          search={{ page: undefined }}
          className="mt-3 inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase text-hint hover:text-burgundy transition-colors no-underline"
        >
          <ArrowLeft size={13} />
          Voltar à lista
        </Link>
      </div>
    );
  }

  if (outcome.kind === "pix") {
    const pix = outcome.resp.pix;
    return (
      <div className="flex flex-col items-center text-center py-6">
        {messageToast && (
          <div className="fixed top-20 left-1/2 -translate-x-1/2 z-[200] bg-dark text-gold-light px-5 py-3 shadow-xl text-[0.82rem] font-heading tracking-wide">
            Em breve: funcionalidade de mensagens será lançada em breve!
          </div>
        )}
        <h2 className="font-display text-[1.4rem] font-bold text-dark mb-2">
          PIX gerado
        </h2>
        <p className="text-[0.9rem] text-dark-warm/80 max-w-[480px] mb-6">
          Aponte a câmera do seu app de banco no QR Code abaixo, ou copie o código PIX.
          O pagamento expira em 30 minutos.
        </p>
        {pix?.qr_code_base64 && (
          <img
            src={`data:image/png;base64,${pix.qr_code_base64}`}
            alt="QR Code PIX"
            className="w-[220px] h-[220px] border border-gold-muted/40 mb-4"
          />
        )}
        {pix?.qr_code && (
          <div className="w-full max-w-[480px]">
            <p className="font-heading text-[0.7rem] font-semibold tracking-[0.1em] uppercase text-hint mb-2">
              Código copia-e-cola
            </p>
            <textarea
              readOnly
              value={pix.qr_code}
              onClick={(e) => e.currentTarget.select()}
              className="w-full text-[0.78rem] font-mono p-3 border border-gold-muted/40 bg-parchment-dark resize-none"
              rows={4}
            />
          </div>
        )}
        <p className="text-[0.75rem] text-hint mt-4 mb-2">
          Quando o pagamento for confirmado, os noivos receberão a notificação automaticamente.
        </p>
        <PostPurchaseCTAs onMessageClick={showMessageComingSoon} />
      </div>
    );
  }

  return (
    <div className="flex flex-col items-center text-center py-10">
      {messageToast && (
        <div className="fixed top-20 left-1/2 -translate-x-1/2 z-[200] bg-dark text-gold-light px-5 py-3 shadow-xl text-[0.82rem] font-heading tracking-wide">
          Em breve: funcionalidade de mensagens será lançada em breve!
        </div>
      )}
      <h2 className="font-display text-[1.4rem] font-bold text-dark mb-2">
        Estamos processando
      </h2>
      <p className="text-[0.95rem] text-dark-warm/80 max-w-[420px]">
        O pagamento está em análise. Você receberá uma confirmação assim que ele for aprovado.
      </p>
      <PostPurchaseCTAs onMessageClick={showMessageComingSoon} />
      <Link
        to="/lista-presentes"
        search={{ page: undefined }}
        className="mt-3 inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase text-hint hover:text-burgundy transition-colors no-underline"
      >
        <ArrowLeft size={13} />
        Voltar à lista
      </Link>
    </div>
  );
}
