# xxdk-WASM

This repository contains the WebAssembly bindings for xxDK. It also includes
examples and a test server to serve the compiled WebAssembly module.

**Note:** If you are updating the version of Go that this repository uses, you
need to ensure that you update the `wasm_exec.js` file as described
[below](#wasm_execjs).

## Updates

The current semantic version of this repository is stored in `SEMVER` in
`version.go`. When making major updates or updates that create an
incompatibility in the storage or databases, the semantic version needs to be
updated and an upgrade path needs to be provided.

## Building

The repository can only be compiled to a WebAssembly binary using `GOOS=js` and
`GOARCH=wasm`.

```shell
$ GOOS=js GOARCH=wasm go build -o xxdk.wasm
```

### Running Unit Tests

Because the bindings use `syscall/js`, tests cannot only be run in a browser. To
automate this process first get
[wasmbrowsertest](https://github.com/agnivade/wasmbrowsertest) and follow their
[installation instructions](https://github.com/agnivade/wasmbrowsertest#quickstart).
Then, tests can be run using the following command.

```shell
$ GOOS=js GOARCH=wasm go test ./...
```

Note, this will fail because `utils/utils_js.s` contains commands only
recognized by the Go WebAssembly compiler and for some reason are not recognized
by the test runner. To get tests to run, temporarily delete the body of
`utils/utils_js.s` during testing.

## Testing and Examples

The `test` directory contains a basic HTTP server that serves an HTML file
running the WebAssembly binary. This is used for basic testing of the binary.

To run the server, first compile the bindings and save them to the `test`
directory. Then run the server

```shell
$ GOOS=js GOARCH=wasm go build -o test/xxdk.wasm
$ cd test/
$ go run main.go
```

Navigate to http://localhost:9090 to see the webpage.

## `wasm_exec.js`

`wasm_exec.js` is provided by Go and is used to import the WebAssembly module in
the browser. It can be retrieved from Go using the following command.

```shell
$ cp "$(go env GOROOT)/misc/wasm/wasm_exec.js" .
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
