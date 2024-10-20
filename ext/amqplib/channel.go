package amqplib

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	log "github.com/sirupsen/logrus"
	"github.com/streadway/amqp"
	"lightbox/ext/util"
	"lightbox/sandbox"
	"os"
	"strconv"
	"sync"
)

type ChannelWrapper struct {
	*amqp.Channel
	conn *amqp.Connection
	tengo.ObjectImpl
	app  *sandbox.Applet
	dump bool
	sync.Once
	props map[string]tengo.Object
}

func (w *ChannelWrapper) TypeName() string {
	return "amqp-channel"
}
func (w *ChannelWrapper) String() string {
	return fmt.Sprintf("amqp-channel:%v", w.Channel)
}
func (w *ChannelWrapper) IndexGet(key tengo.Object) (tengo.Object, error) {
	k, _ := tengo.ToString(key)
	w.Do(func() {
		w.props = map[string]tengo.Object{
			"qos": &tengo.UserFunction{
				Value: stdlib.FuncAIIRE(func(i int, i2 int) error {
					return w.Qos(i, 0, false)
				})},
			"exchange_declare": &tengo.UserFunction{Value: w.exchangeDeclare},
			"queue_declare":    &tengo.UserFunction{Value: w.queueDeclare},
			"queue_bind":       &tengo.UserFunction{Value: w.queueBind},
			"dump": &tengo.UserFunction{Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) == 0 {
					return tengo.FromInterface(w.dump)
				} else {
					if b, ok := tengo.ToBool(args[0]); ok {
						w.dump = b
					}
					return nil, nil
				}
			}},
			"consume": &tengo.UserFunction{Value: w.consume},
		}
	})
	if v, ok := w.props[k]; ok {
		return v, nil
	}
	return nil, fmt.Errorf("%s not found", k)
}

func (w *ChannelWrapper) consume(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) < 2 {
		return nil, errors.New("require least 1 args")
	}
	var (
		script    string
		queue     string
		consumer  string
		autoAck   = true
		exclusive = false
		noLocal   = false
		noWait    = false
		argTable  amqp.Table
	)
	for idx, arg := range args {
		switch idx {
		case 0:
			script, _ = tengo.ToString(arg)
		case 1:
			queue, _ = tengo.ToString(arg)
		case 2:
			consumer, _ = tengo.ToString(arg)
		case 3:
			autoAck, _ = tengo.ToBool(arg)
		case 4:
			exclusive, _ = tengo.ToBool(arg)
		case 5:
			noLocal, _ = tengo.ToBool(arg)
		case 6:
			noWait, _ = tengo.ToBool(arg)
		case 7:
			switch m := arg.(type) {
			case *tengo.Map:
				argTable = util.ToMap[interface{}](m.Value)
			case *tengo.ImmutableMap:
				argTable = util.ToMap[interface{}](m.Value)
			}
		}
	}

	msgs, err := w.Channel.Consume(queue, consumer, autoAck, exclusive, noLocal, noWait, argTable)
	if err != nil {
		log.Error("consume error:", err)
		return
	}
	for msg := range msgs {
		if w.dump {
			buf, _ := json.Marshal(msg)
			if werr := os.WriteFile(strconv.FormatUint(msg.DeliveryTag, 10)+".json", buf, 0666); werr != nil {
				log.Error("dump message error:", err)
			}
		}
		wrappedMsg, err := util.WrapObject(&msg, "delivery_message")
		if err != nil {
			log.Error(err)
			continue
		}
		c, err := w.app.GetCompiled(script, util.PlaceHolders{"msg": nil})
		if err != nil {
			log.Error(err)
			continue
		}
		err = c.Set("msg", wrappedMsg)
		if err != nil {
			log.Error(err)
			continue
		}

		err = c.Run()
		if err != nil {
			log.Error(err)
		}
	}
	log.Info("consumer closed")
	return nil, nil
}

//snippet:name=channel.queue_declare;prefix=queue_declare;body=queue_declare(${name},true,false,false,false,{})/*name,durable,autoDelete,exclusive,noWait,args*/;
func (w *ChannelWrapper) queueDeclare(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 1 {
		return nil, errors.New("require least 1 args")
	}
	var (
		name       string
		durable    = true
		autoDelete = false
		exclusive  = false
		noWait     = false
		argTable   amqp.Table
	)
	for idx, arg := range args {
		switch idx {
		case 0:
			name, _ = tengo.ToString(arg)
		case 1:
			durable, _ = tengo.ToBool(arg)
		case 2:
			autoDelete, _ = tengo.ToBool(arg)
		case 3:
			exclusive, _ = tengo.ToBool(arg)
		case 4:
			noWait, _ = tengo.ToBool(arg)
		case 5:
			switch m := arg.(type) {
			case *tengo.Map:
				argTable = util.ToMap[interface{}](m.Value)
			case *tengo.ImmutableMap:
				argTable = util.ToMap[interface{}](m.Value)
			}
		}
	}
	_, err := w.Channel.QueueDeclare(name, durable, autoDelete, exclusive, noWait, argTable)
	if err != nil {
		return util.Error(err), nil
	}
	return nil, nil
}

//snippet:name=channel.queue_bind;prefix=queue_bind;body=queue_bind(${name},${key},${exchange},false,{})/*name,key,exchange,noWait,args*/;
func (w *ChannelWrapper) queueBind(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 1 {
		return nil, errors.New("require least 1 args")
	}
	var (
		name     string
		key      string
		exchange string
		noWait   = false
		argTable amqp.Table
	)
	for idx, arg := range args {
		switch idx {
		case 0:
			name, _ = tengo.ToString(arg)
		case 1:
			key, _ = tengo.ToString(arg)
		case 2:
			exchange, _ = tengo.ToString(arg)
		case 3:
			noWait, _ = tengo.ToBool(arg)
		case 4:
			switch m := arg.(type) {
			case *tengo.Map:
				argTable = util.ToMap[interface{}](m.Value)
			case *tengo.ImmutableMap:
				argTable = util.ToMap[interface{}](m.Value)
			}
		}
	}
	err := w.QueueBind(name, key, exchange, noWait, argTable)
	if err != nil {
		return util.Error(err), nil
	}
	return nil, nil
}

//snippet:name=channel.exchange_declare;prefix=exchange_declare;body=exchange_declare(${name},${kind},true,false,false,false,{})/*name,kind,durable,autoDelete,internal,noWait,args*/;
func (w *ChannelWrapper) exchangeDeclare(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 1 {
		return nil, errors.New("require least 1 args")
	}
	var (
		name       string
		kind       string
		durable    = true
		autoDelete = false
		internal   = false
		noWait     = false
		argTable   amqp.Table
	)
	for idx, arg := range args {
		switch idx {
		case 0:
			name, _ = tengo.ToString(arg)
		case 1:
			kind, _ = tengo.ToString(arg)
		case 2:
			durable, _ = tengo.ToBool(arg)
		case 3:
			autoDelete, _ = tengo.ToBool(arg)
		case 4:
			internal, _ = tengo.ToBool(arg)
		case 5:
			noWait, _ = tengo.ToBool(arg)
		case 6:
			switch m := arg.(type) {
			case *tengo.Map:
				argTable = util.ToMap[interface{}](m.Value)
			case *tengo.ImmutableMap:
				argTable = util.ToMap[interface{}](m.Value)
			}
		}
	}
	err := w.Channel.ExchangeDeclare(name, kind, durable, autoDelete, internal, noWait, argTable)
	if err != nil {
		return util.Error(err), nil
	}
	return nil, nil
}
