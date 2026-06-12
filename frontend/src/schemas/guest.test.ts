import { describe, test, expect } from "bun:test";
import {
  guestSchema,
  batchConfirmInputSchema,
  updateGuestInputSchema,
} from "./guest";

const baseGuest = {
  id: 1,
  first_name: "João",
  last_name: "Silva",
  relationship: "P" as const,
  family_group: 1,
  created_by: "system",
  updated_by: "system",
  created_at: "2024-01-01T00:00:00Z",
  updated_at: "2024-01-01T00:00:00Z",
};

describe("guestSchema — campo attending tri-estado", () => {
  test("aceita attending=true (vai comparecer)", () => {
    const result = guestSchema.parse({ ...baseGuest, attending: true });
    expect(result.attending).toBe(true);
  });

  test("aceita attending=false (não comparecerá)", () => {
    const result = guestSchema.parse({ ...baseGuest, attending: false });
    expect(result.attending).toBe(false);
  });

  test("aceita attending=null (aguardando)", () => {
    const result = guestSchema.parse({ ...baseGuest, attending: null });
    expect(result.attending).toBeNull();
  });

  test("rejeita campo confirmed (contrato antigo)", () => {
    expect(() =>
      guestSchema.parse({ ...baseGuest, attending: null, confirmed: true }),
    ).not.toThrow();
    const result = guestSchema.parse({ ...baseGuest, attending: null });
    expect(Object.keys(result)).not.toContain("confirmed");
  });
});

describe("batchConfirmInputSchema — campo attending", () => {
  test("aceita attending=true com guest_ids", () => {
    const result = batchConfirmInputSchema.parse({
      guest_ids: [1, 2, 3],
      attending: true,
    });
    expect(result.attending).toBe(true);
    expect(result.guest_ids).toEqual([1, 2, 3]);
  });

  test("aceita attending=false (regressão: decline grava false, não null)", () => {
    const result = batchConfirmInputSchema.parse({
      guest_ids: [4],
      attending: false,
    });
    expect(result.attending).toBe(false);
  });

  test("rejeita attending=null (batch só aceita boolean)", () => {
    expect(() =>
      batchConfirmInputSchema.parse({ guest_ids: [1], attending: null }),
    ).toThrow();
  });

  test("rejeita campo confirmed (contrato antigo)", () => {
    expect(() =>
      batchConfirmInputSchema.parse({ guest_ids: [1], confirmed: true }),
    ).toThrow();
  });

  test("rejeita guest_ids vazio", () => {
    expect(() =>
      batchConfirmInputSchema.parse({ guest_ids: [], attending: true }),
    ).toThrow();
  });
});

describe("updateGuestInputSchema — attending nullable opcional", () => {
  test("aceita attending=true", () => {
    const result = updateGuestInputSchema.parse({ attending: true });
    expect(result.attending).toBe(true);
  });

  test("aceita attending=false", () => {
    const result = updateGuestInputSchema.parse({ attending: false });
    expect(result.attending).toBe(false);
  });

  test("aceita attending=null (reset para aguardando)", () => {
    const result = updateGuestInputSchema.parse({ attending: null });
    expect(result.attending).toBeNull();
  });

  test("aceita attending ausente (campo opcional)", () => {
    const result = updateGuestInputSchema.parse({ first_name: "Maria" });
    expect(result.attending).toBeUndefined();
  });

  test("strip campo confirmed (contrato antigo — chave desconhecida descartada)", () => {
    const result = updateGuestInputSchema.safeParse({ confirmed: true });
    expect(result.success).toBe(false);
  });
});
