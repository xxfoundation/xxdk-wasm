# xxdk-WASM

This repository contains the WebAssembly bindings for xxDK. It also includes a
test server to serve the compiled WebAssembly module.

## Building

The repository can only be compiled to a WebAssembly binary using `GOOS=js` and
`GOARCH=wasm`.

```shell
$ GOOS=js GOARCH=wasm go build -o xxdk.wasm
```

### Running Unit Tests

Because the bindings use `syscall/js`, tests cannot only be run in a browser. To
automate this process first install
[wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest). Then, tests can
be run using the following command.

```shell
$ GOOS=js GOARCH=wasm go test ./...
```

Note, this will fail because `utils/utils_js.s` contains commands only recognized
by the Go WebAssembly compiler and for some reason not recognized by the test
runner. To get tests to run, temporarily delete the body of `utils/utils_js.s`
during testing.

## Testing

The `test` directory contains `assets`, a simple web page to run the Javascript,
and `server`, which runs a simple Go HTTP server to deliver the webpage.

To run the server, first compile the bindings and save them to the `assets`
directory. Then run the server

```shell
$ GOOS=js GOARCH=wasm go build -o test/assets/xxdk.wasm
$ cd test/server/
$ go run main.go
```

Navigate to http://localhost:9090 to see the web page.

## `wasm_exec.js`

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
                'gitlab.com/elixxir/xxdk-wasm/utils.throw': (sp) => {
                    const exception = loadString(sp + 8)
                    const message = loadString(sp + 24)
                    throw globalThis[exception](message)
                },
            }
        }
    }
}
```
