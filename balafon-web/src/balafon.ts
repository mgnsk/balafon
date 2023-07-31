import init from "./main.wasm?init";
import "./wasm_exec.js";
const go = new globalThis.Go();

enum ResponseKind {
  Error = "error",
  XML = "xml",
}

type ErrorResponse = {
  kind: ResponseKind.Error;
  message: string;
};

type XMLResponse = {
  kind: ResponseKind.XML;
  message: string;
};

type ConvertFunction = (input: string) => ErrorResponse | XMLResponse;

export class Balafon {
  private fn: ConvertFunction;

  async init(): Promise<Balafon> {
    return new Promise((resolve) => {
      init(go.importObject).then((instance) => {
        go.run(instance);

        // console.log("Result:", globalThis.convert(":assign c 60")); // call the 'add' function defined in the Go program

        this.fn = globalThis.convert;
        resolve(this);
      });
    });
  }

  convert(input: string): string {
    let result = this.fn(input);
    switch (result.kind) {
      case ResponseKind.Error:
        throw new Error(result.message);

      case ResponseKind.XML:
        return result.message;
    }
  }
}
