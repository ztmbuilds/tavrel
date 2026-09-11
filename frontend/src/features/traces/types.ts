const TraceStatusFilter = {
  All: "all",
  Error: "error",
  Ok: "ok",
} as const;
export type TraceStatusFilter =
  (typeof TraceStatusFilter)[keyof typeof TraceStatusFilter];

const TraceOrder = {
  Asc: "asc",
  Desc: "desc",
} as const;
export type TraceOrder = (typeof TraceOrder)[keyof typeof TraceOrder];

export interface TracesFilters {
  search: string;
  status: TraceStatusFilter;
  service: string;
  environment: string;
  order: TraceOrder;
  start?: string;
  end?: string;
}

export interface TracesListParams extends TracesFilters {
  cursor?: string;
}

export const DEFAULT_TRACES_FILTERS: TracesFilters = {
  search: "",
  status: "all",
  service: "all",
  environment: "all",
  order: "desc",
};
