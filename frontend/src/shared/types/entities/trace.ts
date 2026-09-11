const TraceStatus = {
  Ok: "ok",
  Error: "error",
} as const;
export type TraceStatus = (typeof TraceStatus)[keyof typeof TraceStatus];

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
