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

  // Debug: Check the fetch response before instantiateStreaming
  fetch(wasm).then(async (response) => {
    console.log("[XXDK] logFileWorker fetch response status:", response.status);
    console.log("[XXDK] logFileWorker fetch response headers:", [...response.headers.entries()]);
    console.log("[XXDK] logFileWorker fetch content-type:", response.headers.get('content-type'));

    if (!response.ok) {
      const text = await response.text();
      console.error("[XXDK] logFileWorker fetch failed. Response body:", text.substring(0, 500));
      throw new Error(`Failed to fetch WASM: ${response.status} ${response.statusText}`);
    }

    return WebAssembly.instantiateStreaming(response, go.importObject);
  }).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] logFileWorker started");
  }).catch((err) => {
    console.error("[XXDK] logFileWorker ERROR:", err);
    console.error("[XXDK] logFileWorker stack:", err.stack);
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

  fetch(wasm).then(async (response) => {
    console.log("[XXDK] channelsIndexedDbWorker fetch response status:", response.status);
    console.log("[XXDK] channelsIndexedDbWorker fetch content-type:", response.headers.get('content-type'));

    if (!response.ok) {
      const text = await response.text();
      console.error("[XXDK] channelsIndexedDbWorker fetch failed. Response body:", text.substring(0, 500));
      throw new Error(`Failed to fetch WASM: ${response.status} ${response.statusText}`);
    }

    return WebAssembly.instantiateStreaming(response, go.importObject);
  }).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] channelsIndexedDbWorker started");
  }).catch((err) => {
    console.error("[XXDK] channelsIndexedDbWorker ERROR:", err);
    console.error("[XXDK] channelsIndexedDbWorker stack:", err.stack);
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

  fetch(wasm).then(async (response) => {
    console.log("[XXDK] dmIndexedDbWorker fetch response status:", response.status);
    console.log("[XXDK] dmIndexedDbWorker fetch content-type:", response.headers.get('content-type'));

    if (!response.ok) {
      const text = await response.text();
      console.error("[XXDK] dmIndexedDbWorker fetch failed. Response body:", text.substring(0, 500));
      throw new Error(`Failed to fetch WASM: ${response.status} ${response.statusText}`);
    }

    return WebAssembly.instantiateStreaming(response, go.importObject);
  }).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] dmIndexedDbWorker started");
  }).catch((err) => {
    console.error("[XXDK] dmIndexedDbWorker ERROR:", err);
    console.error("[XXDK] dmIndexedDbWorker stack:", err.stack);
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

  fetch(wasm).then(async (response) => {
    console.log("[XXDK] stateIndexedDbWorker fetch response status:", response.status);
    console.log("[XXDK] stateIndexedDbWorker fetch content-type:", response.headers.get('content-type'));

    if (!response.ok) {
      const text = await response.text();
      console.error("[XXDK] stateIndexedDbWorker fetch failed. Response body:", text.substring(0, 500));
      throw new Error(`Failed to fetch WASM: ${response.status} ${response.statusText}`);
    }

    return WebAssembly.instantiateStreaming(response, go.importObject);
  }).then(async (result) => {
    go.run(result.instance);
    await isReady;
    console.info("[XXDK] stateIndexedDbWorker started");
  }).catch((err) => {
    console.error("[XXDK] stateIndexedDbWorker ERROR:", err);
    console.error("[XXDK] stateIndexedDbWorker stack:", err.stack);
  });
}

