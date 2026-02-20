import { useEffect, useRef, useState } from "react";
import { useQuery, useQueryClient } from "@tanstack/react-query";
import { useNavigate } from "@tanstack/react-router";
import X from "lucide-react/dist/esm/icons/x";
import Search from "lucide-react/dist/esm/icons/search";
import UserCog from "lucide-react/dist/esm/icons/user-cog";
import LogOut from "lucide-react/dist/esm/icons/log-out";
import TriangleAlert from "lucide-react/dist/esm/icons/triangle-alert";
import { API_BASE, IS_DEV } from "../config";
import { clearUserRacf, setUserRacf, type UserListItem } from "../lib/api";
import { useCurrentRacf, useUserMeQuery } from "../lib/user-queries";

function roleLabel(role: string): string {
  if (role === "groom") return "Noivo";
  if (role === "bride") return "Noiva";
  return "Convidado";
}

function roleBadge(role: string): string {
  if (role === "groom")
    return "bg-burgundy/10 border border-burgundy/25 text-burgundy";
  if (role === "bride")
    return "bg-gold/15 border border-gold/30 text-gold-dark";
  return "bg-hint/10 border border-hint/20 text-hint";
}

// Inner component — only ever mounted when IS_DEV is true, so hooks run unconditionally.
function ImpersonationModalInner() {
  const [modalOpen, setModalOpen] = useState(false);
  const [confirmLogout, setConfirmLogout] = useState(false);
  const [search, setSearch] = useState("");
  const searchRef = useRef<HTMLInputElement>(null);
  const queryClient = useQueryClient();
  const navigate = useNavigate();
  const currentRacf = useCurrentRacf();
  const { data: userMe } = useUserMeQuery(!!currentRacf);

  const { data: users = [] } = useQuery<UserListItem[]>({
    queryKey: ["dev-users-list"],
    queryFn: async () => {
      const res = await fetch(`${API_BASE}/api/users`);
      if (!res.ok) throw new Error(`Erro ${res.status}`);
      return res.json() as Promise<UserListItem[]>;
    },
    staleTime: 5 * 60 * 1000,
    retry: false,
  });

  // Keyboard shortcut: ⌘⌥I (Mac) or Ctrl+Alt+I (Win/Linux)
  useEffect(() => {
    function handleKeyDown(e: KeyboardEvent) {
      if (e.key.toLowerCase() === "i" && (e.metaKey || e.ctrlKey) && e.altKey) {
        e.preventDefault();
        setModalOpen((prev) => !prev);
      }
    }
    window.addEventListener("keydown", handleKeyDown);
    return () => window.removeEventListener("keydown", handleKeyDown);
  }, []);

  // ESC to close
  useEffect(() => {
    if (!modalOpen && !confirmLogout) return;
    function handleEsc(e: KeyboardEvent) {
      if (e.key === "Escape") {
        setModalOpen(false);
        setConfirmLogout(false);
      }
    }
    window.addEventListener("keydown", handleEsc);
    return () => window.removeEventListener("keydown", handleEsc);
  }, [modalOpen, confirmLogout]);

  // Auto-focus search on open; clear on close
  useEffect(() => {
    if (modalOpen) {
      setTimeout(() => searchRef.current?.focus(), 50);
    } else {
      setSearch("");
    }
  }, [modalOpen]);

  function handleImpersonate(user: UserListItem) {
    setUserRacf(user.uracf);
    queryClient.invalidateQueries({ queryKey: ["user-me"] });
    setModalOpen(false);
  }

  function handleExitConfirmed() {
    clearUserRacf();
    queryClient.clear();
    setConfirmLogout(false);
    void navigate({ to: "/" });
  }

  const filteredUsers = users.filter((u) => {
    const q = search.toLowerCase().trim();
    if (!q) return true;
    return (
      u.uracf.toLowerCase().includes(q) ||
      u.role.toLowerCase().includes(q) ||
      roleLabel(u.role).toLowerCase().includes(q) ||
      `${u.first_name} ${u.last_name}`.toLowerCase().includes(q)
    );
  });

  return (
    <>
      {/* ── Dev banner ─────────────────────────────────────────── */}
      {currentRacf ? (
        <div
          className="fixed bottom-0 left-0 right-0 z-[990] flex items-center gap-3 px-4 py-2 select-none"
          style={{
            backgroundColor: "var(--color-parchment)",
            borderTop: "1.5px solid rgba(196,169,109,0.35)",
            boxShadow: "0 -2px 16px rgba(28,30,20,0.06)",
          }}
        >
          {/* DEV tag */}
          <span
            className="shrink-0 px-1.5 py-0.5 text-[0.5rem] font-bold tracking-[0.2em] uppercase font-mono rounded-[2px]"
            style={{
              backgroundColor: "var(--color-dark)",
              color: "var(--color-gold-light)",
            }}
          >
            DEV
          </span>

          <UserCog size={12} style={{ color: "var(--color-gold)", opacity: 0.7 }} className="shrink-0" />

          <span
            className="hidden sm:block text-[0.7rem] tracking-wide"
            style={{ fontFamily: "Cinzel, serif", color: "var(--color-hint)" }}
          >
            Impersonando
          </span>

          <span
            className="text-[0.78rem] font-semibold tracking-widest"
            style={{ fontFamily: "Cinzel, serif", color: "var(--color-dark)" }}
          >
            {currentRacf}
          </span>

          {userMe?.role && (
            <span
              className={`text-[0.62rem] px-1.5 py-0.5 rounded-[2px] tracking-wide ${roleBadge(userMe.role)}`}
              style={{ fontFamily: "Cinzel, serif" }}
            >
              {roleLabel(userMe.role)}
            </span>
          )}

          <div className="flex items-center gap-2 ml-auto">
            <button
              type="button"
              onClick={() => setModalOpen(true)}
              className="text-[0.68rem] tracking-wide transition-colors duration-200 px-2.5 py-1 rounded-[2px] cursor-pointer"
              style={{
                fontFamily: "Cinzel, serif",
                color: "var(--color-dark-warm)",
                border: "1px solid rgba(196,169,109,0.4)",
                backgroundColor: "transparent",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.backgroundColor = "rgba(196,169,109,0.1)";
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.backgroundColor = "transparent";
              }}
            >
              Trocar ↕
            </button>
            <button
              type="button"
              onClick={() => setConfirmLogout(true)}
              className="inline-flex items-center gap-1.5 text-[0.68rem] tracking-wide transition-colors duration-200 px-2.5 py-1 rounded-[2px] cursor-pointer"
              style={{
                fontFamily: "Cinzel, serif",
                color: "var(--color-burgundy)",
                border: "1px solid rgba(152,159,91,0.35)",
                backgroundColor: "transparent",
              }}
              onMouseEnter={(e) => {
                (e.currentTarget as HTMLElement).style.backgroundColor = "rgba(152,159,91,0.08)";
              }}
              onMouseLeave={(e) => {
                (e.currentTarget as HTMLElement).style.backgroundColor = "transparent";
              }}
            >
              <LogOut size={11} />
              Sair
            </button>
          </div>
        </div>
      ) : (
        /* Hint bar — no active user */
        <div
          className="fixed bottom-0 left-0 right-0 z-[990] flex items-center gap-2.5 px-4 py-1.5 select-none"
          style={{
            backgroundColor: "var(--color-parchment)",
            borderTop: "1px solid rgba(196,169,109,0.2)",
          }}
        >
          <span
            className="shrink-0 px-1 py-0.5 text-[0.48rem] font-bold tracking-[0.2em] uppercase font-mono rounded-[2px] opacity-60"
            style={{
              backgroundColor: "var(--color-dark)",
              color: "var(--color-gold-light)",
            }}
          >
            DEV
          </span>
          <button
            type="button"
            onClick={() => setModalOpen(true)}
            className="text-[0.65rem] tracking-wide transition-colors duration-200 cursor-pointer"
            style={{
              fontFamily: "Cinzel, serif",
              color: "var(--color-hint)",
              opacity: 0.7,
            }}
            onMouseEnter={(e) => {
              (e.currentTarget as HTMLElement).style.color = "var(--color-gold-dark)";
              (e.currentTarget as HTMLElement).style.opacity = "1";
            }}
            onMouseLeave={(e) => {
              (e.currentTarget as HTMLElement).style.color = "var(--color-hint)";
              (e.currentTarget as HTMLElement).style.opacity = "0.7";
            }}
          >
            ⌘⌥I / Ctrl+Alt+I — Impersonar usuário
          </button>
        </div>
      )}

      {/* ── Logout confirmation modal ───────────────────────────── */}
      {confirmLogout && (
        <div
          className="fixed inset-0 z-[1010] flex items-center justify-center p-4"
          style={{ backgroundColor: "rgba(28,30,20,0.5)", backdropFilter: "blur(3px)" }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setConfirmLogout(false);
          }}
        >
          <div
            className="w-full max-w-sm rounded-[3px] overflow-hidden"
            style={{
              backgroundColor: "var(--color-ivory)",
              border: "1.5px solid rgba(196,169,109,0.35)",
              boxShadow: "0 8px 40px rgba(28,30,20,0.18)",
            }}
          >
            {/* Header */}
            <div
              className="flex items-center justify-between px-5 py-3.5"
              style={{ borderBottom: "1px solid rgba(196,169,109,0.2)" }}
            >
              <div className="flex items-center gap-2.5">
                <span
                  className="px-1.5 py-0.5 text-[0.48rem] font-bold tracking-[0.2em] uppercase font-mono rounded-[2px]"
                  style={{
                    backgroundColor: "var(--color-dark)",
                    color: "var(--color-gold-light)",
                  }}
                >
                  DEV
                </span>
                <span
                  className="text-[0.82rem] tracking-wide"
                  style={{ fontFamily: "Cinzel, serif", color: "var(--color-dark)" }}
                >
                  Confirmar Saída
                </span>
              </div>
              <button
                type="button"
                onClick={() => setConfirmLogout(false)}
                className="transition-colors duration-200 cursor-pointer"
                style={{ color: "var(--color-hint)", opacity: 0.5 }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.opacity = "1";
                  (e.currentTarget as HTMLElement).style.color = "var(--color-dark)";
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.opacity = "0.5";
                  (e.currentTarget as HTMLElement).style.color = "var(--color-hint)";
                }}
              >
                <X size={15} />
              </button>
            </div>

            {/* Body */}
            <div className="px-5 py-5 flex items-start gap-3">
              <TriangleAlert
                size={15}
                className="shrink-0 mt-0.5"
                style={{ color: "var(--color-burgundy)" }}
              />
              <p
                className="text-[0.88rem] leading-relaxed"
                style={{ fontFamily: "Cormorant Garamond, serif", color: "var(--color-dark-warm)" }}
              >
                Você será desconectado de{" "}
                <span
                  className="font-semibold tracking-widest"
                  style={{ fontFamily: "Cinzel, serif", color: "var(--color-dark)" }}
                >
                  {currentRacf}
                </span>{" "}
                e redirecionado para a tela inicial.
              </p>
            </div>

            {/* Actions */}
            <div
              className="flex items-center justify-end gap-2 px-5 py-3"
              style={{ borderTop: "1px solid rgba(196,169,109,0.2)" }}
            >
              <button
                type="button"
                onClick={() => setConfirmLogout(false)}
                className="text-[0.68rem] tracking-wide px-3 py-1.5 rounded-[2px] cursor-pointer transition-colors duration-200"
                style={{
                  fontFamily: "Cinzel, serif",
                  color: "var(--color-hint)",
                  border: "1px solid rgba(125,122,88,0.25)",
                  backgroundColor: "transparent",
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.backgroundColor = "rgba(125,122,88,0.06)";
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.backgroundColor = "transparent";
                }}
              >
                Cancelar
              </button>
              <button
                type="button"
                onClick={handleExitConfirmed}
                className="inline-flex items-center gap-1.5 text-[0.68rem] tracking-wide px-3 py-1.5 rounded-[2px] cursor-pointer transition-colors duration-200"
                style={{
                  fontFamily: "Cinzel, serif",
                  color: "var(--color-gold-light)",
                  backgroundColor: "var(--color-burgundy)",
                  border: "1px solid var(--color-burgundy)",
                }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.backgroundColor = "var(--color-burgundy-deep)";
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.backgroundColor = "var(--color-burgundy)";
                }}
              >
                <LogOut size={11} />
                Sair e ir para início
              </button>
            </div>
          </div>
        </div>
      )}

      {/* ── User selection modal ────────────────────────────────── */}
      {modalOpen && (
        <div
          className="fixed inset-0 z-[1000] flex items-center justify-center p-4"
          style={{ backgroundColor: "rgba(28,30,20,0.5)", backdropFilter: "blur(3px)" }}
          onClick={(e) => {
            if (e.target === e.currentTarget) setModalOpen(false);
          }}
        >
          <div
            className="w-full max-w-md rounded-[3px] overflow-hidden"
            style={{
              backgroundColor: "var(--color-ivory)",
              border: "1.5px solid rgba(196,169,109,0.35)",
              boxShadow: "0 8px 40px rgba(28,30,20,0.18)",
            }}
          >
            {/* Header */}
            <div
              className="flex items-center justify-between px-5 py-3.5"
              style={{ borderBottom: "1px solid rgba(196,169,109,0.18)" }}
            >
              <div className="flex items-center gap-2.5">
                <span
                  className="px-1.5 py-0.5 text-[0.48rem] font-bold tracking-[0.2em] uppercase font-mono rounded-[2px]"
                  style={{
                    backgroundColor: "var(--color-dark)",
                    color: "var(--color-gold-light)",
                  }}
                >
                  DEV
                </span>
                <span
                  className="text-[0.82rem] tracking-wide"
                  style={{ fontFamily: "Cinzel, serif", color: "var(--color-dark)" }}
                >
                  Impersonar Usuário
                </span>
              </div>
              <button
                type="button"
                onClick={() => setModalOpen(false)}
                className="transition-colors duration-200 cursor-pointer"
                style={{ color: "var(--color-hint)", opacity: 0.5 }}
                onMouseEnter={(e) => {
                  (e.currentTarget as HTMLElement).style.opacity = "1";
                  (e.currentTarget as HTMLElement).style.color = "var(--color-dark)";
                }}
                onMouseLeave={(e) => {
                  (e.currentTarget as HTMLElement).style.opacity = "0.5";
                  (e.currentTarget as HTMLElement).style.color = "var(--color-hint)";
                }}
              >
                <X size={15} />
              </button>
            </div>

            {/* Search */}
            <div
              className="px-4 py-3"
              style={{ borderBottom: "1px solid rgba(196,169,109,0.12)" }}
            >
              <div className="relative">
                <Search
                  size={13}
                  className="absolute left-3 top-1/2 -translate-y-1/2"
                  style={{ color: "var(--color-gold)", opacity: 0.5 }}
                />
                <input
                  ref={searchRef}
                  type="text"
                  value={search}
                  onChange={(e) => setSearch(e.target.value)}
                  placeholder="Buscar por nome, URACF ou função…"
                  className="w-full pl-8 pr-3 py-2 text-[0.82rem] rounded-[2px] outline-none transition-all duration-200"
                  style={{
                    fontFamily: "Cormorant Garamond, serif",
                    backgroundColor: "var(--color-parchment)",
                    border: "1px solid rgba(196,169,109,0.3)",
                    color: "var(--color-dark-warm)",
                  }}
                  onFocus={(e) => {
                    e.target.style.borderColor = "rgba(152,159,91,0.6)";
                    e.target.style.boxShadow = "0 0 0 2px rgba(152,159,91,0.1)";
                  }}
                  onBlur={(e) => {
                    e.target.style.borderColor = "rgba(196,169,109,0.3)";
                    e.target.style.boxShadow = "none";
                  }}
                />
              </div>
            </div>

            {/* User list */}
            <div className="max-h-72 overflow-y-auto">
              {filteredUsers.length === 0 ? (
                <div
                  className="px-5 py-8 text-center text-[0.8rem] tracking-wide"
                  style={{ fontFamily: "Cormorant Garamond, serif", color: "var(--color-hint)", opacity: 0.6 }}
                >
                  {users.length === 0 ? "Carregando usuários…" : "Nenhum usuário encontrado."}
                </div>
              ) : (
                filteredUsers.map((u) => (
                  <button
                    key={u.uracf}
                    type="button"
                    onClick={() => handleImpersonate(u)}
                    className="w-full text-left flex items-center gap-3 px-5 py-3 transition-colors duration-150 cursor-pointer"
                    style={{ borderBottom: "1px solid rgba(196,169,109,0.1)" }}
                    onMouseEnter={(e) => {
                      (e.currentTarget as HTMLElement).style.backgroundColor =
                        "rgba(196,169,109,0.08)";
                    }}
                    onMouseLeave={(e) => {
                      (e.currentTarget as HTMLElement).style.backgroundColor = "transparent";
                    }}
                  >
                    <div className="flex-1 min-w-0">
                      <div
                        className="text-[0.88rem] leading-snug truncate"
                        style={{ fontFamily: "Cormorant Garamond, serif", color: "var(--color-dark-warm)" }}
                      >
                        {u.first_name || u.last_name
                          ? `${u.first_name} ${u.last_name}`.trim()
                          : roleLabel(u.role)}
                      </div>
                      <div
                        className="text-[0.62rem] font-mono mt-0.5 tracking-wider"
                        style={{ color: "var(--color-hint)" }}
                      >
                        {u.uracf}
                      </div>
                    </div>

                    <span
                      className={`shrink-0 text-[0.62rem] px-1.5 py-0.5 rounded-[2px] tracking-wide ${roleBadge(u.role)}`}
                      style={{ fontFamily: "Cinzel, serif" }}
                    >
                      {roleLabel(u.role)}
                    </span>

                    {u.uracf === currentRacf && (
                      <span
                        className="shrink-0 text-[0.58rem] tracking-widest font-mono"
                        style={{ color: "var(--color-gold)", opacity: 0.7 }}
                      >
                        atual
                      </span>
                    )}
                  </button>
                ))
              )}
            </div>

            {/* Footer hint */}
            <div
              className="px-5 py-2 text-right text-[0.6rem] tracking-wide font-mono"
              style={{
                borderTop: "1px solid rgba(196,169,109,0.12)",
                color: "var(--color-hint)",
                opacity: 0.4,
              }}
            >
              ESC para fechar · ⌘⌥I / Ctrl+Alt+I para abrir
            </div>
          </div>
        </div>
      )}
    </>
  );
}

// Outer wrapper: returns null immediately in production — no hooks called.
export function ImpersonationModal() {
  if (!IS_DEV) return null;
  return <ImpersonationModalInner />;
}
