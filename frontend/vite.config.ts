import tailwindcss from "@tailwindcss/vite";
import vue from "@vitejs/plugin-vue";
import { fileURLToPath, URL } from "node:url";
import { defineConfig, type PluginOption } from "vite";
import svgLoader from "vite-svg-loader";

const DEFAULT_APP_PORT = "5678";

function resolveAppPort(): string {
  return process.env.APP_PORT?.trim() || DEFAULT_APP_PORT;
}

// https://vite.dev/config/
export default defineConfig(() => {
  const appPort = resolveAppPort();
  // vite-svg-loader types its plugin as Plugin<any>, which collapses mixed arrays to any[].
  const svgPlugin = svgLoader() as PluginOption;

  return {
    plugins: [vue(), svgPlugin, tailwindcss()],
    resolve: {
      alias: {
        "@": fileURLToPath(new URL("./src", import.meta.url)),
      },
    },
    server: {
      port: 5173,
      proxy: {
        "/api/v1": {
          target: `http://localhost:${appPort}`,
          changeOrigin: true,
        },
      },
    },
    build: {
      outDir: "dist",
      emptyOutDir: true,
    },
  };
});
