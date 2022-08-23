import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import sveltePreprocess from "svelte-preprocess";
import * as path from "path";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    svelte({
      preprocess: sveltePreprocess(),
    }),
  ],
  resolve: {
    alias: {
      "@components": path.resolve(__dirname, "./src/components"),
      "@lib": path.resolve(__dirname, "./src/lib"),
      "@Modal": path.resolve(__dirname, "./src/components/Modal.svelte"),
    },
  },
  build: {
    outDir: "../api/frontend-dist",
    emptyOutDir: true,
  },
  server: {
    port: 3000,
  },
});
