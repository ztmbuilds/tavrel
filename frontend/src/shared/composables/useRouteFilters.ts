import { onUnmounted, ref, watch, type Ref } from "vue";
import { useRouter, type RouteLocationNormalizedLoaded } from "vue-router";

export function readQueryParam(
  route: RouteLocationNormalizedLoaded,
  name: string,
): string | undefined {
  const raw = route.query[name];
  if (typeof raw === "string" && raw !== "") return raw;
  return undefined;
}

export function useDebouncedSearch(
  searchInput: Ref<string>,
  delayMs: number,
): Ref<string> {
  const debounced = ref(searchInput.value);
  let timer: ReturnType<typeof setTimeout> | undefined;
  watch(searchInput, (value) => {
    if (timer !== undefined) clearTimeout(timer);
    timer = setTimeout(() => {
      debounced.value = value;
    }, delayMs);
  });
  onUnmounted(() => {
    if (timer !== undefined) clearTimeout(timer);
  });
  return debounced;
}

export function useRouteQuerySync(
  filters: Ref<object>,
  toQuery: () => Record<string, string>,
): void {
  const router = useRouter();
  watch(
    filters,
    () => {
      void router.replace({ query: toQuery() });
    },
    { deep: true },
  );
}
