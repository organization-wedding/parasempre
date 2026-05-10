import { Outlet, Link, useNavigate } from "@tanstack/react-router";
import LogOut from "lucide-react/dist/esm/icons/log-out";
import { CoatOfArms } from "./CoatOfArms";
import { DashboardTabs } from "./DashboardTabs";
import { COUPLE } from "../config";
import { clearAuth } from "../lib/auth";
import { useUserMeQuery } from "../lib/user-queries";
import { UnauthorizedPage } from "../pages/UnauthorizedPage";

function AdminHeader() {
  const navigate = useNavigate();

  function handleLogout() {
    clearAuth();
    void navigate({ to: "/" });
  }

  return (
    <header className="fixed inset-x-0 top-0 z-[100] border-b border-gold-muted backdrop-blur-[12px] bg-[rgba(249,247,239,0.95)]">
      <div className="mx-auto flex max-w-[1280px] items-center justify-between px-6 py-2.5">
        <Link
          to="/admin"
          className="flex shrink-0 items-center gap-3 no-underline text-burgundy"
        >
          <CoatOfArms className="w-[42px] h-auto" />
          <span className="font-display text-lg font-bold tracking-wider text-burgundy">
            {COUPLE.name1} & {COUPLE.name2}
          </span>
        </Link>

        <button
          type="button"
          onClick={handleLogout}
          className="inline-flex items-center gap-[0.45rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.15rem] cursor-pointer transition-all duration-300 whitespace-nowrap bg-transparent text-burgundy border border-burgundy hover:bg-burgundy hover:text-gold-light"
        >
          <LogOut size={15} />
          Sair
        </button>
      </div>
      <div className="absolute -bottom-[3px] left-0 right-0 h-0.5 bg-linear-to-r from-transparent via-gold to-transparent opacity-50 pointer-events-none" />
    </header>
  );
}

export function AdminLayout() {
  const { data: userMe, isLoading: roleLoading } = useUserMeQuery();
  const isAuthorized = userMe?.role === "groom" || userMe?.role === "bride";

  if (!roleLoading && userMe && !isAuthorized) {
    return <UnauthorizedPage />;
  }

  return (
    <div className="min-h-dvh bg-parchment">
      <AdminHeader />
      <main className="mx-auto max-w-[1280px] px-6 pt-24 pb-16">
        {roleLoading ? (
          <div className="flex flex-col items-center justify-center py-20 text-hint">
            <div className="w-8 h-8 border-2 border-gold-muted/30 border-t-burgundy rounded-full animate-spin mb-4" />
            <span className="text-[0.85rem]">Verificando permissões...</span>
          </div>
        ) : (
          <>
            <DashboardTabs />
            <Outlet />
          </>
        )}
      </main>
    </div>
  );
}
