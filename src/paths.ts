import { BundleVersion } from './version';

// TODO: should this global decl be in like the general types folder?
// we reference it everywhere and it woudl be useful to specify all
// the window properties we need in a single place.
declare global {
  interface Window {
      xxdkBasePath: URL;
  }
}

// TODO: change this to a real cdn
export const xxdk_s3_path = "http://elixxir-bins.s3-us-west-1.amazonaws.com/wasm/";
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
// NOTE: we do not use require here, because these are defined entry points in the webpack config
export function logFileWorkerPath(): URL {
  return new URL(window!.xxdkBasePath + 'dist/logFileWorker.js');
}

export function channelsIndexedDbWorkerPath(): URL {
  return new URL(window!.xxdkBasePath + 'dist/channelsIndexedDbWorker.js');
}

export function dmIndexedDbWorkerPath(): URL {
  return new URL(window!.xxdkBasePath + 'dist/dmIndexedDbWorker.js');
}

export function stateIndexedDbWorkerPath() {
  return new URL(window!.xxdkBasePath + 'dist/stateIndexedDbWorker.js');
}
