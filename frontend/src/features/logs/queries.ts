import {
  useInfiniteQuery,
  useQuery,
  type InfiniteData,
  type QueryClient,
  type UseInfiniteQueryReturnType,
  type UseQueryReturnType,
} from "@tanstack/vue-query";
import { computed, ref, type Ref } from "vue";

import { getLogs, getServices, type LogsPage } from "./api";
import type { LogsFilters } from "./types";

export const logKeys = {
  all: ["logs"] as const,
  lists: () => [...logKeys.all, "list"] as const,
  list: (params: LogsFilters) => [...logKeys.lists(), params] as const,
  services: () => [...logKeys.all, "services"] as const,
};

const LIVE_REFETCH_MS = 10000;

export function useInfiniteLogsQuery(
  params: Ref<LogsFilters>,
  live: Ref<boolean> = ref(false),
): UseInfiniteQueryReturnType<InfiniteData<LogsPage>, Error> {
  return useInfiniteQuery({
    queryKey: computed(() => logKeys.list(params.value)),
    queryFn: ({
      pageParam,
    }: {
      pageParam: string | undefined;
    }): Promise<LogsPage> =>
      getLogs(
        pageParam === undefined
          ? { ...params.value }
          : { ...params.value, cursor: pageParam },
      ),
    initialPageParam: undefined as string | undefined,
    getNextPageParam: (lastPage) => lastPage.nextCursor ?? undefined,
    refetchInterval: computed(() => (live.value ? LIVE_REFETCH_MS : false)),
  });
}

export function useLogsServicesQuery(): UseQueryReturnType<string[], Error> {
  return useQuery({
    queryKey: logKeys.services(),
    queryFn: getServices,
  });
}

export async function prefetchLogs(queryClient: QueryClient): Promise<void> {
  await queryClient.prefetchQuery({
    queryKey: logKeys.services(),
    queryFn: getServices,
  });
}
