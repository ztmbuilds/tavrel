<script setup lang="ts">
import { toRef } from "vue";

import Button from "@/shared/components/ui/Button.vue";
import CopyButton from "@/shared/components/CopyButton.vue";
import SegmentedControl from "@/shared/components/ui/SegmentedControl.vue";
import AttributeTreeNode from "./AttributeTreeNode.vue";
import {
  useAttributeTree,
  type AttributeViewMode,
} from "../composables/useAttributeTree";
import type { Attributes } from "@/shared/api/otel";

const props = defineProps<{
  attributes: Attributes;
}>();

const VIEW_OPTIONS: readonly {
  value: AttributeViewMode;
  label: string;
}[] = [
  { value: "nested", label: "Nested" },
  { value: "flat", label: "Flat" },
];

const attributesRef = toRef(props, "attributes");

const {
  mode,
  nestedNodes,
  flatRows,
  hasGroups,
  isEmpty,
  virtualStyle,
  setMode,
  toggle,
  expandAll,
  collapseAll,
} = useAttributeTree(attributesRef);
</script>

<template>
  <div v-if="isEmpty" class="px-2 py-6 text-center">
    <div class="text-primary text-[0.78125rem] font-semibold">
      No attributes on this span
    </div>
  </div>
  <div v-else>
    <div class="mb-2 flex items-center justify-between">
      <SegmentedControl
        :model-value="mode"
        :options="VIEW_OPTIONS"
        aria-label="Attribute view mode"
        @update:model-value="setMode"
      />
      <div
        v-if="mode === 'nested' && hasGroups"
        class="flex items-center gap-3"
      >
        <Button
          variant="ghost"
          size="sm"
          class="h-auto p-0 text-[0.6875rem]"
          @click="expandAll"
        >
          Expand all
        </Button>
        <Button
          variant="ghost"
          size="sm"
          class="h-auto p-0 text-[0.6875rem]"
          @click="collapseAll"
        >
          Collapse all
        </Button>
      </div>
    </div>

    <div v-if="mode === 'nested'">
      <div class="border-subtle border-x-0 border-t border-b-0">
        <AttributeTreeNode
          v-for="node in nestedNodes"
          :key="node.fullPath"
          :node="node"
          :row-style="virtualStyle"
          @toggle="toggle"
        />
      </div>
    </div>

    <div v-else>
      <div class="border-subtle border-x-0 border-t border-b-0">
        <div
          v-for="row in flatRows"
          :key="row.path"
          :style="virtualStyle"
          class="group border-subtle flex items-center gap-2 border-x-0 border-t-0 border-b py-2"
        >
          <span class="w-44 shrink-0 font-mono text-xs wrap-break-word"
            ><span class="text-tertiary">{{ row.namespace }}</span
            ><span class="text-secondary">{{ row.leaf }}</span></span
          >
          <span
            class="text-primary min-w-0 flex-1 font-mono text-xs break-all tabular-nums"
            >{{ row.value }}</span
          >
          <CopyButton
            :label="`Copy ${row.path}`"
            title="Copy key and value"
            :text="`${row.path}: ${row.value}`"
          />
        </div>
      </div>
    </div>
  </div>
</template>
