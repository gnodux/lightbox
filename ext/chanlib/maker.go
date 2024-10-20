package chanlib

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"lightbox/env"
	"lightbox/ext/util"
)

type ChanOption struct {
	Name       string
	Async      bool
	BufferSize int
}

type Channel struct {
	option ChanOption
	ch     chan tengo.Object
}

func (c *Channel) Read() tengo.Object {
	s, ok := <-c.ch
	if !ok {
		return util.Error(fmt.Errorf("channel %s is not availibale", c.option.Name))
	}
	return s
}
func (c *Channel) Write(object tengo.Object) {
	c.ch <- object
}

var chanGenerator = env.NewCache(func(option ChanOption) (*Channel, error) {
	if option.BufferSize == 0 {
		option.BufferSize = 1
	}
	return &Channel{
		option: option,
		ch:     make(chan tengo.Object, option.BufferSize),
	}, nil
})

func NewChannel(opt ChanOption) (*Channel, error) {
	return chanGenerator.Get(opt.Name, opt)
}
