
default: binaries

PACKAGES := $(shell go list -f {{.Dir}} ./...)
GOFILES  := $(addsuffix /*.go,$(PACKAGES))
GOFILES  := $(wildcard $(GOFILES))

VER =
ifndef VER
	VER := $(shell ./bin/incr_build ./VERSION)
endif

setuplocal:
	mkdir -p build/tmp

build/smug: setuplocal main.go
	$(info setting VER to $(VER))
	go build -ldflags "-X main.version=$(VER)" -o build/smug main.go

run: build/smug
	./bin/run_local

clean: setuplocal
	go clean
	rm -rf build

test: export TMPDIR=build/tmp
test: export CGO_ENABLED=0
test: setuplocal
	go test -v ./...

binaries: build/smug-linux-amd64 build/smug-macos-amd64 build/smug-linux-arm64

build/smug-linux-amd64: $(GOFILES)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VER)" -o build/smug-linux-amd64 main.go

build/smug-macos-amd64: $(GOFILES)
	GOOS=darwin GOARCH=amd64 go build -ldflags "-X main.version=$(VER)" -o build/smug-macos-amd64 main.go

build/smug-linux-arm64: $(GOFILES)
	GOOS=linux GOARCH=amd64 go build -ldflags "-X main.version=$(VER)" -o build/smug-linux-arm64  main.go


build-docker: build/smug-linux-amd64
	docker build -t threatsimple/smug:`cat VERSION` .
	docker build -t threatsimple/smug:latest .

tagproj:
	git tag -a v`cat VERSION`
	git push --tags



