package vm

import (
	"fmt"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"io/fs"
	"lightbox/ext"
	"lightbox/ext/modman"
	"lightbox/loghub"
	"lightbox/sandbox"
	"path/filepath"
)

const (
	privateLibPath = "/lib"
)

type AppOption struct {
	sandbox.Option
	Modules  []string          `json:"stdModules,omitempty" yaml:"stdModules"` //启用的模块
	Requires []*modman.Require `json:"requires,omitempty" yaml:"requires"`
}

type VirtualHost struct {
	manager    ConcurrencyMap[string, *sandbox.Applet]
	RepoSource string //仓库目录(zip压缩包的目录）
	RepoDest   string //导入文件的目录
	rootFS     fs.FS
	publicFS   []fs.FS
}

func (v *VirtualHost) NewApplet(opt AppOption) (*sandbox.Applet, error) {
	if app, ok := v.manager.Get(opt.Name); ok {
		return app, fmt.Errorf("sandbox [%s] exists", opt.Name)
	}
	f, err := fs.Sub(v.rootFS, opt.Root)
	if err != nil {
		return nil, err
	}
	app, err := sandbox.NewWithFS(opt.Name, f)
	if err != nil {
		return app, err
	}
	mm, transpiler, hooks := ext.RegistryTable.GetAll(app, opt.Modules...)
	var importers modman.ImportChain
	//第三方包导入路径初始化
	for _, req := range opt.Requires {
		ii, err := modman.NewZipImporterWithDest(filepath.Join(v.RepoSource, req.ZipName()), v.RepoDest, app.Context, transpiler, app.DefaultExt)
		if err != nil {
			log.WithField(sandboxName, opt.Name)
			return nil, err
		}
		importers = append(importers, ii)
	}
	//公共路径
	for _, pfs := range v.publicFS {
		importers = append(importers, modman.NewFSImporter(pfs, app.Context, transpiler, app.DefaultExt))
	}
	app.WithModule(mm, importers).WithTranspiler(transpiler...).WithHook(hooks...)
	return app, nil
}
func (v *VirtualHost) Shutdown(name, reason string) error {
	app, ok := v.manager.Get(name)
	if !ok {
		return fmt.Errorf("sandbox [%s] not exists", name)
	}
	app.Shutdown(reason)
	v.manager.Remove(name)
	return nil
}

var manager = new(VirtualHost)
var subscriber loghub.LogSubscriber

func RegisterAPI(router *mux.Router) {
	if router == nil {
		return
	}
	RegisterAdminAPI(router.PathPrefix("/console").Subrouter())
	RegisterUserAPI(router.PathPrefix("/applet/{sandbox}/").Subrouter())
}

func RegisterAdminAPI(router *mux.Router) {
	if router == nil {
		return
	}
	for p, h := range adminAPIs {
		router.Handle(p, h)
	}
}
func RegisterUserAPI(router *mux.Router) {
	if router == nil {
		return
	}
	for p, h := range appletAPIs {
		router.Handle(p, h)
	}
}
