import { motion } from "framer-motion";
import BookOpen from "lucide-react/dist/esm/icons/book-open";
import { CoatOfArms } from "../CoatOfArms";
import { unlockAudio } from "../../lib/page-flip-sound";

interface Props {
  open: boolean;
  onOpen?: () => void;
  onOpenComplete?: () => void;
  onClosed?: () => void;
}

export function BookCover({ open, onOpen, onOpenComplete, onClosed }: Props) {
  const handleOpenClick = () => {
    unlockAudio();
    onOpen?.();
  };
  return (
    <motion.div
      aria-hidden="true"
      className="book-face absolute inset-0 z-30 book-leather border-2 border-gold-dark/70 shadow-[0_10px_40px_rgba(0,0,0,0.45)]"
      style={{ transformOrigin: "left center" }}
      initial={false}
      animate={{ rotateY: open ? -160 : 0 }}
      transition={{ duration: 1.4, ease: [0.65, 0, 0.35, 1] }}
      onAnimationComplete={() => {
        if (open) onOpenComplete?.();
        else onClosed?.();
      }}
    >
      <div className="absolute inset-3 border border-gold/40 pointer-events-none" />
      <div className="absolute inset-5 border border-gold/25 pointer-events-none" />

      <div className="relative w-full h-full flex flex-col items-center justify-between p-8 md:p-12 text-center">
        <div className="flex flex-col items-center gap-3">
          <div className="flex items-center gap-2">
            <div className="h-px w-10 bg-gold/40" />
            <span className="font-heading text-[0.55rem] tracking-[0.4em] uppercase text-gold/70">
              Anno Domini
            </span>
            <div className="h-px w-10 bg-gold/40" />
          </div>
          <CoatOfArms className="w-16 h-auto opacity-90 mt-2" />
        </div>

        <div className="flex flex-col items-center gap-4">
          <h1 className="font-display text-[1.4rem] md:text-[2rem] font-bold text-gold leading-tight">
            Convidados Secretos
          </h1>
          <div className="flex items-center gap-3">
            <span className="block w-1.5 h-1.5 rotate-45 bg-gold/60" />
            <span className="block w-8 h-px bg-gold/40" />
            <span className="block w-2 h-2 rotate-45 bg-gold/70" />
            <span className="block w-8 h-px bg-gold/40" />
            <span className="block w-1.5 h-1.5 rotate-45 bg-gold/60" />
          </div>
          <p className="font-script text-gold-light/85 text-[1.4rem] md:text-[1.6rem] leading-tight">
            Livro de Presenças
          </p>
        </div>

        {onOpen && (
          <button
            type="button"
            onClick={handleOpenClick}
            className="inline-flex items-center gap-2 font-heading text-[0.68rem] font-semibold tracking-[0.15em] uppercase text-gold-light py-2.5 px-5 border border-gold/55 bg-transparent hover:bg-gold/15 hover:border-gold transition-all duration-300 cursor-pointer"
          >
            <BookOpen size={13} />
            Abrir o Livro
          </button>
        )}
      </div>
    </motion.div>
  );
}
