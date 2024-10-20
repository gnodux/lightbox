package canallib

import (
	"github.com/go-mysql-org/go-mysql/canal"
	"github.com/go-mysql-org/go-mysql/mysql"
	"github.com/go-mysql-org/go-mysql/replication"
	log "github.com/sirupsen/logrus"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

type ScriptEventHandler struct {
	canal.DummyEventHandler
	Handler HandlerConfig
	*sandbox.Applet
}

const (
	Event      = "event"
	Action     = "action"
	Schema     = "schema"
	Table      = "table"
	NextPos    = "next_pos"
	NextPos2   = "pos"
	QueryEvent = "query_event"
)

var rowEventPlaceHolder = util.PlaceHolders{
	Event:  nil,
	Action: nil,
}
var tableChangeEventPlaceHolder = util.PlaceHolders{
	Schema: nil,
	Table:  nil,
}
var ddlEventPlaceHolder = util.PlaceHolders{
	NextPos:    nil,
	NextPos2:   nil,
	QueryEvent: nil,
	Event:      nil,
}

func (h *ScriptEventHandler) String() string {
	return "script-event-handler"
}

func (h *ScriptEventHandler) OnRow(rowsEvent *canal.RowsEvent) error {
	return util.CallWith(func() error {
		if h.Handler.OnRow == "" {
			return nil
		}
		compiled, err := h.Applet.GetCompiled(h.Handler.OnRow, rowEventPlaceHolder)
		if err != nil {
			return err
		}
		event := &RowsEventWrapper{
			app:       h.Applet,
			RowsEvent: rowsEvent,
		}
		eventObj, err := util.WrapObject(event, "rows-event")
		if err != nil {
			return err
		}
		err = compiled.Set(Event, eventObj)
		if err != nil {
			return err
		}

		_ = compiled.Set(Action, event.Action)
		return compiled.Run()
	}, func(err error) bool {
		log.Errorf("on row event error:%s", err)
		return h.Handler.IgnoreError
	})
}
func (h *ScriptEventHandler) OnRotate(event *replication.RotateEvent) error {
	return util.CallWithIgnoreError(func() error {
		if h.Handler.OnRotate == "" {
			return nil
		}
		compiled, err := h.Applet.GetCompiled(h.Handler.OnRotate, util.PlaceHolders{Event: nil})
		if err != nil {
			log.Errorf("compile rotate event script %s error %s", h.Handler.OnRotate, err)
			return err
		}
		wrappedEvent, err := util.FromInterface(event)
		if err != nil {
			return err
		}
		if err = compiled.Set(Event, wrappedEvent); err != nil {
			return err
		}
		return compiled.Run()
	}, h.Handler.IgnoreError)
}
func (h *ScriptEventHandler) OnTableChanged(schema string, table string) error {
	return util.CallWith(func() error {
		if h.Handler.OnTableChanged == "" {
			return nil
		}
		compiled, err := h.Applet.GetCompiled(h.Handler.OnTableChanged, tableChangeEventPlaceHolder)
		if err != nil {
			log.Errorf("compile table changed script %s error %v", h.Handler.OnTableChanged, err)
			return err
		}
		if err = compiled.Set(Schema, schema); err != nil {
			return err
		}
		if err = compiled.Set(Table, table); err != nil {
			return err
		}
		return compiled.Run()
	}, func(err error) bool {
		log.Errorf("on table changed event error:%s", err)
		return h.Handler.IgnoreError
	})

}
func (h *ScriptEventHandler) OnDDL(nextPos mysql.Position, queryEvent *replication.QueryEvent) error {
	return util.CallWith(func() error {
		if h.Handler.OnDDL == "" {
			return nil
		}
		compiled, err := h.Applet.GetCompiled(h.Handler.OnDDL, ddlEventPlaceHolder)
		if err != nil {
			log.Errorf("compile ddl script %s error %v", h.Handler.OnDDL, err)
			return err
		}
		wrappedPos, err := util.FromInterface(&nextPos)
		if err != nil {
			return err
		}
		wrappedEvent, err := util.FromInterface(queryEvent)
		if err != nil {
			return err
		}
		err = compiled.Set(NextPos, wrappedPos)
		if err != nil {
			return err
		}
		err = compiled.Set(NextPos2, wrappedPos)
		if err != nil {
			return err
		}
		err = compiled.Set(QueryEvent, wrappedEvent)
		if err != nil {
			return err
		}
		err = compiled.Set(Event, wrappedEvent)
		if err != nil {
			return err
		}
		return compiled.Run()
	}, func(err error) bool {
		log.Errorf("on ddl event error: %s", err)
		return h.Handler.IgnoreError
	})
}
