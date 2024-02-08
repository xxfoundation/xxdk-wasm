.PHONY: update master release update_master update_release build clean binary tests wasm_tests go_tests

clean:
	go mod tidy
	go mod vendor -e
	rm *.wasm
	rm assets/wasm/*

update:
	-GOFLAGS="" go get all

build:
	GOOS=js GOARCH=wasm go build ./...

update_release:
	GOFLAGS="" go get gitlab.com/elixxir/wasm-utils@release
	GOFLAGS="" go get gitlab.com/xx_network/primitives@release
	GOFLAGS="" go get gitlab.com/elixxir/primitives@release
	GOFLAGS="" go get gitlab.com/xx_network/crypto@release
	GOFLAGS="" go get gitlab.com/elixxir/crypto@release
	GOFLAGS="" go get -d gitlab.com/elixxir/client/v4@release

update_master:
	GOFLAGS="" go get gitlab.com/elixxir/wasm-utils@master
	GOFLAGS="" go get gitlab.com/xx_network/primitives@master
	GOFLAGS="" go get gitlab.com/elixxir/primitives@master
	GOFLAGS="" go get gitlab.com/xx_network/crypto@master
	GOFLAGS="" go get gitlab.com/elixxir/crypto@master
	GOFLAGS="" go get -d gitlab.com/elixxir/client/v4@master

binary:
	mkdir -p assets/wasm
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk.wasm main.go
	cp xxdk.wasm assets/wasm/


worker_binaries:
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-channelsIndexedDkWorker.wasm ./indexedDb/impl/channels/...
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-dmIndexedDkWorker.wasm ./indexedDb/impl/dm/...
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-stateIndexedDkWorker.wasm ./indexedDb/impl/state/...
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-logFileWorker.wasm ./logging/workerThread/...
	mkdir -p assets/wasm
	cp xxdk-*.wasm assets/wasm/

binaries: binary worker_binaries

wasmException = "vendor/gitlab.com/elixxir/wasm-utils/exception"

wasm_tests:
	cp $(wasmException)/throws.go $(wasmException)/throws.go.bak
	cp $(wasmException)/throws.dev $(wasmException)/throws.go
	GOOS=js GOARCH=wasm go test -v ./...
	mv $(wasmException)/throws.go.bak $(wasmException)/throws.go

go_tests:
	go test ./... -v

master: update_master clean build

release: update_release clean build

tests: wasm_tests go_tests
