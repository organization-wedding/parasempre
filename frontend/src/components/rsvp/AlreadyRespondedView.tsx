import Sparkles from "lucide-react/dist/esm/icons/sparkles";

export function AlreadyRespondedView() {
  return (
    <div className="mb-4 flex items-center gap-3 border border-gold/40 bg-gold/10 px-4 py-2.5">
      <Sparkles size={14} className="text-gold-dark shrink-0" />
      <p className="text-[0.78rem] text-dark-warm/80 italic font-body leading-snug">
        Vossa resposta já foi registrada nas crônicas. Podeis alterá-la abaixo se desejardes.
      </p>
    </div>
  );
}
