import {
  keepPreviousData,
  useQuery,
  type QueryClient,
  type UseQueryReturnType,
} from "@tanstack/vue-query";
import { computed, ref, type Ref } from "vue";

import { getServices, getSpans, type SpansPage } from "./api";
import type { SpansListParams } from "./types";

export const spanKeys = {
  all: ["spans"] as const,
  lists: () => [...spanKeys.all, "list"] as const,
  list: (params: SpansListParams) => [...spanKeys.lists(), params] as const,
  services: () => [...spanKeys.all, "services"] as const,
};

const LIVE_REFETCH_MS = 10000;

export function useSpansQuery(
  params: Ref<SpansListParams>,
  live: Ref<boolean> = ref(false),
): UseQueryReturnType<SpansPage, Error> {
  return useQuery({
    queryKey: computed(() => spanKeys.list(params.value)),
    queryFn: (): Promise<SpansPage> => getSpans(params.value),
    placeholderData: keepPreviousData,
    refetchInterval: computed(() => (live.value ? LIVE_REFETCH_MS : false)),
  });
}

export function useServicesQuery(): UseQueryReturnType<string[], Error> {
  return useQuery({
    queryKey: spanKeys.services(),
    queryFn: getServices,
  });
}

export async function prefetchSpans(queryClient: QueryClient): Promise<void> {
  await queryClient.prefetchQuery({
    queryKey: spanKeys.services(),
    queryFn: getServices,
  });
}
