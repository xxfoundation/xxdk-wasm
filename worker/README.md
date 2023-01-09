# Web Worker API

This package allows you to create a [Javascript Web Worker](https://developer.mozilla.org/en-US/docs/Web/API/Web_Workers_API)
from WASM and facilitates communication between the web worker and the main
thread using a messaging system.

## Creating a Worker

Web workers have two sides. There is the main side that creates the worker
thread and the thread itself. The thread needs to be compiled into its own WASM
binary and must have a corresponding Javascript file to launch it.

Example `main.go`:

```go
package main

import (
	"fmt"
	"gitlab.com/elixxir/xxdk-wasm/utils/worker"
)

func main() {
	fmt.Println("Starting WebAssembly Worker.")
	_ = worker.NewThreadManager("exampleWebWorker")
	<-make(chan bool)
}
```


Example WASM start file:

```javascript
importScripts('wasm_exec.js');

const go = new Go();
const binPath = 'xxdk-exampleWebWorker.wasm'
WebAssembly.instantiateStreaming(fetch(binPath), go.importObject).then((result) => {
    go.run(result.instance);
}).catch((err) => {
    console.error(err);
});
```

To start the worker, call `worker.NewManager` with the Javascript file to launch
the worker.

```go
wm, err := worker.NewManager("workerWasm.js", "exampleWebWorker")
if err != nil {
	return nil, err
}
```