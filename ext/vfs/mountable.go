package vfs

import (
	"errors"
	"fmt"
	"io/fs"
	"strings"
)

type MountFS interface {
	Mount(string, fs.FS) error
}

func Mount(root fs.FS, prefix string, sub fs.FS) error {
	if f, ok := root.(MountFS); ok {
		return f.Mount(prefix, sub)
	}
	return errors.New("fs is not mountable")
}

type VirtualFS struct {
	rootFS    fs.FS
	mountedFS map[string]fs.FS
}

func (f *VirtualFS) Stat(name string) (fs.FileInfo, error) {
	for prefix, fsys := range f.mountedFS {
		if strings.HasPrefix(name, prefix) {
			return fs.Stat(fsys, name[len(prefix):])
		}
	}
	return fs.Stat(f.rootFS, name)
}

func (f *VirtualFS) Open(name string) (fs.File, error) {
	for prefix, fsys := range f.mountedFS {
		if strings.HasPrefix(name, prefix) {
			return fsys.Open(name[len(prefix):])
		}
	}
	return f.rootFS.Open(name)
}

func (f *VirtualFS) Mount(mountPath string, fsys fs.FS) error {
	if mountPath == "" {
		return fmt.Errorf("mount path is empty")
	}
	if !fs.ValidPath(mountPath) {
		return fmt.Errorf("mount path %s is not a valid path", mountPath)
	}
	if fsys == nil {
		return errors.New("fs is nil")
	}
	f.mountedFS[mountPath] = fsys
	return nil
}

func NewVirtualFS(rootFS fs.FS) fs.FS {
	return &VirtualFS{rootFS: rootFS, mountedFS: nil}
}
