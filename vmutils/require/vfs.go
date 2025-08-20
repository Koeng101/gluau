// vmutils/require/vfs.go
package require

import (
	"io/fs"
	"strings"
)

type Vfs interface {
	fs.ReadDirFS

	Cwd() string
	Join(path ...string) string
	NormalizePath(path string) string
	IsAbsolutePath(path string) bool
}

// A UnixVfs is a Vfs implementation that uses the standard library's ReadDirFS
// and normalizes paths/handles absolute paths in a Unix-like manner.
type UnixVfs struct {
	fs fs.ReadDirFS
}

func NewUnixVfs(fs fs.ReadDirFS) *UnixVfs {
	return &UnixVfs{fs: fs}
}

func (v *UnixVfs) Open(name string) (fs.File, error) {
	//fmt.Println("Opening file:", name)
	return v.fs.Open(name)
}

func (v *UnixVfs) ReadDir(name string) ([]fs.DirEntry, error) {
	return v.fs.ReadDir(name)
}

func (v *UnixVfs) Cwd() string {
	// Return the current working directory
	return "" // TODO: Support getting the current working directory
}

func (v *UnixVfs) Join(paths ...string) string {
	// Join the given paths using the VFS's path separator
	return strings.Join(paths, "/")
}

func (v *UnixVfs) NormalizePath(path string) string {
	return unixnormalizePath(path)
}

func (v *UnixVfs) IsAbsolutePath(path string) bool {
	return unixisAbsolutePath(path)
}

func vfsIsFile(vfs Vfs, path string) bool {
	// Check if the file exists in the VFS
	_, err := vfs.Open(path)
	return err == nil
}

func vfsIsDir(vfs Vfs, path string) bool {
	// Check if the path is a directory in the VFS
	_, err := vfs.ReadDir(path)
	return err == nil
}
