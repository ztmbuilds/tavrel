<script setup lang="ts" generic="T extends string">
import type { HTMLAttributes } from "vue";

import { TabsContent, TabsList, TabsRoot, TabsTrigger } from "reka-ui";

import { cn } from "@/shared/utils/cn";

const props = defineProps<{
  items: readonly {
    value: T;
    label: string;
  }[];
  ariaLabel?: string;
  class?: HTMLAttributes["class"];
  listClass?: HTMLAttributes["class"];
  contentClass?: HTMLAttributes["class"];
}>();
const model = defineModel<T>({ required: true });
</script>

<template>
  <TabsRoot v-model="model" :class="props.class">
    <TabsList
      :aria-label="props.ariaLabel"
      :class="cn('border-subtle flex gap-7 border-b', props.listClass)"
    >
      <TabsTrigger
        v-for="tab in props.items"
        :key="tab.value"
        :value="tab.value"
        :class="
          cn(
            'text-secondary hover:text-heading relative h-9 cursor-pointer bg-transparent p-0 font-medium',
            'data-[state=active]:text-heading',
            'data-[state=active]:after:bg-signal after:absolute after:inset-x-0 after:-bottom-px after:h-0.5 after:rounded-full after:bg-transparent',
          )
        "
      >
        {{ tab.label }}
      </TabsTrigger>
    </TabsList>
    <TabsContent
      v-for="tab in props.items"
      :key="tab.value"
      :value="tab.value"
      :class="cn('pt-5 outline-none', props.contentClass)"
    >
      <slot :name="tab.value" />
    </TabsContent>
  </TabsRoot>
</template>
