.PHONY: update master release update_master update_release build clean binary tests wasm_tests go_tests

clean:
	go mod tidy
	go mod vendor -e
	-rm *.wasm
	-rm assets/wasm/*

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
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-channelsIndexedDbWorker.wasm ./indexedDb/impl/channels/...
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-dmIndexedDbWorker.wasm ./indexedDb/impl/dm/...
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-stateIndexedDbWorker.wasm ./indexedDb/impl/state/...
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-logFileWorker.wasm ./logging/workerThread/...
	mkdir -p assets/wasm
	cp xxdk-*.wasm assets/wasm/

binaries: binary worker_binaries

wasm_tests:
	cp $(shell go env GOROOT)/misc/wasm/wasm_exec.js wasm_exec.js.bak
	cp wasm_exec.js $(shell go env GOROOT)/misc/wasm/wasm_exec.js
	- GOOS=js GOARCH=wasm go test -v ./...
	mv wasm_exec.js.bak $(shell go env GOROOT)/misc/wasm/wasm_exec.js

go_tests:
	go test ./... -v

master: update_master clean build

release: update_release clean build

tests: wasm_tests go_tests
