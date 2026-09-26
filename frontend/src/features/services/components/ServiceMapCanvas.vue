<script setup lang="ts">
import { computed } from "vue";
import { Background } from "@vue-flow/background";
import { ControlButton, Controls } from "@vue-flow/controls";
import { VueFlow } from "@vue-flow/core";

import ServiceMapNode from "@/features/services/components/ServiceMapNode.vue";

import type { MapLayout } from "../layout";
import { useServiceMapFlow } from "../composables/useServiceMapFlow";

const props = defineProps<{
  layout: MapLayout;
  selected: string | null;
  query: string;
}>();

const emit = defineEmits<{
  select: [id: string];
}>();

const flow = useServiceMapFlow(
  computed(() => props.layout),
  computed(() => props.selected),
  computed(() => props.query),
  (id) => emit("select", id),
);

const {
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
} = flow;
</script>

<template>
  <div
    class="service-map-surface relative h-full min-h-[32.5rem] overflow-hidden"
  >
    <VueFlow
      v-model:nodes="nodes"
      :edges="edges"
      :nodes-draggable="true"
      :nodes-connectable="false"
      :nodes-focusable="false"
      :elements-selectable="false"
      :select-nodes-on-drag="false"
      :pan-on-drag="true"
      :zoom-on-scroll="false"
      :zoom-on-pinch="true"
      :fit-view-on-init="true"
      class="h-full w-full bg-transparent"
      @node-mouse-enter="handleNodeMouseEnter"
      @node-mouse-leave="handleNodeMouseLeave"
      @edge-click="handleEdgeClick"
    >
      <template #node-service="nodeProps">
        <ServiceMapNode v-bind="nodeProps" />
      </template>
      <Background
        variant="dots"
        :gap="24"
        :size="1"
        color="var(--color-strong)"
      />
      <Controls
        position="top-right"
        :show-zoom="false"
        :show-fit-view="false"
        :show-interactive="false"
      >
        <ControlButton
          title="Zoom in"
          aria-label="Zoom in"
          :disabled="!canZoomIn"
          @click.stop="handleZoomIn"
        >
          +
        </ControlButton>
        <ControlButton
          title="Zoom out"
          aria-label="Zoom out"
          :disabled="!canZoomOut"
          @click.stop="handleZoomOut"
        >
          −
        </ControlButton>
        <ControlButton
          title="Fit map to view"
          aria-label="Fit map to view"
          @click.stop="handleFitView"
        >
          ⌖
        </ControlButton>
        <ControlButton
          title="Reset node layout"
          aria-label="Reset node layout"
          @click.stop="resetLayout"
        >
          ↺
        </ControlButton>
      </Controls>
    </VueFlow>
    <p class="sr-only" role="status" aria-live="polite">
      {{ movementAnnouncement }}
    </p>
  </div>
</template>

<style scoped>
:deep(.vue-flow) {
  --vf-node-bg: transparent;
  --vf-node-text: var(--color-primary);
  --vf-connection-path: var(--color-strong);
  --vf-handle: transparent;
}

:deep(.vue-flow__controls) {
  gap: 1px;
  overflow: hidden;
  margin: 0.75rem;
  border: 1px solid var(--color-default);
  border-radius: 0.75rem;
  background: rgb(20 27 35 / 0.94);
  box-shadow: 0 0.5rem 1.5rem rgb(0 0 0 / 0.28);
  backdrop-filter: blur(0.5rem);
}

:deep(.vue-flow__controls-button) {
  width: 2.5rem;
  height: 2.5rem;
  border: 0;
  background: transparent;
  color: var(--color-secondary);
  fill: currentColor;
  transition-property: color, background-color;
  transition-duration: 120ms;
}

:deep(.vue-flow__controls-button:hover) {
  background: var(--color-hover);
  color: var(--color-heading);
}

:deep(.vue-flow__controls-button:disabled) {
  cursor: not-allowed;
  opacity: 0.45;
}

:deep(.vue-flow__controls-button svg) {
  max-width: 1rem;
  max-height: 1rem;
}

:deep(.vue-flow__controls-button:focus-visible) {
  outline: 2px solid var(--color-signal);
  outline-offset: -2px;
}
</style>
