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
  type NodeMouseEvent,
} from "@vue-flow/core";

import {
  MAP_NODE_H,
  MAP_NODE_W,
  type MapLayout,
  type MapNodePosition,
} from "../layout";
import type { ServiceMapNodeData } from "../types";

const FIT_PADDING = { top: 0.08, right: 0.2, bottom: 0.08, left: 0.08 };
const FIT_MAX_ZOOM = 1.2;
const EDGE_DIMMED_OPACITY = 0.18;
const EDGE_DEFAULT_OPACITY = 0.85;
const EDGE_HOT_WIDTH = 1.8;
const EDGE_DEFAULT_WIDTH = 1.4;

type EdgeHandles = {
  sourceHandle: string;
  targetHandle: string;
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
type ServiceMapFlowEdge = Edge;

interface UseServiceMapFlowResult {
  nodes: Ref<ServiceMapFlowNode[]>;
  edges: ComputedRef<ServiceMapFlowEdge[]>;
  movementAnnouncement: Ref<string>;
  canZoomIn: ComputedRef<boolean>;
  canZoomOut: ComputedRef<boolean>;
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
  const nodePositions = computed((): Record<string, MapNodePosition> =>
    Object.fromEntries(nodes.value.map(({ id, position }) => [id, position])),
  );
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
        id: edgeId(edge.from, edge.to),
        source: edge.from,
        target: edge.to,
        sourceHandle: handles.sourceHandle,
        targetHandle: handles.targetHandle,
        markerEnd: {
          type: MarkerType.ArrowClosed,
          color: isHot ? "var(--color-secondary)" : "var(--color-strong)",
          width: 9,
          height: 9,
        },
        style: {
          stroke: isHot ? "var(--color-secondary)" : "var(--color-strong)",
          strokeWidth: isHot ? EDGE_HOT_WIDTH : EDGE_DEFAULT_WIDTH,
          opacity: isDimmed ? EDGE_DIMMED_OPACITY : EDGE_DEFAULT_OPACITY,
        },
        type: "default",
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
    updateNode(id, { position: nextPosition });
    movementAnnouncement.value = `Moved ${id}.`;
  }

  function resetNodesFromLayout(): void {
    nodes.value = layout.value.nodeIds.map((id) => ({
      id,
      type: "service",
      position: layout.value.positions[id] ?? { x: 0, y: 0 },
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

  function handleNodeMouseLeave(): void {
    hovered.value = null;
  }

  function handleEdgeClick(event: EdgeMouseEvent): void {
    selectService(event.edge.target);
  }

  function handleFitView(): void {
    void fitView({ padding: FIT_PADDING, maxZoom: FIT_MAX_ZOOM });
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

function getEdgeHandles(
  source: MapNodePosition | undefined,
  target: MapNodePosition | undefined,
  sourceWidth: number,
  targetWidth: number,
): EdgeHandles {
  if (source === undefined || target === undefined) {
    return {
      sourceHandle: "source-right",
      targetHandle: "target-left",
    };
  }
  const dx = source.x + sourceWidth / 2 - (target.x + targetWidth / 2);
  const dy = source.y + MAP_NODE_H / 2 - (target.y + MAP_NODE_H / 2);
  if (Math.abs(dx) >= Math.abs(dy)) {
    return dx <= 0
      ? { sourceHandle: "source-right", targetHandle: "target-left" }
      : { sourceHandle: "source-left", targetHandle: "target-right" };
  }
  return dy <= 0
    ? { sourceHandle: "source-bottom", targetHandle: "target-top" }
    : { sourceHandle: "source-top", targetHandle: "target-bottom" };
}
