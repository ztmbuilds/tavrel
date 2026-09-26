import dagre, {
  type EdgeLabel,
  type GraphLabel,
  type NodeLabel,
} from "@dagrejs/dagre";

import type { ServiceEdge } from "@/shared/types/entities/service";

export const MAP_NODE_W = 184;
export const MAP_NODE_H = 52;

const SERVICE_LABEL_CHAR_WIDTH = 8;
const NODE_HORIZONTAL_PADDING = 36;
const MAP_RANK_GAP = 104;
const MAP_NODE_GAP = 48;
const MAP_EDGE_GAP = 32;
const MAP_MARGIN = 48;

export interface MapNodePosition {
  x: number;
  y: number;
}

export interface MapLayout {
  nodeIds: string[];
  positions: Record<string, MapNodePosition>;
  nodeWidths: Record<string, number>;
  edges: MapEdgeLayout[];
}

export interface MapEdgeLayout extends ServiceEdge {
  points: MapNodePosition[];
  feedback: boolean;
}

export function computeMapLayout(
  nodeIds: string[],
  edges: ServiceEdge[],
): MapLayout {
  const ids = [
    ...new Set([...nodeIds, ...edges.flatMap(({ from, to }) => [from, to])]),
  ]
    .filter((id) => id !== "")
    .sort();
  const idSet = new Set(ids);
  const graphEdges = edges
    .filter((edge) => idSet.has(edge.from) && idSet.has(edge.to))
    .sort((a, b) => a.from.localeCompare(b.from) || a.to.localeCompare(b.to));
  const nodeWidths: Record<string, number> = {};
  const graph = new dagre.graphlib.Graph<GraphLabel, NodeLabel, EdgeLabel>();

  graph.setGraph({
    rankdir: "LR",
    ranksep: MAP_RANK_GAP,
    nodesep: MAP_NODE_GAP,
    edgesep: MAP_EDGE_GAP,
    marginx: MAP_MARGIN,
    marginy: MAP_MARGIN,
  });
  graph.setDefaultEdgeLabel((): EdgeLabel => ({}));

  for (const id of ids) {
    const width = Math.max(
      MAP_NODE_W,
      Math.ceil(id.length * SERVICE_LABEL_CHAR_WIDTH + NODE_HORIZONTAL_PADDING),
    );
    nodeWidths[id] = width;
    graph.setNode(id, { width, height: MAP_NODE_H });
  }

  for (const edge of graphEdges)
    graph.setEdge(edge.from, edge.to, { weight: 1 });

  dagre.layout(graph);

  const positions: Record<string, MapNodePosition> = {};
  for (const id of ids) {
    const node = graph.node(id);
    if (node?.x === undefined || node.y === undefined) continue;
    positions[id] = {
      x: Math.round(node.x - (nodeWidths[id] ?? MAP_NODE_W) / 2),
      y: Math.round(node.y - MAP_NODE_H / 2),
    };
  }

  const edgeLayouts = graphEdges.map((edge) => {
    const route = graph.edge(edge.from, edge.to);
    const source = positions[edge.from];
    const target = positions[edge.to];
    return {
      ...edge,
      points: route?.points?.map(({ x, y }) => ({ x, y })) ?? [],
      feedback:
        source !== undefined &&
        target !== undefined &&
        target.x + (nodeWidths[edge.to] ?? MAP_NODE_W) / 2 <
          source.x + (nodeWidths[edge.from] ?? MAP_NODE_W) / 2,
    };
  });

  return { nodeIds: ids, positions, nodeWidths, edges: edgeLayouts };
}
