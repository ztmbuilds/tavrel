import {
  computed,
  nextTick,
  ref,
  watch,
  type ComputedRef,
  type Ref,
} from "vue";
import {
  MarkerType,
  useVueFlow,
  type Edge,
  type EdgeMouseEvent,
  type NodeDragEvent,
  type NodeMouseEvent,
} from "@vue-flow/core";

import {
  MAP_NODE_H,
  MAP_NODE_W,
  type MapLayout,
  type MapNodePosition,
} from "../layout";
import type { ServiceMapEdgeData, ServiceMapNodeData } from "../types";

const FIT_PADDING = 0.14;
const EDGE_DIMMED_OPACITY = 0.07;
const EDGE_DEFAULT_OPACITY = 0.72;
const EDGE_HOT_WIDTH = 1.8;
const EDGE_DEFAULT_WIDTH = 1.4;
const FEEDBACK_ROUTE_MARGIN = 48;
const FEEDBACK_ROUTE_LANE_GAP = 28;
const FEEDBACK_VIEW_MARGIN = 40;
const BACKWARD_EDGE_EPSILON = 1;

type MapSide = "left" | "right" | "top" | "bottom";
type EdgeHandles = {
  sourceHandle: string;
  targetHandle: string;
  sourceSide: MapSide;
  targetSide: MapSide;
};
type ServiceMapFlowNode = {
  id: string;
  type: "service";
  position: MapNodePosition;
  width: number;
  height: number;
  draggable: boolean;
  selectable: boolean;
  focusable: boolean;
  connectable: boolean;
  data: ServiceMapNodeData;
};
type ServiceMapFlowEdge = Edge<ServiceMapEdgeData>;

interface UseServiceMapFlowResult {
  nodes: Ref<ServiceMapFlowNode[]>;
  edges: ComputedRef<ServiceMapFlowEdge[]>;
  movementAnnouncement: Ref<string>;
  canZoomIn: ComputedRef<boolean>;
  canZoomOut: ComputedRef<boolean>;
  handleNodeDrag: (event: NodeDragEvent) => void;
  handleNodeMouseEnter: (event: NodeMouseEvent) => void;
  handleNodeMouseLeave: () => void;
  handleEdgeClick: (event: EdgeMouseEvent) => void;
  handleZoomIn: () => void;
  handleZoomOut: () => void;
  handleFitView: () => void;
  resetLayout: () => void;
}

