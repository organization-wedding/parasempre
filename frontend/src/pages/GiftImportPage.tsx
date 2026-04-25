import { useMemo, useRef, useState } from "react";
import { Link, useNavigate } from "@tanstack/react-router";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";
import Upload from "lucide-react/dist/esm/icons/upload";
import Download from "lucide-react/dist/esm/icons/download";
import FileSpreadsheet from "lucide-react/dist/esm/icons/file-spreadsheet";
import CheckCircle2 from "lucide-react/dist/esm/icons/check-circle-2";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import X from "lucide-react/dist/esm/icons/x";
import Gift from "lucide-react/dist/esm/icons/gift";
import { Header } from "../components/Header";
import {
  useCommitGiftImportMutation,
  usePreviewGiftImportMutation,
} from "../lib/gift-queries";
import { useUserMeQuery } from "../lib/user-queries";
import { UnauthorizedPage } from "./UnauthorizedPage";
import { formatBRL } from "../lib/format";
import type {
  CSVPreview,
  CSVPreviewRow,
  CommitImportResponse,
  CreateGiftInput,
} from "../types/gift";

const MAX_FILE_SIZE = 5 * 1024 * 1024; // 5 MB

const SAMPLE_CSV_HEADER = ["name", "description", "price_brl", "image_url", "store_url"];
const SAMPLE_CSV_DELIMITER = ";";

function downloadSampleCSV() {
  const content = SAMPLE_CSV_HEADER.join(SAMPLE_CSV_DELIMITER) + "\r\n";
  const blob = new Blob(["﻿" + content], { type: "text/csv;charset=utf-8" });
  const url = URL.createObjectURL(blob);
  const link = document.createElement("a");
  link.href = url;
  link.download = "modelo-presentes.csv";
  document.body.appendChild(link);
  link.click();
  document.body.removeChild(link);
  URL.revokeObjectURL(url);
}

function rowToInput(row: CSVPreviewRow): CreateGiftInput {
  const i = row.input;
  return {
    name: i.name,
    price_cents: i.price_cents,
    description: i.description ?? undefined,
    image_url: i.image_url ?? undefined,
    store_url: i.store_url ?? undefined,
  };
}

