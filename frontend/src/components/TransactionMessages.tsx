import { useState } from "react";
import MessageSquare from "lucide-react/dist/esm/icons/message-square";
import { useMyTransactionMessageQuery } from "../lib/message-queries";
import { GiftMessageForm } from "./GiftMessageForm";
import { GiftMessageView } from "./GiftMessageView";

interface Props {
  transactionId: number;
  status: string;
  defaultAuthorName?: string;
}

export function TransactionMessages({ transactionId, status, defaultAuthorName }: Props) {
  const enabled = status === "approved";
  const { data: msg, isLoading } = useMyTransactionMessageQuery(transactionId, enabled);
  const [editing, setEditing] = useState(false);

  if (!enabled) return null;
  if (isLoading) {
    return (
      <div className="flex items-center gap-2 text-hint text-[0.8rem]">
        <div className="w-4 h-4 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin" />
        Carregando recado…
      </div>
    );
  }

  if (msg) {
    return (
      <div className="border-t border-gold-muted/30 pt-4">
        <p className="font-heading text-[0.65rem] font-semibold tracking-[0.12em] uppercase text-hint mb-3">
          Seu recado
        </p>
        <GiftMessageView message={msg} showAuthor={false} />
        <p className="mt-2 text-[0.72rem] text-hint italic">
          Enviado em {new Date(msg.created_at).toLocaleDateString("pt-BR")}.
        </p>
      </div>
    );
  }

  if (editing) {
    return (
      <div className="border-t border-gold-muted/30 pt-4">
        <GiftMessageForm
          transactionId={transactionId}
          defaultAuthorName={defaultAuthorName}
          onCreated={() => setEditing(false)}
          onCancel={() => setEditing(false)}
        />
      </div>
    );
  }

  return (
    <div className="border-t border-gold-muted/30 pt-3">
      <button
        type="button"
        onClick={() => setEditing(true)}
        className="inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.45rem] px-3 border border-gold-muted/60 text-hint hover:border-burgundy hover:text-burgundy transition-all duration-200 cursor-pointer"
      >
        <MessageSquare size={13} />
        Deixar recado para os noivos
      </button>
    </div>
  );
}
