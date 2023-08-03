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

const startedPromise = new Promise<void>((resolve) => {
  globalThis.resolveStartedPromise = resolve;
});

export class Balafon {
  started: Promise<void>;
  convert: ConvertFn;
  listPorts: ListPortsFn;
  selectPort: SelectPortFn;
  play: PlayFn;

  constructor() {
    init(go.importObject).then((instance) => {
      go.run(instance).then(() => {
        throw new Error("balafon instance exited");
      });
    });
  }

  async init() {
    await startedPromise;
    this.convert = globalThis.convert;
    this.listPorts = globalThis.listPorts;
    this.selectPort = globalThis.selectPort;
    this.play = globalThis.play;
  }
}

export const balafon = new Balafon();