export function useServiceMapFlow(
  layout: ComputedRef<MapLayout>,
  selected: ComputedRef<string | null>,
  query: ComputedRef<string>,
  selectService: (id: string) => void,
): UseServiceMapFlowResult {
  const { fitView, maxZoom, minZoom, updateNode, viewport, zoomIn, zoomOut } =
    useVueFlow();
  const hovered = ref<string | null>(null);
  const nodes = ref<ServiceMapFlowNode[]>([]);
  const nodePositions = ref<Record<string, MapNodePosition>>({});
  const hasManualChanges = ref(false);
  const movementAnnouncement = ref("");
  const trimmedQuery = computed(() => query.value.trim().toLowerCase());
  const canZoomIn = computed(() => viewport.value.zoom < maxZoom.value);
  const canZoomOut = computed(() => viewport.value.zoom > minZoom.value);

  const neighbors = computed((): Set<string> | null => {
    const focus = hovered.value ?? selected.value;
    if (focus === null) return null;
    const connected = new Set<string>([focus]);
    for (const edge of layout.value.edges) {
      if (edge.from === focus) connected.add(edge.to);
      if (edge.to === focus) connected.add(edge.from);
    }
    return connected;
  });

  const edges = computed((): ServiceMapFlowEdge[] => {
    const focus = hovered.value ?? selected.value;
    const backwardEdges = hasManualChanges.value
      ? layout.value.edges
          .filter((edge) => edge.feedback || isBackwardEdge(edge.from, edge.to))
          .sort(
            (a, b) => a.from.localeCompare(b.from) || a.to.localeCompare(b.to),
          )
      : [];
    const feedbackLanes = new Map(
      backwardEdges.map(
        (edge, index) => [edgeId(edge.from, edge.to), index] as const,
      ),
    );
    const graphTop = getGraphTop();

    return layout.value.edges.map((edge) => {
      const sourcePosition = nodePositions.value[edge.from];
      const targetPosition = nodePositions.value[edge.to];
      const sourceWidth = layout.value.nodeWidths[edge.from] ?? MAP_NODE_W;
      const targetWidth = layout.value.nodeWidths[edge.to] ?? MAP_NODE_W;
      const lane = feedbackLanes.get(edgeId(edge.from, edge.to));
      const handles = hasManualChanges.value
        ? lane === undefined
          ? getEdgeHandles(
              sourcePosition,
              targetPosition,
              sourceWidth,
              targetWidth,
            )
          : createHandles("top", "top")
        : getDagreEdgeHandles(
            edge.points,
            sourcePosition,
            targetPosition,
            sourceWidth,
            targetWidth,
          );
      const routePoints = hasManualChanges.value
        ? lane === undefined ||
          sourcePosition === undefined ||
          targetPosition === undefined
          ? []
          : feedbackRoutePoints(
              sourcePosition,
              targetPosition,
              sourceWidth,
              targetWidth,
              handles,
              graphTop,
              lane,
            )
        : edge.points.length > 2
          ? edge.points.slice(1, -1)
          : [];
      const isHot =
        trimmedQuery.value === "" &&
        focus !== null &&
        (edge.from === focus || edge.to === focus);
      const isDimmed =
        trimmedQuery.value !== ""
          ? !matches(edge.from, trimmedQuery.value) &&
            !matches(edge.to, trimmedQuery.value)
          : focus !== null && !isHot;

      return {
        id: edgeId(edge.from, edge.to),
        source: edge.from,
        target: edge.to,
        sourceHandle: handles.sourceHandle,
        targetHandle: handles.targetHandle,
        type: "service",
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: "var(--color-secondary)",
          width: 9,
          height: 9,
        },
        style: {
          stroke: isHot ? "var(--color-secondary)" : "var(--color-strong)",
          strokeWidth: isHot ? EDGE_HOT_WIDTH : EDGE_DEFAULT_WIDTH,
          opacity: isDimmed ? EDGE_DIMMED_OPACITY : EDGE_DEFAULT_OPACITY,
        },
        data: { routePoints },
        selectable: false,
        focusable: false,
        interactionWidth: 12,
        ariaLabel: `${edge.from} calls ${edge.to}`,
      };
    });
  });

  function matches(id: string, search: string): boolean {
    return id.toLowerCase().includes(search);
  }

  function isNodeDimmed(id: string): boolean {
    if (trimmedQuery.value !== "") return !matches(id, trimmedQuery.value);
    return neighbors.value !== null && !neighbors.value.has(id);
  }

  function moveNode(id: string, delta: MapNodePosition): void {
    const position = nodePositions.value[id];
    if (position === undefined) return;
    const nextPosition = {
      x: position.x + delta.x,
      y: position.y + delta.y,
    };
    nodePositions.value = { ...nodePositions.value, [id]: nextPosition };
    hasManualChanges.value = true;
    updateNode(id, { position: nextPosition });
    movementAnnouncement.value = `Moved ${id}.`;
  }

  function resetNodesFromLayout(): void {
    hasManualChanges.value = false;
    nodePositions.value = { ...layout.value.positions };
    nodes.value = layout.value.nodeIds.map((id) => ({
      id,
      type: "service",
      position: nodePositions.value[id] ?? { x: 0, y: 0 },
      width: layout.value.nodeWidths[id] ?? MAP_NODE_W,
      height: MAP_NODE_H,
      draggable: true,
      selectable: false,
      focusable: false,
      connectable: false,
      data: {
        dimmed: isNodeDimmed(id),
        focused: selected.value === id,
        onSelect: () => selectService(id),
        onMove: (delta: MapNodePosition) => moveNode(id, delta),
      },
    }));
  }

  function syncNodeAppearance(): void {
    nodes.value = nodes.value.map((node) => ({
      ...node,
      data: {
        ...node.data,
        dimmed: isNodeDimmed(node.id),
        focused: selected.value === node.id,
      },
    }));
  }

  function handleNodeMouseEnter(event: NodeMouseEvent): void {
    hovered.value = event.node.id;
  }

  function handleNodeDrag(event: NodeDragEvent): void {
    hasManualChanges.value = true;
    nodePositions.value = {
      ...nodePositions.value,
      [event.node.id]: event.node.position,
    };
  }

  function handleNodeMouseLeave(): void {
    hovered.value = null;
  }

  function handleEdgeClick(event: EdgeMouseEvent): void {
    selectService(event.edge.target);
  }

  function handleFitView(): void {
    const feedbackCount = getBackwardEdgeCount();
    const padding =
      hasManualChanges.value && feedbackCount > 0
        ? {
            top:
              FEEDBACK_VIEW_MARGIN +
              FEEDBACK_ROUTE_MARGIN +
              (feedbackCount - 1) * FEEDBACK_ROUTE_LANE_GAP,
            right: FEEDBACK_VIEW_MARGIN,
            bottom: FEEDBACK_VIEW_MARGIN,
            left: FEEDBACK_VIEW_MARGIN,
          }
        : FIT_PADDING;
    void fitView({ padding });
  }

  function handleZoomIn(): void {
    void zoomIn();
  }

  function handleZoomOut(): void {
    void zoomOut();
  }

  function resetLayout(): void {
    resetNodesFromLayout();
    void nextTick(handleFitView);
  }

  function getBackwardEdgeCount(): number {
    return layout.value.edges.filter(
      (edge) => edge.feedback || isBackwardEdge(edge.from, edge.to),
    ).length;
  }

  function isBackwardEdge(from: string, to: string): boolean {
    const source = nodePositions.value[from];
    const target = nodePositions.value[to];
    if (source === undefined || target === undefined) return false;
    const sourceX =
      source.x + (layout.value.nodeWidths[from] ?? MAP_NODE_W) / 2;
    const targetX = target.x + (layout.value.nodeWidths[to] ?? MAP_NODE_W) / 2;
    return targetX < sourceX - BACKWARD_EDGE_EPSILON;
  }

  function getGraphTop(): number {
    return Object.values(nodePositions.value).reduce(
      (top, position) => Math.min(top, position.y),
      Number.POSITIVE_INFINITY,
    );
  }

  watch(
    layout,
    () => {
      resetNodesFromLayout();
    },
    { immediate: true },
  );
  watch(
    layout,
    () => {
      void nextTick(handleFitView);
    },
    { flush: "post" },
  );
  watch([selected, query, hovered], syncNodeAppearance);

  return {
    nodes,
    edges,
    movementAnnouncement,
    canZoomIn,
    canZoomOut,
    handleNodeDrag,
    handleNodeMouseEnter,
    handleNodeMouseLeave,
    handleEdgeClick,
    handleZoomIn,
    handleZoomOut,
    handleFitView,
    resetLayout,
  };
}

