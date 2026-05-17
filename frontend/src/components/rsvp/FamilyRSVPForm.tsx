import { useEffect, useMemo, useRef, useState } from "react";
import Check from "lucide-react/dist/esm/icons/check";
import Clock from "lucide-react/dist/esm/icons/clock";
import Users from "lucide-react/dist/esm/icons/users";
import { AlreadyRespondedView } from "./AlreadyRespondedView";
import type { Guest } from "../../types/guest";

export type RSVPAction =
  | { kind: "confirm-whole" }
  | { kind: "confirm-selected"; ids: number[] }
  | { kind: "decline-selected"; ids: number[] };

interface Props {
  family: Guest[];
  currentGuestID: number | null;
  onSubmit: (action: RSVPAction) => void;
  disabled?: boolean;
}

export function FamilyRSVPForm({ family, currentGuestID, onSubmit, disabled = false }: Props) {
  const [selected, setSelected] = useState<Record<number, boolean>>(() => {
    const initial: Record<number, boolean> = {};
    for (const g of family) initial[g.id] = g.confirmed;
    return initial;
  });
  const [validationError, setValidationError] = useState<string | null>(null);
  const firstCheckboxRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    firstCheckboxRef.current?.focus();
  }, []);

  const selectedIds = useMemo(
    () => family.filter((g) => selected[g.id]).map((g) => g.id),
    [selected, family],
  );

  const allSelected = selectedIds.length === family.length && family.length > 0;
  const hasAnyConfirmed = family.some((g) => g.confirmed);
  const standalone = family.length <= 1;

  function toggle(id: number) {
    setSelected((prev) => ({ ...prev, [id]: !prev[id] }));
    setValidationError(null);
  }

  function toggleAll() {
    if (allSelected) {
      setSelected({});
    } else {
      const next: Record<number, boolean> = {};
      for (const g of family) next[g.id] = true;
      setSelected(next);
    }
    setValidationError(null);
  }

  function handleConfirmWhole() {
    onSubmit({ kind: "confirm-whole" });
  }

  function handleConfirmSelected() {
    if (selectedIds.length === 0) {
      setValidationError("Selecione ao menos um nome.");
      return;
    }
    onSubmit({ kind: "confirm-selected", ids: selectedIds });
  }

  function handleDeclineSelected() {
    if (selectedIds.length === 0) {
      setValidationError("Selecione ao menos um nome.");
      return;
    }
    onSubmit({ kind: "decline-selected", ids: selectedIds });
  }

  return (
    <div className="w-full text-dark-warm">
      <div className="flex items-center justify-between gap-3 mb-3 pb-3 border-b border-gold-muted/40">
        <div className="flex items-center gap-2">
          <Users size={14} className="text-gold-dark" />
          <span className="font-heading text-[0.66rem] tracking-[0.2em] uppercase text-dark-warm/70">
            Família ({family.length})
          </span>
        </div>
        {!standalone && (
          <button
            type="button"
            onClick={toggleAll}
            disabled={disabled}
            className="text-[0.7rem] font-heading tracking-[0.1em] uppercase text-burgundy hover:text-burgundy-deep cursor-pointer disabled:opacity-50"
          >
            {allSelected ? "Desmarcar todos" : "Selecionar todos"}
          </button>
        )}
      </div>

      {hasAnyConfirmed && <AlreadyRespondedView />}

      <ul className="divide-y divide-gold-muted/25 mb-5">
        {family.map((guest, idx) => {
          const isMe = currentGuestID === guest.id;
          return (
            <li key={guest.id} className="flex items-center gap-3 py-2.5">
              <label className="flex items-center gap-3 cursor-pointer flex-1 min-w-0">
                <input
                  ref={idx === 0 ? firstCheckboxRef : undefined}
                  type="checkbox"
                  checked={!!selected[guest.id]}
                  onChange={() => toggle(guest.id)}
                  disabled={disabled}
                  className="w-4 h-4 cursor-pointer accent-[#989F5B]"
                />
                <span className="font-body text-[1.1rem] text-dark-warm truncate">
                  {guest.first_name} {guest.last_name}
                  {isMe && (
                    <span className="ml-2 font-heading text-[0.6rem] tracking-[0.15em] uppercase text-gold-dark align-middle">
                      (você)
                    </span>
                  )}
                </span>
              </label>
              {guest.confirmed ? (
                <span className="inline-flex items-center gap-1 text-[0.66rem] font-bold text-burgundy bg-burgundy/10 border border-burgundy/25 px-2 py-0.5 font-heading tracking-wide uppercase">
                  <Check size={10} />
                  Confirmado
                </span>
              ) : (
                <span className="inline-flex items-center gap-1 text-[0.66rem] font-bold text-gold-dark bg-gold/10 border border-gold/30 px-2 py-0.5 font-heading tracking-wide uppercase">
                  <Clock size={10} />
                  Aguardando
                </span>
              )}
            </li>
          );
        })}
      </ul>

      {validationError && (
        <p role="alert" className="mb-3 text-[0.78rem] text-[#c25550] font-body italic">
          {validationError}
        </p>
      )}

      <div className="flex flex-col sm:flex-row gap-2.5">
        {!standalone && (
          <button
            type="button"
            onClick={handleConfirmWhole}
            disabled={disabled}
            className="flex-1 inline-flex items-center justify-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.1em] uppercase py-[0.7rem] px-4 bg-burgundy text-gold-light border border-burgundy transition-all duration-300 hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Check size={14} />
            Confirmar Todos
          </button>
        )}
        <button
          type="button"
          onClick={handleConfirmSelected}
          disabled={disabled}
          className={`flex-1 inline-flex items-center justify-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.1em] uppercase py-[0.7rem] px-4 transition-all duration-300 cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed ${
            standalone
              ? "bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)]"
              : "bg-transparent text-burgundy border border-burgundy hover:bg-burgundy hover:text-gold-light"
          }`}
        >
          <Check size={14} />
          {standalone ? "Confirmar minha presença" : "Confirmar Selecionados"}
        </button>
        <button
          type="button"
          onClick={handleDeclineSelected}
          disabled={disabled}
          className="flex-1 inline-flex items-center justify-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.1em] uppercase py-[0.7rem] px-4 bg-transparent text-dark-warm/70 border border-gold-muted/60 transition-all duration-300 hover:border-[#c25550] hover:text-[#c25550] cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
        >
          Recusar Selecionados
        </button>
      </div>
    </div>
  );
}
