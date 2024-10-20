package sandbox

import (
	"context"
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"io/fs"
	"lightbox/env"
	"lightbox/ext/transpile"
	"lightbox/ext/util"
	"lightbox/ext/vfs"
	"os"
	"strings"
	"sync"
	"time"
)

/**
done:
1. Applet的PrivateLib path怎么做？(完成)
2. 全局Transpiler和Applet引用模块的Transpiler要整合进去（完成）
*/

type Signal int

const (
	SigEmpty Signal = 0
	SigStart Signal = 1 << iota
	SigInitialized
	SigBeforeRun
	SigStop
)

type SignalHookFn func(applet *Applet, signal Signal)

func NewHook(signal Signal, fn func(applet *Applet) error) SignalHookFn {
	return func(applet *Applet, sig Signal) {
		if sig == signal {
			if err := fn(applet); err != nil {
				applet.Logger.Error("call hook error", err)
			}
		}
	}
}

type moduleGroup struct {
	getters []tengo.ModuleGetter
}

func (m *moduleGroup) Get(name string) tengo.Importable {
	for _, getter := range m.getters {
		if mod := getter.Get(name); mod != nil {
			return mod
		}
	}
	return nil
}

// Applet 轻量级Applet
type Applet struct {
	Option
	modules             *moduleGroup     //module
	Context             *env.Environment //应用执行环境(实例容器、变量等)
	Logger              *log.Entry       //日志入口
	util.CompileService                  //编译服务
	transpiler          transpile.Group  //转译服务
	hooks               []SignalHookFn   //applet 生命周期的hooks
	//pool                sync.Pool
	config   map[string]interface{} //应用配置
	mx       sync.Mutex
	initOnce sync.Once
}

// WithModule 注册模块
func (app *Applet) WithModule(getters ...tengo.ModuleGetter) *Applet {
	app.mx.Lock()
	defer app.mx.Unlock()
	app.modules.getters = append(app.modules.getters, getters...)
	return app
}

// WithHook 注册AppHook
func (app *Applet) WithHook(hooks ...SignalHookFn) *Applet {
	app.mx.Lock()
	defer app.mx.Unlock()
	app.hooks = append(app.hooks, hooks...)
	return app
}

// WithTranspiler 注册转译器
func (app *Applet) WithTranspiler(trans ...transpile.TransFunc) *Applet {
	app.transpiler = append(app.transpiler, trans...)
	return app
}

// DoNotify 通知已经注册的Hook
func (app *Applet) DoNotify(signals ...Signal) {
	for _, hook := range app.hooks {
		for _, sig := range signals {
			hook(app, sig)
		}
	}
}

// Mount 挂载文件系统
func (app *Applet) Mount(prefix string, sub fs.FS) error {
	return vfs.Mount(app.fileSystem, prefix, sub)
}

// Open 打开文件
func (app *Applet) Open(name string) (fs.File, error) {
	if app == nil || app.fileSystem == nil {
		return nil, fmt.Errorf("sandbox %s not support filesystem", app.Name)
	}
	return app.fileSystem.Open(name)
}

func (app *Applet) Stat(name string) (fs.FileInfo, error) {
	if app == nil || app.fileSystem == nil {
		return nil, fmt.Errorf("sandbox %s not support filesystem", app.Name)
	}
	return fs.Stat(app.fileSystem, name)
}

func (app *Applet) Shutdown(reason string) {
	entry := log.WithField("sandbox", app.Name)
	entry.WithField("reason", reason).Info("shutting down")
	app.DoNotify(SigStop)
	entry.Info("stopped")

}

func New(opt Option) (*Applet, error) {
	if opt.Name == "" {
		return nil, errors.New("require name")
	}
	app := &Applet{
		Option:  opt,
		Logger:  log.WithField("sandbox", opt.Name),
		modules: &moduleGroup{},
	}
	app.CompileService = util.NewScriptCache(time.Second*10, app.fileSystem, app.Compile)
	//注册全局的transpiler
	app.WithTranspiler(transpile.G...)

	app.Context = new(env.Environment)
	//继承自系统的所有环境变量(默认不继承，如果真需要继承，则由sandbox之外的控制器负责)
	//初始化私有环境变量
	for k, v := range opt.Environ {
		app.Context.Set(k, v)
	}
	app.Logger.WithField("env", opt.Environ).Info("initialize environment")
	if v, ok := env.Get[string]("profile"); !ok || v == "" {
		app.Context.Set("profile", "test")
	}
	if app.DefaultExt == "" {
		app.DefaultExt = tengo.SourceFileExtDefault
	}
	app.DoNotify(SigStart)
	return app, nil
}
func NewWithDir(name string, root string) (*Applet, error) {
	return NewWithFS(name, os.DirFS(root))
}
func NewWithFS(name string, f fs.FS) (*Applet, error) {
	return New(Option{
		Name:       name,
		DefaultExt: tengo.SourceFileExtDefault,
		fileSystem: vfs.NewVirtualFS(f),
	})
}

