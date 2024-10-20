package kvstore

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	log "github.com/sirupsen/logrus"
	"golang.org/x/sync/singleflight"
	"strings"
	"sync"
	"sync/atomic"
)

var databases = &sync.Map{}

func Shutdown() {
	databases.Range(func(key, value any) bool {
		if value != nil {
			if db, ok := value.(*badger.DB); ok && db != nil {
				log.Info("shutdown badger database :", key)
				if !db.IsClosed() {
					db.Close()
				}
			}
		}
		return true
	})
}

func init() {
	SetSmallSize()
}

const (
	defaultMaxValueThreshold int64 = 1 << 20
)

var (
	sg = &singleflight.Group{}
	// 10M*2的ValueLog
	defaultValueLogSize int64
	defaultMemTableSize int64
	maxValueThreshold   int64
)

func SetSizeOption(valLog int64, memTable int64, valThreshold int64) {
	atomic.StoreInt64(&defaultValueLogSize, valLog)
	atomic.StoreInt64(&defaultMemTableSize, memTable)
	atomic.StoreInt64(&maxValueThreshold, valThreshold)

}
func SetMiniSize() {
	SetSizeOption(1<<20, 64<<10, 1<<10)
}
func SetSmallSize() {
	SetSizeOption(1<<25, 64<<18, 1<<18)
}

func SetLargeSize() {
	SetSizeOption(1<<30-1, 64<<20, 1<<20)
}

func DefaultOptions(path string) badger.Options {
	opt := badger.DefaultOptions(path)
	return opt.WithValueLogFileSize(defaultValueLogSize).WithMemTableSize(defaultMemTableSize).WithValueThreshold(maxValueThreshold)
}
func OpenWith(name, dir string) (*badger.DB, error) {
	return Open(name, DefaultOptions(dir))
}
func Open(name string, options badger.Options) (*badger.DB, error) {
	if strings.HasPrefix(name, ":") {
		//如果前缀带有`:`,则不落盘
		options.Dir = ""
	}
	if options.Dir == "" {
		options.InMemory = true
	}
	log.Debugf("open badger %s:%s", name, options.Dir)
	d, e, _ := sg.Do(name, func() (interface{}, error) {
		if existsDb, ok := databases.Load(name); !ok {
			log.Debugf("open a new badger %s:%s", name, options.Dir)
			db, err := badger.Open(options)
			if err == nil {
				databases.Store(name, db)
			}
			return db, err
		} else {
			log.Debugf("badger %s:%s already opened", name, options.Dir)
			return existsDb, nil
		}
	})
	if e == nil {
		db := d.(*badger.DB)
		return db, nil
	}
	return nil, e
}
func Get(name string) (*badger.DB, error) {
	d, e, _ := sg.Do(name, func() (interface{}, error) {
		if existsDb, ok := databases.Load(name); !ok {
			return nil, fmt.Errorf("%s  not exists", name)
		} else {
			return existsDb, nil
		}
	})
	if e == nil {
		return d.(*badger.DB), nil
	}
	return nil, e
}
func Close(name string) error {
	if d, ok := databases.Load(name); ok && d != nil {
		databases.Delete(name)
		err := d.(*badger.DB).Close()
		if err != nil {
			return err
		}

	}
	return nil
}

func OpenDefault() (*badger.DB, error) {
	return OpenWith("DEFAULT", "data")
}
func CloseDefault() error {
	return Close("DEFAULT")
}
