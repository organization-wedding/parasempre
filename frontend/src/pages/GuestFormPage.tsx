import { useState, useEffect } from "react";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";
import Save from "lucide-react/dist/esm/icons/save";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import X from "lucide-react/dist/esm/icons/x";
import { Header } from "../components/Header";
import { getGuest, createGuest, updateGuest, getUserRacf } from "../lib/api";

interface Props {
  guestId?: number;
}

export function GuestFormPage({ guestId }: Props) {
  const isEdit = guestId !== undefined;

  const [loading, setLoading] = useState(isEdit);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const [firstName, setFirstName] = useState("");
  const [lastName, setLastName] = useState("");
  const [phone, setPhone] = useState("");
  const [relationship, setRelationship] = useState("P");
  const [familyGroup, setFamilyGroup] = useState("1");
  const [confirmed, setConfirmed] = useState(false);

  useEffect(() => {
    if (!isEdit) return;
    getGuest(guestId)
      .then((guest) => {
        setFirstName(guest.first_name);
        setLastName(guest.last_name);
        setPhone(guest.phone || "");
        setRelationship(guest.relationship);
        setFamilyGroup(String(guest.family_group));
        setConfirmed(guest.confirmed);
      })
      .catch((err) => {
        setError(err instanceof Error ? err.message : "Erro ao carregar convidado");
      })
      .finally(() => setLoading(false));
  }, [guestId, isEdit]);

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault();

    if (!getUserRacf()) {
      setError(
        "Configure sua identificação (RACF) na página de Lista de Presença antes de salvar.",
      );
      return;
    }

    if (!firstName.trim() || !lastName.trim()) {
      setError("Nome e sobrenome são obrigatórios.");
      return;
    }

    const fg = Number(familyGroup);
    if (!fg || fg < 1) {
      setError("Grupo familiar deve ser um número válido.");
      return;
    }

    setSaving(true);
    setError(null);

    try {
      if (isEdit) {
        await updateGuest(guestId, {
          first_name: firstName.trim(),
          last_name: lastName.trim(),
          phone: phone.trim() || undefined,
          relationship,
          family_group: fg,
          confirmed,
        });
      } else {
        await createGuest({
          first_name: firstName.trim(),
          last_name: lastName.trim(),
          phone: phone.trim(),
          relationship,
          family_group: fg,
        });
      }
      window.location.href = "/lista-presenca";
    } catch (err) {
      setError(err instanceof Error ? err.message : "Erro ao salvar");
    } finally {
      setSaving(false);
    }
  }

  // ─── Shared input styles ───
  const inputClass =
    "w-full px-3.5 py-2.5 text-[0.88rem] border border-gold-muted/40 bg-ivory text-dark-warm placeholder:text-hint/40 outline-none focus:border-burgundy transition-colors";
  const labelClass =
    "block font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase text-hint mb-1.5";

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />

      <main className="mx-auto max-w-[640px] px-6 pt-24 pb-16">
        {/* ─── Back Link ─── */}
        <a
          href="/lista-presenca"
          className="inline-flex items-center gap-1.5 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase text-hint no-underline mb-6 transition-colors hover:text-burgundy"
        >
          <ArrowLeft size={15} />
          Voltar para lista
        </a>

        {/* ─── Title ─── */}
        <h1 className="font-display text-[1.4rem] md:text-[1.7rem] font-bold text-dark mb-8">
          {isEdit ? "Editar Convidado" : "Novo Convidado"}
        </h1>

        {/* ─── Error ─── */}
        {error && (
          <div className="mb-6 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
            <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
            <span className="text-[0.82rem] text-[#7a2e2b] flex-1">
              {error}
            </span>
            <button
              type="button"
              onClick={() => setError(null)}
              className="text-[#c25550]/60 hover:text-[#c25550] cursor-pointer"
            >
              <X size={14} />
            </button>
          </div>
        )}

        {loading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="text-[0.85rem]">Carregando dados...</span>
          </div>
        ) : (
          <form
            onSubmit={handleSubmit}
            className="border border-gold-muted/40 bg-ivory p-6 md:p-8 rounded"
          >
            {/* ─── Name Fields ─── */}
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-5">
              <div>
                <label className={labelClass}>Nome</label>
                <input
                  type="text"
                  value={firstName}
                  onChange={(e) => setFirstName(e.target.value)}
                  placeholder="Ex: João"
                  required
                  className={inputClass}
                />
              </div>
              <div>
                <label className={labelClass}>Sobrenome</label>
                <input
                  type="text"
                  value={lastName}
                  onChange={(e) => setLastName(e.target.value)}
                  placeholder="Ex: Silva"
                  required
                  className={inputClass}
                />
              </div>
            </div>

            {/* ─── Phone ─── */}
            <div className="mb-5">
              <label className={labelClass}>Telefone</label>
              <input
                type="text"
                value={phone}
                onChange={(e) =>
                  setPhone(e.target.value.replace(/\D/g, "").slice(0, 11))
                }
                placeholder="11999999999"
                maxLength={11}
                className={`${inputClass} font-mono tracking-wider`}
              />
              <p className="text-[0.72rem] text-hint/60 mt-1">
                Formato: DDD + número (11 dígitos). Opcional.
              </p>
            </div>

            {/* ─── Relationship ─── */}
            <div className="mb-5">
              <label className={labelClass}>Lado</label>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setRelationship("P")}
                  className={`flex-1 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-2.5 border transition-all duration-200 cursor-pointer ${
                    relationship === "P"
                      ? "bg-burgundy text-gold-light border-burgundy"
                      : "bg-transparent text-hint border-gold-muted/50 hover:border-burgundy hover:text-burgundy"
                  }`}
                >
                  Lado do Noivo
                </button>
                <button
                  type="button"
                  onClick={() => setRelationship("R")}
                  className={`flex-1 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-2.5 border transition-all duration-200 cursor-pointer ${
                    relationship === "R"
                      ? "bg-gold text-dark border-gold"
                      : "bg-transparent text-hint border-gold-muted/50 hover:border-gold hover:text-gold-dark"
                  }`}
                >
                  Lado da Noiva
                </button>
              </div>
            </div>

            {/* ─── Family Group ─── */}
            <div className="mb-5">
              <label className={labelClass}>Grupo Familiar</label>
              <input
                type="number"
                value={familyGroup}
                onChange={(e) => setFamilyGroup(e.target.value)}
                min={1}
                required
                className={`${inputClass} w-32`}
              />
              <p className="text-[0.72rem] text-hint/60 mt-1">
                Número que agrupa membros da mesma família.
              </p>
            </div>

            {/* ─── Confirmed (edit only) ─── */}
            {isEdit && (
              <div className="mb-6">
                <label className="flex items-center gap-3 cursor-pointer group">
                  <div className="relative">
                    <input
                      type="checkbox"
                      checked={confirmed}
                      onChange={(e) => setConfirmed(e.target.checked)}
                      className="sr-only peer"
                    />
                    <div className="w-5 h-5 border border-gold-muted/50 bg-ivory peer-checked:bg-burgundy peer-checked:border-burgundy transition-all duration-200 flex items-center justify-center">
                      {confirmed && (
                        <svg
                          viewBox="0 0 16 16"
                          className="w-3 h-3 text-gold-light"
                          fill="none"
                          stroke="currentColor"
                          strokeWidth="2.5"
                          strokeLinecap="round"
                          strokeLinejoin="round"
                        >
                          <polyline points="3 8 7 12 13 4" />
                        </svg>
                      )}
                    </div>
                  </div>
                  <span className="font-heading text-[0.75rem] font-semibold tracking-[0.06em] uppercase text-dark-warm group-hover:text-burgundy transition-colors">
                    Presença confirmada
                  </span>
                </label>
              </div>
            )}

            {/* ─── Actions ─── */}
            <div className="flex flex-col-reverse sm:flex-row gap-3 sm:justify-end pt-4 border-t border-gold-muted/20">
              <a
                href="/lista-presenca"
                className="inline-flex items-center justify-center font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 border border-gold-muted/50 text-hint bg-transparent transition-all duration-200 hover:border-burgundy hover:text-burgundy no-underline cursor-pointer"
              >
                Cancelar
              </a>
              <button
                type="submit"
                disabled={saving}
                className="inline-flex items-center justify-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-burgundy text-gold-light border border-burgundy transition-all duration-200 hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <Save size={14} />
                {saving ? "Salvando..." : "Salvar"}
              </button>
            </div>
          </form>
        )}
      </main>
    </div>
  );
}
