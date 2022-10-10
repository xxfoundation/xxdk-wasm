.PHONY: update master release update_master update_release build clean binary tests wasm_tests go_tests

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
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -trimpath -o xxdk.wasm main.go

wasm_tests:
	cp utils/utils_js.s utils/utils_js.s.bak
	> utils/utils_js.s
	-GOOS=js GOARCH=wasm go test ./... -v
	mv utils/utils_js.s.bak utils/utils_js.s

go_tests:
	go test ./... -v

master: update_master clean build

release: update_release clean build

tests: wasm_tests go_tests
