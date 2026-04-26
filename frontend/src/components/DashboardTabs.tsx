import { Link } from "@tanstack/react-router";

type Tab = "convidados" | "presentes" | "pagamentos";

const tabs: { id: Tab; label: string; to: string }[] = [
  { id: "convidados", label: "Convidados", to: "/dashboard" },
  { id: "presentes", label: "Presentes", to: "/dashboard/presentes" },
  { id: "pagamentos", label: "Pagamentos", to: "/dashboard/pagamentos" },
];

export function DashboardTabs({ active }: { active: Tab }) {
  return (
    <nav className="mb-6 flex gap-6 border-b border-gold-muted/30">
      {tabs.map((tab) =>
        tab.id === active ? (
          <span
            key={tab.id}
            className="font-heading text-[0.72rem] font-semibold tracking-[0.12em] uppercase text-burgundy py-2 border-b-2 border-burgundy -mb-px"
          >
            {tab.label}
          </span>
        ) : (
          <Link
            key={tab.id}
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
