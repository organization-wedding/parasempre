import { Link, useRouterState } from "@tanstack/react-router";

const tabs = [
  { label: "Convidados", to: "/admin" },
  { label: "Presentes", to: "/admin/presentes" },
  { label: "Pagamentos", to: "/admin/pagamentos" },
  { label: "Recados", to: "/admin/recados" },
] as const;

export function DashboardTabs() {
  const pathname = useRouterState({ select: (s) => s.location.pathname });
  const activeTo = tabs.find((tab) => tab.to === pathname)?.to;

  if (!activeTo) return null;

  return (
    <nav className="mb-6 flex gap-6 border-b border-gold-muted/30">
      {tabs.map((tab) =>
        tab.to === activeTo ? (
          <span
            key={tab.to}
            className="font-heading text-[0.72rem] font-semibold tracking-[0.12em] uppercase text-burgundy py-2 border-b-2 border-burgundy -mb-px"
          >
            {tab.label}
          </span>
        ) : (
          <Link
            key={tab.to}
            to={tab.to}
            className="font-heading text-[0.72rem] font-semibold tracking-[0.12em] uppercase text-hint py-2 no-underline transition-colors hover:text-burgundy"
          >
            {tab.label}
          </Link>
        ),
      )}
    </nav>
  );
}
