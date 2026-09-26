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
import type { ServiceMapNodeData } from "../types";

const FIT_PADDING = 0.14;
const DIMMED_EDGE_OPACITY = 0.07;
const DEFAULT_EDGE_OPACITY = 0.72;
const HOT_EDGE_WIDTH = 1.8;
const DEFAULT_EDGE_WIDTH = 1.4;

type MapSide = "left" | "right" | "top" | "bottom";
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

interface UseServiceMapFlowResult {
  nodes: Ref<ServiceMapFlowNode[]>;
  edges: ComputedRef<Edge[]>;
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

  const edges = computed(() => {
    const focus = hovered.value ?? selected.value;

    return layout.value.edges.map((edge) => {
      const sourcePosition = nodePositions.value[edge.from];
      const targetPosition = nodePositions.value[edge.to];
      const sourceWidth = layout.value.nodeWidths[edge.from] ?? MAP_NODE_W;
      const targetWidth = layout.value.nodeWidths[edge.to] ?? MAP_NODE_W;
      const handles = getEdgeHandles(
        sourcePosition,
        targetPosition,
        sourceWidth,
        targetWidth,
      );
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
        id: `${edge.from}→${edge.to}`,
        source: edge.from,
        target: edge.to,
        sourceHandle: handles.sourceHandle,
        targetHandle: handles.targetHandle,
        type: "default",
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: "var(--color-secondary)",
          width: 9,
          height: 9,
        },
        style: {
          stroke: isHot ? "var(--color-secondary)" : "var(--color-strong)",
          strokeWidth: isHot ? HOT_EDGE_WIDTH : DEFAULT_EDGE_WIDTH,
          opacity: isDimmed ? DIMMED_EDGE_OPACITY : DEFAULT_EDGE_OPACITY,
        },
        selectable: false,
        focusable: false,
        interactionWidth: 12,
        ariaLabel: `${edge.from} calls ${edge.to}`,
      } satisfies Edge;
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
    updateNode(id, {
      position: nextPosition,
    });
    movementAnnouncement.value = `Moved ${id}.`;
  }

  function syncNodes(preservePositions = true): void {
    const positions: Record<string, MapNodePosition> = {};
    for (const id of layout.value.nodeIds) {
      positions[id] = preservePositions
        ? (nodePositions.value[id] ??
          layout.value.positions[id] ?? { x: 0, y: 0 })
        : (layout.value.positions[id] ?? { x: 0, y: 0 });
    }
    nodePositions.value = positions;

    nodes.value = layout.value.nodeIds.map((id) => ({
      id,
      type: "service",
      position: positions[id] ?? { x: 0, y: 0 },
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

  function handleNodeMouseEnter(event: NodeMouseEvent): void {
    hovered.value = event.node.id;
  }

  function handleNodeDrag(event: NodeDragEvent): void {
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
    void fitView({ padding: FIT_PADDING });
  }

  function handleZoomIn(): void {
    void zoomIn();
  }

  function handleZoomOut(): void {
    void zoomOut();
  }

  function resetLayout(): void {
    syncNodes(false);
    void nextTick(handleFitView);
  }

  watch(layout, () => syncNodes(), { immediate: true });
  watch(
    layout,
    () => {
      void nextTick(handleFitView);
    },
    { flush: "post" },
  );

  watch([selected, query, hovered], () => syncNodes());

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

function getEdgeHandles(
  source: MapNodePosition | undefined,
  target: MapNodePosition | undefined,
  sourceWidth: number,
  targetWidth: number,
): { sourceHandle: string; targetHandle: string } {
  if (source === undefined || target === undefined) {
    return { sourceHandle: "source-right", targetHandle: "target-left" };
  }

  const dx = source.x + sourceWidth / 2 - (target.x + targetWidth / 2);
  const dy = source.y + MAP_NODE_H / 2 - (target.y + MAP_NODE_H / 2);
  let sourceSide: MapSide;
  let targetSide: MapSide;

  if (Math.abs(dx) >= Math.abs(dy)) {
    sourceSide = dx <= 0 ? "right" : "left";
    targetSide = dx <= 0 ? "left" : "right";
  } else {
    sourceSide = dy <= 0 ? "bottom" : "top";
    targetSide = dy <= 0 ? "top" : "bottom";
  }

  return {
    sourceHandle: `source-${sourceSide}`,
    targetHandle: `target-${targetSide}`,
  };
}
