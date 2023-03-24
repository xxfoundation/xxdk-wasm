.PHONY: update master release update_master update_release build clean binary tests wasm_tests go_tests

clean:
	go mod tidy
	go mod vendor -e

update:
	-GOFLAGS="" go get all

build:
	GOOS=js GOARCH=wasm go build ./...

update_release:
	GOFLAGS="" go get -d gitlab.com/elixxir/client/v4@release
	GOFLAGS="" go get gitlab.com/elixxir/crypto@release
	GOFLAGS="" go get gitlab.com/elixxir/primitives@release
	GOFLAGS="" go get gitlab.com/xx_network/crypto@release
	GOFLAGS="" go get gitlab.com/xx_network/primitives@release

update_master:
	GOFLAGS="" go get -d gitlab.com/elixxir/client@master
	GOFLAGS="" go get gitlab.com/elixxir/crypto@master
	GOFLAGS="" go get gitlab.com/elixxir/primitives@master
	GOFLAGS="" go get gitlab.com/xx_network/crypto@master
	GOFLAGS="" go get gitlab.com/xx_network/primitives@master

binary:
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk.wasm ./src/api/main.go

worker_binaries:
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-channelsIndexedDkWorker.wasm ./src/api/indexedDb/impl/channels/...
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-dmIndexedDkWorker.wasm ./src/api/indexedDb/impl/dm/...
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk-logFileWorker.wasm ./src/api/logging/workerThread/...

emojis:
	go run -ldflags '-w -s' -trimpath ./emoji/... -o emojiSet.json

binaries: binary worker_binaries

wasm_tests:
	GOOS=js GOARCH=wasm go test -v ./...

go_tests:
	go test ./... -v

master: update_master clean build

release: update_release clean build

tests: wasm_tests go_tests
