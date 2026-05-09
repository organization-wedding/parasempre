import type { PublicMessage } from "../schemas/giftMessage";

interface Props {
  message: PublicMessage;
  showAuthor?: boolean;
  className?: string;
}

export function GiftMessageView({ message, showAuthor = true, className = "" }: Props) {
  const date = new Date(message.created_at).toLocaleDateString("pt-BR", {
    day: "2-digit",
    month: "long",
    year: "numeric",
  });

  return (
    <div className={`flex flex-col gap-3 ${className}`}>
      {showAuthor && (
        <div className="flex items-baseline justify-between gap-3">
          <p className="font-heading text-[0.78rem] font-semibold text-burgundy">
            {message.author_name}
          </p>
          <span className="text-[0.7rem] text-hint">{date}</span>
        </div>
      )}
      <p className="text-[0.9rem] text-dark-warm/85 leading-relaxed whitespace-pre-line">
        {message.content}
      </p>
      {message.media_url && message.media_kind === "image" && (
        <img
          src={message.media_url}
          alt="anexo da mensagem"
          loading="lazy"
          className="max-h-[420px] w-auto object-contain border border-gold-muted/40 bg-parchment-dark"
        />
      )}
      {message.media_url && message.media_kind === "audio" && (
        <audio controls src={message.media_url} className="w-full" />
      )}
      {message.media_url && message.media_kind === "video" && (
        <video
          controls
          src={message.media_url}
          className="max-h-[420px] w-full bg-dark/5 border border-gold-muted/40"
        />
      )}
    </div>
  );
}
