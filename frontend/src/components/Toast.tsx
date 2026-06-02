import { useEffect, useRef, useState, type ComponentType } from "react";
import AlertTriangle from "lucide-react/dist/esm/icons/alert-triangle";
import Check from "lucide-react/dist/esm/icons/check";

type ToastVariant = "error" | "success";

interface ToastProps {
  title: string;
  message: string;
  variant?: ToastVariant;
  onDismiss: () => void;
}

const ENTER_DELAY_MS = 20;
const VISIBLE_MS = 4700;
const EXIT_MS = 300;

type IconComponent = ComponentType<{ size?: number; className?: string }>;

const VARIANT_STYLES: Record<
  ToastVariant,
  {
    container: string;
    icon: string;
    title: string;
    Icon: IconComponent;
  }
> = {
  error: {
    container: "border-[#c25550]/50 bg-[#c25550]/20",
    icon: "text-[#c25550]",
    title: "text-[#fca5a5]",
    Icon: AlertTriangle as IconComponent,
  },
  success: {
    container: "border-burgundy/50 bg-burgundy/20",
    icon: "text-burgundy",
    title: "text-gold-light",
    Icon: Check as IconComponent,
  },
};

export function Toast({ title, message, variant = "error", onDismiss }: ToastProps) {
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

  const styles = VARIANT_STYLES[variant];
  const Icon = styles.Icon;

  return (
    <div
      role="alert"
      className={`pointer-events-none fixed bottom-12 right-6 z-50 flex max-w-[380px] items-start gap-3 border backdrop-blur-md px-5 py-4 rounded-[2px] shadow-[0_12px_40px_rgba(0,0,0,0.5)] ${styles.container}`}
      style={{
        transform: show ? "translate(0, 0)" : "translate(0, 1rem)",
        opacity: show ? 1 : 0,
        transition: "transform 300ms ease-out, opacity 300ms ease-out",
      }}
    >
      <Icon size={18} className={`shrink-0 mt-[2px] ${styles.icon}`} />
      <div className="flex flex-col gap-1 min-w-0">
        <span
          className={`font-heading text-[0.72rem] font-semibold tracking-[0.1em] uppercase ${styles.title}`}
        >
          {title}
        </span>
        <span className="text-[0.82rem] leading-relaxed text-ivory/80">{message}</span>
      </div>
    </div>
  );
}
