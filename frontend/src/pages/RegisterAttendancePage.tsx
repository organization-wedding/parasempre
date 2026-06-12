import { useCallback, useEffect, useMemo, useRef, useState } from "react";
import { Link } from "@tanstack/react-router";
import { useReducedMotion } from "framer-motion";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";
import Gift from "lucide-react/dist/esm/icons/gift";
import { Header } from "../components/Header";
import { Footer } from "../components/Footer";
import { Toast } from "../components/Toast";
import { AdminHostView } from "../components/rsvp/AdminHostView";
import { FamilyRSVPForm, type RSVPAction } from "../components/rsvp/FamilyRSVPForm";
import { QuillSignature } from "../components/rsvp/QuillSignature";
import { SignatureBook } from "../components/rsvp/SignatureBook";
import { useUserMeQuery } from "../lib/user-queries";
import { useMyFamilyQuery, useBatchFamilyMutation, useConfirmWholeFamilyMutation, useCancelWholeFamilyMutation } from "../lib/family-queries";
import { STAGE_LABELS, STAGE_TIMING, type RsvpStage } from "../lib/rsvp-stage";

type ToastState = { title: string; message: string; variant?: "error" | "success" } | null;

export function RegisterAttendancePage() {
  const reducedMotion = useReducedMotion();
  const { data: me, isLoading: meLoading } = useUserMeQuery();
  const guestID = me?.guest_id ?? null;
  const familyGroup = me?.family_group ?? null;
  const isGuestUser = guestID !== null && guestID !== undefined;

  const { data: family = [], isLoading: familyLoading } = useMyFamilyQuery(isGuestUser);
  const batchMutation = useBatchFamilyMutation();
  const confirmAllMutation = useConfirmWholeFamilyMutation();
  const cancelAllMutation = useCancelWholeFamilyMutation();

  const [stage, setStage] = useState<RsvpStage>("cover");
  const [toast, setToast] = useState<ToastState>(null);
  const pendingAction = useRef<RSVPAction | null>(null);
  const liveRegionRef = useRef<HTMLDivElement>(null);

  const isMutating =
    batchMutation.isPending || confirmAllMutation.isPending || cancelAllMutation.isPending;

  // Announce stage changes to screen readers
  useEffect(() => {
    if (liveRegionRef.current) {
      liveRegionRef.current.textContent = STAGE_LABELS[stage];
    }
  }, [stage]);

  const firstName = useMemo(() => {
    if (me?.first_name) return me.first_name;
    const meGuest = family.find((g) => g.id === guestID);
    return meGuest?.first_name ?? "Convidado";
  }, [me?.first_name, family, guestID]);

  const runMutation = useCallback(
    async (action: RSVPAction): Promise<boolean> => {
      try {
        if (action.kind === "confirm-whole") {
          if (familyGroup === null || familyGroup === undefined) {
            throw new Error("Família não identificada");
          }
          await confirmAllMutation.mutateAsync(familyGroup);
        } else if (action.kind === "confirm-selected") {
          if (familyGroup !== null && familyGroup !== undefined && action.ids.length === family.length) {
            await confirmAllMutation.mutateAsync(familyGroup);
          } else {
            await batchMutation.mutateAsync({ guest_ids: action.ids, attending: true });
          }
        } else {
          if (familyGroup !== null && familyGroup !== undefined && action.ids.length === family.length) {
            await cancelAllMutation.mutateAsync(familyGroup);
          } else {
            await batchMutation.mutateAsync({ guest_ids: action.ids, attending: false });
          }
        }
        return true;
      } catch (err) {
        const message = err instanceof Error ? err.message : "Não foi possível registrar sua resposta.";
        setToast({ title: "Erro ao registrar", message });
        return false;
      }
    },
    [familyGroup, family.length, confirmAllMutation, cancelAllMutation, batchMutation],
  );

  const handleFormSubmit = useCallback(
    (action: RSVPAction) => {
      pendingAction.current = action;
      if (reducedMotion) {
        // Skip animations — fire mutation immediately, jump to done
        void (async () => {
          const ok = await runMutation(action);
          if (ok) {
            const isDecline = action.kind === "decline-selected";
            setToast({
              title: isDecline ? "Resposta registrada" : "Presença registrada",
              message: isDecline
                ? "Sua ausência foi registrada no livro dos convidados."
                : "Sua resposta foi salva no livro dos convidados.",
              variant: "success",
            });
            setStage("done");
          }
        })();
        return;
      }
      setStage("signing");
    },
    [reducedMotion, runMutation],
  );

  const handleSigningComplete = useCallback(() => {
    void (async () => {
      const action = pendingAction.current;
      if (!action) {
        setStage("closing-pages");
        return;
      }
      const ok = await runMutation(action);
      if (ok) {
        const isDecline = action.kind === "decline-selected";
        setToast({
          title: isDecline ? "Resposta registrada" : "Presença registrada",
          message: isDecline
            ? "Sua ausência foi registrada no livro dos convidados."
            : "Sua resposta foi salva no livro dos convidados.",
          variant: "success",
        });
        setStage("closing-pages");
      } else {
        setStage("family");
      }
      pendingAction.current = null;
    })();
  }, [runMutation]);

  const handleOpen = useCallback(() => setStage("opening"), []);
  const handleCoverOpened = useCallback(() => setStage("pages-flipping"), []);
  const handlePagesFlipped = useCallback(() => setStage("family"), []);
  const handlePagesClosed = useCallback(() => setStage("closing-cover"), []);
  const handleCoverClosed = useCallback(() => setStage("done"), []);

  // Loading state — render the cover only
  if (meLoading || (isGuestUser && familyLoading)) {
    return (
      <div className="min-h-dvh bg-parchment flex flex-col">
        <Header />
        <main className="flex-1 flex items-center justify-center px-6 pt-24 pb-16">
          <div className="flex flex-col items-center text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="font-heading text-[0.75rem] tracking-[0.15em] uppercase">
              Carregando livro
            </span>
          </div>
        </main>
        <Footer />
      </div>
    );
  }

  // Admin / no guest_id — friendly host view
  if (me && !isGuestUser) {
    return (
      <div className="min-h-dvh bg-parchment flex flex-col">
        <Header />
        <main className="flex-1 flex items-center justify-center px-6 pt-24 pb-16">
          <AdminHostView />
        </main>
        <Footer />
      </div>
    );
  }

  // Reduced motion — render the form directly
  if (reducedMotion) {
    return (
      <div className="min-h-dvh bg-parchment flex flex-col">
        <Header />
        <main className="flex-1 flex items-center justify-center px-6 pt-24 pb-16">
          {stage === "done" ? (
            <DoneCard firstName={firstName} />
          ) : (
            <div className="w-full max-w-[480px] book-parchment border-2 border-gold-muted/60 shadow-[0_8px_40px_rgba(28,20,16,0.18)] p-8 md:p-10">
              <h1 className="font-display text-[1.4rem] md:text-[1.7rem] font-bold text-dark-warm text-center mb-1">
                Livro dos Convidados
              </h1>
              <p className="font-heading text-[0.62rem] tracking-[0.3em] uppercase text-gold-dark text-center mb-6">
                Confirme sua presença
              </p>
              <FamilyRSVPForm
                family={family}
                currentGuestID={guestID}
                onSubmit={handleFormSubmit}
                disabled={isMutating}
              />
            </div>
          )}
        </main>
        {toast && (
          <Toast
            title={toast.title}
            message={toast.message}
            variant={toast.variant}
            onDismiss={() => setToast(null)}
          />
        )}
        <Footer />
      </div>
    );
  }

  return (
    <div className="min-h-dvh bg-parchment flex flex-col">
      <Header />

      <main className="flex-1 flex items-center justify-center px-4 pt-24 pb-16">
        <div ref={liveRegionRef} aria-live="polite" className="sr-only" />

        {stage === "done" ? (
          <DoneCard firstName={firstName} />
        ) : (
          <SignatureBook
            stage={stage}
            onOpen={handleOpen}
            onCoverOpened={handleCoverOpened}
            onPagesFlipped={handlePagesFlipped}
            onPagesClosed={handlePagesClosed}
            onCoverClosed={handleCoverClosed}
          >
            {stage === "family" && (
              <div>
                <p className="font-heading text-[0.6rem] tracking-[0.3em] uppercase text-gold-dark text-center mb-1">
                  Capítulo I
                </p>
                <h2 className="font-display text-[1.1rem] md:text-[1.3rem] font-bold text-dark-warm text-center mb-5">
                  Vossas Presenças
                </h2>
                <FamilyRSVPForm
                  family={family}
                  currentGuestID={guestID}
                  onSubmit={handleFormSubmit}
                  disabled={isMutating}
                />
              </div>
            )}
            {stage === "signing" && (
              <QuillSignature firstName={firstName} onComplete={handleSigningComplete} />
            )}
          </SignatureBook>
        )}
      </main>

      {toast && (
        <Toast
          title={toast.title}
          message={toast.message}
          variant={toast.variant}
          onDismiss={() => setToast(null)}
        />
      )}

      <Footer />
    </div>
  );
}

