import { useMutation } from "@tanstack/react-query";
import { createPurchase } from "./api";
import type { PurchaseRequest } from "../schemas/payment";

export function useCreatePurchaseMutation(giftId: number) {
  return useMutation({
    mutationFn: (input: PurchaseRequest) => createPurchase(giftId, input),
  });
}
