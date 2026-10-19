<script setup lang="ts">
import { computed, ref } from "vue";

import arrowUpRightIcon from "@/assets/icons/arrow-up-right.svg?component";
import chevronDownIcon from "@/assets/icons/chevron-down.svg?component";
import chevronRightIcon from "@/assets/icons/chevron-right.svg?component";
import { formatAttributeValue } from "@/shared/api/otel";
import CopyButton from "@/shared/components/CopyButton.vue";
import type { SpanLink } from "@/shared/types/entities/span-link";
import { serviceColor } from "@/shared/utils/service-color";

const props = defineProps<{
  link: SpanLink;
}>();

const emit = defineEmits<{
  open: [];
}>();

const expanded = ref(false);

const attributes = computed(() =>
  Object.entries(props.link.attributes).map(([key, value]) => ({
    key,
    value: formatAttributeValue(value),
  })),
);
</script>

<template>
  <div class="border-subtle border-b last:border-b-0">
    <div
      class="hover:bg-hover flex cursor-pointer items-center gap-2 px-1 py-2.5 transition-colors duration-120 ease-out"
      @click="expanded = !expanded"
    >
      <button
        type="button"
        :aria-expanded="expanded"
        :aria-label="expanded ? 'Collapse link details' : 'Expand link details'"
        class="text-secondary hover:text-primary flex size-5 shrink-0 cursor-pointer items-center justify-center rounded-sm border-none bg-transparent p-0 transition-colors duration-120 ease-out"
        @click.stop="expanded = !expanded"
      >
        <component
          :is="expanded ? chevronDownIcon : chevronRightIcon"
          aria-hidden="true"
          class="size-3.5"
        />
      </button>

      <template v-if="link.target">
        <span
          class="size-2 shrink-0 rounded-full"
          :style="{ background: serviceColor(link.target.service) }"
          aria-hidden="true"
        />
        <button
          type="button"
          class="text-primary group/name hover:text-signal flex min-w-0 cursor-pointer items-center gap-1 border-none bg-transparent p-0 text-left text-[0.8125rem] leading-snug font-medium transition-colors duration-120 ease-out hover:underline hover:underline-offset-2"
          :title="link.target.name"
          @click.stop="emit('open')"
        >
          <span class="min-w-0 truncate">{{ link.target.name }}</span>
          <component
            :is="arrowUpRightIcon"
            aria-hidden="true"
            class="size-3 shrink-0 opacity-0 group-hover/name:opacity-100 group-focus-visible/name:opacity-100"
          />
        </button>
        <span class="sr-only">, status {{ link.target.status }}</span>
      </template>
      <span
        v-else
        class="text-tertiary min-w-0 truncate text-[0.8125rem] leading-snug"
        :title="link.spanId"
      >
        Linked span not in store
      </span>
    </div>

    <div v-if="expanded" class="border-strong mb-2 ml-3 border-l pl-3.5">
      <div
        class="group border-subtle flex items-center gap-3 border-b py-2 font-mono text-xs"
      >
        <span class="text-secondary w-24 shrink-0">trace_id</span>
        <span class="text-primary min-w-0 flex-1 break-all">{{
          link.traceId
        }}</span>
        <CopyButton
          :text="link.traceId"
          label="Copy linked trace ID"
          title="Copy linked trace ID"
        />
      </div>
      <div
        class="group border-subtle flex items-center gap-3 border-b py-2 font-mono text-xs"
      >
        <span class="text-secondary w-24 shrink-0">span_id</span>
        <span class="text-primary min-w-0 flex-1 break-all">{{
          link.spanId
        }}</span>
        <CopyButton
          :text="link.spanId"
          label="Copy linked span ID"
          title="Copy linked span ID"
        />
      </div>
      <div
        v-for="attr in attributes"
        :key="attr.key"
        class="border-subtle flex items-center gap-3 border-b py-2 font-mono text-xs last:border-b-0"
      >
        <span class="text-secondary w-24 shrink-0 break-words">{{
          attr.key
        }}</span>
        <span class="text-primary min-w-0 flex-1 break-all">{{
          attr.value
        }}</span>
      </div>
    </div>
  </div>
</template>
