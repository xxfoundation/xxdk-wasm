////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// NOTE: wasm_exec.js must always be in the same directory as this script.
importScripts('./wasm_exec.js');
// NOTE: This relative path must be preserved in distribution.
const binPath = require('../wasm/xxdk-logFileWorker.wasm');

const isReady = new Promise((resolve) => {
    self.onWasmInitialized = resolve;
});

const go = new Go();
// NOTE: This is wonky, we are in the right path inside the publicPath
// (usually `dist` folder), which is where webpack paths start. They also
// always prefix with a /, so we are going up a directory to come right back
// into e.g., ../dist/assets/wasm/[somefile].wasm
WebAssembly.instantiateStreaming(fetch('..' + binPath), go.importObject).then(async (result) => {
    go.run(result.instance);
    await isReady;
}).catch((err) => {
    console.error(err);
});
