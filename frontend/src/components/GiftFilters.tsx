import { useEffect, useRef, useState } from "react";
import ArrowUpDown from "lucide-react/dist/esm/icons/arrow-up-down";
import Check from "lucide-react/dist/esm/icons/check";
import Search from "lucide-react/dist/esm/icons/search";
import X from "lucide-react/dist/esm/icons/x";
import type { GiftSort } from "../lib/gift-queries";

interface Props {
  search: string;
  priceMin: string;
  priceMax: string;
  sort: "" | GiftSort;
  onSearchChange: (value: string) => void;
  onPriceMinChange: (value: string) => void;
  onPriceMaxChange: (value: string) => void;
  onSortChange: (value: "" | GiftSort) => void;
  onClear: () => void;
}

const fieldClass =
  "border border-gold-muted/50 bg-ivory text-dark-warm text-[0.82rem] px-3 py-2 outline-none focus:border-burgundy transition-colors";
const labelClass =
  "font-heading text-[0.65rem] font-semibold tracking-[0.1em] uppercase text-hint";

const noSpinnerClass =
  "[appearance:textfield] [&::-webkit-outer-spin-button]:appearance-none [&::-webkit-inner-spin-button]:appearance-none";

const sortOptions: Array<{ value: "" | GiftSort; label: string }> = [
  { value: "", label: "Mais recentes" },
  { value: "price_asc", label: "Menor preço" },
  { value: "price_desc", label: "Maior preço" },
];

export function GiftFilters({
  search,
  priceMin,
  priceMax,
  sort,
  onSearchChange,
  onPriceMinChange,
  onPriceMaxChange,
  onSortChange,
  onClear,
}: Props) {
  const [sortOpen, setSortOpen] = useState(false);
  const sortRef = useRef<HTMLDivElement>(null);
  const hasFilter =
    search !== "" || priceMin !== "" || priceMax !== "" || sort !== "";
  const selectedSort = sortOptions.find((option) => option.value === sort) ?? sortOptions[0];

  useEffect(() => {
    function handler(event: MouseEvent) {
      if (sortRef.current && !sortRef.current.contains(event.target as Node)) {
        setSortOpen(false);
      }
    }

    document.addEventListener("mousedown", handler);
    return () => document.removeEventListener("mousedown", handler);
  }, []);

  return (
    <div className="flex flex-wrap gap-3 items-end">
      <div className="flex flex-col gap-1 flex-1 min-w-[180px]">
        <label className={labelClass} htmlFor="gift-search">
          Buscar por nome
        </label>
        <div className="relative">
          <Search
            size={15}
            className="absolute left-3 top-1/2 -translate-y-1/2 text-hint pointer-events-none"
          />
          <input
            id="gift-search"
            type="text"
            value={search}
            onChange={(e) => onSearchChange(e.target.value)}
            placeholder="Ex.: batedeira"
            className={`${fieldClass} w-full pl-9`}
          />
        </div>
      </div>

      <div className="flex flex-col gap-1">
        <label className={labelClass} htmlFor="gift-price-min">
          Preço de (R$)
        </label>
        <input
          id="gift-price-min"
          type="number"
          min="0"
          inputMode="decimal"
          value={priceMin}
          onChange={(e) => onPriceMinChange(e.target.value)}
          placeholder="0"
          className={`${fieldClass} ${noSpinnerClass} w-28`}
        />
      </div>

      <div className="flex flex-col gap-1">
        <label className={labelClass} htmlFor="gift-price-max">
          Preço até (R$)
        </label>
        <input
          id="gift-price-max"
          type="number"
          min="0"
          inputMode="decimal"
          value={priceMax}
          onChange={(e) => onPriceMaxChange(e.target.value)}
          placeholder="∞"
          className={`${fieldClass} ${noSpinnerClass} w-28`}
        />
      </div>

      <div ref={sortRef} className="relative flex flex-col gap-1">
        <span className={labelClass} id="gift-sort-label">
          Ordenar por
        </span>
        <button
          type="button"
          aria-haspopup="listbox"
          aria-expanded={sortOpen}
          aria-labelledby="gift-sort-label"
          onClick={() => setSortOpen((open) => !open)}
          className={`inline-flex min-w-[150px] items-center justify-between gap-2 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-2.5 px-4 border transition-all duration-200 cursor-pointer whitespace-nowrap ${
            sortOpen || sort !== ""
              ? "bg-burgundy text-gold-light border-burgundy"
              : "bg-ivory text-dark-warm border-gold-muted/50 hover:border-burgundy hover:text-burgundy"
          }`}
        >
          <span className="inline-flex items-center gap-2">
            <ArrowUpDown size={14} />
            {selectedSort.label}
          </span>
        </button>

        {sortOpen && (
          <div
            role="listbox"
            aria-labelledby="gift-sort-label"
            className="absolute top-full right-0 mt-1 z-[60] bg-ivory border border-gold-muted/50 shadow-[0_8px_24px_rgba(28,20,16,0.12)] w-48 p-2"
          >
            {sortOptions.map((option) => {
              const selected = sort === option.value;
              return (
                <button
                  key={option.value || "recent"}
                  type="button"
                  role="option"
                  aria-selected={selected}
                  onClick={() => {
                    onSortChange(option.value);
                    setSortOpen(false);
                  }}
                  className={`w-full inline-flex items-center justify-between gap-2 font-heading text-[0.67rem] font-semibold tracking-[0.06em] uppercase py-2 px-3 border transition-all duration-150 cursor-pointer ${
                    selected
                      ? "bg-burgundy text-gold-light border-burgundy"
                      : "bg-transparent text-dark-warm border-transparent hover:border-burgundy hover:text-burgundy"
                  }`}
                >
                  {option.label}
                  {selected && <Check size={12} />}
                </button>
              );
            })}
          </div>
        )}
      </div>

      {hasFilter && (
        <button
          type="button"
          onClick={onClear}
          className="inline-flex items-center gap-1.5 font-heading text-[0.7rem] font-semibold tracking-[0.08em] uppercase py-2 px-3 border border-gold-muted/50 text-hint hover:border-burgundy hover:text-burgundy transition-all duration-200 cursor-pointer"
        >
          <X size={12} />
          Limpar
        </button>
      )}
    </div>
  );
}
