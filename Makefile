.PHONY: update master release update_master update_release build clean binary tests wasm_tests go_tests

clean:
	go mod tidy
	go mod vendor -e
	go clean -cache
	-rm -f *.wasm
	-rm -rf assets/wasm/*
	-rm -rf dist/

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
	cp "$$(go env GOROOT)/lib/wasm/wasm_exec.js" ./wasm_exec.js
	@echo "Building combined WASM binary with all workers..."
	GOOS=js GOARCH=wasm go build \
		-mod=vendor \
		-gcflags=all="-l -B -wb=false" \
		-ldflags '-w -s' \
		-trimpath \
		-tags netgo \
		-o xxdk.wasm
	@if command -v wasm-opt >/dev/null 2>&1; then \
		echo "Optimizing with wasm-opt..."; \
		cp xxdk.wasm xxdk.wasm.orig; \
		wasm-opt xxdk.wasm.orig --enable-bulk-memory -Oz -o xxdk.wasm; \
		orig_size=$$(du -h xxdk.wasm.orig 2>/dev/null | cut -f1 || echo "?"); \
		new_size=$$(du -h xxdk.wasm | cut -f1); \
		echo "  Before: $$orig_size  After: $$new_size"; \
		rm -f xxdk.wasm.orig; \
	fi
	cp xxdk.wasm assets/wasm/
	@echo "Done! Combined binary size: $$(du -h xxdk.wasm | cut -f1)"

binaries: binary

wasm_tests:
	@echo "Running WASM tests (requires wasmbrowsertest)"
	@if ! command -v wasmbrowsertest >/dev/null 2>&1; then \
		echo "Error: wasmbrowsertest not found. Install with:"; \
		echo "  go install github.com/agnivade/wasmbrowsertest@latest"; \
		exit 1; \
	fi
	GOOS=js GOARCH=wasm go test -exec=wasmbrowsertest -v ./...

go_tests:
	go test ./... -v

master: update_master clean build

release: update_release clean build

tests: wasm_tests go_tests
