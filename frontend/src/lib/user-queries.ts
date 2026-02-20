import { useEffect, useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { getUserMe, getUserRacf } from "./api";

export function useCurrentRacf() {
  const [racf, setRacf] = useState(getUserRacf());
  useEffect(() => {
    function onChanged() {
      setRacf(getUserRacf());
    }
    window.addEventListener("racf-changed", onChanged);
    return () => window.removeEventListener("racf-changed", onChanged);
  }, []);
  return racf;
}

export function useUserMeQuery(enabled = true) {
  const racf = useCurrentRacf();
  return useQuery({
    queryKey: ["user-me", racf],
    queryFn: getUserMe,
    enabled: enabled && !!racf,
    staleTime: 5 * 60 * 1000,
    retry: false,
  });
}
