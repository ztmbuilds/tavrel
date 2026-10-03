import type { Log } from "@/shared/types/entities/log";
import type { Trace } from "@/shared/types/entities/trace";

export const DASHBOARD_WINDOWS = [
  { value: "15m", label: "Last 15 minutes" },
  { value: "1h", label: "Last 1 hour" },
  { value: "6h", label: "Last 6 hours" },
  { value: "24h", label: "Last 24 hours" },
  { value: "7d", label: "Last 7 days" },
] as const;

export const DEFAULT_DASHBOARD_RANGE = "24h";

export type DashboardRange = (typeof DASHBOARD_WINDOWS)[number]["value"];

export interface DashboardOverview {
  traceCount: number;
  errorTraceCount: number;
  activeServiceCount: number;
}

export interface TraceTimeBucket {
  startTime: string;
  endTime: string;
  okCount: number;
  errorCount: number;
}

export interface DashboardSummary {
  overview: DashboardOverview;
  traceBuckets: TraceTimeBucket[];
  recentErrorTraces: Trace[];
  recentErrorLogs: Log[];
}
