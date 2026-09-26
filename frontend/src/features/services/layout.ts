import type { ServiceEdge } from "@/shared/types/entities/service";

export const MAP_NODE_W = 138;
export const MAP_NODE_H = 38;
export const MAP_COL_W = 172;
export const MAP_MARGIN_X = 26;
const MAP_ROW_GAP = 26;
const MAP_MARGIN_TOP = 24;
const MAP_MARGIN_BOTTOM = 24;
const MAX_TIER_ITERATIONS = 500;

export interface MapNodePosition {
  x: number;
  y: number;
}

export interface MapEdgePath {
  from: string;
  to: string;
  d: string;
}

export interface MapLayout {
  width: number;
  height: number;
  positions: Record<string, MapNodePosition>;
  tiers: Record<string, number>;
  columns: string[][];
  maxTier: number;
  edgePaths: MapEdgePath[];
}

function edgePathD(from: MapNodePosition, to: MapNodePosition): string {
  const x1 = from.x + MAP_NODE_W;
  const y1 = from.y + MAP_NODE_H / 2;
  const x2 = to.x;
  const y2 = to.y + MAP_NODE_H / 2;
  const mx = (x1 + x2) / 2;
  return `M${x1} ${y1} C${mx} ${y1}, ${mx} ${y2}, ${x2} ${y2}`;
}

export function computeMapLayout(
  nodeIds: string[],
  edges: ServiceEdge[],
): MapLayout {
  const empty: MapLayout = {
    width: MAP_MARGIN_X * 2,
    height: MAP_MARGIN_TOP + MAP_MARGIN_BOTTOM,
    positions: {},
    tiers: {},
    columns: [],
    maxTier: 0,
    edgePaths: [],
  };
  const ids = [...new Set(nodeIds)].sort();
  if (ids.length === 0) return empty;

  const outgoing = new Map<string, string[]>();
  const incoming = new Map<string, number>();
  for (const id of ids) {
    outgoing.set(id, []);
    incoming.set(id, 0);
  }
  for (const edge of edges) {
    if (edge.from === edge.to) continue;
    const targets = outgoing.get(edge.from);
    const count = incoming.get(edge.to);
    if (targets === undefined || count === undefined) continue;
    targets.push(edge.to);
    incoming.set(edge.to, count + 1);
  }

  const tiers = new Map<string, number>();
  const queue: string[] = [];
  for (const id of ids) {
    if ((incoming.get(id) ?? 0) === 0) {
      tiers.set(id, 0);
      queue.push(id);
    }
  }
  if (queue.length === 0) {
    for (const id of ids) {
      tiers.set(id, 0);
      queue.push(id);
    }
  }
  let guard = 0;
  while (queue.length > 0 && guard < MAX_TIER_ITERATIONS) {
    guard += 1;
    const current = queue.shift();
    if (current === undefined) break;
    const currentTier = tiers.get(current) ?? 0;
    for (const next of outgoing.get(current) ?? []) {
      const candidate = currentTier + 1;
      if (candidate > (tiers.get(next) ?? -1)) {
        tiers.set(next, candidate);
        queue.push(next);
      }
    }
  }

  let maxTier = 0;
  for (const id of ids) {
    const tier = tiers.get(id) ?? 0;
    if (tier > maxTier) maxTier = tier;
  }
  const columns: string[][] = Array.from({ length: maxTier + 1 }, (_, index) =>
    ids.filter((id) => (tiers.get(id) ?? 0) === index),
  );

  let rowsMax = 0;
  for (const column of columns) {
    if (column.length > rowsMax) rowsMax = column.length;
  }
  const width = MAP_MARGIN_X * 2 + maxTier * MAP_COL_W + MAP_NODE_W;
  const height =
    MAP_MARGIN_TOP +
    MAP_MARGIN_BOTTOM +
    Math.max(rowsMax, 1) * (MAP_NODE_H + MAP_ROW_GAP) -
    MAP_ROW_GAP;

  const positions: Record<string, MapNodePosition> = {};
  columns.forEach((column, columnIndex) => {
    const columnHeight =
      column.length * (MAP_NODE_H + MAP_ROW_GAP) - MAP_ROW_GAP;
    const startY =
      MAP_MARGIN_TOP +
      (height - MAP_MARGIN_TOP - MAP_MARGIN_BOTTOM - columnHeight) / 2;
    column.forEach((id, rowIndex) => {
      positions[id] = {
        x: MAP_MARGIN_X + columnIndex * MAP_COL_W,
        y: startY + rowIndex * (MAP_NODE_H + MAP_ROW_GAP),
      };
    });
  });

  const tiersRecord: Record<string, number> = {};
  for (const id of ids) tiersRecord[id] = tiers.get(id) ?? 0;

  const edgePaths: MapEdgePath[] = [];
  for (const edge of edges) {
    const from = positions[edge.from];
    const to = positions[edge.to];
    if (from === undefined || to === undefined) continue;
    edgePaths.push({ from: edge.from, to: edge.to, d: edgePathD(from, to) });
  }

  return {
    width,
    height,
    positions,
    tiers: tiersRecord,
    columns,
    maxTier,
    edgePaths,
  };
}
