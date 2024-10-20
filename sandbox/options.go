package sandbox

import (
	"io/fs"
	"lightbox/ext/vfs"
	"os"
)

type Option struct {
	Name       string            `json:"name,omitempty"`       //名称
	DefaultExt string            `json:"defaultExt,omitempty"` //默认扩展文件名
	Volumes    map[string]string `json:"volumes,omitempty"`    //映射卷
	Environ    map[string]string `json:"environ,omitempty"`    //环境变量
	Root       string            `json:"rootDir,omitempty"`    //根目录(根目录不等于实际的磁盘目录),有可能是容器的"子目录"
	fileSystem fs.FS             //文件系统
}

func (opt *Option) WithFS(f fs.FS) *Option {
	opt.fileSystem = vfs.NewVirtualFS(f)
	return opt
}
func (opt *Option) WithRoot(p string) *Option {
	opt.Root = p
	return opt.WithFS(os.DirFS(p))
}
