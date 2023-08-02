import init from "./main.wasm?init";
import "./wasm_exec.js";
const go = new globalThis.Go();

type Pos = {
  offset: number;
  line: number;
  column: number;
};

export type Response = {
  written?: number;
  err: string;
  pos: Pos;
};

type ConvertFunction = (dst: Uint8Array, input: string) => Response;

export class Balafon {
  private fn: ConvertFunction;

  async init(): Promise<Balafon> {
    return new Promise((resolve) => {
      init(go.importObject).then((instance) => {
        go.run(instance);

        this.fn = globalThis.convert;
        resolve(this);
      });
    });
  }

  convert(dst: Uint8Array, input: string): Response {
    return this.fn(dst, input);
  }
}
