import type { QueryClient } from "@tanstack/vue-query";

import { queryClient } from "@/app/query-client";
import router from "@/app/router";

export interface PreloadMeta {
  prefetchQueries: () => Promise<(client: QueryClient) => Promise<void>>;
}

const preloaded = new Set<string>();

export function useRoutePreload(): {
  preloadRoute: (routeName: string) => Promise<void>;
} {
  async function preloadRoute(routeName: string): Promise<void> {
    if (preloaded.has(routeName)) return;
    preloaded.add(routeName);

    const route = router.getRoutes().find((r) => r.name === routeName);
    if (!route) return;

    await Promise.all(
      Object.values(route.components ?? {})
        .filter(
          (component): component is () => Promise<unknown> =>
            typeof component === "function",
        )
        .map((component) => component()),
    );

    const preload = route.meta.preload as PreloadMeta | undefined;
    if (preload) {
      const prefetchFn = await preload.prefetchQueries();
      await prefetchFn(queryClient);
    }
  }

  return { preloadRoute };
}
