import { useQuery } from "@tanstack/react-query";
import {
  listMyPurchases,
  getMyPurchase,
  listAdminTransactions,
  adminTransactionsSummary,
} from "./api";

export const txKeys = {
  myList: (params?: { page?: number; limit?: number }) => ["my-purchases", params] as const,
  myDetail: (id: number) => ["my-purchase", id] as const,
  adminList: (filter?: { status?: string; gift_id?: number; page?: number; limit?: number }) =>
    ["admin-transactions", filter] as const,
  adminSummary: () => ["admin-summary"] as const,
};

export function useMyPurchasesQuery(params?: { page?: number; limit?: number }) {
  return useQuery({
    queryKey: txKeys.myList(params),
    queryFn: () => listMyPurchases(params),
    placeholderData: (prev) => prev,
  });
}

export function useMyPurchaseQuery(id: number, enabled: boolean) {
  return useQuery({
    queryKey: txKeys.myDetail(id),
    queryFn: () => getMyPurchase(id),
    enabled,
  });
}

export function useAdminTransactionsQuery(filter?: {
  status?: string;
  gift_id?: number;
  page?: number;
  limit?: number;
}) {
  return useQuery({
    queryKey: txKeys.adminList(filter),
    queryFn: () => listAdminTransactions(filter),
    placeholderData: (prev) => prev,
  });
}

export function useAdminSummaryQuery() {
  return useQuery({
    queryKey: txKeys.adminSummary(),
    queryFn: () => adminTransactionsSummary(),
    staleTime: 30 * 1000,
  });
}
