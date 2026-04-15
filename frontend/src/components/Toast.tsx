import { useEffect, useRef, useState } from "react";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";

interface ToastProps {
  title: string;
  message: string;
  onDismiss: () => void;
}

const ENTER_DELAY_MS = 20;
const VISIBLE_MS = 4700;
const EXIT_MS = 300;

export function Toast({ title, message, onDismiss }: ToastProps) {
  const [show, setShow] = useState(false);
  const dismissRef = useRef(onDismiss);
  dismissRef.current = onDismiss;

  useEffect(() => {
    const enter = setTimeout(() => setShow(true), ENTER_DELAY_MS);
    const exit = setTimeout(() => setShow(false), ENTER_DELAY_MS + VISIBLE_MS);
    const remove = setTimeout(() => dismissRef.current(), ENTER_DELAY_MS + VISIBLE_MS + EXIT_MS);
    return () => {
      clearTimeout(enter);
      clearTimeout(exit);
      clearTimeout(remove);
    };
  }, []);

  return (
    <div
      role="alert"
      className="pointer-events-none fixed bottom-12 right-6 z-50 flex max-w-[380px] items-start gap-3 border border-[#c25550]/50 bg-[#c25550]/20 backdrop-blur-md px-5 py-4 rounded-[2px] shadow-[0_12px_40px_rgba(0,0,0,0.5)]"
      style={{
        transform: show ? "translate(0, 0)" : "translate(0, 1rem)",
        opacity: show ? 1 : 0,
        transition: "transform 300ms ease-out, opacity 300ms ease-out",
      }}
    >
      <AlertTriangle size={18} className="text-[#c25550] shrink-0 mt-[2px]" />
      <div className="flex flex-col gap-1 min-w-0">
        <span className="font-heading text-[0.72rem] font-semibold tracking-[0.1em] uppercase text-[#fca5a5]">
          {title}
        </span>
        <span className="text-[0.82rem] leading-relaxed text-ivory/80">
          {message}
        </span>
      </div>
    </div>
  );
}
