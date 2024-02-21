import { BundleVersion } from './version';

import { startLogFileWorker, startChannelsIndexedDbWorker,
  startDmIndexedDbWorker, startStateIndexedDbWorker } from './workers.js';

// wasmExec is the path to wasm_exec.js, which is also the root at the
// publicPath in webpack set as an entry point in the config.
// FIXME: I could not derive this via require/import and it is not
//        clear from docs how to do so. We don't need to load it because we
//        might downloaded it from a CDN. Hardcoded for now since it does not
//        change. In an ideal world, we'd be able to get this as a hash name and
//        have a special loader that checks the hash.
const wasmExec = '/dist/wasm_exec.js';

// TODO: should this global decl be in like the general types folder?
// we reference it everywhere and it woudl be useful to specify all
// the window properties we need in a single place.
declare global {
  interface Window {
      xxdkBasePath: URL;
      xxdkWasmExecBlobURL: URL;
  }
}

// TODO: change this to a real cdn
export const xxdk_s3_path = "https://elixxir-bins.s3-us-west-1.amazonaws.com/wasm/";
export const default_xxdk_path = new URL(BundleVersion, xxdk_s3_path);
// TODO: docstring?
export let xxdkBasePath : URL | undefined;

if (typeof window! !== 'undefined') {
  if (typeof window!.xxdkBasePath == 'undefined') {
    window!.xxdkBasePath = default_xxdk_path;
  }
}
if (xxdkBasePath === undefined) {
  xxdkBasePath = default_xxdk_path;
}

export function setXXDKBasePath(newPath: URL) {
  if (typeof window! !== 'undefined') {
    window!.xxdkBasePath = newPath;
  }
  xxdkBasePath = newPath;
}

// TODO: These functions should be kept internal to this package but the current API and legacy
// apps need to maintain access
export async function logFileWorkerPath(): Promise<URL> {
  const binPath = require('../assets/wasm/xxdk-logFileWorker.wasm');
  const wasm = new URL(window!.xxdkBasePath + binPath.toString());
  console.info("Loading logFileWorker (" + wasm + ")");
  return downloadWorkerToBlobURL(wasm, startLogFileWorker);
}

export async function channelsIndexedDbWorkerPath(): Promise<URL> {
  const binPath = require('../assets/wasm/xxdk-channelsIndexedDbWorker.wasm');
  const wasm = new URL(window!.xxdkBasePath + binPath.toString());
  console.info("Loading channelsIndexedDbWorker (" + wasm + ")");
  return downloadWorkerToBlobURL(wasm, startChannelsIndexedDbWorker);
}

export async function dmIndexedDbWorkerPath(): Promise<URL> {
  const binPath = require('../assets/wasm/xxdk-dmIndexedDbWorker.wasm');
  const wasm = new URL(window!.xxdkBasePath + binPath.toString());
  console.info("Loading dmIndexedDbWorker (" + wasm + ")");
  return downloadWorkerToBlobURL(wasm, startDmIndexedDbWorker);
}

export async function stateIndexedDbWorkerPath(): Promise<URL> {
  const binPath = require('../assets/wasm/xxdk-stateIndexedDbWorker.wasm');
  const wasm = new URL(window!.xxdkBasePath + binPath.toString());
  console.info("Loading stateIndexedDbWorker (" + wasm + ")");
  return downloadWorkerToBlobURL(wasm, startStateIndexedDbWorker);
}

// NOTE: Whereas we can load the workers via function.toString(), they all need
// to reference the wasm_exec.js import script, preferably the same one. To handle this,
// we download that script into it's own blob url, then explicitly call importScrips(blob url).
async function downloadWorkerToBlobURL(wasm: URL, workerFn: (wasm: URL) => any): Promise<URL> {
  const wasmExec = await wasmExecBlob();
  const blobElems = [
    "importScripts('" + wasmExec + "');  ",
    "(" + workerFn.toString() + ")('" + wasm + "');"
  ];
  const urlStr = URL.createObjectURL(new Blob(blobElems, {type: 'text/javascript'}));
  console.info("[XXDK] Loaded " + wasm.toString() + " worker at: " + urlStr.toString());
  console.trace("[XXDK] worker contents: " + blobElems.toString());
  return new URL(urlStr);
}

// wasmExecBlob downloads the wasm_exec.js file from the xxdkBasePath, which could
// be locally hosted or on a CDN, then makes it available via a blob url.
async function wasmExecBlob(): Promise<URL> {
  if (window!.xxdkWasmExecBlobURL !== undefined) {
    return window!.xxdkWasmExecBlobURL;
  }

  const url = new URL(window!.xxdkBasePath + wasmExec.toString());
  console.trace("[XXDK] wasm_exec.js download url: " + url.toString());
  try {
    const response = await fetch(url);
    const data = await response.text();
    const urlStr = URL.createObjectURL(new Blob([data], {type: 'text/javascript'}));
    window!.xxdkWasmExecBlobURL = new URL(urlStr);
  } catch (x) {
    console.error("[XXDK] Unable to load wasm_exec.js into a blob url: " + x);
    throw(x);
  }
  console.info("[XXDK] wasm_exec.js loaded at: " + window!.xxdkWasmExecBlobURL);
  return window!.xxdkWasmExecBlobURL;
}
