////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// NOTE: wasm_exec.js must always be in the same directory as this script.
importScripts(require('./wasm_exec.js'));
// NOTE: This relative path must be preserved in distribution.
const binPath = require('../wasm/xxdk-dmIndexedDkWorker.wasm');

const isReady = new Promise((resolve) => {
    self.onWasmInitialized = resolve;
});

const go = new Go();
go.argv = [
    '--logLevel=2',
    '--threadLogLevel=2',
]
WebAssembly.instantiateStreaming(fetch(binPath), go.importObject).then(async (result) => {
    go.run(result.instance);
    await isReady;
}).catch((err) => {
    console.error(err);
});
