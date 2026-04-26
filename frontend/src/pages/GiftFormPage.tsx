import { useEffect, useState } from "react";
import { Link, useNavigate } from "@tanstack/react-router";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";
import Save from "lucide-react/dist/esm/icons/save";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import Sparkles from "lucide-react/dist/esm/icons/sparkles";
import X from "lucide-react/dist/esm/icons/x";
import { Header } from "../components/Header";
import {
  useCreateGiftMutation,
  useGiftQuery,
  useScrapeGiftURLMutation,
  useUpdateGiftMutation,
} from "../lib/gift-queries";
import { useUserMeQuery } from "../lib/user-queries";
import { UnauthorizedPage } from "./UnauthorizedPage";
import type { CreateGiftInput, UpdateGiftInput } from "../types/gift";

interface Props {
  giftId?: number;
}

const giftFormSchema = z.object({
  name: z.string().trim().min(1, "Nome é obrigatório.").max(200, "Nome muito longo."),
  description: z.string().trim().max(2000, "Descrição muito longa."),
  price: z
    .string()
    .trim()
    .min(1, "Preço é obrigatório.")
    .refine((v) => {
      const n = Number(v.replace(",", "."));
      return Number.isFinite(n) && n > 0;
    }, "Preço deve ser maior que zero."),
  image_url: z
    .string()
    .trim()
    .refine((v) => v === "" || /^https:\/\/[^\s]+$/.test(v), "URL inválida — use HTTPS."),
  store_url: z
    .string()
    .trim()
    .refine((v) => v === "" || /^https:\/\/[^\s]+$/.test(v), "URL inválida — use HTTPS."),
});

type GiftFormValues = z.infer<typeof giftFormSchema>;

function priceToCents(value: string): number {
  const normalized = value.replace(",", ".").replace(/[^\d.]/g, "");
  const n = Number(normalized);
  return Math.round(n * 100);
}

function centsToPriceInput(cents: number): string {
  return (cents / 100).toFixed(2);
}

