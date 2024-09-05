# xxdk-WASM

This repository contains the WebAssembly bindings for xxDK. It also includes
examples and a test server to serve the compiled WebAssembly module.

**Note:** If you are updating the version of Go that this repository uses, you
need to ensure that you update the `wasm_exec.js` file as described
[below](#wasm_execjs).

Examples for how to use this package can be found at:

https://git.xx.network/xx_network/xxdk-examples

## How to use xxDK in your Web Application

To install with npm, run the following:

```
npm i --save xxdk-wasm
```

By default, xxdk uses a CDN hosted by the [xx
foundation](http://xxfoundation.org). If you would like to self-host
the wasm binaries, you can include the following postinstall script in
your package.json file:

```
{
  ...
  "scripts": {
    ...
    "postinstall": "mkdir -p public && cp -r node_modules/xxdk-wasm public/xxdk-wasm"
  },
  ...
}
```

You may also run this command manually after installation to
accomplish the same result.

Then you will need to override the base path with `setXXDKBasePath`
inside your useEffect statement to load the wasm file from your local
public assets:

```
    xxdk.setXXDKBasePath(window!.location.href + "xxdk-wasm");
```

See the
[xxdk-examples](https://git.xx.network/xx_network/xxdk-examples) repository for
examples in react and plain html/javascript.

For support, please reach out on [Discord](https://disxord.xx.network) or
[Telegram](https://t.me/xxnetwork). We also monitor
[xx general chat](http://alpha.speakeasy.tech/join?0Name=xxGeneralChat&1Description=Talking+about+the+xx+network&2Level=Public&3Created=1674152234202224215&e=%2FqE8BEgQQkXC6n0yxeXGQjvyklaRH6Z%2BWu8qvbFxiuw%3D&k=RMfN%2B9pD%2FJCzPTIzPk%2Bpf0ThKPvI425hye4JqUxi3iA%3D&l=368&m=0&p=1&s=rb%2BrK0HsOYcPpTF6KkpuDWxh7scZbj74kVMHuwhgUR0%3D&v=1)
on xx network.

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
                'gitlab.com/elixxir/wasm-utils/utils.throw': (sp) => {
                    const exception = loadString(sp + 8)
                    const message = loadString(sp + 24)
                    throw globalThis[exception](message)
                },
            }
        }
    }
}
```

# NPM instructions

## Test options

### Link

Run this to make the package globally available:

```
npm link
```

Then in your app for testing:

```
npm link xxdk-wasm
```

### Pack (tarball)

```
npm pack
```

This will create a file like `xxdk-wasm-0.3.19.tgz`. Then you can
install it in your project as follows:

```
npm i ../xxdk-wasm/xxdk-wasm-0.3.19.tgz
```

Some tools cache (e.g., nextjs), and you'll need to remove that when
doing continuous updates between. You can do that in your repo with:

```
git clean -ffdx
#O
rm -rf .next/cache
```

## Publishing

```
npm login
npm publish
```
