import { type ReactNode } from "react";
import { AnimatePresence, motion } from "framer-motion";
import { BookCover } from "./BookCover";
import { BookPages } from "./BookPages";
import { MagicParticles } from "./MagicParticles";
import type { RsvpStage } from "../../lib/rsvp-stage";

interface Props {
  stage: RsvpStage;
  onOpen: () => void;
  onCoverOpened: () => void;
  onPagesFlipped: () => void;
  onPagesClosed: () => void;
  onCoverClosed: () => void;
  children?: ReactNode;
}

export function SignatureBook({
  stage,
  onOpen,
  onCoverOpened,
  onPagesFlipped,
  onPagesClosed,
  onCoverClosed,
  children,
}: Props) {
  const coverOpen = !(stage === "cover" || stage === "closing-cover" || stage === "done");
  const showPages = stage === "pages-flipping" || stage === "closing-pages";
  const pagesDirection: "forward" | "backward" =
    stage === "closing-pages" ? "backward" : "forward";
  const showContent = stage === "family" || stage === "signing";
  const particlesActive =
    stage === "opening" ||
    stage === "pages-flipping" ||
    stage === "closing-pages" ||
    stage === "closing-cover";

  return (
    <div className="book-scene w-full max-w-[560px] md:max-w-[640px] mx-auto">
     <div className="relative">
      <MagicParticles active={particlesActive} />
      <div className="book-root relative aspect-[3/4] md:aspect-[4/5] w-full">
        {/* Inside spread (always rendered behind cover) */}
        <div
          aria-hidden={stage === "cover"}
          className="absolute inset-0 z-10 book-parchment border-2 border-gold-muted/50 shadow-[0_6px_30px_rgba(28,20,16,0.18)] overflow-hidden"
        >
          <div className="absolute inset-3 border border-gold/25 pointer-events-none" />
          <div className="absolute inset-0 book-spine-shadow pointer-events-none" />

          <AnimatePresence mode="wait">
            {showContent && (
              <motion.div
                key={stage}
                initial={{ opacity: 0, y: 12 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -8 }}
                transition={{ duration: 0.35, ease: "easeOut" }}
                className="absolute inset-0 flex items-center justify-center p-6 md:p-10 overflow-y-auto"
              >
                <div className="w-full max-w-[440px]">{children}</div>
              </motion.div>
            )}
          </AnimatePresence>
        </div>

        {/* Flipping pages overlay */}
        {showPages && (
          <BookPages
            pageCount={9}
            direction={pagesDirection}
            onComplete={() => {
              if (stage === "pages-flipping") onPagesFlipped();
              else if (stage === "closing-pages") onPagesClosed();
            }}
          />
        )}

        {/* Cover */}
        <BookCover
          open={coverOpen}
          onOpen={stage === "cover" ? onOpen : undefined}
          onOpenComplete={() => {
            if (stage === "opening") onCoverOpened();
          }}
          onClosed={() => {
            if (stage === "closing-cover") onCoverClosed();
          }}
        />
      </div>
     </div>
    </div>
  );
}
