import { useQuery, type UseQueryReturnType } from "@tanstack/vue-query";
import { computed, type Ref } from "vue";

import {
  resolveWindow,
  type SharedTime,
} from "@/shared/composables/useGlobalTimeRange";
import type { ServiceGraph } from "@/shared/types/entities/service";

import { getServiceGraph } from "./api";

export const serviceGraphKeys = {
  all: ["service-graph"] as const,
  graph: (params: SharedTime) => [...serviceGraphKeys.all, params] as const,
};

export function useServiceGraphQuery(
  params: Ref<SharedTime>,
): UseQueryReturnType<ServiceGraph, Error> {
  return useQuery({
    queryKey: computed(() => serviceGraphKeys.graph(params.value)),
    queryFn: (): Promise<ServiceGraph> => {
      const window = resolveWindow(params.value);
      return getServiceGraph(window);
    },
  });
}
