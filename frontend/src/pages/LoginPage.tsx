import { useEffect, useRef, useState } from "react";
import { Link, useNavigate, useSearch } from "@tanstack/react-router";
import Phone from "lucide-react/dist/esm/icons/phone";
import ArrowLeft from "lucide-react/dist/esm/icons/arrow-left";
import ShieldCheck from "lucide-react/dist/esm/icons/shield-check";
import Loader from "lucide-react/dist/esm/icons/loader-circle";
import { isAuthenticated, setAuth } from "../lib/auth";
import { useSendOtpMutation, useVerifyOtpMutation } from "../lib/auth-queries";
import { devLogin, OtpApiError, OtpRateLimitError } from "../lib/api";
import { Toast } from "../components/Toast";
import { IS_DEV } from "../config";

function titleForStatus(status: number): string {
  if (status === 400) return "Dados inválidos";
  if (status === 401) return "Código inválido";
  if (status === 404) return "Telefone não encontrado";
  if (status === 429) return "Muitas tentativas";
  if (status >= 500) return "Erro do servidor";
  return "Erro";
}

function formatPhoneDisplay(digits: string): string {
  if (digits.length === 0) return "";
  if (digits.length <= 2) return digits;
  if (digits.length <= 7) return `(${digits.slice(0, 2)}) ${digits.slice(2)}`;
  return `(${digits.slice(0, 2)}) ${digits.slice(2, 7)}-${digits.slice(7)}`;
}

function maskPhone(phone: string): string {
  if (phone.length < 11) return phone;
  return `(${phone.slice(0, 2)}) ${phone.slice(2, 5)}**-**${phone.slice(9)}`;
}

function KeySVG({ className = "" }: { className?: string }) {
  return (
    <svg
      viewBox="0 0 200 220"
      className={className}
      fill="none"
      xmlns="http://www.w3.org/2000/svg"
      aria-hidden="true"
    >
      <circle cx="100" cy="100" r="88" stroke="var(--color-gold)" strokeWidth="1" opacity="0.25" />
      <circle cx="100" cy="100" r="82" stroke="var(--color-gold)" strokeWidth="0.5" opacity="0.15" />
      <path
        d="M100 24 L164 48 L164 106 Q164 158 100 180 Q36 158 36 106 L36 48 Z"
        fill="rgba(152,159,91,0.05)"
        stroke="var(--color-gold)"
        strokeWidth="1.5"
        opacity="0.85"
      />
      <path
        d="M100 34 L156 56 L156 106 Q156 152 100 170 Q44 152 44 106 L44 56 Z"
        fill="none"
        stroke="var(--color-gold-muted)"
        strokeWidth="0.8"
        opacity="0.35"
      />
      {/* Key shape */}
      <circle cx="100" cy="80" r="18" fill="none" stroke="var(--color-gold)" strokeWidth="2" opacity="0.9" />
      <circle cx="100" cy="80" r="8" fill="none" stroke="var(--color-gold)" strokeWidth="1.5" opacity="0.5" />
      <line x1="100" y1="98" x2="100" y2="140" stroke="var(--color-gold)" strokeWidth="2" opacity="0.9" strokeLinecap="round" />
      <line x1="100" y1="120" x2="112" y2="120" stroke="var(--color-gold)" strokeWidth="2" opacity="0.7" strokeLinecap="round" />
      <line x1="100" y1="130" x2="110" y2="130" stroke="var(--color-gold)" strokeWidth="2" opacity="0.7" strokeLinecap="round" />
      {/* Corner ornaments */}
      <text x="52" y="56" fill="var(--color-gold)" fontSize="9" opacity="0.3" textAnchor="middle">✦</text>
      <text x="148" y="56" fill="var(--color-gold)" fontSize="9" opacity="0.3" textAnchor="middle">✦</text>
      <text x="52" y="160" fill="var(--color-gold)" fontSize="9" opacity="0.3" textAnchor="middle">✦</text>
      <text x="148" y="160" fill="var(--color-gold)" fontSize="9" opacity="0.3" textAnchor="middle">✦</text>
    </svg>
  );
}

const RESEND_COOLDOWN = 60;