export function GiftImportPage() {
  const navigate = useNavigate();
  const { data: userMe, isLoading: roleLoading } = useUserMeQuery();
  const isAuthorized = userMe?.role === "groom" || userMe?.role === "bride";

  const previewMutation = usePreviewGiftImportMutation();
  const commitMutation = useCommitGiftImportMutation();

  const [preview, setPreview] = useState<CSVPreview | null>(null);
  const [selectedLines, setSelectedLines] = useState<Set<number>>(new Set());
  const [uiError, setUiError] = useState<string | null>(null);
  const [commitResult, setCommitResult] = useState<CommitImportResponse | null>(null);
  const fileInputRef = useRef<HTMLInputElement>(null);

  async function handleFile(event: React.ChangeEvent<HTMLInputElement>) {
    const file = event.target.files?.[0];
    if (!file) return;

    setUiError(null);
    setPreview(null);
    setCommitResult(null);

    if (file.size > MAX_FILE_SIZE) {
      setUiError("Arquivo maior que 5 MB — divida em partes menores.");
      if (fileInputRef.current) fileInputRef.current.value = "";
      return;
    }

    try {
      const result = await previewMutation.mutateAsync(file);
      setPreview(result);
      const newLines = result.rows
        .filter((r) => r.status === "new")
        .map((r) => r.line_number);
      setSelectedLines(new Set(newLines));
    } catch (err) {
      setUiError(err instanceof Error ? err.message : "Erro ao processar CSV");
    } finally {
      if (fileInputRef.current) fileInputRef.current.value = "";
    }
  }

  function toggleLine(lineNumber: number) {
    setSelectedLines((prev) => {
      const next = new Set(prev);
      if (next.has(lineNumber)) next.delete(lineNumber);
      else next.add(lineNumber);
      return next;
    });
  }

  function toggleAllNew() {
    if (!preview) return;
    const newLines = preview.rows.filter((r) => r.status === "new").map((r) => r.line_number);
    const allSelected = newLines.every((ln) => selectedLines.has(ln));
    if (allSelected) {
      setSelectedLines(new Set());
    } else {
      setSelectedLines(new Set(newLines));
    }
  }

  async function handleCommit() {
    if (!preview) return;
    const rowsToCommit = preview.rows
      .filter((r) => r.status === "new" && selectedLines.has(r.line_number))
      .map(rowToInput);

    if (rowsToCommit.length === 0) {
      setUiError("Selecione ao menos um presente para importar.");
      return;
    }

    setUiError(null);
    try {
      const result = await commitMutation.mutateAsync(rowsToCommit);
      setCommitResult(result);
    } catch (err) {
      setUiError(err instanceof Error ? err.message : "Erro ao importar presentes");
    }
  }

  function resetImport() {
    setPreview(null);
    setSelectedLines(new Set());
    setUiError(null);
    setCommitResult(null);
    previewMutation.reset();
    commitMutation.reset();
  }

  const summary = preview?.summary;
  const hasSelectable = useMemo(
    () => (preview?.rows ?? []).some((r) => r.status === "new"),
    [preview],
  );
  const selectedCount = selectedLines.size;

  if (!roleLoading && userMe && !isAuthorized) {
    return <UnauthorizedPage />;
  }

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />

      <main className="mx-auto max-w-[1100px] px-6 pt-24 pb-16">
        <Link
          to="/dashboard/presentes"
          className="inline-flex items-center gap-1.5 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase text-hint no-underline mb-6 transition-colors hover:text-burgundy"
        >
          <ArrowLeft size={15} />
          Voltar para lista
        </Link>

        <h1 className="font-display text-[1.5rem] md:text-[1.8rem] font-bold text-dark mb-8">
          Importar Presentes via CSV
        </h1>

        {uiError && (
          <div className="mb-6 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
            <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
            <span className="text-[0.82rem] text-[#7a2e2b] flex-1">{uiError}</span>
            <button
              type="button"
              onClick={() => setUiError(null)}
              className="text-[#c25550]/60 hover:text-[#c25550] cursor-pointer"
            >
              <X size={14} />
            </button>
          </div>
        )}

        {commitResult ? (
          <ResultCard
            result={commitResult}
            onDone={() => void navigate({ to: "/dashboard/presentes" })}
            onImportAnother={resetImport}
          />
        ) : preview ? (
          <>
            <SummaryCard summary={summary!} />
            <PreviewTable
              rows={preview.rows}
              selectedLines={selectedLines}
              onToggleLine={toggleLine}
              onToggleAllNew={toggleAllNew}
              hasSelectable={hasSelectable}
            />

            <div className="mt-6 flex flex-col-reverse sm:flex-row gap-3 sm:justify-end">
              <button
                type="button"
                onClick={resetImport}
                disabled={commitMutation.isPending}
                className="inline-flex items-center justify-center font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 border border-gold-muted/50 text-hint bg-transparent transition-all duration-200 hover:border-burgundy hover:text-burgundy cursor-pointer disabled:opacity-50"
              >
                Cancelar
              </button>
              <button
                type="button"
                onClick={() => void handleCommit()}
                disabled={commitMutation.isPending || selectedCount === 0}
                className="inline-flex items-center justify-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-burgundy text-gold-light border border-burgundy transition-all duration-200 hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <CheckCircle2 size={14} />
                {commitMutation.isPending
                  ? "Importando..."
                  : `Confirmar importação (${selectedCount})`}
              </button>
            </div>
          </>
        ) : (
          <UploadCard
            fileInputRef={fileInputRef}
            onFile={handleFile}
            loading={previewMutation.isPending}
          />
        )}
      </main>
    </div>
  );
}

