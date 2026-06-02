import { useEffect, useRef, useState } from "react";
import {
  getMercadoPagoInstance,
  type BrickController,
  type PaymentBrickFormData,
} from "../lib/mercado-pago";

interface Props {
  amount: number;
  containerId?: string;
  onSubmit: (formData: PaymentBrickFormData) => Promise<void>;
}

export function PaymentBrick({ amount, containerId = "payment-brick-container", onSubmit }: Props) {
  const [error, setError] = useState<string | null>(null);
  const controllerRef = useRef<BrickController | null>(null);
  const onSubmitRef = useRef(onSubmit);

  useEffect(() => {
    onSubmitRef.current = onSubmit;
  }, [onSubmit]);

  useEffect(() => {
    let cancelled = false;

    (async () => {
      try {
        const mp = await getMercadoPagoInstance();
        if (cancelled) return;

        const builder = mp.bricks();
        const controller = await builder.create("payment", containerId, {
          initialization: { amount },
          customization: {
            paymentMethods: {
              creditCard: "all",
              bankTransfer: ["pix"],
              maxInstallments: 12,
            },
            visual: { style: { theme: "default" } },
          },
          callbacks: {
            onReady: () => setError(null),
            onSubmit: async ({ formData }) => {
              await onSubmitRef.current(formData);
            },
            onError: (err) => {
              const message = err instanceof Error ? err.message : "Erro no pagamento.";
              setError(message);
            },
          },
        });
        if (cancelled) {
          controller.unmount();
          return;
        }
        controllerRef.current = controller;
      } catch (err) {
        if (cancelled) return;
        const message = err instanceof Error ? err.message : "Não foi possível carregar o checkout.";
        setError(message);
      }
    })();

    return () => {
      cancelled = true;
      if (controllerRef.current) {
        try {
          controllerRef.current.unmount();
        } catch {
        }
        controllerRef.current = null;
      }
    };
  }, [amount, containerId]);

  return (
    <div>
      {error && (
        <div className="mb-4 rounded border border-[#c25550]/30 bg-[#fef2f1] px-4 py-3 text-[0.85rem] text-[#7a2e2b]">
          {error}
        </div>
      )}
      <div id={containerId} />
    </div>
  );
}
