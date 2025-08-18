ROOT_DIR := $(dir $(realpath $(lastword $(MAKEFILE_LIST))))

# Before running this, ensure you have the necessary targets installed:
# rustup default nightly # build_std needs nightly
# rustup target add x86_64-unknown-linux-gnu
# rustup target add aarch64-unknown-linux-gnu
# sudo apt install g++-aarch64-linux-gnu
library:
	$(MAKE) -C rustlib/rustlib build_normal
	cp target/release/librustlib.a rustlib/librustlib_linux_amd64.a

# Before running this, ensure you have the necessary targets installed:
# rustup default nightly # build_std needs nightly
# rustup target add x86_64-unknown-linux-gnu
# rustup target add aarch64-unknown-linux-gnu
# sudo apt install g++-aarch64-linux-gnu
distlib:
	$(MAKE) -C rustlib/rustlib build_linux_amd64
	cp target/x86_64-unknown-linux-gnu/release/librustlib.a rustlib/librustlib_linux_amd64.a
	$(MAKE) -C rustlib/rustlib build_linux_arm64
	cp target/aarch64-unknown-linux-gnu/release/librustlib.a rustlib/librustlib_linux_arm64.a

build:
	go build -o go-rust
	GOOS=linux GOARCH=amd64 go build -o go-rust-linux-amd64
	CGO_ENABLED=1 GOOS=linux GOARCH=arm64 CC=aarch64-linux-gnu-gcc go build -o go-rust-linux-arm64

all: library build

run: build
	./go-rust

clean:
	rm -rf ./rustlib/rustlib/target ./target
	rm ./rustlib/*.a go-rust