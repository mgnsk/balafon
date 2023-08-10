<script lang="ts">
  import { Balafon, type Port, type ConvertResponse } from "./balafon";
  import { onMount } from "svelte";
  import CodeMirror from "../node_modules/svelte-codemirror-editor/src/lib";
  import { linter, type Diagnostic } from "@codemirror/lint";
  import {
    AlignRestOption,
    Cursor,
    OpenSheetMusicDisplay,
  } from "opensheetmusicdisplay";

  const balafon = new Balafon();
  let osmd: OpenSheetMusicDisplay = null;
  const cache = new Uint8Array(1 << 20); // 1M

  let inputValue = "";
  let errorMessage = "";

  let ports: Port[] = [];
  let selectedPort = 0;

  let response: ConvertResponse = null;

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
    const portsResponse = balafon.listPorts();
    if (portsResponse.err) {
      showError(portsResponse.err);
    } else {
      ports = portsResponse.ports;
    }

    osmd = new OpenSheetMusicDisplay("osmdContainer");
    osmd.setOptions({
      alignRests: AlignRestOption.Auto,
      autoBeam: true,
      backend: "svg",
      drawTitle: false,
      followCursor: true,
    });

    // https://github.com/opensheetmusicdisplay/opensheetmusicdisplay/issues/1254#issuecomment-1282613439
    osmd.EngravingRules.PercussionOneLineCutoff = 0;

    let cachedInput = localStorage.getItem("balafon-web-input");
    if (cachedInput !== null) {
      inputValue = cachedInput;
      await onInput();
    }

    let cachedPort = localStorage.getItem("balafon-web-port");
    if (cachedPort !== null) {
      selectedPort = parseInt(cachedPort);
      onSelectPort();
    }
  });

  async function onInput() {
    let input = inputValue;

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
      localStorage.setItem("balafon-web-input", inputValue);
    } else {
      if (osmd.Drawer) {
        osmd.clear();
      }
      showError(response.err);
    }
  }

  function showError(err: string) {
    console.error(response);
    errorMessage = err;
  }

  function onSelectPort() {
    const resp = balafon.selectPort(selectedPort);
    if (resp.err) {
      showError(resp.err);
    } else {
      localStorage.setItem("balafon-web-port", selectedPort.toString());
    }
  }

  async function onPlay() {
    const resp = balafon.play(inputValue);
    if (resp.err) {
      showError(resp.err);
    }
  }
</script>

<main>
  <div class="error">{errorMessage}</div>

  <div class="toolbar">
    <label for="ports">Choose MIDI out port:</label>
    <select
      name="ports"
      id="ports"
      bind:value={selectedPort}
      on:change={onSelectPort}
    >
      {#each ports as port}
        <option value={port.number}>{port.number}: {port.name}</option>
      {/each}
    </select>
    <a href="#" on:click={onPlay} title="Play">Play</a>
  </div>

  <div class="editor">
    <CodeMirror
      bind:value={inputValue}
      extensions={[balafonLinter]}
      on:change={onInput}
    />
  </div>

  <div class="score" id="osmdContainer"></div>
</main>

<style>
  @import "normalize.css";

  main {
    position: absolute;
    top: 0;
    left: 0;
    right: 0;
    bottom: 0;
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
</style>
