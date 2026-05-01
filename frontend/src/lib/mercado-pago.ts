import { MERCADO_PAGO_PUBLIC_KEY } from "../config";

const SDK_URL = "https://sdk.mercadopago.com/js/v2";

declare global {
  interface Window {
    MercadoPago?: MercadoPagoCtor;
  }
}

interface MercadoPagoCtor {
  new(publicKey: string, opts?: { locale?: string }): MercadoPagoInstance;
}

export interface MercadoPagoInstance {
  bricks(): BricksBuilder;
}

export interface BricksBuilder {
  create(
    brickType: "payment",
    containerId: string,
    settings: PaymentBrickSettings,
  ): Promise<BrickController>;
}

export interface BrickController {
  unmount(): void;
}

export interface PaymentBrickFormData {
  token?: string;
  payment_method_id: string;
  issuer_id?: string;
  installments?: number;
  payer: {
    email: string;
    identification: {
      type: string;
      number: string;
    };
  };
}

export interface PaymentBrickSettings {
  initialization: {
    amount: number;
    payer?: { email?: string };
  };
  customization?: {
    paymentMethods?: {
      creditCard?: "all" | string[];
      bankTransfer?: "all" | string[];
      maxInstallments?: number;
    };
    visual?: {
      style?: { theme?: "default" | "dark" | "bootstrap" | "flat" };
    };
  };
  callbacks: {
    onReady?: () => void;
    onSubmit: (param: { formData: PaymentBrickFormData }) => Promise<void>;
    onError?: (error: unknown) => void;
  };
}

let loaderPromise: Promise<MercadoPagoCtor> | null = null;

export function loadMercadoPagoSDK(): Promise<MercadoPagoCtor> {
  if (typeof window === "undefined") {
    return Promise.reject(new Error("Mercado Pago SDK only available in browser"));
  }
  if (window.MercadoPago) {
    return Promise.resolve(window.MercadoPago);
  }
  if (loaderPromise) return loaderPromise;

  loaderPromise = new Promise((resolve, reject) => {
    const existing = document.querySelector<HTMLScriptElement>(
      `script[src="${SDK_URL}"]`,
    );
    const onLoad = () => {
      if (window.MercadoPago) resolve(window.MercadoPago);
      else reject(new Error("MercadoPago not exposed after SDK load"));
    };
    if (existing) {
      existing.addEventListener("load", onLoad, { once: true });
      existing.addEventListener(
        "error",
        () => reject(new Error("Failed to load Mercado Pago SDK")),
        { once: true },
      );
      return;
    }
    const script = document.createElement("script");
    script.src = SDK_URL;
    script.async = true;
    script.crossOrigin = "anonymous";
    script.addEventListener("load", onLoad, { once: true });
    script.addEventListener(
      "error",
      () => {
        loaderPromise = null;
        reject(new Error("Failed to load Mercado Pago SDK"));
      },
      { once: true },
    );
    document.head.appendChild(script);
  });
  return loaderPromise;
}

export async function getMercadoPagoInstance(): Promise<MercadoPagoInstance> {
  if (!MERCADO_PAGO_PUBLIC_KEY) {
    throw new Error("Mercado Pago public key not configured.");
  }
  const Ctor = await loadMercadoPagoSDK();
  return new Ctor(MERCADO_PAGO_PUBLIC_KEY, { locale: "pt-BR" });
}
