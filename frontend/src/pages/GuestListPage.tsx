import { useState, useEffect, useMemo, useRef } from "react";
import Search from "lucide-react/dist/esm/icons/search";
import Plus from "lucide-react/dist/esm/icons/plus";
import Upload from "lucide-react/dist/esm/icons/upload";
import Pencil from "lucide-react/dist/esm/icons/pencil";
import Trash2 from "lucide-react/dist/esm/icons/trash-2";
import X from "lucide-react/dist/esm/icons/x";
import Check from "lucide-react/dist/esm/icons/check";
import Clock from "lucide-react/dist/esm/icons/clock";
import Users from "lucide-react/dist/esm/icons/users";
import FileSpreadsheet from "lucide-react/dist/esm/icons/file-spreadsheet";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import LogIn from "lucide-react/dist/esm/icons/log-in";
import LogOut from "lucide-react/dist/esm/icons/log-out";
import { Header } from "../components/Header";
import {
  listGuests,
  deleteGuest as apiDeleteGuest,
  importGuests as apiImportGuests,
  getUserRacf,
  setUserRacf,
  clearUserRacf,
} from "../lib/api";
import type { Guest, ImportResult } from "../types/guest";

function formatPhone(phone: string | null): string {
  if (!phone) return "\u2014";
  if (phone.length === 11) {
    return `(${phone.slice(0, 2)}) ${phone.slice(2, 7)}-${phone.slice(7)}`;
  }
  return phone;
}

type RelationshipFilter = "" | "P" | "R";
type ConfirmedFilter = "" | "yes" | "no";

