import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  createGuest,
  deleteGuest,
  getGuest,
  importGuests,
  listGuests,
  updateGuest,
} from "./api";
import type { CreateGuestInput, UpdateGuestInput } from "../types/guest";

export const guestQueryKeys = {
  all: ["guests"] as const,
  detail: (id: number) => ["guest", id] as const,
};

export function useGuestsQuery() {
  return useQuery({
    queryKey: guestQueryKeys.all,
    queryFn: listGuests,
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
      queryClient.invalidateQueries({ queryKey: guestQueryKeys.all });
    },
  });
}

export function useUpdateGuestMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: number; input: UpdateGuestInput }) => updateGuest(id, input),
    onSuccess: (guest) => {
      queryClient.setQueryData(guestQueryKeys.detail(guest.id), guest);
      queryClient.invalidateQueries({ queryKey: guestQueryKeys.all });
    },
  });
}

export function useDeleteGuestMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (id: number) => deleteGuest(id),
    onSuccess: (_, id) => {
      queryClient.removeQueries({ queryKey: guestQueryKeys.detail(id) });
      queryClient.invalidateQueries({ queryKey: guestQueryKeys.all });
    },
  });
}

export function useImportGuestsMutation() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: (file: File) => importGuests(file),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: guestQueryKeys.all });
    },
  });
}
