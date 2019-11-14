

default: build/smug
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

