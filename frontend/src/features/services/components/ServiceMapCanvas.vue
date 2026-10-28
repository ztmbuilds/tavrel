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
    class="service-map-surface relative h-full min-h-[32.5rem] overflow-hidden max-[53.75rem]:min-h-0"
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
