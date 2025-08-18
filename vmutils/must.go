//go:build cgo

package vmutils

// Panics if err != nil, else returns obj
func Must[T any](obj T, err error) T {
	if err != nil {
		panic(err)
	}
	return obj
}

// Panics if err != nil, else returns obj
func MustOk(err error) {
	if err != nil {
		panic(err)
	}
}
