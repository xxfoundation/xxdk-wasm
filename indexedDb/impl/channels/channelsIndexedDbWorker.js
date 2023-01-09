////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

importScripts('wasm_exec.js');

const go = new Go();
const binPath = 'xxdk-channelsIndexedDkWorker.wasm'
WebAssembly.instantiateStreaming(fetch(binPath), go.importObject).then((result) => {
    go.run(result.instance);
    LogLevel(1);
}).catch((err) => {
    console.error(err);
});