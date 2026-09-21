import {
  useInfiniteQuery,
  type InfiniteData,
  type UseInfiniteQueryReturnType,
} from "@tanstack/vue-query";
import { computed, ref, type Ref } from "vue";

import { resolveWindow } from "@/shared/composables/useGlobalTimeRange";

import { getSpans, type SpansPage } from "./api";
import type { SpansFilters } from "./types";

export const spanKeys = {
  all: ["spans"] as const,
  lists: () => [...spanKeys.all, "list"] as const,
  list: (params: SpansFilters) => [...spanKeys.lists(), params] as const,
};

const LIVE_REFETCH_MS = 10000;

export function useInfiniteSpansQuery(
  params: Ref<SpansFilters>,
  live: Ref<boolean> = ref(false),
): UseInfiniteQueryReturnType<InfiniteData<SpansPage>, Error> {
  return useInfiniteQuery({
    queryKey: computed(() => spanKeys.list(params.value)),
    queryFn: ({
      pageParam,
    }: {
      pageParam: string | undefined;
    }): Promise<SpansPage> => {
      const window = resolveWindow(params.value);
      return getSpans(
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
