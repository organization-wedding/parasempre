import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import {
  batchConfirmFamily,
  cancelGuest,
  cancelWholeFamily,
  confirmGuest,
  confirmWholeFamily,
  getMyFamily,
} from "./api";
import { useAuth } from "./auth-queries";
import { guestQueryKeys } from "./guest-queries";
import type { BatchConfirmInput } from "../types/guest";

export const familyQueryKeys = {
  my: ["my-family"] as const,
};

export function useMyFamilyQuery(enabled = true) {
  const { token } = useAuth();
  return useQuery({
    queryKey: familyQueryKeys.my,
    queryFn: getMyFamily,
    enabled: enabled && !!token,
    staleTime: 30 * 1000,
  });
}

function useInvalidateFamily() {
  const queryClient = useQueryClient();
  return () => {
    queryClient.invalidateQueries({ queryKey: familyQueryKeys.my });
    queryClient.invalidateQueries({ queryKey: guestQueryKeys.all() });
  };
}

export function useConfirmGuestMutation() {
  const invalidate = useInvalidateFamily();
  return useMutation({
    mutationFn: (id: number) => confirmGuest(id),
    onSuccess: () => invalidate(),
  });
}

export function useCancelGuestMutation() {
  const invalidate = useInvalidateFamily();
  return useMutation({
    mutationFn: (id: number) => cancelGuest(id),
    onSuccess: () => invalidate(),
  });
}

export function useBatchFamilyMutation() {
  const invalidate = useInvalidateFamily();
  return useMutation({
    mutationFn: (input: BatchConfirmInput) => batchConfirmFamily(input),
    onSuccess: () => invalidate(),
  });
}

export function useConfirmWholeFamilyMutation() {
  const invalidate = useInvalidateFamily();
  return useMutation({
    mutationFn: (familyGroup: number) => confirmWholeFamily(familyGroup),
    onSuccess: () => invalidate(),
  });
}

export function useCancelWholeFamilyMutation() {
  const invalidate = useInvalidateFamily();
  return useMutation({
    mutationFn: (familyGroup: number) => cancelWholeFamily(familyGroup),
    onSuccess: () => invalidate(),
  });
}
