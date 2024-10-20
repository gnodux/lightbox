package vfs

import (
	"archive/zip"
	"fmt"
	"io/fs"
)

type ZipFS struct {
	zipFile string
}

type zipFileWrap struct {
	zr *zip.ReadCloser
	fs.File
}

func (z *zipFileWrap) Close() error {
	fmt.Println("close")
	if z != nil && z.zr != nil {
		_ = z.zr.Close()
	}
	return z.File.Close()
}

func (z *ZipFS) Open(name string) (fs.File, error) {
	zr, err := zip.OpenReader(z.zipFile)
	if err != nil {
		return nil, err
	}
	f, err := zr.Open(name)
	if err != nil {
		return nil, err
	}
	return &zipFileWrap{
		zr:   zr,
		File: f,
	}, nil
}

func (z *ZipFS) String() string {
	return "ZipFS(" + z.zipFile + ")"
}

func NewZipFS(zipFile string) *ZipFS {
	return &ZipFS{zipFile: zipFile}
}
