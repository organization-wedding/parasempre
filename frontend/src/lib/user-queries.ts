import { useQuery } from "@tanstack/react-query";
import { getUserMe } from "./api";
import { useAuth } from "./auth-queries";

export function useUserMeQuery(enabled = true) {
  const { token } = useAuth();
  return useQuery({
    queryKey: ["user-me", token],
    queryFn: getUserMe,
    enabled: enabled && !!token,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
