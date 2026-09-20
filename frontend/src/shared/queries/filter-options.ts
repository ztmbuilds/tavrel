import {
  useQuery,
  type QueryClient,
  type UseQueryReturnType,
} from "@tanstack/vue-query";

import { apiFetch } from "@/shared/api/client";
import type { StringListResponse } from "@/shared/api/otel";

export const servicesKey = ["services"] as const;

export const environmentsKey = ["environments"] as const;

export async function getServices(): Promise<string[]> {
  const response = await apiFetch<StringListResponse>("/api/v1/services");
  return response.data;
}

export async function getEnvironments(): Promise<string[]> {
  const response =
    await apiFetch<StringListResponse>("/api/v1/environments");
  return response.data;
}

export function useServicesQuery(): UseQueryReturnType<string[], Error> {
  return useQuery({
    queryKey: servicesKey,
    queryFn: getServices,
  });
}

export function useEnvironmentsQuery(): UseQueryReturnType<string[], Error> {
  return useQuery({
    queryKey: environmentsKey,
    queryFn: getEnvironments,
  });
}

export async function prefetchFilterOptions(
  queryClient: QueryClient,
): Promise<void> {
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: servicesKey,
      queryFn: getServices,
    }),
    queryClient.prefetchQuery({
      queryKey: environmentsKey,
      queryFn: getEnvironments,
    }),
  ]);
}
