package cronlib

import (
	"encoding/json"
	"github.com/dgraph-io/badger/v3"
	"github.com/robfig/cron/v3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
	"lightbox/contract"
	"lightbox/sandbox"
	"sync"
)

const (
	defaultName        = "_CRON_SERVICE"
	defaultStorageName = "DEFAULT_CRON_STORAGE"
)

var defaultStorageDir = "cron_storage"
var cronServices = &sync.Map{}
var cronSingle = &singleflight.Group{}

//func GetAllCronService() []*CronService {
//	var services []*CronService
//	cronServices.Range(func(key, value interface{}) bool {
//		switch c := value.(type) {
//		case *CronService:
//			services = append(services, c)
//		}
//		return true
//	})
//	return services
//}

func NewCronServiceWithDB(config *ServiceConfig, db *badger.DB) (*CronService, error) {
	if err := contract.Require(
		contract.NotNil(config, "config is nil"),
		contract.NotNil(config.app, "app is nil"),
	); err != nil {
		return nil, err
	}
	if config.Name == "" {
		config.Name = defaultName
	}
	result, err, _ := cronSingle.Do(config.Name, func() (interface{}, error) {
		instance, ok := cronServices.Load(config.Name)
		if ok {
			return instance, nil
		}
		slog := &logWrapper{
			Entry: config.app.Logger,
		}
		opts := []cron.Option{
			cron.WithLogger(slog), cron.WithSeconds(),
		}
		newCron := &CronService{
			Cron:          cron.New(opts...),
			ServiceConfig: *config,
			db:            db,
			log:           slog,
			mutex:         &sync.Mutex{},
			//objectIndex:   map[string]tengo.Object{},
		}
		if db != nil {
			if err := newCron.Save(); err != nil {
				return nil, err
			}
		}
		//newCron.wrap()
		cronServices.Store(newCron.Name, newCron)
		return newCron, nil
	})
	if err == nil {
		return result.(*CronService), nil
	}
	return nil, err
}

func NewCronService(config *ServiceConfig) (*CronService, error) {
	return NewCronServiceWithDB(config, nil)
}

func BootCronServices(db *badger.DB, applet *sandbox.Applet) error {
	var configs [][]byte
	err := db.View(func(txn *badger.Txn) error {
		option := badger.DefaultIteratorOptions
		option.Prefix = []byte(ServiceConfigPrefix)
		iterator := txn.NewIterator(option)
		defer iterator.Close()
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			itm := iterator.Item()
			err := itm.Value(func(val []byte) error {
				configs = append(configs, val)
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return err
	}
	for _, configData := range configs {
		cfg := &ServiceConfig{}
		err = json.Unmarshal(configData, cfg)
		if err == nil {
			log.Infof("restore cron service %s", cfg.Name)
			cfg.app = applet
			service, err := NewCronService(cfg)
			service.app = applet
			if err != nil {
				return err
			} else {
				service.db = db
				err := service.Restore()
				if err != nil {
					return err
				}
			}
		} else {
			return err
		}
	}
	return nil
}