function UploadCard({
  fileInputRef,
  onFile,
  loading,
}: {
  fileInputRef: React.RefObject<HTMLInputElement | null>;
  onFile: (e: React.ChangeEvent<HTMLInputElement>) => void;
  loading: boolean;
}) {
  return (
    <div className="border border-gold-muted/40 bg-ivory p-8 md:p-10 rounded">
      <div className="flex flex-col items-center text-center">
        <FileSpreadsheet size={42} className="text-gold-muted mb-4" />
        <h2 className="font-heading text-[1rem] font-semibold text-dark-warm mb-2">
          Selecione o arquivo CSV
        </h2>
        <p className="text-[0.88rem] text-hint max-w-[480px] mb-6 leading-relaxed">
          O arquivo deve ter cabeçalho. Colunas obrigatórias: <code className="text-dark-warm">name</code>, <code className="text-dark-warm">price_brl</code>. Opcionais: <code className="text-dark-warm">description</code>, <code className="text-dark-warm">image_url</code>, <code className="text-dark-warm">store_url</code>. O preço pode usar <code className="text-dark-warm">150.00</code> ou <code className="text-dark-warm">150,00</code>.
        </p>
        <div className="flex flex-col sm:flex-row gap-3 items-center">
          <button
            type="button"
            onClick={downloadSampleCSV}
            disabled={loading}
            className="inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 border border-burgundy text-burgundy bg-transparent transition-all duration-200 hover:bg-burgundy hover:text-gold-light cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed"
          >
            <Download size={14} />
            Baixar modelo
          </button>
          <label
            className={`inline-flex items-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-burgundy text-gold-light border border-burgundy transition-all duration-200 hover:bg-burgundy-deep cursor-pointer ${loading ? "opacity-50 cursor-not-allowed" : ""
              }`}
          >
            <Upload size={14} />
            {loading ? "Processando..." : "Escolher arquivo"}
            <input
              ref={fileInputRef}
              type="file"
              accept=".csv"
              className="hidden"
              onChange={(e) => void onFile(e)}
              disabled={loading}
            />
          </label>
        </div>
      </div>
    </div>
  );
}

function SummaryCard({
  summary,
}: {
  summary: NonNullable<CSVPreview["summary"]>;
}) {
  return (
    <div className="mb-6 flex flex-wrap gap-6 border border-gold-muted/40 bg-ivory px-6 py-4 rounded">
      <SummaryStat label="Total" value={summary.total} tone="neutral" />
      <SummaryStat label="Novos" value={summary.new} tone="success" />
      <SummaryStat label="Duplicados" value={summary.duplicate} tone="warning" />
      <SummaryStat label="Inválidos" value={summary.invalid} tone="error" />
    </div>
  );
}

function SummaryStat({
  label,
  value,
  tone,
}: {
  label: string;
  value: number;
  tone: "neutral" | "success" | "warning" | "error";
}) {
  const color = {
    neutral: "text-dark-warm",
    success: "text-burgundy",
    warning: "text-gold-dark",
    error: "text-[#c25550]",
  }[tone];
  return (
    <div className="flex flex-col gap-0.5">
      <span className="font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
        {label}
      </span>
      <span className={`font-display text-[1.4rem] font-bold ${color}`}>{value}</span>
    </div>
  );
}

function PreviewTable({
  rows,
  selectedLines,
  onToggleLine,
  onToggleAllNew,
  hasSelectable,
}: {
  rows: CSVPreviewRow[];
  selectedLines: Set<number>;
  onToggleLine: (lineNumber: number) => void;
  onToggleAllNew: () => void;
  hasSelectable: boolean;
}) {
  const newRowsCount = rows.filter((r) => r.status === "new").length;
  const allSelected = newRowsCount > 0 && newRowsCount === selectedLines.size;

  return (
    <div className="overflow-x-auto rounded border border-gold-muted/50 shadow-[0_2px_8px_rgba(28,20,16,0.06)]">
      <table className="w-full min-w-[720px]">
        <thead>
          <tr className="border-b-2 border-gold-muted/40 bg-dark/[0.04]">
            <th className="py-3 px-3 w-10">
              {hasSelectable && (
                <input
                  type="checkbox"
                  checked={allSelected}
                  onChange={onToggleAllNew}
                  aria-label="Selecionar todos"
                  className="cursor-pointer"
                />
              )}
            </th>
            <th className="py-3 px-3 w-16 text-left font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
              Linha
            </th>
            <th className="py-3 px-3 w-24 text-left font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
              Status
            </th>
            <th className="py-3 px-3 text-left font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
              Nome
            </th>
            <th className="py-3 px-3 w-28 text-left font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
              Preço
            </th>
            <th className="py-3 px-3 text-left font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint">
              Observações
            </th>
          </tr>
        </thead>
        <tbody>
          {rows.map((row) => {
            const canSelect = row.status === "new";
            const isSelected = selectedLines.has(row.line_number);
            return (
              <tr
                key={row.line_number}
                className={`border-b border-gold-muted/20 last:border-b-0 ${canSelect && isSelected ? "bg-burgundy/[0.04]" : "bg-ivory"
                  }`}
              >
                <td className="py-2 px-3">
                  <input
                    type="checkbox"
                    checked={canSelect && isSelected}
                    disabled={!canSelect}
                    onChange={() => onToggleLine(row.line_number)}
                    aria-label={`Linha ${row.line_number}`}
                    className="cursor-pointer disabled:cursor-not-allowed disabled:opacity-30"
                  />
                </td>
                <td className="py-2 px-3 text-[0.82rem] text-hint">{row.line_number}</td>
                <td className="py-2 px-3">
                  <StatusBadge status={row.status} />
                </td>
                <td className="py-2 px-3 text-[0.88rem] text-dark-warm">
                  {row.input.name || <span className="text-hint/60">—</span>}
                </td>
                <td className="py-2 px-3 text-[0.88rem] text-dark-warm">
                  {row.input.price_cents > 0 ? (
                    formatBRL(row.input.price_cents)
                  ) : (
                    <span className="text-hint/60">—</span>
                  )}
                </td>
                <td className="py-2 px-3 text-[0.78rem] text-hint">
                  {row.errors && row.errors.length > 0 ? (
                    <span className="text-[#7a2e2b]">{row.errors.join("; ")}</span>
                  ) : row.status === "duplicate" ? (
                    <span className="text-gold-dark">Já existe na lista</span>
                  ) : (
                    ""
                  )}
                </td>
              </tr>
            );
          })}
        </tbody>
      </table>
    </div>
  );
}