function DoneCard({ firstName }: { firstName: string }) {
  return (
    <div className="w-full max-w-[460px] book-parchment border-2 border-gold-muted/60 shadow-[0_8px_40px_rgba(28,20,16,0.18)] p-10 text-center anim-fade-in-up">
      <div className="absolute inset-3 border border-gold/30 pointer-events-none" />
      <p className="font-heading text-[0.62rem] tracking-[0.3em] uppercase text-gold-dark mb-3">
        Selado
      </p>
      <h1 className="font-display text-[1.5rem] md:text-[1.9rem] font-bold text-dark-warm mb-2">
        Resposta registrada
      </h1>
      <p className="font-script text-burgundy text-[2.4rem] leading-tight mb-3">
        {firstName}
      </p>
      <p className="font-body italic text-[1rem] text-dark-warm/75 mb-8">
        Vosso nome consta no livro dos convidados.
      </p>

      <div className="flex flex-col sm:flex-row gap-2.5 justify-center">
        <Link
          to="/"
          className="inline-flex items-center justify-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.1em] uppercase py-[0.65rem] px-5 border border-burgundy text-burgundy bg-transparent hover:bg-burgundy hover:text-gold-light transition-all duration-300 no-underline"
        >
          <ArrowLeft size={14} />
          Voltar ao início
        </Link>
        <Link
          to="/lista-presentes"
          search={{ page: undefined }}
          className="inline-flex items-center justify-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.1em] uppercase py-[0.65rem] px-5 bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] transition-all duration-300 no-underline"
        >
          <Gift size={14} />
          Comprar Presente
        </Link>
      </div>
    </div>
  );
}
