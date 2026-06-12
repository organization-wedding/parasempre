import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  commitGiftImport,
  createGift,
  deleteGift,
  getGift,
  isNotFoundError,
  listGifts,
  previewGiftImport,
  scrapeGiftURL,
  updateGift,
} from "./api";
import type { CreateGiftInput, UpdateGiftInput } from "../types/gift";

export type GiftSort = "price_asc" | "price_desc";

export type GiftListParams = {
  page?: number;
  limit?: number;
  search?: string;
  price_min?: number;
  price_max?: number;
  sort?: GiftSort;
};

export const giftQueryKeys = {
  all: (params?: GiftListParams) => ["gifts", params] as const,
  detail: (id: number) => ["gift", id] as const,
};

export function useGiftsQuery(params?: GiftListParams) {
  return useQuery({
    queryKey: giftQueryKeys.all(params),
    queryFn: () => listGifts(params),
    // Preserva a página anterior enquanto a próxima carrega — evita flash
    // de spinner ao paginar numa lista pública.
    placeholderData: (prev) => prev,
  });
}

export function useGiftQuery(id: number, enabled = true) {
  return useQuery({
    queryKey: giftQueryKeys.detail(id),
    queryFn: () => getGift(id),
    enabled,
    retry: (failureCount, error) => !isNotFoundError(error) && failureCount < 2,
  });
}

export function useCreateGiftMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateGiftInput) => createGift(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ["gifts"] });
    },
  });
}

export function useUpdateGiftMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: UpdateGiftInput }) =>
      updateGift(id, input),
    onSuccess: (gift) => {
      queryClient.setQueryData(giftQueryKeys.detail(gift.id), gift);
      queryClient.invalidateQueries({ queryKey: ["gifts"] });
    },
  });
}

export function useDeleteGiftMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteGift(id),
    onSuccess: (_, id) => {
      queryClient.removeQueries({ queryKey: giftQueryKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: ["gifts"] });
    },
  });
}

export function usePreviewGiftImportMutation() {
  return useMutation({
    mutationFn: (file: File) => previewGiftImport(file),
  });
}

export function useCommitGiftImportMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (rows: CreateGiftInput[]) => commitGiftImport(rows),
    onSuccess: (resp) => {
      if (resp.created > 0) {
        queryClient.invalidateQueries({ queryKey: ["gifts"] });
      }
    },
  });
}

export function useScrapeGiftURLMutation() {
  return useMutation({
    mutationFn: (url: string) => scrapeGiftURL(url),
  });
}
