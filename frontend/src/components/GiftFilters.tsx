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
  const hasFilter =
    search !== "" || priceMin !== "" || priceMax !== "" || sort !== "";

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

      <div className="flex flex-col gap-1">
        <label className={labelClass} htmlFor="gift-sort">
          Ordenar por
        </label>
        <select
          id="gift-sort"
          value={sort}
          onChange={(e) => onSortChange(e.target.value as "" | GiftSort)}
          className={fieldClass}
        >
          <option value="">Mais recentes</option>
          <option value="price_asc">Menor preço</option>
          <option value="price_desc">Maior preço</option>
        </select>
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
