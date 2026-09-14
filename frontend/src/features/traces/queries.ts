import {
  keepPreviousData,
  useQuery,
  type QueryClient,
  type UseQueryReturnType,
} from "@tanstack/vue-query";
import { computed, ref, type Ref } from "vue";

import { ApiError } from "@/shared/api/client";
import type { Log } from "@/shared/types/entities/log";

import {
  getEnvironments,
  getServices,
  getSpanLogs,
  getTraceDetail,
  getTraces,
  type TracesPage,
} from "./api";
import { DEFAULT_TRACES_FILTERS, type SpanLogWindow, type TraceDetail, type TracesListParams } from "./types";

export const traceKeys = {
  all: ["traces"] as const,
  lists: () => [...traceKeys.all, "list"] as const,
  list: (params: TracesListParams) => [...traceKeys.lists(), params] as const,
  details: () => [...traceKeys.all, "detail"] as const,
  detail: (id: string) => [...traceKeys.details(), id] as const,
  spanLogs: (traceId: string, window: SpanLogWindow | null) =>
    [...traceKeys.all, "span-logs", traceId, window] as const,
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

export function useTraceDetailQuery(
  id: Ref<string>,
): UseQueryReturnType<TraceDetail, Error> {
  return useQuery({
    queryKey: computed(() => traceKeys.detail(id.value)),
    queryFn: (): Promise<TraceDetail> => getTraceDetail(id.value),
    retry: (count: number, error: Error): boolean => {
      if (error instanceof ApiError && error.isNotFound()) return false;
      return count < 1;
    },
  });
}

export function useTraceSpanLogsQuery(
  traceId: Ref<string>,
  window: Ref<SpanLogWindow | null>,
): UseQueryReturnType<Log[], Error> {
  return useQuery({
    queryKey: computed(() => traceKeys.spanLogs(traceId.value, window.value)),
    queryFn: async (): Promise<Log[]> => {
      const current = window.value;
      if (current === null) return [];
      return getSpanLogs(traceId.value, current.spanId, current.start, current.end);
    },
    enabled: computed(() => window.value !== null),
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