// GetDefaultExt get default script extension
func (app *Applet) GetDefaultExt() string {
	return app.DefaultExt
}

var configFileName = []string{
	"application_{profile}.yml",
	"application.yml",
}
var configDir = []string{
	"config_{profile}",
	"config",
}

func (app *Applet) Initialize() {
	app.initOnce.Do(func() {
		app.DoNotify(SigInitialized)
	})
}

func (app *Applet) Config() map[string]interface{} {
	app.mx.Lock()
	defer app.mx.Unlock()
	if app.config == nil {
		var allConfig []map[string]interface{}
		logger := log.WithField("sandbox", app.Name)
		//读取配置文件,有限读取指定profile
		for _, cfgName := range configFileName {
			realCfg, err := app.Context.Parse(cfgName)
			if err != nil {
				logger.Info("parse environment error:", err)
				continue
			}
			if _, err = fs.Stat(app, realCfg); err != nil {
				logger.Debug(realCfg, " not exists, skipped")
				continue
			}
			logger.Info("read config file from:", realCfg)
			data, err := fs.ReadFile(app, realCfg)
			if err != nil {
				logger.Info("read config error:", err)
				continue
			}
			cfg := make(map[string]interface{})
			if err = yaml.Unmarshal(data, &cfg); err == nil {
				allConfig = append(allConfig, cfg)
				break
			} else {
				logger.Errorf("load yml config %s error:%s ", realCfg, err)
			}

		}

		//寻找配置目录
		for _, cfgDir := range configDir {
			dirName, _ := app.Context.Parse(cfgDir)
			if fi, err := app.Stat(dirName); err != nil || !fi.IsDir() {
				logger.Info(dirName, ":", err)
				continue
			}
			cfgDirFS, err := fs.Sub(app, dirName)
			if err != nil {
				logger.Debug("get sub filesystem error:", err)
				continue
			}
			matches, err := fs.Glob(cfgDirFS, "*.yml")
			if err != nil {
				logger.Debug("get sub config files error:", err)
				continue
			}
			for _, mf := range matches {
				if data, err := fs.ReadFile(cfgDirFS, mf); err == nil {
					var m map[string]interface{}
					if err = yaml.Unmarshal(data, &m); err == nil {
						allConfig = append(allConfig, m)
					} else {
						logger.Error("unmarshal config file ", cfgDir, "/", mf, " error:", err)
					}
				} else {
					logger.Error("read config file error:", err)
				}
			}
		}
		app.config = util.MergeMap(allConfig...)
	}
	return app.config
}
func (app *Applet) GetConfig(key string) (result interface{}, ok bool) {
	keys := strings.Split(key, ".")
	currentMap := app.Config()
	var currentResult interface{}
	var currentOk bool
	for _, k := range keys {
		currentResult, currentOk = currentMap[k]
		if !currentOk {
			break
		} else {
			result = currentResult
			ok = currentOk
		}
		if currentMap, currentOk = result.(map[string]interface{}); !currentOk {
			break
		}
	}
	return
}

func (app *Applet) Transpile(src []byte) ([]byte, error) {
	return app.transpiler.Transpile(src)
}

func (app *Applet) Compile(src []byte, placeHolder util.PlaceHolders, fileName string) (*tengo.Compiled, error) {
	//在第一次编译的时候，初始化应用(SigInitialized)
	app.Initialize()
	var err error
	src, err = app.transpiler.Transpile(src)
	if err != nil {
		return nil, err
	}
	script := tengo.NewScriptWith(src, fileName, app.DefaultExt)
	for k, _ := range placeHolder {
		//编译时，只设置占位符(变量定义，不做实际的值)
		if err := script.Add(k, nil); err != nil {
			return nil, err
		}
	}
	script.SetImports(app.modules)
	return script.Compile()
}

func (app *Applet) RunFileContext(ctx context.Context, fileName string, args map[string]interface{}) (*tengo.Compiled, error) {
	if args == nil {
		args = map[string]interface{}{}
	}
	compiled, err := app.CompileService.GetCompiled(fileName, args)
	if err != nil {
		return compiled, err
	}
	for k, v := range args {
		if err = compiled.Set(k, v); err != nil {
			return nil, err
		}
	}
	err = compiled.RunContext(ctx)
	return compiled, err
}

func (app *Applet) RunFile(fileName string, args map[string]interface{}) (*tengo.Compiled, error) {
	return app.RunFileContext(context.Background(), fileName, args)
}

func (app *Applet) RunContext(ctx context.Context, src []byte, args map[string]interface{}, fileName string) (*tengo.Compiled, error) {
	if args == nil {
		args = map[string]interface{}{}
	}
	compiled, err := app.Compile(src, args, fileName)
	if err != nil {
		return nil, err
	}
	for k, v := range args {
		if err = compiled.Set(k, v); err != nil {
			return nil, err
		}
	}
	err = compiled.RunContext(ctx)
	return compiled, err
}

func (app *Applet) Run(src []byte, args map[string]interface{}, fileName string) (*tengo.Compiled, error) {
	return app.RunContext(context.Background(), src, args, fileName)
}
