build_linux_amd64:
	cargo build-std-release -vvv --target x86_64-unknown-linux-gnu
	x86_64-linux-gnu-strip --strip-debug ../../target/x86_64-unknown-linux-gnu/release/librustlib.a
build_linux_arm64:
	cargo build-std-release -vvv --target aarch64-unknown-linux-gnu
	aarch64-linux-gnu-strip --strip-debug ../../target/aarch64-unknown-linux-gnu/release/librustlib.a

build_normal:
	cargo build --release -vvv