package util

import (
	"github.com/d5/tengo/v2"
	"golang.org/x/sync/singleflight"
	"io/fs"
	"sort"
	"strings"
	"sync"
	"time"
)

type CachedItem struct {
	Compiled   *tengo.Compiled
	LastModify time.Time
	LastCheck  time.Time
	Error      error
}

const ScriptName = "__name__"

var DefaultPlaceHolder = PlaceHolders{
	ScriptName: "",
}

type PlaceHolders map[string]interface{}

func (p PlaceHolders) Names() []string {
	if p != nil {
		var names []string
		for k, _ := range p {
			names = append(names, k)
		}
		//map中key可能是无序的,导致缓存多个版本，简单的做饭就是names排序
		sort.Strings(names)
		return names
	}
	return []string{}
}

type CompileFunc func(src []byte, placeHolder PlaceHolders, fileName string) (*tengo.Compiled, error)

//type SourceLoaderFunc func(srcFile string) ([]byte, error)
//type SourceStateFunc func(srcFile string) (time.Time, error)

//var DefaultScriptCache = NewScriptCache(500*time.Millisecond, FileStat, ioutil.ReadFile, DefaultCompiler)

//
//func NewCompiler(modules tengo.ModuleGetter, extendFc ...func(script *tengo.Script)) CompileFunc {
//	return func(src []byte, placeHolder []string, fileName string) (*tengo.Compiled, error) {
//		script := tengo.NewScriptWith(src, fileName, filepath.Ext(fileName))
//		script.SetImports(modules)
//		for _, f := range extendFc {
//			f(script)
//		}
//		for _, p := range placeHolder {
//			err := script.WithRegistry(p, tengo.UndefinedValue)
//			if err != nil {
//				return nil, err
//			}
//		}
//		for k, v := range DefaultPlaceHolder {
//			err := script.WithRegistry(k, v)
//			if err != nil {
//				return nil, err
//			}
//		}
//		compiled, err := script.Compile()
//		if err != nil {
//			log.Infof("compile %s error:%v", fileName, err)
//		}
//		return compiled, err
//	}
//}

//func DefaultCompiler(src []byte, placeHolder []string, fileName string) (*tengo.Compiled, error) {
//	script := tengo.NewScriptWith(src, fileName, filepath.Ext(fileName))
//	for _, p := range placeHolder {
//		err := script.Add(p, tengo.UndefinedValue)
//		if err != nil {
//			return nil, err
//		}
//	}
//	for k, v := range DefaultPlaceHolder {
//		err := script.Add(k, v)
//		if err != nil {
//			return nil, err
//		}
//	}
//	compiled, err := script.Compile()
//
//	if err != nil {
//		log.Infof("compile %s error:%v", fileName, err)
//	}
//	return compiled, err
//}
//func FileStat(srcFile string) (time.Time, error) {
//	if fi, err := os.Stat(srcFile); err != nil {
//		return time.Now(), err
//	} else {
//		return fi.ModTime(), nil
//	}
//}

func NewScriptCache(dur time.Duration, f fs.FS, compileFunc CompileFunc) *ScriptCache {
	return &ScriptCache{
		compileFunc:   compileFunc,
		fileSystem:    f,
		checkDuration: dur,
		m:             &sync.Map{},
		g:             &singleflight.Group{},
	}
}

type CompileService interface {
	GetCompiled(srcFile string, placeHolder PlaceHolders) (*tengo.Compiled, error)
}

type ScriptCache struct {
	m             *sync.Map
	g             *singleflight.Group
	compileFunc   CompileFunc
	fileSystem    fs.FS
	checkDuration time.Duration
}

func (c *ScriptCache) Clean() {
	c.m.Range(func(key, value interface{}) bool {
		c.m.Delete(key)
		return true
	})
}

func (c *ScriptCache) state(srcFile string) (time.Time, error) {
	fi, err := fs.Stat(c.fileSystem, srcFile)
	if err != nil {
		return time.Now(), err
	}
	return fi.ModTime(), nil
}

func (c *ScriptCache) GetCompiled(srcFile string, placeHolder PlaceHolders) (*tengo.Compiled, error) {
	key := srcFile + "[" + strings.Join(placeHolder.Names(), ",") + "]"
	result, _, _ := c.g.Do(key, func() (interface{}, error) {
		var cachedItem *CachedItem
		itm, ok := c.m.Load(key)
		if !ok {
			//缓存中不存在
			cachedItem = &CachedItem{
				LastCheck: time.Now(),
			}
			cachedItem.LastModify, cachedItem.Error = c.state(srcFile)
			if cachedItem.Error == nil {
				if src, err := fs.ReadFile(c.fileSystem, srcFile); err == nil {
					cachedItem.Compiled, cachedItem.Error = c.compileFunc(src, placeHolder, srcFile)
				} else {
					cachedItem.Error = err
				}
			}
			c.m.Store(key, cachedItem)
		} else {
			//缓存中存在
			cachedItem = itm.(*CachedItem)
			if time.Since(cachedItem.LastCheck) > c.checkDuration {
				cachedItem.LastCheck = time.Now()
				lastModify, err := c.state(srcFile)
				if err == nil {
					if lastModify.After(cachedItem.LastModify) {
						//文件是否过期
						if src, err := fs.ReadFile(c.fileSystem, srcFile); err == nil {
							cachedItem.Compiled, cachedItem.Error = c.compileFunc(src, placeHolder, srcFile)
						} else {
							cachedItem.Error = err
						}
					}
				} else {
					cachedItem.Error = err
				}
			}
			return cachedItem, nil
		}
		return cachedItem, nil
	})
	item := result.(*CachedItem)
	if item.Error != nil {
		return nil, item.Error
	}

	return item.Compiled.Clone(), item.Error
}
