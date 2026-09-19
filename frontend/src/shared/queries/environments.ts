import { useQuery, type UseQueryReturnType } from "@tanstack/vue-query";

import { apiFetch } from "@/shared/api/client";
import type { StringListResponse } from "@/shared/api/otel";

export const environmentsKey = ["environments"] as const;

export async function getEnvironments(): Promise<string[]> {
  const response =
    await apiFetch<StringListResponse>("/api/v1/environments");
  return response.data;
}

export function useEnvironmentsQuery(): UseQueryReturnType<string[], Error> {
  return useQuery({
    queryKey: environmentsKey,
    queryFn: getEnvironments,
  });
}
