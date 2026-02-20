import { useEffect, useMemo, useRef, useState } from "react";
import { Link } from "@tanstack/react-router";
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
import SlidersHorizontal from "lucide-react/dist/esm/icons/sliders-horizontal";
import { Header } from "../components/Header";
import {
  useDeleteGuestMutation,
  useDeleteGuestsMutation,
  useGuestsQuery,
  useImportGuestsMutation,
} from "../lib/guest-queries";
import { useUserMeQuery } from "../lib/user-queries";
import { UnauthorizedPage } from "./UnauthorizedPage";
import type { Guest, ImportResult } from "../types/guest";

function formatPhone(phone: string | null): string {
  if (!phone) return "—";
  if (phone.length === 11) {
    return `(${phone.slice(0, 2)}) ${phone.slice(2, 7)}-${phone.slice(7)}`;
  }
  return phone;
}

type RelationshipFilter = "" | "P" | "R";
type ConfirmedFilter = "" | "yes" | "no";

export function GuestListPage() {
  const { data: guests = [], isLoading, error, refetch } = useGuestsQuery();
  const deleteMutation = useDeleteGuestMutation();
  const deletesMutation = useDeleteGuestsMutation();
  const importMutation = useImportGuestsMutation();

  const [uiError, setUiError] = useState<string | null>(null);
  const [search, setSearch] = useState("");
  const [filterRel, setFilterRel] = useState<RelationshipFilter>("");
  const [filterConf, setFilterConf] = useState<ConfirmedFilter>("");
  const [filterOpen, setFilterOpen] = useState(false);
  const filterRef = useRef<HTMLDivElement>(null);
  const [deleteTarget, setDeleteTarget] = useState<Guest | null>(null);
  const [importOpen, setImportOpen] = useState(false);
  const [importResult, setImportResult] = useState<ImportResult | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  // Bulk selection
  const [selectedIds, setSelectedIds] = useState<Set<number>>(new Set());
  const [bulkDeleteOpen, setBulkDeleteOpen] = useState(false);
  const selectAllRef = useRef<HTMLInputElement>(null);

  // Role check — RACF comes from the impersonation modal
  const { data: userMe, isLoading: roleLoading } = useUserMeQuery();
  const isAuthorized = userMe?.role === "groom" || userMe?.role === "bride";

  const effectiveError = uiError ?? (error instanceof Error ? error.message : null);

  const filtered = useMemo(() => {
    return guests.filter((g) => {
      const q = search.toLowerCase();
      const matchesSearch =
        !q ||
        `${g.first_name} ${g.last_name}`.toLowerCase().includes(q) ||
        (g.phone && g.phone.includes(search));
      const matchesRel = !filterRel || g.relationship === filterRel;
      const matchesConf = !filterConf || (filterConf === "yes" ? g.confirmed : !g.confirmed);
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

  const activeFilterCount = (filterRel !== "" ? 1 : 0) + (filterConf !== "" ? 1 : 0);
  const allSelected = filtered.length > 0 && selectedIds.size === filtered.length;
  const someSelected = selectedIds.size > 0 && selectedIds.size < filtered.length;

  useEffect(() => {
    if (selectAllRef.current) {
      selectAllRef.current.indeterminate = someSelected;
    }
  }, [someSelected]);

  useEffect(() => {
    setSelectedIds(new Set());
  }, [search, filterRel, filterConf]);

  // Close filter dropdown on outside click
  useEffect(() => {
    function handler(e: MouseEvent) {
      if (filterRef.current && !filterRef.current.contains(e.target as Node)) {
        setFilterOpen(false);
      }
    }
    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  function toggleSelect(id: number) {
    setSelectedIds((prev) => {
      const next = new Set(prev);
      if (next.has(id)) next.delete(id);
      else next.add(id);
      return next;
    });
  }

  function toggleSelectAll() {
    if (allSelected) setSelectedIds(new Set());
    else setSelectedIds(new Set(filtered.map((g) => g.id)));
  }

  async function handleDelete() {
    if (!deleteTarget) return;
    try {
      await deleteMutation.mutateAsync(deleteTarget.id);
      setDeleteTarget(null);
    } catch (mutationError) {
      setUiError(mutationError instanceof Error ? mutationError.message : "Erro ao excluir convidado");
      setDeleteTarget(null);
    }
  }

  async function handleBulkDelete() {
    try {
      await deletesMutation.mutateAsync(Array.from(selectedIds));
      setSelectedIds(new Set());
      setBulkDeleteOpen(false);
    } catch (mutationError) {
      setUiError(mutationError instanceof Error ? mutationError.message : "Erro ao excluir convidados");
      setBulkDeleteOpen(false);
    }
  }

  async function handleImportFile(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;
    setImportResult(null);
    try {
      const result = await importMutation.mutateAsync(file);
      setImportResult(result);
      if (result.imported > 0) void refetch();
    } catch (mutationError) {
      setImportResult({
        imported: 0,
        errors: [mutationError instanceof Error ? mutationError.message : "Erro na importação"],
        total: 0,
      });
    } finally {
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  function closeImportModal() {
    setImportOpen(false);
    setImportResult(null);
  }

  function openImport() {
    setImportOpen(true);
  }

  // Unauthorized — show when role check finished and user is not groom/bride
  if (!roleLoading && !isAuthorized) {
    return <UnauthorizedPage />;
  }

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />

      <main className="mx-auto max-w-[1280px] px-6 pt-24 pb-16">
        {roleLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="text-[0.85rem]">Verificando permissões...</span>
          </div>
        ) : (
        <>
        {/* Page header */}
        <div className="flex flex-col gap-4 mb-8 sm:flex-row sm:items-start sm:justify-between">
          <div>
            <h1 className="font-display text-[1.5rem] md:text-[1.8rem] font-bold text-dark">Gerenciar Convidados</h1>
            <div className="flex flex-wrap gap-x-5 gap-y-1 mt-2.5">
              <span className="inline-flex items-center gap-1.5 text-[0.82rem] text-dark-warm/70">
                <Users size={14} />
                {stats.total} convidados
              </span>
              <span className="inline-flex items-center gap-1.5 text-[0.82rem] text-burgundy font-medium">
                <Check size={14} />
                {stats.confirmed} confirmados
              </span>
              <span className="inline-flex items-center gap-1.5 text-[0.82rem] text-gold-dark font-medium">
                <Clock size={14} />
                {stats.pending} pendentes
              </span>
            </div>
          </div>

          <div className="flex gap-2.5 shrink-0">
            <button
              type="button"
              onClick={openImport}
              className="inline-flex items-center gap-[0.4rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.1rem] cursor-pointer no-underline transition-all duration-300 whitespace-nowrap bg-transparent text-burgundy border border-burgundy hover:bg-burgundy hover:text-gold-light"
            >
              <Upload size={14} />
              Importar
            </button>
            <Link
              to="/dashboard/novo"
              className="inline-flex items-center gap-[0.4rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.1rem] cursor-pointer no-underline transition-all duration-300 whitespace-nowrap bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] hover:-translate-y-px"
            >
              <Plus size={14} />
              Novo Convidado
            </Link>
          </div>
        </div>

        {/* Search + Filter button */}
        <div className="flex flex-col gap-3 mb-6 sm:flex-row sm:items-center">
          <div className="relative flex-1">
            <Search size={16} className="absolute left-3.5 top-1/2 -translate-y-1/2 text-hint/50 pointer-events-none" />
            <input
              type="text"
              value={search}
              onChange={(event) => setSearch(event.target.value)}
              placeholder="Buscar por nome ou telefone..."
              className="w-full pl-10 pr-4 py-2.5 text-[0.85rem] border border-gold-muted/50 bg-ivory text-dark-warm placeholder:text-hint/40 outline-none focus:border-burgundy transition-colors"
            />
          </div>

          {/* Filter dropdown */}
          <div ref={filterRef} className="relative shrink-0">
            <button
              type="button"
              onClick={() => setFilterOpen((v) => !v)}
              className={`inline-flex items-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-2.5 px-4 border transition-all duration-200 cursor-pointer whitespace-nowrap ${
                filterOpen || activeFilterCount > 0
                  ? "bg-burgundy text-gold-light border-burgundy"
                  : "bg-ivory text-dark-warm border-gold-muted/50 hover:border-burgundy hover:text-burgundy"
              }`}
            >
              <SlidersHorizontal size={14} />
              Filtros
              {activeFilterCount > 0 && (
                <span className="inline-flex items-center justify-center w-4 h-4 rounded-full bg-gold-light/30 text-[0.6rem] font-bold leading-none">
                  {activeFilterCount}
                </span>
              )}
            </button>

            {filterOpen && (
              <div className="absolute top-full right-0 mt-1 z-[60] bg-ivory border border-gold-muted/50 shadow-[0_8px_24px_rgba(28,20,16,0.12)] w-64 p-4">
                {/* Lado */}
                <div className="mb-4">
                  <p className="font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint mb-2">
                    Lado
                  </p>
                  <div className="flex gap-1.5 flex-wrap">
                    {(["", "P", "R"] as RelationshipFilter[]).map((val) => (
                      <button
                        key={val}
                        type="button"
                        onClick={() => setFilterRel(val)}
                        className={`font-heading text-[0.67rem] font-semibold tracking-[0.06em] uppercase py-1 px-3 border transition-all duration-150 cursor-pointer ${
                          filterRel === val
                            ? "bg-burgundy text-gold-light border-burgundy"
                            : "bg-transparent text-dark-warm border-gold-muted/40 hover:border-burgundy hover:text-burgundy"
                        }`}
                      >
                        {val === "" ? "Todos" : val === "P" ? "Noivo" : "Noiva"}
                      </button>
                    ))}
                  </div>
                </div>

                {/* Status */}
                <div className="mb-4">
                  <p className="font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint mb-2">
                    Status
                  </p>
                  <div className="flex gap-1.5 flex-wrap">
                    {(["", "yes", "no"] as ConfirmedFilter[]).map((val) => (
                      <button
                        key={val}
                        type="button"
                        onClick={() => setFilterConf(val)}
                        className={`font-heading text-[0.67rem] font-semibold tracking-[0.06em] uppercase py-1 px-3 border transition-all duration-150 cursor-pointer ${
                          filterConf === val
                            ? "bg-burgundy text-gold-light border-burgundy"
                            : "bg-transparent text-dark-warm border-gold-muted/40 hover:border-burgundy hover:text-burgundy"
                        }`}
                      >
                        {val === "" ? "Todos" : val === "yes" ? "Confirmados" : "Pendentes"}
                      </button>
                    ))}
                  </div>
                </div>

                {activeFilterCount > 0 && (
                  <button
                    type="button"
                    onClick={() => { setFilterRel(""); setFilterConf(""); }}
                    className="w-full font-heading text-[0.65rem] font-semibold tracking-[0.08em] uppercase py-1.5 text-hint hover:text-burgundy transition-colors cursor-pointer border-t border-gold-muted/30 pt-3 mt-1"
                  >
                    Limpar filtros
                  </button>
                )}
              </div>
            )}
          </div>
        </div>

        {effectiveError && (
          <div className="mb-4 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
            <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
            <span className="text-[0.82rem] text-[#7a2e2b] flex-1">{effectiveError}</span>
            <button type="button" onClick={() => setUiError(null)} className="text-[#c25550]/60 hover:text-[#c25550] cursor-pointer">
              <X size={14} />
            </button>
          </div>
        )}

        {isLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="text-[0.85rem]">Carregando convidados...</span>
          </div>
        ) : guests.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-20 text-center">
            <Users size={48} className="text-gold-muted/40 mb-4" />
            <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">Nenhum convidado cadastrado</h2>
            <p className="text-[0.88rem] text-hint max-w-[360px] mb-6">
              Comece adicionando convidados individualmente ou importe uma lista via arquivo CSV.
            </p>
            <div className="flex flex-col sm:flex-row gap-3 justify-center">
              <button
                type="button"
                onClick={openImport}
                className="inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-transparent text-burgundy border border-burgundy transition-all duration-300 hover:bg-burgundy hover:text-gold-light cursor-pointer"
              >
                <Upload size={14} />
                Importar Lista
              </button>
              <Link
                to="/dashboard/novo"
                className="inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-burgundy text-gold-light border border-burgundy transition-all duration-300 hover:bg-burgundy-deep no-underline"
              >
                <Plus size={14} />
                Adicionar Primeiro Convidado
              </Link>
            </div>
          </div>
        ) : filtered.length === 0 ? (
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <Search size={36} className="text-gold-muted/40 mb-3" />
            <p className="text-[0.88rem] text-hint">Nenhum convidado encontrado com os filtros aplicados.</p>
          </div>
        ) : (
          <>
            <div className="flex items-center justify-between mb-2">
              <span className="text-[0.75rem] text-dark-warm/60">
                {filtered.length === guests.length
                  ? `${filtered.length} convidados`
                  : `${filtered.length} de ${guests.length} convidados`}
              </span>
              {selectedIds.size > 0 && (
                <div className="flex items-center gap-3">
                  <span className="text-[0.75rem] text-dark-warm/60">
                    {selectedIds.size} selecionado{selectedIds.size !== 1 ? "s" : ""}
                  </span>
                  <button
                    type="button"
                    onClick={() => setBulkDeleteOpen(true)}
                    className="inline-flex items-center gap-1.5 font-heading text-[0.68rem] font-semibold tracking-[0.06em] uppercase py-1 px-3 border border-[#c25550]/60 text-[#c25550] bg-transparent hover:bg-[#c25550] hover:text-white transition-all duration-200 cursor-pointer"
                  >
                    <Trash2 size={12} />
                    Excluir selecionados
                  </button>
                  <button
                    type="button"
                    onClick={() => setSelectedIds(new Set())}
                    className="text-[0.72rem] text-hint/70 hover:text-dark-warm transition-colors cursor-pointer"
                  >
                    Desmarcar
                  </button>
                </div>
              )}
            </div>

            <div className="overflow-x-auto rounded border border-gold-muted/50 shadow-[0_2px_8px_rgba(28,20,16,0.06)]">
              <table className="w-full min-w-[740px]">
                <thead>
                  <tr className="border-b-2 border-gold-muted/40 bg-dark/[0.04]">
                    <th className="py-3 px-3 w-10">
                      <input
                        ref={selectAllRef}
                        type="checkbox"
                        checked={allSelected}
                        onChange={toggleSelectAll}
                        className="w-4 h-4 cursor-pointer accent-[#989F5B]"
                      />
                    </th>
                    <th className="text-left font-heading text-[0.67rem] font-bold tracking-[0.12em] uppercase text-dark-warm/70 py-3 px-4">Nome</th>
                    <th className="text-left font-heading text-[0.67rem] font-bold tracking-[0.12em] uppercase text-dark-warm/70 py-3 px-4">Telefone</th>
                    <th className="text-center font-heading text-[0.67rem] font-bold tracking-[0.12em] uppercase text-dark-warm/70 py-3 px-4">Lado</th>
                    <th className="text-center font-heading text-[0.67rem] font-bold tracking-[0.12em] uppercase text-dark-warm/70 py-3 px-4">Grupo</th>
                    <th className="text-center font-heading text-[0.67rem] font-bold tracking-[0.12em] uppercase text-dark-warm/70 py-3 px-4">Status</th>
                    <th className="text-center font-heading text-[0.67rem] font-bold tracking-[0.12em] uppercase text-dark-warm/70 py-3 px-4 w-24">Ações</th>
                  </tr>
                </thead>
                <tbody className="bg-ivory">
                  {filtered.map((guest, idx) => (
                    <tr
                      key={guest.id}
                      className={`border-b border-gold-muted/25 transition-colors hover:bg-parchment ${
                        selectedIds.has(guest.id) ? "bg-burgundy/[0.04]" : idx % 2 === 1 ? "bg-parchment/50" : ""
                      }`}
                    >
                      <td className="py-3 px-3">
                        <input
                          type="checkbox"
                          checked={selectedIds.has(guest.id)}
                          onChange={() => toggleSelect(guest.id)}
                          className="w-4 h-4 cursor-pointer accent-[#989F5B]"
                        />
                      </td>
                      <td className="py-3 px-4">
                        <span className="text-[0.88rem] font-semibold text-dark">
                          {guest.first_name} {guest.last_name}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-[0.84rem] text-dark-warm/70 font-mono tracking-wide">
                        {formatPhone(guest.phone)}
                      </td>
                      <td className="py-3 px-4 text-center">
                        <span
                          className={`inline-block font-heading text-[0.63rem] font-bold tracking-[0.08em] uppercase py-[3px] px-2.5 ${
                            guest.relationship === "P"
                              ? "bg-burgundy/15 text-burgundy border border-burgundy/30"
                              : "bg-gold/20 text-gold-dark border border-gold/35"
                          }`}
                        >
                          {guest.relationship === "P" ? "Noivo" : "Noiva"}
                        </span>
                      </td>
                      <td className="py-3 px-4 text-center text-[0.84rem] font-semibold text-dark-warm/60">{guest.family_group}</td>
                      <td className="py-3 px-4 text-center">
                        {guest.confirmed ? (
                          <span className="inline-flex items-center gap-1.5 text-[0.72rem] font-bold text-burgundy bg-burgundy/10 border border-burgundy/20 px-2.5 py-1">
                            <Check size={11} />
                            Confirmado
                          </span>
                        ) : (
                          <span className="inline-flex items-center gap-1.5 text-[0.72rem] font-bold text-gold-dark bg-gold/10 border border-gold/25 px-2.5 py-1">
                            <Clock size={11} />
                            Pendente
                          </span>
                        )}
                      </td>
                      <td className="py-3 px-4">
                        <div className="flex items-center justify-center gap-1">
                          <Link
                            to="/dashboard/$guestId"
                            params={{ guestId: String(guest.id) }}
                            className="p-1.5 text-dark-warm/40 hover:text-burgundy transition-colors"
                            title="Editar"
                          >
                            <Pencil size={15} />
                          </Link>
                          <button
                            type="button"
                            onClick={() => setDeleteTarget(guest)}
                            className="p-1.5 text-dark-warm/40 hover:text-[#c25550] transition-colors cursor-pointer"
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
      </>
      )}
      </main>

      {/* Single delete modal */}
      {deleteTarget && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-dark/50 backdrop-blur-[2px]" onClick={() => !deleteMutation.isPending && setDeleteTarget(null)} />
          <div className="relative bg-ivory border border-gold-muted/50 shadow-xl max-w-[400px] w-full p-6 anim-fade-in-up">
            <div className="flex flex-col items-center text-center">
              <div className="w-12 h-12 rounded-full bg-[#fef2f1] border border-[#c25550]/15 flex items-center justify-center mb-4">
                <AlertTriangle size={22} className="text-[#c25550]" />
              </div>
              <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">Excluir Convidado</h2>
              <p className="text-[0.88rem] text-hint mb-6">
                Tem certeza que deseja excluir <strong className="text-dark-warm">{deleteTarget.first_name} {deleteTarget.last_name}</strong>? Esta ação não pode ser desfeita.
              </p>
              <div className="flex gap-3 w-full">
                <button type="button" onClick={() => setDeleteTarget(null)} disabled={deleteMutation.isPending}
                  className="flex-1 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-4 border border-gold-muted/60 text-hint bg-transparent hover:border-burgundy hover:text-burgundy cursor-pointer disabled:opacity-50 transition-colors">
                  Cancelar
                </button>
                <button type="button" onClick={() => void handleDelete()} disabled={deleteMutation.isPending}
                  className="flex-1 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-4 border border-[#c25550] bg-[#c25550] text-white hover:bg-[#a83f3b] cursor-pointer disabled:opacity-50 transition-colors">
                  {deleteMutation.isPending ? "Excluindo..." : "Excluir"}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Bulk delete modal */}
      {bulkDeleteOpen && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-dark/50 backdrop-blur-[2px]" onClick={() => !deletesMutation.isPending && setBulkDeleteOpen(false)} />
          <div className="relative bg-ivory border border-gold-muted/50 shadow-xl max-w-[400px] w-full p-6 anim-fade-in-up">
            <div className="flex flex-col items-center text-center">
              <div className="w-12 h-12 rounded-full bg-[#fef2f1] border border-[#c25550]/15 flex items-center justify-center mb-4">
                <AlertTriangle size={22} className="text-[#c25550]" />
              </div>
              <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">Excluir Convidados</h2>
              <p className="text-[0.88rem] text-hint mb-6">
                Tem certeza que deseja excluir <strong className="text-dark-warm">{selectedIds.size} convidado{selectedIds.size !== 1 ? "s" : ""}</strong>? Esta ação não pode ser desfeita.
              </p>
              <div className="flex gap-3 w-full">
                <button type="button" onClick={() => setBulkDeleteOpen(false)} disabled={deletesMutation.isPending}
                  className="flex-1 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-4 border border-gold-muted/60 text-hint bg-transparent hover:border-burgundy hover:text-burgundy cursor-pointer disabled:opacity-50 transition-colors">
                  Cancelar
                </button>
                <button type="button" onClick={() => void handleBulkDelete()} disabled={deletesMutation.isPending}
                  className="flex-1 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-4 border border-[#c25550] bg-[#c25550] text-white hover:bg-[#a83f3b] cursor-pointer disabled:opacity-50 transition-colors">
                  {deletesMutation.isPending ? "Excluindo..." : "Excluir"}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Import modal */}
      {importOpen && (
        <div className="fixed inset-0 z-[200] flex items-center justify-center p-4">
          <div className="absolute inset-0 bg-dark/50 backdrop-blur-[2px]" onClick={() => !importMutation.isPending && closeImportModal()} />
          <div className="relative bg-ivory border border-gold-muted/50 shadow-xl max-w-[480px] w-full p-6 anim-fade-in-up">
            <button type="button" onClick={closeImportModal} disabled={importMutation.isPending}
              className="absolute top-4 right-4 text-hint/60 hover:text-dark-warm transition-colors cursor-pointer">
              <X size={18} />
            </button>
            <div className="flex items-center gap-3 mb-5">
              <div className="w-10 h-10 bg-burgundy/10 border border-burgundy/15 flex items-center justify-center">
                <FileSpreadsheet size={20} className="text-burgundy" />
              </div>
              <div>
                <h2 className="font-heading text-[1rem] font-semibold text-dark-warm">Importar Convidados</h2>
                <p className="text-[0.75rem] text-hint">Formatos aceitos: CSV ou XLSX</p>
              </div>
            </div>
            <div className="mb-5">
              <p className="text-[0.8rem] text-dark-warm/70 mb-3">
                O arquivo deve conter as colunas: <code className="text-[0.75rem] bg-parchment-dark/60 px-1.5 py-0.5 font-mono">first_name, last_name, phone, relationship, family_group</code>
              </p>
              <label className="flex flex-col items-center justify-center gap-2 py-8 border-2 border-dashed border-gold-muted/40 bg-parchment/50 cursor-pointer hover:border-burgundy/40 hover:bg-parchment transition-colors">
                <Upload size={24} className="text-hint/50" />
                <span className="text-[0.82rem] text-hint">
                  {importMutation.isPending ? "Importando..." : "Clique para selecionar um arquivo"}
                </span>
                <input ref={fileInputRef} type="file" accept=".csv,.xlsx"
                  onChange={(event) => void handleImportFile(event)}
                  disabled={importMutation.isPending} className="hidden" />
              </label>
            </div>
            {importMutation.isPending && (
              <div className="flex items-center justify-center py-4">
                <div className="w-6 h-6 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin" />
              </div>
            )}
            {importResult && (
              <div className="border border-gold-muted/40 bg-parchment p-4">
                <div className="flex flex-wrap gap-4 mb-2">
                  <span className="text-[0.82rem]"><strong className="text-burgundy">{importResult.imported}</strong> importados</span>
                  <span className="text-[0.82rem]"><strong className="text-dark-warm/60">{importResult.total}</strong> no arquivo</span>
                  {importResult.errors.length > 0 && (
                    <span className="text-[0.82rem]"><strong className="text-[#c25550]">{importResult.errors.length}</strong> erros</span>
                  )}
                </div>
                {importResult.errors.length > 0 && (
                  <div className="mt-2 max-h-32 overflow-y-auto">
                    {importResult.errors.map((importError, index) => (
                      <p key={index} className="text-[0.75rem] text-[#c25550] leading-relaxed">{importError}</p>
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
