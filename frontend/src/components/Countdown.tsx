import { useState, useEffect } from "react";
import { WEDDING_DATE } from "../config";

function getTimeLeft() {
  const diff = WEDDING_DATE.getTime() - Date.now();
  if (diff <= 0) return { days: 0, hours: 0, minutes: 0, seconds: 0 };
  return {
    days: Math.floor(diff / (1000 * 60 * 60 * 24)),
    hours: Math.floor((diff / (1000 * 60 * 60)) % 24),
    minutes: Math.floor((diff / (1000 * 60)) % 60),
    seconds: Math.floor((diff / 1000) % 60),
  };
}

const UNITS = [
  { key: "days", label: "Dias" },
  { key: "hours", label: "Horas" },
  { key: "minutes", label: "Min" },
  { key: "seconds", label: "Seg" },
] as const;

export function Countdown() {
  const [timeLeft, setTimeLeft] = useState(getTimeLeft);

  useEffect(() => {
    const interval = setInterval(() => setTimeLeft(getTimeLeft()), 1000);
    return () => clearInterval(interval);
  }, []);

  return (
    <div
      className="anim-fade-in-up flex justify-center gap-3 sm:gap-4 md:gap-6"
      style={{ animationDelay: "1.3s" }}
    >
      {UNITS.map((unit) => (
        <div key={unit.key} className="flex flex-col items-center gap-2">
          <span className="font-heading text-[1.6rem] sm:text-[2rem] md:text-[2.5rem] font-bold text-ivory bg-[rgba(201,168,76,0.08)] border border-[rgba(201,168,76,0.2)] w-14 h-14 sm:w-[68px] sm:h-[68px] md:w-[84px] md:h-[84px] flex items-center justify-center leading-none">
            {String(timeLeft[unit.key]).padStart(2, "0")}
          </span>
          <span className="font-heading text-[0.55rem] sm:text-[0.6rem] md:text-[0.65rem] font-semibold tracking-[0.18em] uppercase text-gold-muted">
            {unit.label}
          </span>
        </div>
      ))}
    </div>
  );
}
