import { useState } from "react";
import { Link } from "@tanstack/react-router";
import Menu from "lucide-react/dist/esm/icons/menu";
import X from "lucide-react/dist/esm/icons/x";
import Gift from "lucide-react/dist/esm/icons/gift";
import UserCheck from "lucide-react/dist/esm/icons/user-check";
import { CoatOfArms } from "./CoatOfArms";
import { NAV_LINKS } from "../config";

export function Header() {
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);

  return (
    <header className="fixed inset-x-0 top-0 z-100 border-b border-gold-muted backdrop-blur-[12px] bg-[rgba(249,247,239,0.95)]">
      <div className="mx-auto flex max-w-[1280px] items-center justify-between px-6 py-2.5">
        <Link
          to="/"
          className="flex shrink-0 items-center gap-3 no-underline text-burgundy"
          onClick={() => setMobileMenuOpen(false)}
        >
          <CoatOfArms className="w-[42px] h-auto" />
          <span className="font-display text-lg font-bold tracking-wider text-burgundy">
            Para Sempre
          </span>
        </Link>

        <nav className="hidden nav:flex gap-9">
          {NAV_LINKS.map((link) => (
            <Link
              key={link.href}
              to={link.href}
              className="group relative font-heading text-[0.72rem] font-semibold tracking-[0.12em] uppercase text-dark-warm no-underline py-1 transition-colors duration-300 hover:text-burgundy"
            >
              {link.label}
              <span className="absolute -bottom-0.5 left-1/2 h-[1.5px] w-0 -translate-x-1/2 bg-gold transition-all duration-300 group-hover:w-full" />
            </Link>
          ))}
        </nav>

        <div className="hidden nav:flex gap-2.5">
          <a
            href="/registrar-presenca"
            className="inline-flex items-center gap-[0.45rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.15rem] cursor-pointer no-underline transition-all duration-300 whitespace-nowrap bg-transparent text-burgundy border border-burgundy hover:bg-burgundy hover:text-gold-light"
          >
            <UserCheck size={15} />
            Registrar Presença
          </a>
          <a
            href="/comprar-presente"
            className="inline-flex items-center gap-[0.45rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.15rem] cursor-pointer no-underline transition-all duration-300 whitespace-nowrap bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] hover:-translate-y-px"
          >
            <Gift size={15} />
            Comprar Presente
          </a>
        </div>

        <button
          className="nav:hidden flex items-center justify-center bg-transparent border-none text-burgundy cursor-pointer p-1.5"
          onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
          aria-label={mobileMenuOpen ? "Fechar menu" : "Abrir menu"}
        >
          {mobileMenuOpen ? <X size={24} /> : <Menu size={24} />}
        </button>
      </div>

      {mobileMenuOpen && (
        <div className="anim-slide-down border-t border-parchment-dark px-6 pt-4 pb-6">
          <nav className="flex flex-col mb-5">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.href}
                to={link.href}
                onClick={() => setMobileMenuOpen(false)}
                className="font-heading text-[0.85rem] font-semibold tracking-[0.1em] uppercase text-dark-warm no-underline py-3 border-b border-parchment-dark transition-colors duration-200 hover:text-burgundy"
              >
                {link.label}
              </Link>
            ))}
          </nav>
          <div className="flex flex-col gap-2.5">
            <a
              href="/registrar-presenca"
              className="inline-flex items-center justify-center gap-[0.45rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.15rem] cursor-pointer no-underline transition-all duration-300 whitespace-nowrap bg-transparent text-burgundy border border-burgundy hover:bg-burgundy hover:text-gold-light w-full"
            >
              <UserCheck size={15} />
              Registrar Presença
            </a>
            <a
              href="/comprar-presente"
              className="inline-flex items-center justify-center gap-[0.45rem] font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-[0.55rem] px-[1.15rem] cursor-pointer no-underline transition-all duration-300 whitespace-nowrap bg-burgundy text-gold-light border border-burgundy hover:bg-burgundy-deep hover:shadow-[0_4px_16px_rgba(97,106,47,0.35)] hover:-translate-y-px w-full"
            >
              <Gift size={15} />
              Comprar Presente
            </a>
          </div>
        </div>
      )}
      <div className="absolute -bottom-[3px] left-0 right-0 h-0.5 bg-linear-to-r from-transparent via-gold to-transparent opacity-50 pointer-events-none" />
    </header>
  );
}
