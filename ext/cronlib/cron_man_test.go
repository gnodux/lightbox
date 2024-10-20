package cronlib

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/sirupsen/logrus"
	"lightbox/kvstore"
	"lightbox/sandbox"
	"os"
	"os/exec"
	"testing"
	"time"
)

var app *sandbox.Applet

func TestMain(t *testing.M) {
	app, _ = sandbox.NewWithDir("default", ".")
	m := stdlib.GetModuleMap(stdlib.AllModuleNames()...)
	m.AddBuiltinModule(Entry.Name(), Entry.GetModule(app, Entry.AllNames()...))
	app.WithModule(m)
	t.Run()
}

func TestNewServiceFrom(t *testing.T) {
	script := `
import(fmt,cron)
svc:=cron.boot()
svc.schedule({
	name:"job1",
	description:"job 描述",
	cron:"@every 10s",
	policy:["skipIfRunning"],
	entry: [{cmd:"ping",cmdArgs:["www.baidu.com","-c","50"]},{"script":"testdata/test1.tengo","args":{arg1:1,arg2:2}}]
})

fmt.println(svc.list())
svc.run()
`
	timeout, cancelFn := context.WithTimeout(context.Background(), 3*time.Minute)
	go func() {
		<-timeout.Done()
		cancelFn()
	}()
	c, err := app.RunContext(timeout, []byte(script), nil, "")
	if err != nil {
		fmt.Println(err)
	} else {

		fmt.Println(c.GetAll())
	}

}

func TestNewCronService(t *testing.T) {
	db, err := kvstore.Open("cron", kvstore.DefaultOptions("testdata/cron_storage"))
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	svc, err := NewCronServiceWithDB(&ServiceConfig{
		Name:        "test1",
		Description: "测试的",
		app:         app,
	}, db)

	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	svc.DoSchedule(JobDetail{
		Name:        "job1",
		Description: "测试的job",
		Cron:        "@every 10s",
		Script:      "testdata/test1.tengo",
	}, true)
	svc.DoSchedule(JobDetail{
		Name: "job2",
		Cron: "@every 10s",
		Entry: []ExecEntry{
			{
				Cmd:     "ping",
				CmdArgs: []string{"www.baidu.com", "-c", "30"},
			},
		},
		Policy: []JobPolicy{
			PolicySkipIfRunning,
		},
	}, true)
	svc.Run()
}

func TestLogrusLevel(t *testing.T) {
	output := `{"level":"debug","policy":["skipIfRunning","delayIfRunning"]}`
	type TestEnum struct {
		Level  logrus.Level `json:"level"`
		Policy []JobPolicy  `json:"policy"`
	}
	ad := &TestEnum{
		Level:  logrus.DebugLevel,
		Policy: []JobPolicy{PolicySkipIfRunning, PolicyDelayIfRunning},
	}
	encoder := json.NewEncoder(os.Stdout)
	err := encoder.Encode(ad)
	if err != nil {
		t.Fatal(err)
	}
	decoder := json.NewDecoder(bytes.NewBuffer([]byte(output)))
	newAd := &TestEnum{}
	err = decoder.Decode(newAd)
	if err != nil {
		t.Fatal(err)
	}
	fmt.Printf("%+v", newAd)
}

func TestCommand(t *testing.T) {
	cmd := exec.Command("ls", "-l")
	err := cmd.Run()
	if err != nil {
		t.Fatal(err)
	}
}
