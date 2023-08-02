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
      drawTitle: true,
      // drawingParameters: "compacttight" // don't display title, composer etc., smaller margins
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
  <div class="div1">
    <!-- <h1>Balafon</h1> -->
    <!-- <div id="editor"> -->
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
    <!-- </div> -->
  </div>
  <div class="div2" id="osmdContainer"></div>

  <div class="div3">{errorMessage}</div>

  <!-- <div id="container"> -->
  <!--   <div class="row text-input"> -->
  <!--     <div class="col"> -->
  <!--     </div> -->
  <!--     <div class="col"> -->
  <!--       <div id="osmdContainer"></div> -->
  <!--     </div> -->
  <!--   </div> -->
  <!-- </div> -->
</main>

<style>
  @import "normalize.css";

  /* main { */
  /*   height: 100%; */
  /* } */

  main {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
    /* height: 100vh; */
    display: grid;
    grid-template-columns: repeat(2, 1fr);
    grid-template-rows: repeat(2, 1fr);
    grid-column-gap: 0px;
    grid-row-gap: 0px;
  }

  .div1 {
    grid-area: 1 / 1 / 2 / 2;
    border: 1px solid gray;
  }
  .div2 {
    grid-area: 1 / 2 / 2 / 3;
    border: 1px solid gray;
  }
  .div3 {
    grid-area: 2 / 1 / 3 / 3;
    border: 1px solid gray;
  }

  /* #container { */
  /*   display: flex; */
  /*   flex-direction: column; */
  /*   height: 100%; */
  /* } */
  /**/
  /* .row { */
  /*   display: flex; */
  /*   flex-direction: row; */
  /* } */
  /**/
  /* .row.text-input { */
  /*   flex-grow: 1; */
  /* } */
  /**/
  /* .row.text-input > .col { */
  /*   display: flex; */
  /*   flex-direction: column; */
  /* } */
  /**/
  /* .col { */
  /*   flex: 1; */
  /*   border: 1px solid grey; */
  /* } */

  /* #editor { */
  /*   flex-grow: 1; */
  /*   height: 100%; */
  /* } */

  #osmdContainer {
    flex-grow: 1;
    height: 100%;
  }

  /* .error { */
  /*   color: red; */
  /*   font-weight: bold; */
  /* } */
</style>
