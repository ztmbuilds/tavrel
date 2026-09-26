import { apiFetch } from "@/shared/api/client";
import type {
  ServiceEdge,
  ServiceGraph,
  ServiceNode,
} from "@/shared/types/entities/service";

interface ServiceNodeDto {
  id: string;
}

interface ServiceEdgeDto {
  from: string;
  to: string;
}

interface ServiceGraphData {
  nodes: ServiceNodeDto[];
  edges: ServiceEdgeDto[];
}

interface ServiceGraphResponse {
  message: string;
  data: ServiceGraphData;
  meta: { nextCursor: string | null };
}

export interface ServiceGraphWindow {
  start?: string;
  end?: string;
}

function toServiceNode(dto: ServiceNodeDto): ServiceNode {
  return { id: dto.id };
}

function toServiceEdge(dto: ServiceEdgeDto): ServiceEdge {
  return { from: dto.from, to: dto.to };
}

export async function getServiceGraph(
  window: ServiceGraphWindow,
): Promise<ServiceGraph> {
  const query: Record<string, string | undefined> = {};
  if (window.start !== undefined) query.start = window.start;
  if (window.end !== undefined) query.end = window.end;

  const response = await apiFetch<ServiceGraphResponse>(
    "/api/v1/service-graph",
    {
      query,
    },
  );
  return {
    nodes: response.data.nodes.map(toServiceNode),
    edges: response.data.edges.map(toServiceEdge),
  };
}
