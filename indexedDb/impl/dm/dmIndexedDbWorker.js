////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

importScripts('wasm_exec.js');

const isReady = new Promise((resolve) => {
    self.onWasmInitialized = resolve;
});

const go = new Go();
go.argv = [
    '--logLevel=2',
    '--threadLogLevel=2',
]
const binPath = 'xxdk-dmIndexedDkWorker.wasm'
WebAssembly.instantiateStreaming(fetch(binPath), go.importObject).then(async (result) => {
    go.run(result.instance);
    await isReady;
}).catch((err) => {
    console.error(err);
});