import { onMounted, onUnmounted, watch, type Ref } from "vue";

const PREFETCH_MARGIN_PX = "600px";

interface LogsInfiniteScrollOptions {
  sentinel: Ref<Element | null>;
  scroller: Ref<Element | null>;
  canLoadMore: Ref<boolean>;
  onLoadMore: () => void;
}

export function useLogsInfiniteScroll(
  options: LogsInfiniteScrollOptions,
): void {
  let observer: IntersectionObserver | undefined;

  function disconnect(): void {
    observer?.disconnect();
    observer = undefined;
  }

  function observe(): void {
    disconnect();
    const target = options.sentinel.value;
    if (target === null || typeof IntersectionObserver === "undefined") return;
    observer = new IntersectionObserver(
      (entries) => {
        if (
          entries[0]?.isIntersecting === true &&
          options.canLoadMore.value
        ) {
          options.onLoadMore();
        }
      },
      { root: options.scroller.value, rootMargin: PREFETCH_MARGIN_PX },
    );
    observer.observe(target);
  }

  watch(
    [() => options.sentinel.value, () => options.scroller.value],
    () => {
      observe();
    },
    { flush: "post" },
  );

  onMounted(observe);
  onUnmounted(disconnect);
}