function edgeId(from: string, to: string): string {
  return `${from}→${to}`;
}

function getDagreEdgeHandles(
  points: MapNodePosition[],
  source: MapNodePosition | undefined,
  target: MapNodePosition | undefined,
  sourceWidth: number,
  targetWidth: number,
): EdgeHandles {
  const first = points[0];
  const last = points[points.length - 1];
  if (
    source === undefined ||
    target === undefined ||
    first === undefined ||
    last === undefined
  ) {
    return getEdgeHandles(source, target, sourceWidth, targetWidth);
  }
  const sourceSide = sideToward(source, sourceWidth, first);
  const targetSide = sideToward(target, targetWidth, last);
  return createHandles(sourceSide, targetSide);
}

function getEdgeHandles(
  source: MapNodePosition | undefined,
  target: MapNodePosition | undefined,
  sourceWidth: number,
  targetWidth: number,
): EdgeHandles {
  if (source === undefined || target === undefined) {
    return createHandles("right", "left");
  }
  const dx = source.x + sourceWidth / 2 - (target.x + targetWidth / 2);
  const dy = source.y + MAP_NODE_H / 2 - (target.y + MAP_NODE_H / 2);
  if (Math.abs(dx) >= Math.abs(dy)) {
    return dx <= 0
      ? createHandles("right", "left")
      : createHandles("left", "right");
  }
  return dy <= 0
    ? createHandles("bottom", "top")
    : createHandles("top", "bottom");
}

function sideToward(
  node: MapNodePosition,
  width: number,
  point: MapNodePosition,
): MapSide {
  const dx = point.x - (node.x + width / 2);
  const dy = point.y - (node.y + MAP_NODE_H / 2);
  if (Math.abs(dx) >= Math.abs(dy)) return dx >= 0 ? "right" : "left";
  return dy >= 0 ? "bottom" : "top";
}

function createHandles(sourceSide: MapSide, targetSide: MapSide): EdgeHandles {
  return {
    sourceHandle: `source-${sourceSide}`,
    targetHandle: `target-${targetSide}`,
    sourceSide,
    targetSide,
  };
}

function feedbackRoutePoints(
  source: MapNodePosition,
  target: MapNodePosition,
  sourceWidth: number,
  targetWidth: number,
  handles: EdgeHandles,
  graphTop: number,
  lane: number,
): MapNodePosition[] {
  const laneY =
    graphTop - FEEDBACK_ROUTE_MARGIN - lane * FEEDBACK_ROUTE_LANE_GAP;
  const sourceAnchor = handleAnchor(source, sourceWidth, handles.sourceSide);
  const targetAnchor = handleAnchor(target, targetWidth, handles.targetSide);
  return [
    { x: sourceAnchor.x, y: laneY },
    { x: targetAnchor.x, y: laneY },
  ];
}

function handleAnchor(
  position: MapNodePosition,
  width: number,
  side: MapSide,
): MapNodePosition {
  if (side === "left") return { x: position.x, y: position.y + MAP_NODE_H / 2 };
  if (side === "right")
    return { x: position.x + width, y: position.y + MAP_NODE_H / 2 };
  if (side === "top") return { x: position.x + width / 2, y: position.y };
  return { x: position.x + width / 2, y: position.y + MAP_NODE_H };
}
