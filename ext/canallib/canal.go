package canallib

import (
	"github.com/d5/tengo/v2"
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

type CanalWrapper struct {
	*canal.Canal
}

func (c *CanalWrapper) RunFromLast() error {
	pos, err := c.GetMasterPos()
	if err != nil {
		return err
	}
	return c.RunFrom(pos)
}
func (c *CanalWrapper) RunFromPos(pos int64, name string) error {
	return c.RunFrom(mysql.Position{Pos: uint32(pos), Name: name})
}
func newConfig(args ...tengo.Object) (tengo.Object, error) {
	return util.WrapObject(&CanalConfig{Config: canal.NewDefaultConfig()}, "config")
}

func newCanalWithApp(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	cfg := &CanalConfig{
		Config: canal.NewDefaultConfig(),
	}
	err := util.StructFromArgs(args, cfg)
	if err != nil {
		return nil, err
	}
	c, err := canal.NewCanal(cfg.Config)
	if err != nil {
		return nil, err
	}
	c.SetEventHandler(&ScriptEventHandler{Handler: cfg.Handler, Applet: app})
	if err != nil {
		return nil, err
	}
	wrapper := &CanalWrapper{
		Canal: c,
	}

	return util.FromInterface(wrapper)
}
