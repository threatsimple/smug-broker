
default: build/smug

setuplocal:
	mkdir -p build/tmp

cleanuplocal:
	rm -rf build

build/smug: setuplocal main.go
	go build -o build/smug main.go

run: build/smug
	./bin/run_it

clean: setuplocal
	go clean
	rm -rf build

test: setuplocal
	go test -v ./...

