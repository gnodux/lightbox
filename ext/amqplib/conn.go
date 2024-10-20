package amqplib

import (
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/streadway/amqp"
	"lightbox/ext/util"
	"lightbox/sandbox"
	"sync"
)

type ConnectionWrapper struct {
	tengo.ObjectImpl
	sync.Once
	props map[string]tengo.Object
	app   *sandbox.Applet
	*amqp.Connection
	url string
}

func (w *ConnectionWrapper) TypeName() string {
	return "amqp-connection"
}
func (w *ConnectionWrapper) String() string {
	return "amqp-connection:" + w.Connection.LocalAddr().String()
}

func (w *ConnectionWrapper) IndexGet(key tengo.Object) (tengo.Object, error) {
	w.Do(func() {
		w.props = map[string]tengo.Object{
			"channel": &tengo.UserFunction{Value: func(args ...tengo.Object) (tengo.Object, error) {
				ch, err := w.Channel()
				if err != nil {
					return nil, err
				}
				return &ChannelWrapper{Channel: ch, app: w.app, conn: w.Connection}, nil
			}},
			"close":     &tengo.UserFunction{Value: stdlib.FuncARE(w.Close)},
			"is_closed": &tengo.UserFunction{Value: stdlib.FuncARB(w.IsClosed)},
			"connect":   &tengo.UserFunction{Value: w.connect},
			"stat": &tengo.UserFunction{Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				stat := w.Connection.ConnectionState()
				return util.WrapObject(stat, "connection-stat")
			}},
			"wait_close": &tengo.UserFunction{Value: w.waitClose},
			"major":      &tengo.Int{Value: int64(w.Major)},
			"minor":      &tengo.Int{Value: int64(w.Minor)},
		}

	})
	k, _ := tengo.ToString(key)
	if v, ok := w.props[k]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("%s not exists", k)
}

func (w *ConnectionWrapper) connect(args ...tengo.Object) (ret tengo.Object, err error) {
	if w.Connection != nil && !w.IsClosed() {
		return
	}
	w.Connection, err = amqp.Dial(w.url)
	if err != nil {
		return util.Error(err), nil
	}
	return nil, nil
}

func (w *ConnectionWrapper) waitClose(args ...tengo.Object) (ret tengo.Object, err error) {
	if w.Connection != nil {
		c := make(chan *amqp.Error)
		closeErr := <-w.Connection.NotifyClose(c)
		return util.Error(closeErr), nil
	}
	return util.Error(errors.New("connection is nil")), nil
}

func Dial(url string, app *sandbox.Applet) (*ConnectionWrapper, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, err
	}
	wrapper := &ConnectionWrapper{Connection: conn, url: url, app: app}
	return wrapper, nil
}
