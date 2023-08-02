<script lang="ts">
  import { Balafon, type Response } from "./balafon";
  import { onMount } from "svelte";
  import CodeMirror from "../node_modules/svelte-codemirror-editor/src/lib";
  // import { syntaxTree } from "@codemirror/language";
  import { linter, type Diagnostic } from "@codemirror/lint";
  import {
    AlignRestOption,
    OpenSheetMusicDisplay,
  } from "opensheetmusicdisplay";

  const balafon = new Balafon();
  let osmd: OpenSheetMusicDisplay = null;
  const cacheKey = "balafon-web-input";
  const cache = new Uint8Array(1 << 20); // 1M

  let value = "";
  let errorMessage = "";

  let response: Response = null;

  const balafonLinter = linter(() => {
    let diagnostics: Diagnostic[] = [];

    if (response && response.err) {
      diagnostics.push({
        from: response.pos.offset,
        to: response.pos.offset + 1,
        severity: "error",
        message: response.err,
      });
    }

    return diagnostics;
  });

  onMount(async () => {
    await balafon.init();

    osmd = new OpenSheetMusicDisplay("osmdContainer");
    osmd.setOptions({
      alignRests: AlignRestOption.Auto,
      autoBeam: true,
      backend: "svg",
      drawTitle: false,
    });

    let cachedInput = localStorage.getItem(cacheKey);
    if (cachedInput !== null) {
      value = cachedInput;
    }

    await onInput();
  });

  async function onInput() {
    localStorage.setItem(cacheKey, value);

    let input = value;

    if (input.trim().length === 0) {
      errorMessage = "";
      return;
    }

    response = balafon.convert(cache, input);
    if (response.written) {
      var blob = new Blob([cache.subarray(0, response.written)], {
        type: "application/vnd.recordare.musicxml+xml",
      });
      var url = URL.createObjectURL(blob);

      await osmd.load(url);
      osmd.render();
      errorMessage = "";
    } else {
      if (osmd.Drawer) {
        osmd.clear();
      }
      console.error(response);
      errorMessage = response.err;
    }
  }
</script>

<main>
  <div class="container">
    <div class="error">{errorMessage}</div>
    <div class="toolbar">Play Stop buttons</div>
    <div class="editor">
      <CodeMirror
        bind:value
        extensions={[balafonLinter]}
        on:change={onInput}
        styles={{
          "&": {
            height: "100%",
            // "flex-grow": "1",
          },
        }}
      />
    </div>
    <div class="score" id="osmdContainer"></div>
  </div>
</main>

<style>
  @import "normalize.css";

  .container {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    /* height: 100vh; */
    display: grid;
    grid-template-columns: 1fr 1fr;
    grid-template-rows: 0.1fr 2.8fr 0.1fr;
    grid-auto-columns: 1fr;
    gap: 0px 0px;
    grid-auto-flow: row;
    grid-template-areas:
      "toolbar toolbar"
      "editor score"
      "error error";
  }

  .error {
    grid-area: error;
    color: red;
    font-weight: bold;
  }

  .toolbar {
    grid-area: toolbar;
  }

  .editor {
    grid-area: editor;
  }

  .score {
    grid-area: score;
  }

  /* #osmdContainer { */
  /*   flex-grow: 1; */
  /*   height: 100%; */
  /* } */

  /* .error { */
  /*   color: red; */
  /*   font-weight: bold; */
  /* } */
</style>
