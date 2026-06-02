import { Link } from "@tanstack/react-router";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";
import Crown from "lucide-react/dist/esm/icons/crown";
import { CoatOfArms } from "../CoatOfArms";

export function AdminHostView() {
  return (
    <div className="relative w-full max-w-[480px] mx-auto book-parchment border-2 border-gold-muted/60 shadow-[0_8px_40px_rgba(28,20,16,0.25)] p-10 text-center">
      <div className="absolute inset-3 border border-gold/30 pointer-events-none" />
      <CoatOfArms className="w-16 h-auto mx-auto mb-5 opacity-80" />

      <div className="flex items-center justify-center gap-3 mb-4">
        <Crown size={18} className="text-gold-dark" />
        <span className="font-heading text-[0.65rem] tracking-[0.3em] uppercase text-gold-dark">
          Anfitrião
        </span>
        <Crown size={18} className="text-gold-dark" />
      </div>

      <h1 className="font-display text-[1.4rem] md:text-[1.8rem] font-bold text-dark-warm mb-3">
        Você é o anfitrião
      </h1>

      <p className="font-body text-[1.05rem] text-dark-warm/75 leading-relaxed mb-8">
        Nenhuma confirmação de presença necessária para os noivos.
        O livro dos convidados secretos é destinado aos visitantes.
      </p>

      <Link
        to="/"
        className="inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.1em] uppercase text-burgundy no-underline py-[0.65rem] px-6 border border-burgundy transition-all duration-300 hover:bg-burgundy hover:text-gold-light"
      >
        <ArrowLeft size={14} />
        Voltar ao início
      </Link>
    </div>
  );
}
