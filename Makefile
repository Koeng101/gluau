ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

library:
	$(MAKE) -C rustlib/rustlib build
	cp target/release/librustlib.a rustlib/librustlib_linux_amd64.a

build:
	go build -o go-rust

all: library build

run: build
	./go-rust

clean:
	rm -rf ./rustlib/rustlib/target
	rm ./rustlib/*.a go-rust