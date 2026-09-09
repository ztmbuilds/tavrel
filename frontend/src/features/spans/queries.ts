import {
  keepPreviousData,
  useQuery,
  type QueryClient,
  type UseQueryReturnType,
} from "@tanstack/vue-query";
import { computed, type Ref } from "vue";

import { getServices, getSpans, type SpansPage } from "./api";
import { DEFAULT_SPANS_FILTERS, type SpansListParams } from "./types";

export const spanKeys = {
  all: ["spans"] as const,
  lists: () => [...spanKeys.all, "list"] as const,
  list: (params: SpansListParams) => [...spanKeys.lists(), params] as const,
  services: () => [...spanKeys.all, "services"] as const,
};

export function useSpansQuery(
  params: Ref<SpansListParams>,
): UseQueryReturnType<SpansPage, Error> {
  return useQuery({
    queryKey: computed(() => spanKeys.list(params.value)),
    queryFn: (): Promise<SpansPage> => getSpans(params.value),
    placeholderData: keepPreviousData,
  });
}

export function useServicesQuery(): UseQueryReturnType<string[], Error> {
  return useQuery({
    queryKey: spanKeys.services(),
    queryFn: getServices,
  });
}

export async function prefetchSpans(queryClient: QueryClient): Promise<void> {
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: spanKeys.list({ ...DEFAULT_SPANS_FILTERS }),
      queryFn: (): Promise<SpansPage> =>
        getSpans({ ...DEFAULT_SPANS_FILTERS }),
    }),
    queryClient.prefetchQuery({
      queryKey: spanKeys.services(),
      queryFn: getServices,
    }),
  ]);
}
