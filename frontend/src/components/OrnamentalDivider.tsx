export function OrnamentalDivider({ variant = "gold" }: { variant?: "gold" | "muted" }) {
  const color = variant === "gold" ? "text-gold" : "text-hint";
  return (
    <div className={`flex items-center justify-center gap-3 my-4 ${color}`}>
      <div className="h-px w-12 md:w-20 bg-linear-to-r from-transparent to-current" />
      <span className="text-base leading-none">&#9884;</span>
      <div className="h-px w-12 md:w-20 bg-linear-to-l from-transparent to-current" />
    </div>
  );
}