export function GuestListPage() {
  const [guests, setGuests] = useState<Guest[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [filterRel, setFilterRel] = useState<RelationshipFilter>("");
  const [filterConf, setFilterConf] = useState<ConfirmedFilter>("");

  // RACF state
  const [racfInput, setRacfInput] = useState("");
  const [racfSaved, setRacfSaved] = useState(!!getUserRacf());

  // Delete modal
  const [deleteTarget, setDeleteTarget] = useState<Guest | null>(null);
  const [deleteLoading, setDeleteLoading] = useState(false);

  // Import modal
  const [importOpen, setImportOpen] = useState(false);
  const [importLoading, setImportLoading] = useState(false);
  const [importResult, setImportResult] = useState<ImportResult | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  useEffect(() => {
    loadGuests();
  }, []);

  async function loadGuests() {
    setLoading(true);
    setError(null);
    try {
      const data = await listGuests();
      setGuests(data);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Erro ao carregar convidados",
      );
    } finally {
      setLoading(false);
    }
  }

  const filtered = useMemo(() => {
    return guests.filter((g) => {
      const q = search.toLowerCase();
      const matchesSearch =
        !q ||
        `${g.first_name} ${g.last_name}`.toLowerCase().includes(q) ||
        (g.phone && g.phone.includes(search));
      const matchesRel = !filterRel || g.relationship === filterRel;
      const matchesConf =
        !filterConf ||
        (filterConf === "yes" ? g.confirmed : !g.confirmed);
      return matchesSearch && matchesRel && matchesConf;
    });
  }, [guests, search, filterRel, filterConf]);

  const stats = useMemo(
    () => ({
      total: guests.length,
      confirmed: guests.filter((g) => g.confirmed).length,
      pending: guests.filter((g) => !g.confirmed).length,
    }),
    [guests],
  );

  function handleSaveRacf() {
    const trimmed = racfInput.trim();
    if (/^[A-Za-z0-9]{5}$/.test(trimmed)) {
      setUserRacf(trimmed);
      setRacfSaved(true);
      setRacfInput("");
    }
  }

  function handleLogout() {
    clearUserRacf();
    setRacfSaved(false);
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    setDeleteLoading(true);
    try {
      await apiDeleteGuest(deleteTarget.id);
      setGuests((prev) => prev.filter((g) => g.id !== deleteTarget.id));
      setDeleteTarget(null);
    } catch (err) {
      setError(
        err instanceof Error ? err.message : "Erro ao excluir convidado",
      );
      setDeleteTarget(null);
    } finally {
      setDeleteLoading(false);
    }
  }

  async function handleImportFile(e: React.ChangeEvent<HTMLInputElement>) {
    const file = e.target.files?.[0];
    if (!file) return;
    setImportLoading(true);
    setImportResult(null);
    try {
      const result = await apiImportGuests(file);
      setImportResult(result);
      if (result.imported > 0) loadGuests();
    } catch (err) {
      setImportResult({
        imported: 0,
        errors: [err instanceof Error ? err.message : "Erro na importação"],
        total: 0,
      });
    } finally {
      setImportLoading(false);
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  function closeImportModal() {
    setImportOpen(false);
    setImportResult(null);
    setImportLoading(false);
  }

  // ─── Filter chip helper ───
  function chip(
    label: string,
    active: boolean,
    onClick: () => void,
  ) {
    return (
      <button
        type="button"
        onClick={onClick}
        className={`font-heading text-[0.68rem] font-semibold tracking-[0.08em] uppercase py-1.5 px-3 border transition-all duration-200 cursor-pointer ${
          active
            ? "bg-burgundy text-gold-light border-burgundy"
            : "bg-transparent text-hint border-gold-muted/50 hover:border-burgundy hover:text-burgundy"
        }`}
      >
        {label}
      </button>
    );
  }

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />

      <main className="mx-auto max-w-[1280px] px-6 pt-24 pb-16">
        {/* ─── RACF Identity Bar ─── */}
        {!racfSaved ? (
          <div className="mb-6 flex flex-col gap-3 sm:flex-row sm:items-center rounded border border-gold/30 bg-ivory px-5 py-4">
            <div className="flex items-center gap-2 flex-1">
              <LogIn size={16} className="text-burgundy shrink-0" />
              <span className="font-heading text-[0.75rem] font-semibold tracking-[0.06em] uppercase text-dark-warm">
                Identificação necessária
              </span>
            </div>
            <div className="flex gap-2">
              <input
                type="text"
                value={racfInput}
                onChange={(e) => setRacfInput(e.target.value.toUpperCase())}
                onKeyDown={(e) => e.key === "Enter" && handleSaveRacf()}
                maxLength={5}
                placeholder="RACF"
                className="w-24 px-3 py-2 text-[0.8rem] font-mono tracking-wider border border-gold-muted/50 bg-parchment text-dark-warm placeholder:text-hint/50 outline-none focus:border-burgundy transition-colors"
              />
              <button
                type="button"
                onClick={handleSaveRacf}
                disabled={!/^[A-Za-z0-9]{5}$/.test(racfInput.trim())}
                className="font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-2 px-4 bg-burgundy text-gold-light border border-burgundy cursor-pointer transition-all duration-200 hover:bg-burgundy-deep disabled:opacity-40 disabled:cursor-not-allowed"
              >
                Entrar
              </button>
            </div>
          </div>
        ) : (
          <div className="mb-6 flex items-center gap-3 text-[0.78rem] text-hint">
            <span>
              Identificado como{" "}
              <strong className="font-mono tracking-wider text-burgundy">
                {getUserRacf()}
              </strong>
            </span>
            <button
              type="button"
              onClick={handleLogout}
              className="inline-flex items-center gap-1 text-[0.72rem] text-hint/70 hover:text-burgundy transition-colors cursor-pointer"
            >
              <LogOut size={12} />
              Trocar
            </button>
          </div>
        )}

        {/* ─── Page Header ─── */}
        <div className="flex flex-col gap-4 mb-8 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h1 className="font-display text-[1.5rem] md:text-[1.8rem] font-bold text-dark">
              Lista de Presença
            </h1>
            <div className="flex flex-wrap gap-x-5 gap-y-1 mt-2.5">
              <span className="inline-flex items-center gap-1.5 text-[0.82rem] text-hint">
                <Users size={14} />
                {stats.total} convidados
              </span>
              <span className="inline-flex items-center gap-1.5 text-[0.82rem] text-burgundy">
                <Check size={14} />
                {stats.confirmed} confirmados
              </span>
              <span className="inline-flex items-center gap-1.5 text-[0.82rem] text-gold-dark">
                <Clock size={14} />
                {stats.pending} pendentes
              </span>
            </div>
          </div>

          <div className="flex gap-2.5 shrink-0">
            <button
              type="button"
              onClick={() => {
                if (!racfSaved) {
                  setError("Configure sua identificação (RACF) antes de importar.");
                  return;
                }
                setImportOpen(true);
              }}
              className="inline-flex items-center gap-[0.4rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.1rem] cursor-pointer no-underline transition-all duration-300 whitespace-nowrap bg-transparent text-burgundy border border-burgundy hover:bg-burgundy hover:text-gold-light"
            >
              <Upload size={14} />
              Importar
            </button>
            <a
              href="/lista-presenca/novo"
              className="inline-flex items-center gap-[0.4rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.1rem] cursor-pointer no-underline transition-all duration-300 whitespace-nowrap bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] hover:-translate-y-px"
            >
              <Plus size={14} />
              Novo Convidado
            </a>
          </div>
        </div>

        {/* ─── Search & Filters ─── */}
        <div className="flex flex-col gap-3 mb-6 md:flex-row md:items-center">
          <div className="relative flex-1">
            <Search
              size={16}
              className="absolute left-3.5 top-1/2 -translate-y-1/2 text-hint/60 pointer-events-none"
            />
            <input
              type="text"
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Buscar por nome ou telefone..."
              className="w-full pl-10 pr-4 py-2.5 text-[0.85rem] border border-gold-muted/40 bg-ivory text-dark-warm placeholder:text-hint/40 outline-none focus:border-burgundy transition-colors"
            />
          </div>

          <div className="flex flex-wrap gap-1.5">
            {chip("Todos", !filterRel && !filterConf, () => {
              setFilterRel("");
              setFilterConf("");
            })}
            {chip("Noivo", filterRel === "P", () =>
              setFilterRel(filterRel === "P" ? "" : "P"),
            )}
            {chip("Noiva", filterRel === "R", () =>
              setFilterRel(filterRel === "R" ? "" : "R"),
            )}
            <span className="w-px bg-gold-muted/30 mx-1 self-stretch" />
            {chip("Confirmados", filterConf === "yes", () =>
              setFilterConf(filterConf === "yes" ? "" : "yes"),
            )}
            {chip("Pendentes", filterConf === "no", () =>
              setFilterConf(filterConf === "no" ? "" : "no"),
            )}
          </div>
        </div>

        {/* ─── Error Banner ─── */}
        {error && (
          <div className="mb-4 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
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

        {/* ─── Content ─── */}
        {loading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="text-[0.85rem]">Carregando convidados...</span>
          </div>
        ) : guests.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <Users size={48} className="text-gold-muted/40 mb-4" />
            <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">
              Nenhum convidado cadastrado
            </h2>
            <p className="text-[0.88rem] text-hint max-w-[360px] mb-6">
              Comece adicionando convidados individualmente ou importe uma lista
              via arquivo CSV.
            </p>
            <a
              href="/lista-presenca/novo"
              className="inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-burgundy text-gold-light border border-burgundy transition-all duration-300 hover:bg-burgundy-deep no-underline"
            >
              <Plus size={14} />
              Adicionar Primeiro Convidado
            </a>
          </div>
        ) : filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <Search size={36} className="text-gold-muted/40 mb-3" />
            <p className="text-[0.88rem] text-hint">
              Nenhum convidado encontrado com os filtros aplicados.
            </p>
          </div>
        ) : (
          <>
            <div className="text-[0.75rem] text-hint mb-2">
              {filtered.length === guests.length
                ? `${filtered.length} convidados`
                : `${filtered.length} de ${guests.length} convidados`}
            </div>

            {/* ─── Table ─── */}
            <div className="overflow-x-auto rounded border border-gold-muted/40 bg-ivory">
              <table className="w-full min-w-[700px]">
                <thead>
                  <tr className="border-b border-gold-muted/30 bg-parchment">
                    <th className="text-left font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-hint py-3 px-4">
                      Nome
                    </th>
                    <th className="text-left font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-hint py-3 px-4">
                      Telefone
                    </th>
                    <th className="text-center font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-hint py-3 px-4">
                      Lado
                    </th>
                    <th className="text-center font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-hint py-3 px-4">
                      Grupo
                    </th>
                    <th className="text-center font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-hint py-3 px-4">
                      Status
                    </th>
                    <th className="text-center font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-hint py-3 px-4 w-24">
                      Ações
                    </th>
                  </tr>
                </thead>
                <tbody>
                  {filtered.map((guest) => (
                    <tr
                      key={guest.id}
                      className="border-b border-parchment-dark/40 transition-colors hover:bg-parchment/60"
                    >
                      <td className="py-3 px-4">
                        <span className="text-[0.88rem] font-medium text-dark-warm">
                          {guest.first_name} {guest.last_name}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-[0.84rem] text-hint font-mono tracking-wide">
                        {formatPhone(guest.phone)}
                      </td>
                      <td className="py-3 px-4 text-center">
                        <span
                          className={`inline-block font-heading text-[0.62rem] font-semibold tracking-[0.1em] uppercase py-1 px-2.5 rounded-sm ${
                            guest.relationship === "P"
                              ? "bg-burgundy/10 text-burgundy border border-burgundy/20"
                              : "bg-gold/15 text-gold-dark border border-gold/25"
                          }`}
                        >
                          {guest.relationship === "P" ? "Noivo" : "Noiva"}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-center text-[0.84rem] text-hint">
                        {guest.family_group}
                      </td>
                      <td className="py-3 px-4 text-center">
                        {guest.confirmed ? (
                          <span className="inline-flex items-center gap-1 text-[0.72rem] font-semibold text-burgundy">
                            <Check size={13} />
                            Confirmado
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1 text-[0.72rem] font-semibold text-gold-dark">
                            <Clock size={13} />
                            Pendente
                          </span>
                        )}
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center justify-center gap-1">
                          <a
                            href={`/lista-presenca/${guest.id}`}
                            className="p-1.5 text-hint hover:text-burgundy transition-colors"
                            title="Editar"
                          >
                            <Pencil size={15} />
                          </a>
                          <button
                            type="button"
                            onClick={() => {
                              if (!racfSaved) {
                                setError(
                                  "Configure sua identificação (RACF) antes de excluir.",
                                );
                                return;
                              }
                              setDeleteTarget(guest);
                            }}
                            className="p-1.5 text-hint hover:text-[#c25550] transition-colors cursor-pointer"
                            title="Excluir"
                          >
                            <Trash2 size={15} />
                          </button>
                        </div>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </>
        )}
      </main>

      {/* ─── Delete Confirmation Modal ─── */}
      {deleteTarget && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-4">
          <div
            className="absolute inset-0 bg-dark/40 backdrop-blur-[2px]"
            onClick={() => !deleteLoading && setDeleteTarget(null)}
          />
          <div className="relative bg-ivory border border-gold-muted/40 shadow-xl max-w-[400px] w-full p-6 anim-fade-in-up">
            <div className="flex flex-col items-center text-center">
              <div className="w-12 h-12 rounded-full bg-[#fef2f1] flex items-center justify-center mb-4">
                <AlertTriangle size={22} className="text-[#c25550]" />
              </div>
              <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">
                Excluir Convidado
              </h2>
              <p className="text-[0.88rem] text-hint mb-6">
                Tem certeza que deseja excluir{" "}
                <strong className="text-dark-warm">
                  {deleteTarget.first_name} {deleteTarget.last_name}
                </strong>
                ? Esta ação não pode ser desfeita.
              </p>
              <div className="flex gap-3 w-full">
                <button
                  type="button"
                  onClick={() => setDeleteTarget(null)}
                  disabled={deleteLoading}
                  className="flex-1 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-4 border border-gold-muted/50 text-hint bg-transparent transition-all duration-200 hover:border-burgundy hover:text-burgundy cursor-pointer disabled:opacity-50"
                >
                  Cancelar
                </button>
                <button
                  type="button"
                  onClick={handleDelete}
                  disabled={deleteLoading}
                  className="flex-1 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-4 border border-[#c25550] bg-[#c25550] text-white transition-all duration-200 hover:bg-[#a83f3b] cursor-pointer disabled:opacity-50"
                >
                  {deleteLoading ? "Excluindo..." : "Excluir"}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* ─── Import Modal ─── */}
      {importOpen && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-4">
          <div
            className="absolute inset-0 bg-dark/40 backdrop-blur-[2px]"
            onClick={() => !importLoading && closeImportModal()}
          />
          <div className="relative bg-ivory border border-gold-muted/40 shadow-xl max-w-[480px] w-full p-6 anim-fade-in-up">
            <button
              type="button"
              onClick={closeImportModal}
              disabled={importLoading}
              className="absolute top-4 right-4 text-hint/60 hover:text-dark-warm transition-colors cursor-pointer"
            >
              <X size={18} />
            </button>

            <div className="flex items-center gap-3 mb-5">
              <div className="w-10 h-10 rounded-full bg-burgundy/10 flex items-center justify-center">
                <FileSpreadsheet size={20} className="text-burgundy" />
              </div>
              <div>
                <h2 className="font-heading text-[1rem] font-semibold text-dark-warm">
                  Importar Convidados
                </h2>
                <p className="text-[0.75rem] text-hint">
                  Formatos aceitos: CSV ou XLSX
                </p>
              </div>
            </div>

            <div className="mb-5">
              <p className="text-[0.8rem] text-hint mb-3">
                O arquivo deve conter as colunas:{" "}
                <code className="text-[0.75rem] bg-parchment-dark/50 px-1.5 py-0.5 rounded-sm font-mono">
                  first_name, last_name, phone, relationship, family_group
                </code>
              </p>
              <label className="flex flex-col items-center justify-center gap-2 py-8 border-2 border-dashed border-gold-muted/40 bg-parchment/50 cursor-pointer transition-colors hover:border-burgundy/40 hover:bg-parchment">
                <Upload size={24} className="text-hint/50" />
                <span className="text-[0.82rem] text-hint">
                  {importLoading
                    ? "Importando..."
                    : "Clique para selecionar um arquivo"}
                </span>
                <input
                  ref={fileInputRef}
                  type="file"
                  accept=".csv,.xlsx"
                  onChange={handleImportFile}
                  disabled={importLoading}
                  className="hidden"
                />
              </label>
            </div>

            {importLoading && (
              <div className="flex items-center justify-center py-4">
                <div className="w-6 h-6 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin" />
              </div>
            )}

            {importResult && (
              <div className="border border-gold-muted/30 bg-parchment rounded p-4">
                <div className="flex flex-wrap gap-4 mb-2">
                  <span className="text-[0.82rem]">
                    <strong className="text-burgundy">
                      {importResult.imported}
                    </strong>{" "}
                    importados
                  </span>
                  <span className="text-[0.82rem]">
                    <strong className="text-hint">
                      {importResult.total}
                    </strong>{" "}
                    no arquivo
                  </span>
                  {importResult.errors.length > 0 && (
                    <span className="text-[0.82rem]">
                      <strong className="text-[#c25550]">
                        {importResult.errors.length}
                      </strong>{" "}
                      erros
                    </span>
                  )}
                </div>
                {importResult.errors.length > 0 && (
                  <div className="mt-2 max-h-32 overflow-y-auto">
                    {importResult.errors.map((err, i) => (
                      <p
                        key={i}
                        className="text-[0.75rem] text-[#c25550] leading-relaxed"
                      >
                        {err}
                      </p>
                    ))}
                  </div>
                )}
              </div>
            )}
          </div>
        </div>
      )}
    </div>
  );
}
