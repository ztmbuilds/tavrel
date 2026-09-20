import {
  keepPreviousData,
  useQuery,
  type UseQueryReturnType,
} from "@tanstack/vue-query";
import { computed, ref, type Ref } from "vue";

import { resolveWindow } from "@/shared/composables/useGlobalTimeRange";

import { getSpans, type SpansPage } from "./api";
import type { SpansListParams } from "./types";

export const spanKeys = {
  all: ["spans"] as const,
  lists: () => [...spanKeys.all, "list"] as const,
  list: (params: SpansListParams) => [...spanKeys.lists(), params] as const,
};

const LIVE_REFETCH_MS = 10000;

export function useSpansQuery(
  params: Ref<SpansListParams>,
  live: Ref<boolean> = ref(false),
): UseQueryReturnType<SpansPage, Error> {
  return useQuery({
    queryKey: computed(() => spanKeys.list(params.value)),
    queryFn: (): Promise<SpansPage> =>
      getSpans({ ...params.value, ...resolveWindow(params.value) }),
    placeholderData: keepPreviousData,
    refetchInterval: computed(() => (live.value ? LIVE_REFETCH_MS : false)),
  });
}
