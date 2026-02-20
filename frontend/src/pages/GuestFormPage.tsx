import { useEffect, useMemo, useRef, useState } from "react";
import { Link, useNavigate } from "@tanstack/react-router";
import { z } from "zod";
import { Controller, useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";
import Save from "lucide-react/dist/esm/icons/save";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import X from "lucide-react/dist/esm/icons/x";
import { Header } from "../components/Header";
import { getUserRacf } from "../lib/api";
import {
  useCreateGuestMutation,
  useGuestQuery,
  useGuestsQuery,
  useUpdateGuestMutation,
} from "../lib/guest-queries";
import { useUserMeQuery } from "../lib/user-queries";
import { UnauthorizedPage } from "./UnauthorizedPage";
import type { Guest } from "../types/guest";

interface Props {
  guestId?: number;
}

const guestFormSchema = z.object({
  firstName: z.string().trim().min(1, "Nome é obrigatório."),
  lastName: z.string().trim().min(1, "Sobrenome é obrigatório."),
  phone: z
    .string()
    .transform((value) => value.replace(/\D/g, ""))
    .refine((value) => value.length === 0 || value.length === 11, {
      message: "Telefone deve ter 11 dígitos.",
    }),
  relationship: z.enum(["P", "R"]),
  familyGroup: z.number().int().min(1).optional(),
  confirmed: z.boolean(),
});

type GuestFormValues = z.infer<typeof guestFormSchema>;

function formatPhoneDisplay(digits: string): string {
  if (digits.length === 0) return "";
  if (digits.length <= 2) return digits;
  if (digits.length <= 7) return `(${digits.slice(0, 2)}) ${digits.slice(2)}`;
  return `(${digits.slice(0, 2)}) ${digits.slice(2, 7)}-${digits.slice(7)}`;
}

export function GuestFormPage({ guestId }: Props) {
  const isEdit = guestId !== undefined;
  const navigate = useNavigate();
  const createMutation = useCreateGuestMutation();
  const updateMutation = useUpdateGuestMutation();
  const guestQuery = useGuestQuery(guestId ?? 0, isEdit);
  const { data: allGuests = [] } = useGuestsQuery();
  const { data: userMe, isLoading: roleLoading } = useUserMeQuery();
  const isAuthorized = userMe?.role === "groom" || userMe?.role === "bride";

  // Family group autocomplete state
  const [familySearch, setFamilySearch] = useState("");
  const [assignedFamilyGroup, setAssignedFamilyGroup] = useState<number | undefined>(undefined);
  const [familyGroupLabel, setFamilyGroupLabel] = useState("");
  const [showFamilyDropdown, setShowFamilyDropdown] = useState(false);
  const familyContainerRef = useRef<HTMLDivElement>(null);

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    control,
    setError,
    clearErrors,
    watch,
    formState: { errors, isSubmitting },
  } = useForm<GuestFormValues>({
    resolver: zodResolver(guestFormSchema),
    defaultValues: {
      firstName: "",
      lastName: "",
      phone: "",
      relationship: "P",
      familyGroup: undefined,
      confirmed: false,
    },
  });

  useEffect(() => {
    if (!guestQuery.data) return;

    reset({
      firstName: guestQuery.data.first_name,
      lastName: guestQuery.data.last_name,
      phone: guestQuery.data.phone ?? "",
      relationship: guestQuery.data.relationship,
      familyGroup: guestQuery.data.family_group,
      confirmed: guestQuery.data.confirmed,
    });

    const fg = guestQuery.data.family_group;
    setAssignedFamilyGroup(fg);
    setFamilyGroupLabel(`Grupo #${fg}`);
  }, [guestQuery.data, reset]);

  // Close dropdown on click outside
  useEffect(() => {
    function handleClickOutside(e: MouseEvent) {
      if (familyContainerRef.current && !familyContainerRef.current.contains(e.target as Node)) {
        setShowFamilyDropdown(false);
      }
    }
    document.addEventListener("mousedown", handleClickOutside);
    return () => document.removeEventListener("mousedown", handleClickOutside);
  }, []);

  // Pre-select relationship based on role in create mode
  useEffect(() => {
    if (isEdit || !userMe?.role) return;
    if (userMe.role === "groom") setValue("relationship", "P");
    else if (userMe.role === "bride") setValue("relationship", "R");
  }, [userMe?.role, isEdit, setValue]);

  const familyOptions = useMemo(() => {
    if (!familySearch.trim()) return [];
    const q = familySearch.toLowerCase().trim();
    return allGuests
      .filter((g) => {
        if (isEdit && g.id === guestId) return false;
        const name = `${g.first_name} ${g.last_name}`.toLowerCase();
        return name.includes(q);
      })
      .slice(0, 8);
  }, [allGuests, familySearch, isEdit, guestId]);

  const familyMembers = useMemo((): Guest[] => {
    if (assignedFamilyGroup === undefined) return [];
    return allGuests.filter((g) => {
      if (isEdit && g.id === guestId) return false;
      return g.family_group === assignedFamilyGroup;
    });
  }, [allGuests, assignedFamilyGroup, isEdit, guestId]);

  function handleSelectFamily(guest: Guest) {
    setAssignedFamilyGroup(guest.family_group);
    setFamilyGroupLabel(`${guest.first_name} ${guest.last_name}`);
    setValue("familyGroup", guest.family_group, { shouldValidate: true });
    setFamilySearch("");
    setShowFamilyDropdown(false);
  }

  function clearFamily() {
    setAssignedFamilyGroup(undefined);
    setFamilyGroupLabel("");
    setFamilySearch("");
    setValue("familyGroup", undefined);
  }

  const mutationError =
    (createMutation.error instanceof Error && createMutation.error.message) ||
    (updateMutation.error instanceof Error && updateMutation.error.message) ||
    (guestQuery.error instanceof Error && guestQuery.error.message) ||
    null;

  async function onSubmit(values: GuestFormValues) {
    clearErrors("root");

    if (!getUserRacf()) {
      setError("root", {
        message: "Configure sua identificação (RACF) na página de Lista de Presença antes de salvar.",
      });
      return;
    }

    try {
      if (isEdit && guestId !== undefined) {
        await updateMutation.mutateAsync({
          id: guestId,
          input: {
            first_name: values.firstName.trim(),
            last_name: values.lastName.trim(),
            phone: values.phone.trim() || undefined,
            relationship: values.relationship,
            family_group: values.familyGroup,
            confirmed: values.confirmed,
          },
        });
      } else {
        await createMutation.mutateAsync({
          first_name: values.firstName.trim(),
          last_name: values.lastName.trim(),
          phone: values.phone.trim(),
          relationship: values.relationship,
          family_group: values.familyGroup,
        });
      }

      await navigate({ to: "/dashboard" });
    } catch (submitError) {
      setError("root", {
        message: submitError instanceof Error ? submitError.message : "Erro ao salvar",
      });
    }
  }

  // Unauthorized guard
  if (getUserRacf() && !roleLoading && userMe && !isAuthorized) {
    return <UnauthorizedPage />;
  }

  const inputClass =
    "w-full px-3.5 py-2.5 text-[0.88rem] border border-gold-muted/40 bg-ivory text-dark-warm placeholder:text-hint/40 outline-none focus:border-burgundy transition-colors";
  const labelClass =
    "block font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase text-hint mb-1.5";
  const loading = isEdit && guestQuery.isLoading;
  const saving = isSubmitting || createMutation.isPending || updateMutation.isPending;
  const relationship = watch("relationship");
  const confirmed = watch("confirmed");
  const errorMessage = errors.root?.message ?? mutationError;

  return (
    <div className="min-h-dvh bg-parchment">
      <Header />

      <main className="mx-auto max-w-[640px] px-6 pt-24 pb-16">
        <Link
          to="/dashboard"
          className="inline-flex items-center gap-1.5 font-heading text-[0.72rem] font-semibold tracking-[0.08em] uppercase text-hint no-underline mb-6 transition-colors hover:text-burgundy"
        >
          <ArrowLeft size={15} />
          Voltar para lista
        </Link>

        <h1 className="font-display text-[1.4rem] md:text-[1.7rem] font-bold text-dark mb-8">
          {isEdit ? "Editar Convidado" : "Novo Convidado"}
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
          <form onSubmit={handleSubmit(onSubmit)} className="border border-gold-muted/40 bg-ivory p-6 md:p-8 rounded">
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-4 mb-5">
              <div>
                <label className={labelClass}>Nome</label>
                <input type="text" placeholder="Ex: João" className={inputClass} {...register("firstName")} />
                {errors.firstName?.message && <p className="text-[0.72rem] text-[#c25550] mt-1">{errors.firstName.message}</p>}
              </div>
              <div>
                <label className={labelClass}>Sobrenome</label>
                <input type="text" placeholder="Ex: Silva" className={inputClass} {...register("lastName")} />
                {errors.lastName?.message && <p className="text-[0.72rem] text-[#c25550] mt-1">{errors.lastName.message}</p>}
              </div>
            </div>

            <div className="mb-5">
              <label className={labelClass}>Telefone</label>
              <Controller
                name="phone"
                control={control}
                render={({ field }) => (
                  <input
                    type="text"
                    value={formatPhoneDisplay(field.value ?? "")}
                    onChange={(e) => {
                      const digits = e.target.value.replace(/\D/g, "").slice(0, 11);
                      field.onChange(digits);
                    }}
                    onBlur={field.onBlur}
                    placeholder="(11) 99999-9999"
                    maxLength={15}
                    className={`${inputClass} font-mono tracking-wider`}
                  />
                )}
              />
              <p className="text-[0.72rem] text-hint/60 mt-1">DDD + número (11 dígitos). Opcional.</p>
              {errors.phone?.message && <p className="text-[0.72rem] text-[#c25550] mt-1">{errors.phone.message}</p>}
            </div>

            <div className="mb-5">
              <label className={labelClass}>Lado</label>
              <div className="flex gap-3">
                <button
                  type="button"
                  onClick={() => setValue("relationship", "P")}
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
                  onClick={() => setValue("relationship", "R")}
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

            {/* Family Group Autocomplete */}
            <div className="mb-5">
              <label className={labelClass}>Grupo Familiar</label>
              <div ref={familyContainerRef} className="relative">
                {assignedFamilyGroup !== undefined ? (
                  <div className="flex items-center justify-between px-3.5 py-2.5 border border-gold-muted/40 bg-ivory">
                    <span className="text-[0.88rem] text-dark-warm">{familyGroupLabel}</span>
                    <button
                      type="button"
                      onClick={clearFamily}
                      className="text-hint/50 hover:text-dark-warm transition-colors cursor-pointer ml-2 shrink-0"
                      title="Remover seleção"
                    >
                      <X size={14} />
                    </button>
                  </div>
                ) : (
                  <input
                    type="text"
                    value={familySearch}
                    onChange={(e) => {
                      setFamilySearch(e.target.value);
                      setShowFamilyDropdown(true);
                    }}
                    onFocus={() => {
                      if (familySearch.trim()) setShowFamilyDropdown(true);
                    }}
                    placeholder="Pesquisar por nome do convidado..."
                    className={inputClass}
                    autoComplete="off"
                  />
                )}

                {showFamilyDropdown && familyOptions.length > 0 && assignedFamilyGroup === undefined && (
                  <div className="absolute top-full left-0 right-0 z-[50] bg-ivory border border-gold-muted/40 border-t-0 shadow-md max-h-52 overflow-y-auto">
                    {familyOptions.map((g) => (
                      <button
                        type="button"
                        key={g.id}
                        onMouseDown={(e) => e.preventDefault()}
                        onClick={() => handleSelectFamily(g)}
                        className="w-full text-left px-3.5 py-2.5 transition-colors hover:bg-parchment border-b border-gold-muted/20 last:border-b-0 cursor-pointer"
                      >
                        <span className="text-[0.85rem] text-dark-warm">
                          {g.first_name} {g.last_name}
                        </span>
                        <span className="text-[0.72rem] text-hint ml-2">Família #{g.family_group}</span>
                      </button>
                    ))}
                  </div>
                )}

                {showFamilyDropdown && familySearch.trim().length > 0 && familyOptions.length === 0 && assignedFamilyGroup === undefined && (
                  <div className="absolute top-full left-0 right-0 z-[50] bg-ivory border border-gold-muted/40 border-t-0 shadow-md px-3.5 py-3">
                    <span className="text-[0.82rem] text-hint">Nenhum convidado encontrado.</span>
                  </div>
                )}
              </div>

              {familyMembers.length > 0 && (
                <div className="mt-3 p-3 bg-parchment border border-gold-muted/25">
                  <p className="text-[0.7rem] font-heading font-semibold tracking-[0.06em] uppercase text-hint mb-2">
                    Outros desta família ({familyMembers.length})
                  </p>
                  <div className="flex flex-wrap gap-1.5">
                    {familyMembers.map((m) => (
                      <span
                        key={m.id}
                        className="inline-flex items-center text-[0.75rem] text-dark-warm bg-ivory border border-gold-muted/30 px-2 py-0.5"
                      >
                        {m.first_name} {m.last_name}
                      </span>
                    ))}
                  </div>
                </div>
              )}

              <p className="text-[0.72rem] text-hint/60 mt-1.5">
                {assignedFamilyGroup !== undefined
                  ? `Grupo #${assignedFamilyGroup} atribuído.`
                  : "Digite um nome para buscar e agrupar com uma família existente. Deixe em branco para criar um novo grupo."}
              </p>
            </div>

            {isEdit && (
              <div className="mb-6">
                <label className="flex items-center gap-3 cursor-pointer group">
                  <div className="relative">
                    <input type="checkbox" className="sr-only peer" {...register("confirmed")} />
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

            <div className="flex flex-col-reverse sm:flex-row gap-3 sm:justify-end pt-4 border-t border-gold-muted/20">
              <Link
                to="/dashboard"
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
