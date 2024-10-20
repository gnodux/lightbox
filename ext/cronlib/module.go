package cronlib

import (
	"errors"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
	"lightbox/ext/util"
	"lightbox/kvstore"
	"lightbox/sandbox"
	"sync"
)

var bootLock = &sync.Mutex{}
var module = map[string]tengo.Object{
	//snippet:name=cron.new;prefix=new;body=new({db_name:${db_name}});
	//snippet:name=cron.set_data_dir;prefix=set_data_dir;body=set_data_dir(${1:dir});description=设置CRON服务器的默认存储目录;
	"set_data_dir": &tengo.UserFunction{Value: stdlib.FuncASRE(func(s string) error {
		defaultStorageDir = s
		return nil
	})},
	//snippet:name=cron.get;prefix=get;body=get(${1:servicename});
	"get": &tengo.UserFunction{
		Name: "get",
		Value: func(args ...tengo.Object) (tengo.Object, error) {
			var (
				name string
				ok   bool
			)
			if len(args) > 0 {
				name, ok = tengo.ToString(args[0])
				if !ok {
					return nil, errors.New("invalidate cron name")
				}
			}
			if instance, ok := cronServices.Load(name); ok && instance != nil {
				return util.NewProxy(instance.(*CronService)).WithConstructor(CronServiceConstructor), nil
			}
			return tengo.UndefinedValue, nil
		},
	},
}
var appModule = map[string]sandbox.UserFunction{
	//snippet:name=cron.boot;prefix=boot;body=boot();
	//snippet:name=cron.boot;prefix=boot;body=boot($1);
	"boot": boot,
	"new":  newService,
	"all":  listAllService,
}

func newService(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	var (
		dbName string
		dbDir  string
		db     *badger.DB
		err    error
	)
	conf := &ServiceConfig{
		app: app,
	}
	if len(args) == 1 {
		err = util.StructFromObject(args[0], conf)
	}
	if dbName == "" {
		dbName = defaultStorageName
	}
	if dbDir == "" {
		dbDir = defaultStorageDir
	}

	db, err = kvstore.Open(dbName, kvstore.DefaultOptions(dbDir))
	if err != nil {
		return nil, err
	}
	svc, err := NewCronServiceWithDB(conf, db)
	return util.NewProxy(svc).WithConstructor(CronServiceConstructor), err
}

func boot(app *sandbox.Applet, args ...tengo.Object) (ret tengo.Object, err error) {
	bootLock.Lock()
	defer bootLock.Unlock()
	log.Infof("bootstrap [%s]  cron service", app.Name)
	name := app.Name + defaultName
	if instance, ok := cronServices.Load(name); ok && instance != nil {
		ret = util.NewProxy(instance.(*CronService)).WithConstructor(CronServiceConstructor)
		log.Infof("default service already initialized")
		return
	}
	storage := defaultStorageDir
	if len(args) > 0 {
		storage, _ = tengo.ToString(args[0])
	}
	var db *badger.DB
	db, err = kvstore.Open(defaultStorageName, kvstore.DefaultOptions(storage))
	if err != nil {
		return util.Error(err), nil
	}
	err = BootCronServices(db, app)
	if err != nil {
		return
	}
	//再次从恢复的数据库中寻找
	if instance, ok := cronServices.Load(name); ok && instance != nil {
		ret = util.NewProxy(instance.(*CronService)).WithConstructor(CronServiceConstructor)
		return
	} else {
		//default service not found，initialize a default cron service
		cfg := &ServiceConfig{
			Name:        name,
			Description: "default cron service",
			app:         app,
		}
		var service *CronService
		service, err = NewCronServiceWithDB(cfg, db)
		if err != nil {
			return util.Error(err), nil
		} else {
			return util.NewProxy(service).WithConstructor(CronServiceConstructor), nil
		}
	}
}

func listAllService(applet *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	var lst = make(map[string]tengo.Object)
	cronServices.Range(func(key, value any) bool {
		if k, ok := key.(string); ok {
			if v, ok := value.(*CronService); ok {
				lst[k] = util.NewProxy(v).WithConstructor(CronServiceConstructor)
			}
		}
		return true
	})
	return &tengo.ImmutableMap{Value: lst}, nil
}

var Entry = sandbox.NewRegistry("cron", module, appModule)
