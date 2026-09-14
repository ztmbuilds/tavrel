export type TraceStatus = "ok" | "error";

export interface Trace {
  traceId: string;
  rootService: string;
  rootName: string;
  environment: string;
  startTime: string;
  endTime: string;
  spanCount: number;
  status: TraceStatus;
}
