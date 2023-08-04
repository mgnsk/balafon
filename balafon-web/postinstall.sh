#!/usr/bin/env bash

set -euo pipefail

cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" src/wasm_exec.js
