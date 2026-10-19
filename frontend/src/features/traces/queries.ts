import {
  useInfiniteQuery,
  useQuery,
  type InfiniteData,
  type UseInfiniteQueryReturnType,
  type UseQueryReturnType,
} from "@tanstack/vue-query";
import { computed, ref, type Ref } from "vue";

import { ApiError } from "@/shared/api/client";
import { resolveWindow } from "@/shared/composables/useGlobalTimeRange";
import type { Log } from "@/shared/types/entities/log";
import type { SpanLink } from "@/shared/types/entities/span-link";

import {
  getSpanLinks,
  getSpanLogs,
  getTraceDetail,
  getTraces,
  type TracesPage,
} from "./api";
import type { SpanLogWindow, TraceDetail, TracesFilters } from "./types";

export const traceKeys = {
  all: ["traces"] as const,
  lists: () => [...traceKeys.all, "list"] as const,
  list: (params: TracesFilters) => [...traceKeys.lists(), params] as const,
  details: () => [...traceKeys.all, "detail"] as const,
  detail: (id: string) => [...traceKeys.details(), id] as const,
  spanLogs: (traceId: string, window: SpanLogWindow | null) =>
    [...traceKeys.all, "span-logs", traceId, window] as const,
  spanLinks: (traceId: string, spanId: string | null) =>
    [...traceKeys.all, "span-links", traceId, spanId] as const,
};

const LIVE_REFETCH_MS = 10000;

export function useInfiniteTracesQuery(
  params: Ref<TracesFilters>,
  live: Ref<boolean> = ref(false),
): UseInfiniteQueryReturnType<InfiniteData<TracesPage>, Error> {
  return useInfiniteQuery({
    queryKey: computed(() => traceKeys.list(params.value)),
    queryFn: ({
      pageParam,
    }: {
      pageParam: string | undefined;
    }): Promise<TracesPage> => {
      const window = resolveWindow(params.value);
      return getTraces(
        pageParam === undefined
          ? { ...params.value, ...window }
          : { ...params.value, ...window, cursor: pageParam },
      );
    },
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
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
      return getSpanLogs(
        traceId.value,
        current.spanId,
        current.start,
        current.end,
      );
    },
    enabled: computed(() => window.value !== null),
  });
}

export function useSpanLinksQuery(
  traceId: Ref<string>,
  spanId: Ref<string | null>,
): UseQueryReturnType<SpanLink[], Error> {
  return useQuery({
    queryKey: computed(() => traceKeys.spanLinks(traceId.value, spanId.value)),
    queryFn: async (): Promise<SpanLink[]> => {
      const current = spanId.value;
      if (current === null) return [];
      return getSpanLinks(traceId.value, current);
    },
    enabled: computed(() => spanId.value !== null),
  });
}
