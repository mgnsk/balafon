import init from "./main.wasm?init";
import "./wasm_exec.js";
const go = new globalThis.Go();

type Pos = {
  offset: number;
  line: number;
  column: number;
};

export type ConvertResponse = {
  written?: number;
  err: string;
  pos: Pos;
};

export type PlayResponse = {
  err: string;
};

type ConvertFunction = (dst: Uint8Array, input: string) => ConvertResponse;
type PlayFunction = () => ConvertResponse;

export class Balafon {
  private convertFn: ConvertFunction;
  private playFn: PlayFunction;

  async init(): Promise<Balafon> {
    return new Promise((resolve) => {
      init(go.importObject).then((instance) => {
        go.run(instance);

        this.convertFn = globalThis.convert;
        this.playFn = globalThis.play;
        resolve(this);
      });
    });
  }

  convert(dst: Uint8Array, input: string): ConvertResponse {
    return this.convertFn(dst, input);
  }

  play(): PlayResponse {
    return this.playFn();
  }
}
