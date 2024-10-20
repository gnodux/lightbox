package loghub

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"net/http"
	"sync"
)

const sandboxName = "sandbox"

type LogSubscriber interface {
	logrus.Hook
	Subscribe(string, chan *logrus.Entry) LogSubscriber
	UnSubscribe(string) LogSubscriber
}

const (
	defaultFilter = `result=(%s)`
)

type SubscriberSetting struct {
	compiled *tengo.Compiled
	Filter   string         `json:"filter"`
	Levels   []logrus.Level `json:"levels"`
	sync.RWMutex
}

func entryToMap(entry *logrus.Entry) map[string]interface{} {
	m := make(map[string]interface{}, len(entry.Data)+4)
	for k, v := range entry.Data {
		m[k] = v
	}
	m["level"] = entry.Level.String()
	m["message"] = entry.Message
	m["time"] = entry.Time
	return m
}

func (l *SubscriberSetting) update() error {
	l.Lock()
	defer l.Unlock()
	src := ""
	if l.Filter != "" {
		src = fmt.Sprintf(defaultFilter, l.Filter)
	}
	if src == "" {
		l.compiled = nil
		return nil
	}
	script := tengo.NewScript([]byte(src))
	err := script.Add("result", false)
	if err != nil {
		return err
	}
	err = script.Add("log", nil)
	if err != nil {
		return err
	}
	l.compiled, err = script.Compile()
	return err
}
func (l *SubscriberSetting) Exec(entry map[string]interface{}) (bool, error) {
	if l.compiled == nil {
		return true, nil
	}
	cc := l.compiled.Clone()
	err := cc.Set("log", entry)
	if err != nil {
		return false, err
	}
	if err = cc.Run(); err != nil {
		return false, err
	}

	v := cc.Get("result")
	if v != nil {
		return v.Bool(), nil
	}
	return false, nil
}

type logHooker struct {
	subscribe map[string]chan<- *logrus.Entry
	levels    []logrus.Level
	sync.Mutex
}

func (f *logHooker) Levels() []logrus.Level {
	return f.levels
}

func (f *logHooker) Fire(entry *logrus.Entry) error {
	go func() {
		defer func() {
			err := recover()
			if err != nil {
				fmt.Println("log hook error:", err)
			}
		}()
		for _, ch := range f.subscribe {
			ch <- entry
		}
	}()
	return nil
}

func (f *logHooker) UnSubscribe(name string) LogSubscriber {
	f.Lock()
	defer f.Unlock()
	delete(f.subscribe, name)
	return f
}
func (f *logHooker) Subscribe(name string, ch chan *logrus.Entry) LogSubscriber {
	f.Lock()
	defer f.Unlock()
	if f.subscribe == nil {
		f.subscribe = map[string]chan<- *logrus.Entry{
			name: ch,
		}
	} else {
		f.subscribe[name] = ch
	}
	return f
}

func NewLogSubscriber() LogSubscriber {
	return NewWithLevels(logrus.StandardLogger(), logrus.TraceLevel, logrus.DebugLevel, logrus.InfoLevel, logrus.WarnLevel, logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel)
}

func NewWithLevels(logger *logrus.Logger, levels ...logrus.Level) LogSubscriber {
	hook := &logHooker{levels: levels}
	logger.AddHook(hook)
	return hook
}

func NewWebsocketSubscribeHandler(subscriber LogSubscriber) http.HandlerFunc {
	return func(w http.ResponseWriter, request *http.Request) {
		if subscriber == nil {
			w.Write([]byte("subscriber not initialize"))
			return
		}
		vars := mux.Vars(request)
		name, ok := vars[sandboxName]

		if !ok {
			return
		}
		filter := &SubscriberSetting{}
		if err := filter.update(); err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		//mutex := &sync.Mutex{}
		upgrader := websocket.Upgrader{
			EnableCompression: true,
			CheckOrigin: func(r *http.Request) bool {
				//do not check origin
				return true
			},
		}
		ws, err := upgrader.Upgrade(w, request, nil)
		if err != nil {
			return
		}
		defer func() {
			ws.Close()
		}()
		ch := make(chan *logrus.Entry, 100)
		subscriber.Subscribe(name, ch)
		defer func() {
			subscriber.UnSubscribe(name)
		}()
		go func() {
			for msg := range ch {
				if n, ok := msg.Data[sandboxName]; !(ok && n == name) {
					continue
				}
				if filter.Levels != nil && len(filter.Levels) > 0 {
					match := false
					for _, l := range filter.Levels {
						if l == msg.Level {
							match = true
							break
						}
					}
					if !match {
						continue
					}
				}
				m := entryToMap(msg)
				err = func() error {
					//mutex.Lock()
					//defer mutex.Unlock()
					if match, merr := filter.Exec(m); match {
						return ws.WriteJSON(m)
					} else if merr != nil {
						_ = ws.WriteJSON(map[string]interface{}{
							"code":  500,
							"error": merr.Error(),
						})
					}
					return nil
				}()
				if err != nil {
					logrus.Error(err)
				}
			}
		}()
		for {
			err1 := func() error {
				//mutex.Lock()
				//defer mutex.Unlock()
				err = ws.ReadJSON(filter)
				if err != nil {
					return err
				}
				if err = filter.update(); err != nil {
					return ws.WriteJSON(map[string]interface{}{
						"code":  "500",
						"error": err.Error(),
					})
				} else {
					return ws.WriteJSON(map[string]interface{}{
						"code":    200,
						"message": "subscriber setting updated",
					})
				}
			}()
			if err1 != nil {
				logrus.Error("read json command", err)
				break
			}
		}
	}
}
