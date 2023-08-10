import { defineConfig } from "vite";
import { svelte } from "@sveltejs/vite-plugin-svelte";
import commonjs from "vite-plugin-commonjs";

// https://vitejs.dev/config/
export default defineConfig({
  plugins: [
    commonjs({
      filter(id) {
        return true;
        // // `node_modules` is exclude by default, so we need to include it explicitly
        // // https://github.com/vite-plugin/vite-plugin-commonjs/blob/v0.7.0/src/index.ts#L125-L127
        // if (id.includes("node_modules/xxx")) {
        //   return true;
        // }
      },
    }),
    svelte(),
  ],
});
