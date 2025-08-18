//go:build amd64 && linux

package vm

/*
#cgo LDFLAGS: -L../rustlib -lrustlib_linux_amd64 -lstdc++ -lm -ldl
*/
import "C"
