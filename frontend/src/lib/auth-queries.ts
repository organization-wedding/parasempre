import { useEffect, useState } from "react";
import { useMutation } from "@tanstack/react-query";
import { getToken, getAuthRole, getAuthUracf, isAuthenticated } from "./auth";
import { sendOtp, verifyOtp, type TokenResponse } from "./api";

export function useAuth() {
  const [state, setState] = useState(() => ({
    token: getToken(),
    role: getAuthRole(),
    uracf: getAuthUracf(),
    isAuthenticated: isAuthenticated(),
  }));

  useEffect(() => {
    function onChanged() {
      setState({
        token: getToken(),
        role: getAuthRole(),
        uracf: getAuthUracf(),
        isAuthenticated: isAuthenticated(),
      });
    }
    window.addEventListener("auth-changed", onChanged);
    return () => window.removeEventListener("auth-changed", onChanged);
  }, []);

  return state;
}

export function useSendOtpMutation() {
  return useMutation({
    mutationFn: (phone: string) => sendOtp(phone),
  });
}

export function useVerifyOtpMutation() {
  return useMutation({
    mutationFn: ({ phone, code }: { phone: string; code: string }) =>
      verifyOtp(phone, code),
  });
}
