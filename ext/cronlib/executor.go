package cronlib

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"github.com/robfig/cron/v3"
	"github.com/sirupsen/logrus"
	"lightbox/contract"
	"lightbox/sandbox"
	"os/exec"
	"strings"
	"sync"
)

type Executor struct {
	service *CronService
	app     *sandbox.Applet
	mux     *sync.Mutex
	version int64
	JobDetail
	Job cron.Job
}

func SaveExecutor(db *badger.DB, e *Executor) error {
	err := contract.Require(contract.NotNil(db, "db is nil"), contract.NotNil(e, "executor is nil"))
	if err != nil {
		return err
	}
	return db.Update(func(txn *badger.Txn) error {
		var val []byte
		val, err = json.Marshal(e.JobDetail)
		if err != nil {
			return err
		}
		entry := &badger.Entry{
			Key:      []byte(fmt.Sprintf(JobConfigKey, e.service.Name, e.Name)),
			Value:    val,
			UserMeta: byte(1),
		}
		return txn.SetEntry(entry)
	})
}
func (e *Executor) WithWrapper(wrappers ...cron.JobWrapper) {
	e.Job = cron.NewChain(wrappers...).Then(cron.FuncJob(e.run))
}
func (e *Executor) Run() {
	if e.Job != nil {
		e.Job.Run()
	} else {
		e.run()
	}
}
func (e *Executor) run() {
	if e.Script != "" {
		if _, err := e.app.RunFile(e.JobDetail.Script, map[string]interface{}{}); err != nil {
			e.app.Logger.Errorf("run job script %s:%s", e, err)
			return
		}
	}
	for _, entry := range e.Entry {
		if entry.Cmd != "" {
			out, err := exec.Command(entry.Cmd, entry.CmdArgs...).Output()
			if err != nil {
				e.app.Logger.Error("run command ", entry.Cmd, "with output", string(out), "with error", err)
			} else {
				e.app.Logger.WithFields(map[string]interface{}{
					"cmd":  entry.Cmd,
					"args": strings.Join(entry.CmdArgs, " "),
					"out":  string(out),
				}).Infof("command line executed")
			}
		} else if entry.Script != "" {
			if _, err := e.app.RunFile(entry.Script, entry.Args); err != nil {
				e.app.Logger.Error("run script ", entry.Script, err)
			}
		} else {
			logrus.Infof("unknown job type %s (command or script not set)", e)
		}
	}

	logrus.Infof("job %s finished", e)
}

func (e *Executor) String() string {
	return fmt.Sprintf("[%s:%s](%s)(%s)", e.service.Name, e.Name, e.Cron, e.Script)
}
