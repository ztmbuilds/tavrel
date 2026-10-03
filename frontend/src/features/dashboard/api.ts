import { apiFetch } from "@/shared/api/client";
import { toLog, toTrace, type LogDto, type TraceDto } from "@/shared/api/otel";
import type { Log } from "@/shared/types/entities/log";

import type { DashboardSummary, TraceTimeBucket } from "./types";

interface DashboardStatsDto {
  overview: {
    traceCount: number;
    errorTraceCount: number;
    activeServiceCount: number;
  };
  traceBuckets: TraceTimeBucket[];
  recentErrorTraces: TraceDto[];
  recentErrorLogs: LogDto[];
}

interface DashboardStatsResponse {
  message: string;
  data: DashboardStatsDto;
  meta: { nextCursor: string | null };
}

export async function getDashboardSummary(
  start: string,
  end: string,
): Promise<DashboardSummary> {
  const response = await apiFetch<DashboardStatsResponse>("/api/v1/stats", {
    query: { start, end },
  });
  return {
    overview: response.data.overview,
    traceBuckets: response.data.traceBuckets,
    recentErrorTraces: response.data.recentErrorTraces.map((trace: TraceDto) =>
      toTrace(trace),
    ),
    recentErrorLogs: response.data.recentErrorLogs.map((log: LogDto): Log =>
      toLog(log),
    ),
  };
}