export function LoginPage() {
  const navigate = useNavigate();
  const search = useSearch({ strict: false }) as { redirect?: string };
  const redirectTo = search.redirect ?? "/admin";

  const [step, setStep] = useState<"phone" | "code">("phone");
  const [phone, setPhone] = useState("");
  const [code, setCode] = useState("");
  const [toast, setToast] = useState<{ id: number; title: string; message: string } | null>(null);
  const [resendCooldown, setResendCooldown] = useState(0);
  const [autoLoginPending, setAutoLoginPending] = useState(IS_DEV);

  const codeInputRef = useRef<HTMLInputElement>(null);
  const autoLoginAttemptedRef = useRef(false);
  const sendOtp = useSendOtpMutation();
  const verifyOtp = useVerifyOtpMutation();

  function showToast(title: string, message: string) {
    setToast({ id: Date.now(), title, message });
  }

  function showErrorToast(err: unknown, fallback: string) {
    if (err instanceof OtpApiError) {
      showToast(titleForStatus(err.status), err.message);
      return;
    }
    showToast("Erro", err instanceof Error ? err.message : fallback);
  }

  // Auto-login em dev/teste: bypass do OTP via /api/auth/dev-login.
  // Dispara em localhost e ambiente de teste — produção segue fluxo normal.
  useEffect(() => {
    if (!IS_DEV) return;
    if (autoLoginAttemptedRef.current) return;
    autoLoginAttemptedRef.current = true;

    if (isAuthenticated()) {
      const target = redirectTo.startsWith("/") ? redirectTo : "/admin";
      void navigate({ to: target });
      return;
    }

    devLogin()
      .then((res) => {
        setAuth(res.token, res.role, res.uracf);
        const target = redirectTo.startsWith("/") ? redirectTo : "/admin";
        void navigate({ to: target });
      })
      .catch((err) => {
        console.warn("dev-login indisponível, fluxo OTP normal:", err);
        setAutoLoginPending(false);
      });
  }, [navigate, redirectTo]);

  // Resend cooldown timer
  useEffect(() => {
    if (resendCooldown <= 0) return;
    const timer = setTimeout(() => setResendCooldown((c) => c - 1), 1000);
    return () => clearTimeout(timer);
  }, [resendCooldown]);

  // Auto-focus code input
  useEffect(() => {
    if (step === "code") {
      setTimeout(() => codeInputRef.current?.focus(), 100);
    }
  }, [step]);

  async function handleSendOtp() {
    if (phone.length !== 11) {
      showToast("Dados inválidos", "Informe um telefone válido com DDD (11 dígitos).");
      return;
    }
    try {
      await sendOtp.mutateAsync(phone);
      setStep("code");
      setCode("");
      setResendCooldown(RESEND_COOLDOWN);
    } catch (err) {
      if (err instanceof OtpRateLimitError) {
        setResendCooldown(err.retryAfterSeconds);
        setStep("code");
      }
      showErrorToast(err, "Erro ao enviar código");
    }
  }

  async function handleVerifyOtp() {
    if (code.length !== 6) {
      showToast("Dados inválidos", "Informe o código de 6 dígitos.");
      return;
    }
    try {
      const result = await verifyOtp.mutateAsync({ phone, code });
      setAuth(result.token, result.role, result.uracf);

      const target = redirectTo.startsWith("/") ? redirectTo : "/admin";
      void navigate({ to: target });
    } catch (err) {
      showErrorToast(err, "Código inválido ou expirado");
    }
  }

  async function handleResend() {
    try {
      await sendOtp.mutateAsync(phone);
      setResendCooldown(RESEND_COOLDOWN);
      setCode("");
    } catch (err) {
      if (err instanceof OtpRateLimitError) {
        setResendCooldown(err.retryAfterSeconds);
      }
      showErrorToast(err, "Erro ao reenviar código");
    }
  }

  function handleBackToPhone() {
    setStep("phone");
    setCode("");
  }

  const isLoading = sendOtp.isPending || verifyOtp.isPending;

  if (autoLoginPending) {
    return (
      <div className="relative flex min-h-dvh flex-col items-center justify-center overflow-hidden bg-dark p-6">
        <Loader size={32} className="animate-spin text-gold" />
      </div>
    );
  }

  return (
    <div className="relative flex min-h-dvh flex-col items-center justify-center overflow-hidden bg-dark p-6">
      <div className="dot-pattern opacity-[0.04]" aria-hidden="true" />
      <div className="vignette" aria-hidden="true" />

      {/* Top decorative line */}
      <div className="absolute top-0 left-0 right-0 h-px bg-linear-to-r from-transparent via-gold/40 to-transparent" />

      <div className="relative z-1 w-full max-w-[420px]">
        <KeySVG className="w-[140px] md:w-[170px] h-auto mx-auto mb-6" />

        {/* Ornamental divider */}
        <div className="flex items-center justify-center gap-3 mb-5">
          <div className="h-px w-12 bg-gold/30" />
          <span className="text-gold/40 text-[0.6rem] tracking-[0.3em] font-heading uppercase">
            {step === "phone" ? "Identificação" : "Verificação"}
          </span>
          <div className="h-px w-12 bg-gold/30" />
        </div>

        <h1 className="font-display text-[1.4rem] md:text-[1.7rem] font-bold text-ivory mb-3 leading-tight text-center">
          {step === "phone" ? "Acesso ao Sistema" : "Código de Verificação"}
        </h1>

        <p className="text-[0.88rem] text-gold-muted leading-relaxed mb-8 text-center">
          {step === "phone"
            ? "Informe seu telefone cadastrado para receber o código de acesso via WhatsApp."
            : (
                <>
                  Enviamos um código para{" "}
                  <span className="font-semibold text-gold">{maskPhone(phone)}</span>.
                  Verifique seu WhatsApp.
                </>
              )}
        </p>

        {/* Step 1: Phone */}
        {step === "phone" && (
          <form
            onSubmit={(e) => {
              e.preventDefault();
              void handleSendOtp();
            }}
          >
            <label className="block font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-gold-muted/70 mb-2">
              <Phone size={12} className="inline mr-1.5 -mt-px" />
              Telefone
            </label>
            <input
              type="text"
              value={formatPhoneDisplay(phone)}
              onChange={(e) => {
                const digits = e.target.value.replace(/\D/g, "").slice(0, 11);
                setPhone(digits);
              }}
              placeholder="(43) 99999-9999"
              maxLength={15}
              autoFocus
              className="w-full px-4 py-3 text-[1rem] font-mono tracking-wider border border-gold-muted/30 bg-dark-warm/30 text-ivory placeholder:text-gold-muted/25 outline-none focus:border-gold/60 transition-colors rounded-[2px]"
            />
            <p className="text-[0.72rem] text-gold-muted/40 mt-1.5 mb-6">
              DDD + número (11 dígitos)
            </p>

            <button
              type="submit"
              disabled={isLoading || phone.length !== 11}
              className="w-full inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.1em] uppercase py-3 px-6 bg-burgundy text-gold-light border border-burgundy transition-all duration-300 hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed rounded-[2px]"
            >
              {sendOtp.isPending ? (
                <Loader size={16} className="animate-spin" />
              ) : (
                <ShieldCheck size={16} />
              )}
              {sendOtp.isPending ? "Enviando..." : "Enviar Código"}
            </button>
          </form>
        )}

        {/* Step 2: Code */}
        {step === "code" && (
          <form
            onSubmit={(e) => {
              e.preventDefault();
              void handleVerifyOtp();
            }}
          >
            <label className="block font-heading text-[0.68rem] font-semibold tracking-[0.1em] uppercase text-gold-muted/70 mb-2">
              <ShieldCheck size={12} className="inline mr-1.5 -mt-px" />
              Código OTP
            </label>
            <input
              ref={codeInputRef}
              type="text"
              inputMode="numeric"
              value={code}
              onChange={(e) => {
                const digits = e.target.value.replace(/\D/g, "").slice(0, 6);
                setCode(digits);
              }}
              placeholder="000000"
              maxLength={6}
              className="w-full px-4 py-3 text-[1.5rem] font-mono tracking-[0.5em] text-center border border-gold-muted/30 bg-dark-warm/30 text-ivory placeholder:text-gold-muted/25 outline-none focus:border-gold/60 transition-colors rounded-[2px]"
            />

            <div className="flex items-center justify-between mt-2 mb-6">
              <button
                type="button"
                onClick={handleBackToPhone}
                className="text-[0.72rem] text-gold-muted/60 hover:text-gold transition-colors cursor-pointer font-heading tracking-wide"
              >
                Trocar telefone
              </button>
              {resendCooldown > 0 ? (
                <span className="text-[0.72rem] text-gold-muted/40 font-heading tracking-wide">
                  Reenviar em {resendCooldown}s
                </span>
              ) : (
                <button
                  type="button"
                  onClick={() => void handleResend()}
                  disabled={sendOtp.isPending}
                  className="text-[0.72rem] text-gold-muted/60 hover:text-gold transition-colors cursor-pointer font-heading tracking-wide disabled:opacity-40"
                >
                  Reenviar código
                </button>
              )}
            </div>

            <button
              type="submit"
              disabled={isLoading || code.length !== 6}
              className="w-full inline-flex items-center justify-center gap-2 font-heading text-[0.72rem] font-semibold tracking-[0.1em] uppercase py-3 px-6 bg-burgundy text-gold-light border border-burgundy transition-all duration-300 hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] cursor-pointer disabled:opacity-40 disabled:cursor-not-allowed rounded-[2px]"
            >
              {verifyOtp.isPending ? (
                <Loader size={16} className="animate-spin" />
              ) : (
                <ShieldCheck size={16} />
              )}
              {verifyOtp.isPending ? "Verificando..." : "Verificar"}
            </button>
          </form>
        )}

        {/* Back to home */}
        <div className="mt-8 text-center">
          <Link
            to="/"
            className="inline-flex items-center gap-1.5 font-heading text-[0.68rem] tracking-[0.08em] uppercase text-gold-muted/50 no-underline transition-colors hover:text-gold"
          >
            <ArrowLeft size={13} />
            Voltar ao início
          </Link>
        </div>
      </div>

      {/* Bottom decorative line */}
      <div className="absolute bottom-0 left-0 right-0 h-px bg-linear-to-r from-transparent via-gold/40 to-transparent" />

      {toast && (
        <Toast
          key={toast.id}
          title={toast.title}
          message={toast.message}
          onDismiss={() => setToast(null)}
        />
      )}
    </div>
  );
}
