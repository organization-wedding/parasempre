import Phone from "lucide-react/dist/esm/icons/phone";
import Mail from "lucide-react/dist/esm/icons/mail";
import { CoatOfArms } from "./CoatOfArms";
import { COUPLE, CONTACT, NAV_LINKS } from "../config";

export function Footer() {
  return (
    <footer className="bg-dark text-gold-muted relative">
      {/* Ornament line */}
      <div className="h-[3px] bg-linear-to-r from-burgundy-deep via-gold to-burgundy-deep" />

      <div className="mx-auto grid max-w-[1100px] grid-cols-1 gap-10 px-6 py-12 md:grid-cols-[1.8fr_1fr_1fr] md:py-14">
        <div className="flex flex-col gap-1">
          <CoatOfArms className="w-22 h-auto mb-2" />
          <span className="font-display text-[1.35rem] font-bold text-gold">
            Para Sempre
          </span>
          <span className="font-heading text-[0.85rem] font-medium tracking-[0.08em] text-gold-muted">
            {COUPLE.name1} & {COUPLE.name2}
          </span>
          <span className="text-[0.95rem] text-hint mt-0.5">
            12 de Outubro de 2026
          </span>
        </div>

        <div className="flex flex-col gap-2.5">
          <h3 className="font-heading text-[0.68rem] font-bold tracking-[0.2em] uppercase text-gold mb-4">
            Navegação
          </h3>
          {NAV_LINKS.map((link) => (
            <a
              key={link.href}
              href={link.href}
              className="inline-flex items-center gap-2 w-fit text-[1.05rem] text-gold-muted no-underline transition-colors duration-300 hover:text-gold"
            >
              {link.label}
            </a>
          ))}
        </div>

        <div className="flex flex-col gap-2.5">
          <h3 className="font-heading text-[0.68rem] font-bold tracking-[0.2em] uppercase text-gold mb-4">
            Contato
          </h3>
          <a
            href={CONTACT.phoneHref}
            className="inline-flex items-center gap-2 w-fit text-[1.05rem] text-gold-muted no-underline transition-colors duration-300 hover:text-gold"
          >
            <Phone size={14} />
            {CONTACT.phone}
          </a>
          <a
            href={`mailto:${CONTACT.email}`}
            className="inline-flex items-center gap-2 w-fit text-[1.05rem] text-gold-muted no-underline transition-colors duration-300 hover:text-gold"
          >
            <Mail size={14} />
            {CONTACT.email}
          </a>
        </div>
      </div>

      <div className="border-t border-gold/10 px-6 py-5 text-center">
        <p className="text-[0.85rem] text-hint">
          &copy; 2026 Para Sempre — {COUPLE.name1} & {COUPLE.name2}. Todos os
          direitos reservados.
        </p>
      </div>
    </footer>
  );
}
