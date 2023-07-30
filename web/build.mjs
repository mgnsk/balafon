import esbuild from "esbuild";

const target = "es2020";

await esbuild.build({
  entryPoints: ["src/index.ts"],
  outdir: "dist",
  bundle: true,
  minify: false,
  sourcemap: false,
  splitting: true,
  format: "esm",
  target: [target],
  // mangleProps: /_$/,
  treeShaking: true,
});
