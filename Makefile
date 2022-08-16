.PHONY: update master release update_master update_release build clean version

version:
	go run main.go generate
	sed -i.bak 's/package\ cmd/package\ xxdk/g' version_vars.go
	mv version_vars.go xxdk/version_vars.go

clean:
	rm -rf vendor/
	go mod vendor -e

update:
	-GOFLAGS="" go get all

build:
	go build ./...
	go mod tidy

update_release:
	GOFLAGS="" go get gitlab.com/elixxir/client@release

update_master:
	GOFLAGS="" go get gitlab.com/elixxir/client@master

binary:
	GOOS=js GOARCH=wasm go build -ldflags '-w -s' -o xxdk.wasm main.go


master: update_master clean build version

release: update_release clean build
