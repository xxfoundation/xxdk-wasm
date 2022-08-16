# xxdk-WASM

WebAssembly bindings for xxDK.

## Building

```shell
$ GOOS=js GOARCH=wasm go build -o xxdk.wasm
```

### Running Tests

To run unit tests, you need to first install
[wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest).

`wasm/wasm_js.s` contains commands only recognized by the Go WebAssembly
compiler and thus cause compile errors when running tests. It needs to be
excluded when running tests.

```shell
$ GOOS=js GOARCH=wasm go test
```

## Testing

The `test` directory contains a website and server to run the compiled
WebAssembly module. `assets` contains the website and `server` contains a small
Go HTTP server.

```shell
$ GOOS=js GOARCH=wasm go build -o test/assets/xxdk.wasm
$ go run test/server/main.go
```

### `wasm_exec.js`

`wasm_exec.js` is provided by Go and is used to import the WebAssembly module in
the browser. It can be retrieved from Go using the following command.

```shell
$ cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" test/assets/
```

Note that this repository makes edits to `wasm_exec.js` and you must either use
the one in this repository or add the following lines in the `go` `importObject`
on `global.Go`.

```javascript
global.Go = class {
    constructor() {
        // ...
        this.importObject = {
            go: {
                // ...
                // func Throw(exception string, message string)
                'gitlab.com/elixxir/client/wasm.Throw': (sp) => {
                    const exception = loadString(sp + 8)
                    const message = loadString(sp + 24)
                    throw globalThis[exception](message)
                },
            }
        }
    }
}
```