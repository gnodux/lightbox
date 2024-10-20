package cronlib

import (
	"encoding/json"
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/dgraph-io/badger/v3"
	"github.com/robfig/cron/v3"
	"lightbox/ext/util"
	"lightbox/sandbox"
	"sync"
	"time"
)

const (
	ServiceConfigPrefix = "service:config:"
	ServiceConfigKey    = "service:config:%s"
	JobConfigPrefix     = "job:%s:"
	JobConfigKey        = "job:%s:%s"
)

type (
	ServiceConfig struct {
		//Name Cron service name
		Name string `json:"name"`
		//Description Cron ServiceConfig description
		Description string `json:"description"`
		//Location Cron service location
		Location time.Location `json:"location"`
		//app owner application
		app *sandbox.Applet
	}

	ExecEntry struct {
		Cmd     string                 `json:"cmd"`
		Script  string                 `json:"script"`
		CmdArgs []string               `json:"cmdArgs"`
		Args    map[string]interface{} `json:"args"`
	}
	ExecEntries []ExecEntry
	JobDetail   struct {
		EntryId int `json:"entry_id"`
		//Name 名称
		Name string `json:"name"`
		//Description 描述
		Description string `json:"desc"`
		//Cron cron表达式
		Cron string `json:"cron"`
		//Script 执行脚本(简单的执行脚本)
		Script string `json:"script"`
		//Entry 入口点
		Entry []ExecEntry `json:"entry"`
		//Policy 执行策略(0:串行,1:并行)
		Policy []JobPolicy `json:"policy"`
		//Next 下次执行时间
		Next time.Time
		//上次执行时间
		Prev time.Time
	}
	CronService struct {
		*cron.Cron
		db *badger.DB
		ServiceConfig
		log   cron.Logger
		mutex *sync.Mutex
	}
)

func (e ExecEntry) ToTengoObject() tengo.Object {
	obj := map[string]tengo.Object{}
	obj["cmd"] = &tengo.String{Value: e.Cmd}
	obj["script"] = &tengo.String{Value: e.Script}
	if e.CmdArgs == nil {
		obj["cmdArgs"] = tengo.UndefinedValue
	} else {
		obj["cmdArgs"], _ = util.ToImmutableArray(e.CmdArgs...)
	}
	obj["args"], _ = tengo.FromInterface(e.Args)
	return &tengo.ImmutableMap{Value: obj}
}

func (entries ExecEntries) ToTengoSlice() *tengo.Array {
	var objs []tengo.Object
	for _, enr := range entries {
		objs = append(objs, enr.ToTengoObject())
	}
	return &tengo.Array{Value: objs}
}

//Restore restore job from database
func (s *CronService) Restore() error {
	if s.db == nil {
		s.app.Logger.Infof("%s store db is nil,store exit", s.Name)
		return nil
	}
	var jobData [][]byte
	err := s.db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.Prefix = []byte(fmt.Sprintf(JobConfigPrefix, s.Name))
		iter := txn.NewIterator(opt)
		defer iter.Close()
		for iter.Rewind(); iter.Valid(); iter.Next() {
			err := iter.Item().Value(func(val []byte) error {
				jobData = append(jobData, val)
				return nil
			})
			if err != nil {
				s.app.Logger.Error("restore job data from database error", err)
				return err
			}
		}
		return nil
	})
	if err != nil {
		s.app.Logger.Error("restore service error", s.Name, err)
	}
	for _, data := range jobData {
		job := &JobDetail{}
		err = json.Unmarshal(data, job)
		if err != nil {
			s.app.Logger.Error("unmarshal job  error", err)
			return err
		}
		err = s.DoSchedule(*job, false)
		if err != nil {
			s.app.Logger.Error("restore job  error", err)
			return err
		} else {
			s.app.Logger.Infof("restore job:%+v", job)
		}
	}
	return nil
}
func (s *CronService) DoSchedule(job JobDetail, store bool) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	executor := &Executor{
		service:   s,
		JobDetail: job,
		app:       s.app,
	}
	removeIds := s.RemoveByName(job.Name)
	if len(removeIds) > 0 {
		s.app.Logger.Info("remove jobs :", removeIds)
	}
	var (
		entryId cron.EntryID
		err     error
	)
	var wrappers = []cron.JobWrapper{cron.Recover(s.log)}
	for _, p := range job.Policy {
		switch p {
		case PolicySkipIfRunning:
			wrappers = append(wrappers, cron.SkipIfStillRunning(s.log))
		case PolicyDelayIfRunning:
			wrappers = append(wrappers, cron.DelayIfStillRunning(s.log))
		}
	}
	executor.WithWrapper(wrappers...)
	entryId, err = s.AddJob(job.Cron, executor)
	executor.JobDetail.EntryId = int(entryId)
	if err == nil && s.db != nil && store {
		return SaveExecutor(s.db, executor)
	}
	return err
}
func (s *CronService) List() []JobDetail {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	var jobs []JobDetail
	for _, entry := range s.Entries() {
		if executor, ok := entry.Job.(*Executor); ok {
			cfg := executor.JobDetail
			cfg.Prev = entry.Prev
			cfg.Next = entry.Next
			cfg.EntryId = int(entry.ID)
			jobs = append(jobs, cfg)
		}
	}
	return jobs
}

