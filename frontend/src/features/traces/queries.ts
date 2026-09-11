import {
  keepPreviousData,
  useQuery,
  type QueryClient,
  type UseQueryReturnType,
} from "@tanstack/vue-query";
import { computed, ref, type Ref } from "vue";

import { getEnvironments, getServices, getTraces, type TracesPage } from "./api";
import { DEFAULT_TRACES_FILTERS, type TracesListParams } from "./types";

export const traceKeys = {
  all: ["traces"] as const,
  lists: () => [...traceKeys.all, "list"] as const,
  list: (params: TracesListParams) => [...traceKeys.lists(), params] as const,
  services: () => [...traceKeys.all, "services"] as const,
  environments: () => [...traceKeys.all, "environments"] as const,
};

const LIVE_REFETCH_MS = 10000;

export function useTracesQuery(
  params: Ref<TracesListParams>,
  live: Ref<boolean> = ref(false),
): UseQueryReturnType<TracesPage, Error> {
  return useQuery({
    queryKey: computed(() => traceKeys.list(params.value)),
    queryFn: (): Promise<TracesPage> => getTraces(params.value),
    placeholderData: keepPreviousData,
    refetchInterval: computed(() => (live.value ? LIVE_REFETCH_MS : false)),
  });
}

export function useTracesServicesQuery(): UseQueryReturnType<string[], Error> {
  return useQuery({
    queryKey: traceKeys.services(),
    queryFn: getServices,
  });
}

export function useTracesEnvironmentsQuery(): UseQueryReturnType<
  string[],
  Error
> {
  return useQuery({
    queryKey: traceKeys.environments(),
    queryFn: getEnvironments,
  });
}

export async function prefetchTraces(queryClient: QueryClient): Promise<void> {
  await Promise.all([
    queryClient.prefetchQuery({
      queryKey: traceKeys.list({ ...DEFAULT_TRACES_FILTERS }),
      queryFn: (): Promise<TracesPage> =>
        getTraces({ ...DEFAULT_TRACES_FILTERS }),
    }),
    queryClient.prefetchQuery({
      queryKey: traceKeys.services(),
      queryFn: getServices,
    }),
    queryClient.prefetchQuery({
      queryKey: traceKeys.environments(),
      queryFn: getEnvironments,
    }),
  ]);
}
