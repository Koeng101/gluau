//go:build cgo

package vm

/*
#cgo linux,amd64 LDFLAGS: -L../rustlib -lrustlib_linux_amd64 -lstdc++ -lm -ldl -lpthread
#cgo linux,arm64 LDFLAGS: -L../rustlib -lrustlib_linux_arm64 -lstdc++ -lm -ldl -lpthread
*/
import "C"
