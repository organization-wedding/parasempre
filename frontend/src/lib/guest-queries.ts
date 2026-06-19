import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createGuest,
  deleteGuest,
  getGuest,
  getGuestStats,
  importGuests,
  listGuests,
  updateGuest,
} from "./api";
import type { CreateGuestInput, UpdateGuestInput } from "../types/guest";

type GuestListParams = {
  page?: number;
  limit?: number;
  search?: string;
  relationship?: string;
  attending?: string;
};

export const guestQueryKeys = {
  root: ["guests"] as const,
  list: (params?: GuestListParams) => ["guests", "list", params] as const,
  stats: () => ["guests", "stats"] as const,
  detail: (id: number) => ["guest", id] as const,
};

export function useGuestsQuery(params?: GuestListParams) {
  return useQuery({
    queryKey: guestQueryKeys.list(params),
    queryFn: () => listGuests(params),
  });
}

export function useGuestStatsQuery() {
  return useQuery({
    queryKey: guestQueryKeys.stats(),
    queryFn: () => getGuestStats(),
  });
}

export function useGuestQuery(id: number, enabled = true) {
  return useQuery({
    queryKey: guestQueryKeys.detail(id),
    queryFn: () => getGuest(id),
    enabled,
  });
}

export function useCreateGuestMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (input: CreateGuestInput) => createGuest(input),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: guestQueryKeys.root });
    },
  });
}

export function useUpdateGuestMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: UpdateGuestInput }) => updateGuest(id, input),
    onSuccess: (guest) => {
      queryClient.setQueryData(guestQueryKeys.detail(guest.id), guest);
      queryClient.invalidateQueries({ queryKey: guestQueryKeys.root });
    },
  });
}

export function useDeleteGuestMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteGuest(id),
    onSuccess: (_, id) => {
      queryClient.removeQueries({ queryKey: guestQueryKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: guestQueryKeys.root });
    },
  });
}

export function useDeleteGuestsMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (ids: number[]) => {
      await Promise.all(ids.map((id) => deleteGuest(id)));
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: guestQueryKeys.root });
    },
  });
}

export function useImportGuestsMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => importGuests(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: guestQueryKeys.root });
    },
  });
}
