////////////////////////////////////////////////////////////////////////////////
// Copyright © 2024 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// NOTE: These javascript functions are used to load web workers. They
// can't call other functions and they must use the webworker
// available APIs to function. Thats why they are duplicate code.

export function startLogFileWorker(wasm) {
  console.trace("[XXDK] logFileWorker loading from: " + wasm.toString());
  const isReady = new Promise((resolve) => {
    self.onWasmInitialized = resolve;
  });

  const go = new Go();
  WebAssembly.instantiateStreaming(fetch(wasm), go.importObject).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] logFileWorker started");
  }).catch((err) => {
    console.error(err);
  });
}

export function startChannelsIndexedDbWorker(wasm) {
  console.trace("[XXDK] channelsIndexedDbWorker loading from: " + wasm.toString());
  const isReady = new Promise((resolve) => {
    self.onWasmInitialized = resolve;
  });
  const go = new Go();
  go.argv = [
    '--logLevel=2',
    '--threadLogLevel=2',
  ]
  WebAssembly.instantiateStreaming(fetch(wasm), go.importObject).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] channelsIndexedDbWorker started");
  }).catch((err) => {
    console.error(err);
  });
}

export function startDmIndexedDbWorker(wasm) {
  console.trace("[XXDK] dmIndexedDbWorker loading from: " + wasm.toString());
  const isReady = new Promise((resolve) => {
    self.onWasmInitialized = resolve;
  });
  const go = new Go();
  go.argv = [
    '--logLevel=2',
    '--threadLogLevel=2',
  ]
  WebAssembly.instantiateStreaming(fetch(wasm), go.importObject).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] dmIndexedDbWorker started");
  }).catch((err) => {
    console.error(err);
  });
}

export function startStateIndexedDbWorker(wasm) {
  console.trace("[XXDK] stateIndexedDbWorker loading from: " + wasm.toString());
  const isReady = new Promise((resolve) => {
    self.onWasmInitialized = resolve;
  });
  const go = new Go();
  go.argv = [
    '--logLevel=2',
    '--threadLogLevel=2',
  ]
  WebAssembly.instantiateStreaming(fetch(wasm), go.importObject).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] stateIndexedDbWorker started");
  }).catch((err) => {
    console.error(err);
  });
}