export function GiftFormPage({ giftId }: Props) {
  const isEdit = giftId !== undefined;
  const navigate = useNavigate();
  const createMutation = useCreateGiftMutation();
  const updateMutation = useUpdateGiftMutation();
  const scrapeMutation = useScrapeGiftURLMutation();
  const giftQuery = useGiftQuery(giftId ?? 0, isEdit);
  const { data: userMe, isLoading: roleLoading } = useUserMeQuery();
  const isAuthorized = userMe?.role === "groom" || userMe?.role === "bride";

  const [scrapeURL, setScrapeURL] = useState("");
  const [scrapeError, setScrapeError] = useState<string | null>(null);
  const [scrapeSuccess, setScrapeSuccess] = useState(false);

  const {
    register,
    handleSubmit,
    reset,
    setError,
    setValue,
    clearErrors,
    formState: { errors, isSubmitting },
  } = useForm<GiftFormValues>({
    resolver: zodResolver(giftFormSchema),
    defaultValues: {
      name: "",
      description: "",
      price: "",
      image_url: "",
      store_url: "",
    },
  });

  useEffect(() => {
    if (!giftQuery.data) return;
    reset({
      name: giftQuery.data.name,
      description: giftQuery.data.description ?? "",
      price: centsToPriceInput(giftQuery.data.price_cents),
      image_url: giftQuery.data.image_url ?? "",
      store_url: giftQuery.data.store_url ?? "",
    });
  }, [giftQuery.data, reset]);

  const mutationError =
    (createMutation.error instanceof Error && createMutation.error.message) ||
    (updateMutation.error instanceof Error && updateMutation.error.message) ||
    (giftQuery.error instanceof Error && giftQuery.error.message) ||
    null;

  async function onSubmit(values: GiftFormValues) {
    clearErrors("root");

    const payload = {
      name: values.name.trim(),
      description: values.description.trim() || undefined,
      price_cents: priceToCents(values.price),
      image_url: values.image_url.trim() || undefined,
      store_url: values.store_url.trim() || undefined,
    };

    try {
      if (isEdit && giftId !== undefined) {
        await updateMutation.mutateAsync({
          id: giftId,
          input: payload as UpdateGiftInput,
        });
      } else {
        await createMutation.mutateAsync(payload as CreateGiftInput);
      }
      await navigate({ to: "/dashboard/presentes" });
    } catch (submitError) {
      setError("root", {
        message: submitError instanceof Error ? submitError.message : "Erro ao salvar",
      });
    }
  }

  async function handleScrape() {
    setScrapeError(null);
    setScrapeSuccess(false);
    const url = scrapeURL.trim();
    if (!/^https:\/\/[^\s]+$/.test(url)) {
      setScrapeError("Cole uma URL https:// válida da loja.");
      return;
    }
    try {
      const data = await scrapeMutation.mutateAsync(url);
      if (data.name) setValue("name", data.name, { shouldDirty: true });
      if (data.description) setValue("description", data.description, { shouldDirty: true });
      if (data.price_cents > 0) setValue("price", centsToPriceInput(data.price_cents), { shouldDirty: true });
      if (data.image_url) setValue("image_url", data.image_url, { shouldDirty: true });
      setValue("store_url", data.store_url, { shouldDirty: true });
      setScrapeSuccess(true);
    } catch (err) {
      setScrapeError(err instanceof Error ? err.message : "Erro ao buscar dados do produto.");
    }
  }

  if (!roleLoading && userMe && !isAuthorized) {
    return <UnauthorizedPage />;
  }

  const inputClass =
    "w-full px-3.5 py-2.5 text-[0.88rem] border border-gold-muted/40 bg-ivory text-dark-warm placeholder:text-hint/40 outline-none focus:border-burgundy transition-colors";
  const labelClass =
    "block font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase text-hint mb-1.5";
  const loading = isEdit && giftQuery.isLoading;
  const saving = isSubmitting || createMutation.isPending || updateMutation.isPending;
  const errorMessage = errors.root?.message ?? mutationError;

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />

      <main className="mx-auto max-w-[640px] px-6 pt-24 pb-16">
        <Link
          to="/dashboard/presentes"
          className="inline-flex items-center gap-1.5 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase text-hint no-underline mb-6 transition-colors hover:text-burgundy"
        >
          <ArrowLeft size={15} />
          Voltar para lista
        </Link>

        <h1 className="font-display text-[1.4rem] md:text-[1.7rem] font-bold text-dark mb-8">
          {isEdit ? "Editar Presente" : "Novo Presente"}
        </h1>

        {errorMessage && (
          <div className="mb-6 flex items-center gap-3 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3">
            <AlertTriangle size={16} className="text-[#c25550] shrink-0" />
            <span className="text-[0.82rem] text-[#7a2e2b] flex-1">{errorMessage}</span>
            <button
              type="button"
              onClick={() => clearErrors("root")}
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
            onSubmit={handleSubmit(onSubmit)}
            className="border border-gold-muted/40 bg-ivory p-6 md:p-8 rounded"
          >
            {!isEdit && (
              <div className="mb-6 pb-6 border-b border-gold-muted/20">
                <div className="flex items-center gap-2 mb-3">
                  <Sparkles size={14} className="text-burgundy" />
                  <h2 className="font-heading text-[0.78rem] font-semibold tracking-[0.08em] uppercase text-burgundy">
                    Buscar por link da loja
                  </h2>
                </div>
                <p className="text-[0.78rem] text-hint mb-3">
                  Cole o link de um produto e tentaremos preencher os campos automaticamente.
                </p>
                <div className="flex flex-col sm:flex-row gap-2.5">
                  <input
                    type="url"
                    placeholder="https://loja.com/produto"
                    value={scrapeURL}
                    onChange={(e) => setScrapeURL(e.target.value)}
                    disabled={scrapeMutation.isPending}
                    className={`${inputClass} flex-1`}
                  />
                  <button
                    type="button"
                    onClick={handleScrape}
                    disabled={scrapeMutation.isPending || scrapeURL.trim() === ""}
                    className="inline-flex items-center justify-center gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.1rem] bg-burgundy text-gold-light border border-burgundy transition-all duration-200 hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] cursor-pointer disabled:opacity-50 disabled:cursor-not-allowed whitespace-nowrap"
                  >
                    {scrapeMutation.isPending ? (
                      <>
                        <div className="w-3.5 h-3.5 border-[1.5px] border-gold-light/30 border-t-gold-light rounded-full animate-spin" />
                        Buscando...
                      </>
                    ) : (
                      <>
                        <Sparkles size={13} />
                        Buscar
                      </>
                    )}
                  </button>
                </div>
                {scrapeError && (
                  <div className="mt-3 flex items-start gap-2 rounded border border-[#c25550]/30 bg-[#fef2f1] px-3 py-2">
                    <AlertTriangle size={14} className="text-[#c25550] shrink-0 mt-0.5" />
                    <span className="text-[0.78rem] text-[#7a2e2b] flex-1">{scrapeError}</span>
                    <button
                      type="button"
                      onClick={() => setScrapeError(null)}
                      className="text-[#c25550]/60 hover:text-[#c25550] cursor-pointer"
                    >
                      <X size={12} />
                    </button>
                  </div>
                )}
                {scrapeSuccess && !scrapeError && (
                  <p className="mt-3 text-[0.78rem] text-burgundy-deep">
                    Dados preenchidos abaixo — revise e ajuste antes de salvar.
                  </p>
                )}
              </div>
            )}

            <div className="mb-5">
              <label className={labelClass}>Nome</label>
              <input
                type="text"
                placeholder="Ex: Jogo de panelas"
                className={inputClass}
                {...register("name")}
              />
              {errors.name?.message && (
                <p className="text-[0.72rem] text-[#c25550] mt-1">{errors.name.message}</p>
              )}
            </div>

            <div className="mb-5">
              <label className={labelClass}>Descrição</label>
              <textarea
                rows={3}
                placeholder="Opcional — detalhes do presente"
                className={`${inputClass} resize-y`}
                {...register("description")}
              />
              {errors.description?.message && (
                <p className="text-[0.72rem] text-[#c25550] mt-1">{errors.description.message}</p>
              )}
            </div>

            <div className="mb-5">
              <label className={labelClass}>Preço (R$)</label>
              <input
                type="text"
                inputMode="decimal"
                placeholder="150.00"
                className={inputClass}
                {...register("price")}
              />
              <p className="text-[0.72rem] text-hint/60 mt-1">Use ponto ou vírgula para decimais. Ex: 150.00 ou 150,00</p>
              {errors.price?.message && (
                <p className="text-[0.72rem] text-[#c25550] mt-1">{errors.price.message}</p>
              )}
            </div>

            <div className="mb-5">
              <label className={labelClass}>URL da imagem</label>
              <input
                type="url"
                placeholder="https://exemplo.com/imagem.jpg"
                className={inputClass}
                {...register("image_url")}
              />
              <p className="text-[0.72rem] text-hint/60 mt-1">Opcional — deve começar com https://</p>
              {errors.image_url?.message && (
                <p className="text-[0.72rem] text-[#c25550] mt-1">{errors.image_url.message}</p>
              )}
            </div>

            <div className="mb-6">
              <label className={labelClass}>URL da loja</label>
              <input
                type="url"
                placeholder="https://loja.com/produto"
                className={inputClass}
                {...register("store_url")}
              />
              <p className="text-[0.72rem] text-hint/60 mt-1">Opcional — link de referência para o produto</p>
              {errors.store_url?.message && (
                <p className="text-[0.72rem] text-[#c25550] mt-1">{errors.store_url.message}</p>
              )}
            </div>

            <div className="flex flex-col-reverse sm:flex-row gap-3 sm:justify-end pt-4 border-t border-gold-muted/20">
              <Link
                to="/dashboard/presentes"
                className="inline-flex items-center justify-center font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.6rem] px-5 border border-gold-muted/50 text-hint bg-transparent transition-all duration-200 hover:border-burgundy hover:text-burgundy no-underline cursor-pointer"
              >
                Cancelar
              </Link>
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
