// SPDX-License-Identifier: MIT
// Provenance: gopuml cmd/plantuml-wasm/plantuml.mjs

// ES module loader of the WebAssembly build of the playground, for browsers
// and Node.js. Usage:
//
//   import { load } from './gogv.mjs';
//   const gogv = await load();
//   const result = await gogv.render('digraph { a -> b }');
//
// gogv.version names the release of the build. result has svg (a
// Uint8Array, or null after an error), warnings (their texts) and error
// (null, or { message, line }).

let loaded;

export function load(wasmURL = new URL('./gogv.wasm', import.meta.url)) {
  loaded ??= start(wasmURL);
  return loaded;
}

async function start(wasmURL) {
  if (!globalThis.Go) {
    await import('./wasm_exec.js');
  }
  const go = new globalThis.Go();
  let instance;
  if (wasmURL.protocol === 'file:' && typeof process !== 'undefined' && process.versions?.node) {
    const { readFile } = await import('node:fs/promises');
    ({ instance } = await WebAssembly.instantiate(await readFile(wasmURL), go.importObject));
  } else {
    ({ instance } = await WebAssembly.instantiateStreaming(fetch(wasmURL), go.importObject));
  }
  go.run(instance);
  const api = globalThis.gogv;
  return { version: api.version, render: (source) => api.render(source) };
}
