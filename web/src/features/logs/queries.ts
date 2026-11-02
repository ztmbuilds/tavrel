import {
  useInfiniteQuery,
  type InfiniteData,
  type UseInfiniteQueryReturnType,
} from "@tanstack/vue-query";
import { computed, ref, type Ref } from "vue";

import { resolveWindow } from "@/shared/composables/useGlobalTimeRange";

import { getLogs, type LogsPage } from "./api";
import type { LogsFilters } from "./types";

export const logKeys = {
  all: ["logs"] as const,
  lists: () => [...logKeys.all, "list"] as const,
  list: (params: LogsFilters) => [...logKeys.lists(), params] as const,
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
    }): Promise<LogsPage> => {
      const window = resolveWindow(params.value);
      return getLogs(
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
