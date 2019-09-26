

default: build/smug
VER =
ifndef VER
	VER := $(shell ./bin/incr_build ./VERSION)
endif

showme:
	$(info ver is $(VER))
	$(info ver is $(VER))
	$(info ver is $(VER))
	$(info ver is $(VER))


setuplocal:
	mkdir -p build/tmp

cleanuplocal:
	rm -rf build

build/smug: setuplocal main.go
	$(info setting VER to $(VER))
	go build -ldflags "-X main.version=$(VER)" -o build/smug main.go

run: build/smug
	./bin/run_local

clean: setuplocal
	go clean
	rm -rf build

test: export TMPDIR=build/tmp
test: setuplocal
	go test -v ./...

