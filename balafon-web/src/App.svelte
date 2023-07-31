<script lang="ts">
  import { Balafon } from "./balafon";
  import CodeMirror from "svelte-codemirror-editor";
  import { OpenSheetMusicDisplay } from "opensheetmusicdisplay";
  import { tick } from "svelte";

  const balafon = new Balafon();
  let value = "";
  let error = "";

  let osmd: OpenSheetMusicDisplay = null;

  async function onInput() {
    if (osmd === null) {
      //        await tick();
      osmd = new OpenSheetMusicDisplay("osmdContainer");
      osmd.setOptions({
        backend: "svg",
        drawTitle: true,
        // drawingParameters: "compacttight" // don't display title, composer etc., smaller margins
      });
    }

    let input = value;

    if (input.trim().length === 0) {
      error = "";
      return;
    }

    try {
      const result = balafon.convert(input);
      error = "";

      //    osmd.clear();
      await osmd.load(result);
      osmd.render();
    } catch (err) {
      error = err.message;
      console.error(err);
    }
  }
</script>

<main>
  <div id="container">
    <div class="row text-input">
      <div class="col">
        <h1>Balafon</h1>
        {#await balafon.init()}
          Loading balafon...
        {:then _}
          <div id="editor">
            <CodeMirror
              bind:value
              on:change={onInput}
              styles={{
                "&": {
                  "flex-grow": "1",
                },
              }}
            />
          </div>
        {:catch err}
          System error: {err.message}.
        {/await}
      </div>
      <div class="col">
        <div id="osmdContainer"></div>
      </div>
    </div>
    <div class="row">
      <div class="error">{error}</div>
    </div>
  </div>
</main>

<style>
  @import "normalize.css";

  main {
    height: 100vh;
  }

  #container {
    display: flex;
    flex-direction: column;
    height: 100%;
  }

  .row {
    display: flex;
    flex-direction: row;
  }

  .row.text-input {
    flex-grow: 1;
  }

  .row.text-input > .col {
    display: flex;
    flex-direction: column;
  }

  .col {
    flex: 1;
    border: 1px solid grey;
  }

  #editor {
    flex-grow: 1;
    height: 100%;
  }

  #osmdContainer {
    flex-grow: 1;
    height: 100%;
  }

  .error {
    color: red;
    font-weight: bold;
  }
</style>
