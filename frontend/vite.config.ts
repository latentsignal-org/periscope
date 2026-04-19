import { execSync } from "node:child_process";
import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import {
  applyDevProxyHeaders,
  getDevProxyTarget,
} from "./dev-proxy";

function gitCommit(): string {
  try {
    return execSync("git rev-parse --short HEAD", {
      encoding: "utf-8",
    }).trim();
  } catch {
    return "unknown";
  }
}

export default defineConfig({
  base: "/",
  plugins: [svelte()],
  define: {
    "import.meta.env.VITE_BUILD_COMMIT": JSON.stringify(
      gitCommit(),
    ),
  },
  resolve: {
    conditions: ["browser"],
  },
  server: {
    proxy: {
      "/api": {
        target: getDevProxyTarget(process.env),
        changeOrigin: true,
        configure(proxy, options) {
          proxy.on("proxyReq", (proxyReq) => {
            applyDevProxyHeaders(proxyReq, String(options.target));
          });
        },
      },
    },
  },
  build: {
    outDir: "dist",
    emptyOutDir: true,
  },
  test: {
    environment: "jsdom",
    exclude: ["e2e/**", "node_modules/**"],
    server: {
      deps: {
        inline: ["svelte"],
      },
    },
  },
});
