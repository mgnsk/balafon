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

export type Port = {
  number: number;
  name: string;
};

export type ListPortsResponse = {
  err?: string;
  ports: Port[];
};

export type SelectPortResponse = {
  err?: string;
};

export type PlayResponse = {
  err?: string;
};

type ConvertFn = (dst: Uint8Array, input: string) => ConvertResponse;
type ListPortsFn = () => ListPortsResponse;
type SelectPortFn = (port: number) => SelectPortResponse;
type PlayFn = (input: string) => PlayResponse;

export class Balafon {
  convert: ConvertFn;
  listPorts: ListPortsFn;
  selectPort: SelectPortFn;
  play: PlayFn;

  async init(): Promise<Balafon> {
    return new Promise((resolve) => {
      init(go.importObject).then((instance) => {
        go.run(instance).then(() => {
          throw new Error("balafon instance exited");
        });

        // setTimeout hack to give go instance chance to set the globals first.
        setTimeout(() => {
          this.convert = globalThis.convert;
          this.listPorts = globalThis.listPorts;
          this.selectPort = globalThis.selectPort;
          this.play = globalThis.play;
          resolve(this);
        }, 0);
      });
    });
  }
}