function StatusBadge({ status }: { status: "new" | "duplicate" | "invalid" }) {
  const config = {
    new: { label: "Novo", classes: "bg-burgundy/10 text-burgundy border-burgundy/30" },
    duplicate: {
      label: "Duplicado",
      classes: "bg-gold/10 text-gold-dark border-gold/40",
    },
    invalid: {
      label: "Inválido",
      classes: "bg-[#c25550]/10 text-[#7a2e2b] border-[#c25550]/30",
    },
  }[status];
  return (
    <span
      className={`inline-flex items-center font-heading text-[0.65rem] font-semibold tracking-[0.08em] uppercase px-2 py-0.5 border ${config.classes}`}
    >
      {config.label}
    </span>
  );
}

function ResultCard({
  result,
  onDone,
  onImportAnother,
}: {
  result: CommitImportResponse;
  onDone: () => void;
  onImportAnother: () => void;
}) {
  const hasCreated = result.created > 0;
  const skipped = result.skipped ?? [];

  return (
    <div className="border border-gold-muted/40 bg-ivory p-8 md:p-10 rounded">
      <div className="flex flex-col items-center text-center">
        {hasCreated ? (
          <CheckCircle2 size={42} className="text-burgundy mb-4" />
        ) : (
          <Gift size={42} className="text-gold-muted mb-4" />
        )}
        <h2 className="font-display text-[1.3rem] font-bold text-dark mb-2">
          {hasCreated
            ? `${result.created} ${result.created === 1 ? "presente importado" : "presentes importados"}`
            : "Nada foi importado"}
        </h2>
        {skipped.length > 0 && (
          <div className="w-full mt-4 border border-gold/30 bg-gold/[0.06] rounded px-4 py-3 text-left">
            <p className="font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-gold-dark mb-2">
              {skipped.length} {skipped.length === 1 ? "linha ignorada" : "linhas ignoradas"}
            </p>
            <ul className="text-[0.82rem] text-dark-warm/80 space-y-1 list-disc pl-4">
              {skipped.map((msg, idx) => (
                <li key={idx}>{msg}</li>
              ))}
            </ul>
          </div>
        )}

        <div className="mt-8 flex flex-col sm:flex-row gap-3">
          <button
            type="button"
            onClick={onImportAnother}
            className="inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 border border-burgundy text-burgundy hover:bg-burgundy hover:text-gold-light transition-all duration-200 cursor-pointer"
          >
            Importar outro CSV
          </button>
          <button
            type="button"
            onClick={onDone}
            className="inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep transition-all duration-200 cursor-pointer"
          >
            Voltar para lista
          </button>
        </div>
      </div>
    </div>
  );
}
