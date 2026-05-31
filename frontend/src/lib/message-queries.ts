import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createGiftMessage,
  deleteGiftMessage,
  getMyTransactionMessage,
  listAdminGiftMessages,
} from "./api";
import type { PublicMessage } from "../schemas/giftMessage";

export const messageKeys = {
  byTransaction: (txId: number) => ["message-by-transaction", txId] as const,
  byGift: (giftId: number, page?: number) => ["messages-by-gift", giftId, page ?? 1] as const,
  byGiftAll: (giftId: number) => ["messages-by-gift", giftId] as const,
  adminAll: (params?: { page?: number; limit?: number }) => ["admin-messages", params] as const,
  adminAllRoot: ["admin-messages"] as const,
};

export function useMyTransactionMessageQuery(transactionId: number, enabled: boolean) {
  return useQuery({
    queryKey: messageKeys.byTransaction(transactionId),
    queryFn: () => getMyTransactionMessage(transactionId),
    enabled,
    staleTime: 60 * 1000,
  });
}

export function useAdminGiftMessagesQuery(params?: { page?: number; limit?: number }) {
  return useQuery({
    queryKey: messageKeys.adminAll(params),
    queryFn: () => listAdminGiftMessages(params),
    placeholderData: (prev) => prev,
    staleTime: 5 * 60 * 1000,
  });
}

export function useCreateGiftMessageMutation(transactionId: number) {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (formData: FormData) => createGiftMessage(transactionId, formData),
    onSuccess: (msg: PublicMessage) => {
      qc.setQueryData(messageKeys.byTransaction(transactionId), msg);
      qc.invalidateQueries({ queryKey: messageKeys.byGiftAll(msg.gift_id) });
      qc.invalidateQueries({ queryKey: messageKeys.adminAllRoot });
    },
  });
}

export function useDeleteGiftMessageMutation() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteGiftMessage(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: messageKeys.adminAllRoot });
      qc.invalidateQueries({ queryKey: ["messages-by-gift"] });
    },
  });
}
