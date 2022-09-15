.PHONY: update master release update_master update_release build clean binary

clean:
	rm -rf vendor/
	go mod vendor -e

update:
	-GOFLAGS="" go get all

build:
	GOOS=js GOARCH=wasm go build ./...
	go mod tidy

update_release:
	GOFLAGS="" go get -d gitlab.com/elixxir/client@release
	GOFLAGS="" go get gitlab.com/elixxir/crypto@release
	GOFLAGS="" go get gitlab.com/xx_network/primitives@release

update_master:
	GOFLAGS="" go get -d gitlab.com/elixxir/client@master
	GOFLAGS="" go get gitlab.com/elixxir/crypto@master
	GOFLAGS="" go get gitlab.com/xx_network/primitives@master

binary:
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -o xxdk.wasm main.go

master: update_master clean build

release: update_release clean build
