<script setup lang="ts">
import {
  computed,
  ref,
  type FunctionalComponent,
  type SVGAttributes,
} from "vue";
import { useRoute } from "vue-router";

import PreloadLink from "@/shared/components/PreloadLink.vue";

import { preserveTimeQuery } from "@/shared/composables/useGlobalTimeRange";

import { cn } from "@/shared/utils/cn";
import Button from "@/shared/components/ui/Button.vue";
import Tooltip from "@/shared/components/ui/Tooltip.vue";
import dashboardIcon from "@/assets/icons/dashboard.svg?component";
import logsIcon from "@/assets/icons/logs.svg?component";
import servicesIcon from "@/assets/icons/services.svg?component";
import spansIcon from "@/assets/icons/spans.svg?component";
import tracesIcon from "@/assets/icons/traces.svg?component";
import sidebarCollapseIcon from "@/assets/icons/sidebar-collapse.svg?component";
import sidebarExpandIcon from "@/assets/icons/sidebar-expand.svg?component";

interface NavItem {
  label: string;
  to: string;
  routeName: string;
  match: readonly string[];
  icon: FunctionalComponent<SVGAttributes>;
}

const OBSERVE_ITEMS: readonly NavItem[] = [
  {
    label: "Dashboard",
    to: "/",
    routeName: "dashboard",
    match: ["dashboard"],
    icon: dashboardIcon,
  },
  {
    label: "Traces",
    to: "/traces",
    routeName: "traces",
    match: ["traces", "trace-detail"],
    icon: tracesIcon,
  },
  {
    label: "Logs",
    to: "/logs",
    routeName: "logs",
    match: ["logs"],
    icon: logsIcon,
  },
  {
    label: "Spans",
    to: "/spans",
    routeName: "spans",
    match: ["spans"],
    icon: spansIcon,
  },
  {
    label: "Services",
    to: "/services",
    routeName: "services",
    match: ["services"],
    icon: servicesIcon,
  },
];

const ITEM_CLASSES =
  "flex h-8 w-auto shrink-0 cursor-pointer items-center justify-start gap-2 rounded-[10px] px-2.5 text-[0.8125rem] font-medium whitespace-nowrap text-tertiary transition-colors duration-120 ease-out hover:bg-hover hover:text-secondary max-[56.25rem]:w-8 max-[56.25rem]:justify-center max-[56.25rem]:px-0 max-[40rem]:w-auto max-[40rem]:justify-start max-[40rem]:px-2.5";

const COLLAPSED_ITEM_CLASSES =
  "w-8 justify-center px-0 max-[40rem]:w-auto max-[40rem]:justify-start max-[40rem]:px-2.5";

const ACTIVE_CLASSES = "bg-active text-ink hover:bg-active hover:text-ink";

const STORAGE_KEY = "app-sidebar-collapsed";

function getInitialCollapsed(): boolean {
  try {
    return localStorage.getItem(STORAGE_KEY) === "1";
  } catch {
    return false;
  }
}

const route = useRoute();
const activeRouteName = computed(() => String(route.name ?? ""));
const isCollapsed = ref(getInitialCollapsed());
const toggleLabel = computed(() =>
  isCollapsed.value ? "Expand sidebar" : "Collapse sidebar",
);
const tooltipText = computed(() => (isCollapsed.value ? "Expand" : "Collapse"));

function isActive(item: NavItem): boolean {
  return item.match.includes(activeRouteName.value);
}

function navTo(path: string): { path: string; query: Record<string, string> } {
  return { path, query: preserveTimeQuery(route.query) };
}

function handleToggle(): void {
  isCollapsed.value = !isCollapsed.value;
  try {
    localStorage.setItem(STORAGE_KEY, isCollapsed.value ? "1" : "0");
  } catch {
    return;
  }
}
</script>

<template>
  <nav
    aria-label="Primary"
    :class="
      cn(
        'border-subtle bg-panel flex w-52 shrink-0 flex-col items-stretch gap-0.5 overflow-y-auto border-r pt-2 pr-4 pb-4 pl-2 transition-[width] duration-200 ease-out max-[56.25rem]:w-13 max-[56.25rem]:items-center max-[56.25rem]:px-1.5 max-[40rem]:w-full max-[40rem]:flex-row max-[40rem]:flex-wrap max-[40rem]:items-center max-[40rem]:gap-1 max-[40rem]:overflow-visible max-[40rem]:border-r-0 max-[40rem]:border-b max-[40rem]:px-3 max-[40rem]:py-2',
        isCollapsed && 'w-13 items-center px-1.5',
      )
    "
  >
    <div
      :class="
        cn(
          'flex items-center justify-end pb-1 max-[56.25rem]:hidden',
          isCollapsed && 'justify-center px-0',
          'max-[40rem]:hidden',
        )
      "
    >
      <Tooltip :content="tooltipText">
        <Button
          variant="ghost"
          size="sm"
          :aria-expanded="!isCollapsed"
          :aria-label="toggleLabel"
          class="size-6 shrink-0 p-0 [&_svg]:size-4"
          @click="handleToggle"
        >
          <component
            :is="isCollapsed ? sidebarExpandIcon : sidebarCollapseIcon"
            aria-hidden="true"
            class="text-tertiary shrink-0"
          />
        </Button>
      </Tooltip>
    </div>

    <PreloadLink
      v-for="item in OBSERVE_ITEMS"
      :key="item.label"
      :to="navTo(item.to)"
      :route-name="item.routeName"
      :title="item.label"
      :aria-label="item.label"
      :aria-current="isActive(item) ? 'page' : undefined"
      :class="
        cn(
          ITEM_CLASSES,
          isCollapsed && COLLAPSED_ITEM_CLASSES,
          isActive(item) && ACTIVE_CLASSES,
        )
      "
    >
      <component
        :is="item.icon"
        aria-hidden="true"
        :class="
          cn('text-tertiary size-4 shrink-0', isActive(item) && 'text-ink')
        "
      />
      <span
        :class="
          cn(
            'overflow-hidden text-ellipsis whitespace-nowrap max-[56.25rem]:hidden',
            isCollapsed && 'hidden',
            'max-[40rem]:inline',
          )
        "
        >{{ item.label }}</span
      >
    </PreloadLink>
  </nav>
</template>
