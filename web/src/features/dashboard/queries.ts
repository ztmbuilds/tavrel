import { useQuery, type UseQueryReturnType } from "@tanstack/vue-query";
import { computed, type Ref } from "vue";

import { resolveWindow } from "@/shared/composables/useGlobalTimeRange";

import { getDashboardSummary } from "./api";
import type { DashboardRange, DashboardSummary } from "./types";

export const dashboardKeys = {
  all: ["dashboard"] as const,
  summary: (range: DashboardRange) => [...dashboardKeys.all, range] as const,
};

const REFRESH_INTERVAL_MS = 30000;

export function useDashboardSummaryQuery(
  range: Ref<DashboardRange>,
): UseQueryReturnType<DashboardSummary, Error> {
  return useQuery({
    queryKey: computed(() => dashboardKeys.summary(range.value)),
    queryFn: (): Promise<DashboardSummary> => {
      const now = new Date();
      const window = resolveWindow({ range: range.value }, now.getTime());
      return getDashboardSummary(
        window.start ?? now.toISOString(),
        now.toISOString(),
      );
    },
    refetchInterval: REFRESH_INTERVAL_MS,
  });
}
