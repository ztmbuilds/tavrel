<script setup lang="ts">
import { computed } from "vue";

import type { Log } from "@/shared/types/entities/log";
import { cn } from "@/shared/utils/cn";
import { SHORT_LIST_ID_LENGTH, truncateTraceId } from "@/shared/utils/format";
import { serviceColor } from "@/shared/utils/service-color";
import {
  formatRelativeTime,
  formatShortDateTimeWithSeconds,
} from "@/shared/utils/time";
import CopyButton from "@/shared/components/CopyButton.vue";
import Badge from "@/shared/components/ui/Badge.vue";
import Button from "@/shared/components/ui/Button.vue";
const props = defineProps<{
  log: Log;
  expanded: boolean;
}>();

defineEmits<{
  toggle: [id: number];
  openTrace: [traceId: string];
}>();

const relativeLabel = computed(() => formatRelativeTime(props.log.timestamp));

const dateTimeLabel = computed(() =>
  formatShortDateTimeWithSeconds(props.log.timestamp),
);

const shortTraceId = computed(() =>
  truncateTraceId(props.log.traceId, SHORT_LIST_ID_LENGTH),
);

const rowLabel = computed(
  () =>
    `${props.log.level}, ${props.log.service}, ${dateTimeLabel.value}: ${props.log.message}`,
);
</script>

<template>
  <div
    :class="
      cn(
        'border-subtle flex min-h-10 cursor-pointer items-start gap-2.5 border-b px-5 py-1.5 transition-colors duration-120 ease-out',
        log.level === 'error'
          ? 'bg-error/[0.05] hover:bg-error/[0.08]'
          : 'hover:bg-white/[0.035]',
      )
    "
    role="listitem"
    tabindex="0"
    :aria-expanded="expanded"
    :aria-label="rowLabel"
    @click="$emit('toggle', log.id)"
    @keydown.enter.prevent="$emit('toggle', log.id)"
    @keydown.space.prevent="$emit('toggle', log.id)"
  >
    <div
      :class="
        cn(
          'mt-0.5 w-0.5 shrink-0 self-stretch rounded-sm',
          log.level === 'error' && 'bg-error',
          log.level === 'warn' && 'bg-warn',
          log.level === 'info' && 'bg-info',
        )
      "
      aria-hidden="true"
    />
    <div
      :class="
        cn(
          'text-tertiary w-11 shrink-0 pt-0.5 text-[0.6875rem] font-semibold tracking-wide uppercase',
          log.level === 'error' && 'text-error-text',
          log.level === 'warn' && 'text-warn',
        )
      "
      aria-hidden="true"
    >
      {{ log.level }}
    </div>
    <div
      class="text-tertiary w-[10.75rem] shrink-0 truncate pt-px text-xs tabular-nums max-[40rem]:hidden"
      :title="relativeLabel"
    >
      {{ dateTimeLabel }}
    </div>
    <div
      class="flex w-[7.375rem] shrink-0 items-center pt-0"
      aria-hidden="true"
    >
      <Badge
        variant="outline"
        class="h-5 max-w-full gap-1 px-2 text-[0.6875rem] font-normal"
        :title="log.service"
      >
        <span
          class="size-1.5 shrink-0 rounded-full"
          :style="{ background: serviceColor(log.service) }"
        />
        <span class="min-w-0 overflow-hidden text-ellipsis">{{
          log.service
        }}</span>
      </Badge>
    </div>
    <div
      class="text-tertiary w-28 shrink-0 truncate pt-0.5 text-xs max-[56.25rem]:hidden"
      :title="log.environment"
      aria-hidden="true"
    >
      {{ log.environment }}
    </div>
    <div
      :class="
        cn(
          'group text-primary flex min-w-0 flex-1 items-center gap-1.5 text-[0.78125rem] leading-relaxed',
          log.level === 'error' && 'text-error-text',
        )
      "
    >
      <span class="min-w-0 flex-1 truncate" :title="log.message">
        {{ log.message }}
      </span>
      <CopyButton
        compact
        :text="log.message"
        label="Copy log message"
        title="Copy log message"
      />
    </div>
    <span
      v-if="log.traceId !== ''"
      class="group inline-flex shrink-0 items-center gap-1 max-[56.25rem]:hidden"
    >
      <Button
        variant="ghost"
        size="sm"
        class="text-tertiary hover:text-heading h-auto p-0 font-mono text-[0.71875rem] hover:bg-transparent"
        :aria-label="`trace:${shortTraceId}, open trace`"
        :title="`Open trace ${log.traceId}`"
        @click.stop="$emit('openTrace', log.traceId)"
      >
        trace:{{ shortTraceId }}
      </Button>
      <CopyButton
        :text="log.traceId"
        label="Copy trace ID"
        title="Copy trace ID"
      />
    </span>
  </div>
  <div
    v-if="expanded"
    class="border-subtle bg-surface border-b py-2.5 pr-5 pl-[3.6875rem] max-[40rem]:px-5"
  >
    <div
      class="group text-secondary bg-canvas relative rounded py-2 pr-9 pl-2.5 text-[0.78125rem] leading-relaxed break-words whitespace-pre-wrap"
    >
      {{ log.message }}
      <CopyButton
        compact
        class="absolute top-1.5 right-1.5"
        :text="log.message"
        label="Copy log message"
        title="Copy log message"
      />
    </div>
    <dl class="mt-2.5 flex flex-wrap gap-x-5 gap-y-2 text-xs leading-relaxed">
      <div class="flex gap-1">
        <dt class="text-tertiary">service&nbsp;</dt>
        <dd class="text-primary m-0">{{ log.service }}</dd>
      </div>
      <div v-if="log.environment !== ''" class="flex gap-1">
        <dt class="text-tertiary">environment&nbsp;</dt>
        <dd class="text-primary m-0">{{ log.environment }}</dd>
      </div>
      <div class="flex gap-1">
        <dt class="text-tertiary">level&nbsp;</dt>
        <dd class="text-primary m-0">{{ log.level }}</dd>
      </div>
      <div v-if="log.traceId !== ''" class="flex gap-1">
        <dt class="text-tertiary">trace_id&nbsp;</dt>
        <dd class="m-0">
          <span class="group inline-flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              class="text-secondary hover:text-heading h-auto p-0 font-mono text-[0.71875rem] hover:bg-transparent"
              :aria-label="`${log.traceId}, open trace`"
              :title="log.traceId"
              @click="$emit('openTrace', log.traceId)"
            >
              {{ log.traceId }}
            </Button>
            <CopyButton
              :text="log.traceId"
              label="Copy trace ID"
              title="Copy trace ID"
            />
          </span>
        </dd>
      </div>
      <div v-if="log.spanId !== ''" class="flex gap-1">
        <dt class="text-tertiary">span_id&nbsp;</dt>
        <dd class="text-primary m-0 font-mono text-[0.71875rem] tabular-nums">
          <span class="group inline-flex items-center gap-1">
            <span>{{ log.spanId }}</span>
            <CopyButton
              :text="log.spanId"
              label="Copy span ID"
              title="Copy span ID"
            />
          </span>
        </dd>
      </div>
      <div v-if="log.host !== ''" class="flex gap-1">
        <dt class="text-tertiary">host&nbsp;</dt>
        <dd class="text-primary m-0 tabular-nums">{{ log.host }}</dd>
      </div>
    </dl>
  </div>
</template>
