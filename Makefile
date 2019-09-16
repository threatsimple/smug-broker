
VER=0.0.1

default: build/smug

setuplocal:
	mkdir -p build/tmp

cleanuplocal:
	rm -rf build

build/smug: setuplocal main.go
	go build -ldflags "-X main.version=$(VER)" -o build/smug main.go

run: build/smug
	./bin/run_local

clean: setuplocal
	go clean
	rm -rf build

test: setuplocal
	go test -v ./...

