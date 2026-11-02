<script setup lang="ts">
import { computed } from "vue";
import { Background } from "@vue-flow/background";
import { ControlButton, Controls } from "@vue-flow/controls";
import { VueFlow } from "@vue-flow/core";

import fitIcon from "@/assets/icons/fit.svg?component";
import historyIcon from "@/assets/icons/history.svg?component";
import minusIcon from "@/assets/icons/minus.svg?component";
import plusIcon from "@/assets/icons/plus.svg?component";
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
  <div class="service-map-surface relative h-full overflow-hidden">
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
      :fit-view-on-init="false"
      class="h-full w-full bg-transparent"
      @nodes-initialized="handleFitView"
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
          <component :is="plusIcon" />
        </ControlButton>
        <ControlButton
          title="Zoom out"
          aria-label="Zoom out"
          :disabled="!canZoomOut"
          @click.stop="handleZoomOut"
        >
          <component :is="minusIcon" />
        </ControlButton>
        <ControlButton
          title="Fit map to view"
          aria-label="Fit map to view"
          @click.stop="handleFitView"
        >
          <component :is="fitIcon" />
        </ControlButton>
        <ControlButton
          title="Reset node layout"
          aria-label="Reset node layout"
          @click.stop="resetLayout"
        >
          <component :is="historyIcon" />
        </ControlButton>
      </Controls>
    </VueFlow>
    <p class="sr-only" role="status" aria-live="polite">
      {{ movementAnnouncement }}
    </p>
  </div>
</template>