func (s *CronService) Find(filter func(executor *Executor) bool) []cron.EntryID {
	var entryIds []cron.EntryID
	for _, entry := range s.Entries() {
		if executor, ok := entry.Job.(*Executor); ok {
			if filter(executor) {
				entryIds = append(entryIds, entry.ID)
			}
		}
	}
	return entryIds
}

func (s *CronService) RemoveBy(filter func(*Executor) bool) []cron.EntryID {
	ids := s.Find(filter)
	for _, id := range ids {
		s.Remove(id)
	}
	return ids
}
func (s *CronService) RemoveByName(jobNames ...string) []cron.EntryID {
	return s.RemoveBy(func(executor *Executor) bool {
		for _, name := range jobNames {
			if executor.Name == name {
				return true
			}
		}
		return false
	})
}

func (s *CronService) Save() error {
	if s.db != nil {
		v, e := json.Marshal(&s.ServiceConfig)
		if e != nil {
			return e
		}
		return s.db.Update(func(txn *badger.Txn) error {

			entry := &badger.Entry{
				Key:      []byte(fmt.Sprintf(ServiceConfigKey, s.Name)),
				Value:    v,
				UserMeta: byte(1),
			}
			return txn.SetEntry(entry)
		})
	}
	return nil
}
func (s *CronService) RemoveFromDB() error {
	if s.db == nil {
		return nil
	}
	return s.db.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(fmt.Sprintf(ServiceConfigKey, s.Name)))
	})
}

func CronServiceConstructor(proxy *util.Proxy[*CronService]) {
	s := proxy.Value
	proxy.Props = map[string]tengo.Object{
		//snippet:
		"start": &tengo.UserFunction{
			Name: "start",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				s.Start()
				return nil, nil
			},
		},
		"stop": &tengo.UserFunction{
			Name: "stop",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				s.Stop()
				return nil, nil
			},
		},
		"run": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				s.Run()
				return nil, nil
			},
		},
		"remove": &tengo.UserFunction{
			Name: "remove",
			Value: util.FuncASsRE(func(names []string) error {
				if names != nil {
					for _, name := range names {
						s.RemoveByName(name)
					}
				}
				return nil
			}),
		},
		"list": &tengo.UserFunction{
			Name: "list",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				configs := s.List()
				var result []tengo.Object
				for _, conf := range configs {
					result = append(result, &tengo.ImmutableMap{Value: map[string]tengo.Object{
						"name":        &tengo.String{Value: conf.Name},
						"cron":        &tengo.String{Value: conf.Cron},
						"script":      &tengo.String{Value: conf.Script},
						"policy":      JobPolicies(conf.Policy).ToTengoSlice(),
						"entry":       ExecEntries(conf.Entry).ToTengoSlice(),
						"description": &tengo.String{Value: conf.Description},
						"prev":        &tengo.Time{Value: conf.Prev},
						"next":        &tengo.Time{Value: conf.Next},
						"entry_id":    &tengo.Int{Value: int64(conf.EntryId)},
					}})
				}
				return &tengo.Array{Value: result}, nil
			},
		},
		"schedule": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				conf := &JobDetail{}
				if len(args) != 1 {
					return nil, tengo.ErrWrongNumArguments
				}
				if err := util.UnmashalObject(args[0], conf); err != nil {
					return nil, err
				}

				err := s.DoSchedule(*conf, true)
				return proxy, err
			},
		},
	}

}
