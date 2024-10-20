package canallib

import (
	"database/sql"
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/go-mysql-org/go-mysql/canal"
	log "github.com/sirupsen/logrus"
	"lightbox/ext/databaselib"
	"lightbox/ext/util"
	"lightbox/sandbox"
	"reflect"
	"strings"
)

type RowsEventWrapper struct {
	*canal.RowsEvent
	app *sandbox.Applet
}

func (e *RowsEventWrapper) Get(row int, column int) interface{} {
	return e.Rows[row][column]
}
func (e *RowsEventWrapper) All(args ...tengo.Object) (tengo.Object, error) {
	var rows []*RowEvent
	switch e.Action {
	case "update":
		for idx := 0; idx < len(e.Rows)/2; idx++ {
			rows = append(rows, NewRowEvent(e.app, e.RowsEvent, e.Rows[idx*2], e.Rows[idx*2+1]))
		}
	case "insert", "delete":
		for _, r := range e.Rows {
			rows = append(rows, NewRowEvent(e.app, e.RowsEvent, nil, r))
		}
	}
	return util.FromInterface(rows)
}

type Row struct {
	tengo.ObjectImpl
	rowsEvent *canal.RowsEvent
	raw       []interface{}
}

func (r *Row) IndexGet(key tengo.Object) (tengo.Object, error) {
	switch k := key.(type) {
	case *tengo.Int:
		if r.raw != nil {
			if int(k.Value) <= len(r.raw) {
				return tengo.FromInterface(r.raw[int(k.Value)])
			} else {
				return nil, errors.New("index out of range")
			}
		}
		return nil, nil
	case *tengo.String:
		if r.rowsEvent == nil || r.raw == nil {
			return nil, nil
		}
		for idx, col := range r.rowsEvent.Table.Columns {
			if col.Name == k.Value {
				return tengo.FromInterface(r.raw[idx])
			}
		}
		return nil, nil

	}
	return nil, errors.New("invalidate key type")
}
func (r *Row) String() string {
	return fmt.Sprintf("%v", r.raw)
}

type RowEvent struct {
	Before *Row
	After  *Row
	Event  *canal.RowsEvent
	app    *sandbox.Applet
}

func (re *RowEvent) rowDiff() map[string]interface{} {
	result := map[string]interface{}{}
	switch re.Event.Action {
	case "delete", "insert":
		for idx, o := range re.After.raw {
			column := re.Event.Table.Columns[idx]
			result[column.Name] = o
		}
	case "update":
		for idx, v := range re.After.raw {
			beforeV := re.Before.raw[idx]
			if !reflect.DeepEqual(v, beforeV) {
				column := re.Event.Table.Columns[idx]
				result[column.Name] = v
			}
		}
	}
	return result
}
func (re *RowEvent) Diff(args ...tengo.Object) (tengo.Object, error) {
	result, err := util.FromInterface(re.rowDiff())
	return result, err
}

func (re *RowEvent) SyncToTable(databaseName string, schema string, table string) (int64, error) {
	switch re.Event.Action {
	case canal.DeleteAction:
		return re.doDelete(databaseName, schema, table)
	case canal.UpdateAction:
		return re.doUpdate(databaseName, schema, table)
	case canal.InsertAction:
		return re.doInsert(databaseName, schema, table)
	}
	return 0, nil
}
func (re *RowEvent) SyncToDb(databaseName string, schema string) (int64, error) {
	return re.SyncToTable(databaseName, schema, re.Event.Table.Name)
}

func (re *RowEvent) SyncTo(databaseName string) (int64, error) {
	return re.SyncToTable(databaseName, re.Event.Table.Schema, re.Event.Table.Name)
}

const (
	DeleteSQL = "DELETE FROM %s.%s WHERE %s"
	UpdateSQL = "UPDATE %s.%s SET %s WHERE %s"
	InsertSQL = "INSERT INTO %s.%s (%s) VALUES (%s)"
)

func (re *RowEvent) doDelete(name, schema, table string) (int64, error) {
	db, err := databaselib.Get(re.app, name)
	if err != nil {
		return -1, fmt.Errorf("database %s not initialized", name)
	}
	if len(re.Event.Table.PKColumns) == 0 {
		return -1, fmt.Errorf("%s no PK Columns", re.Event.Table)
	}
	suffix := ""
	where := ""
	for _, colIdx := range re.Event.Table.PKColumns {
		column := re.Event.Table.Columns[colIdx]
		where = where + suffix + column.Name + "=?"
		suffix = " and "
	}

	args, err := re.Event.Table.GetPKValues(re.After.raw)
	if err != nil {
		return -1, err
	}
	query := fmt.Sprintf(DeleteSQL, schema, table, where)
	var result sql.Result
	if result, err = db.Exec(query, args...); err != nil {
		return -1, err
	}
	return result.RowsAffected()
}

func (re *RowEvent) doUpdate(name, schema, table string) (int64, error) {
	db, err := databaselib.Get(re.app, name)
	if err != nil {
		return -1, fmt.Errorf("database %s not initialized", name)
	}
	suffix := ""
	where := ""
	for _, colIdx := range re.Event.Table.PKColumns {
		column := re.Event.Table.Columns[colIdx]
		where = where + suffix + column.Name + "=?"
		suffix = " and "
	}
	args, err := re.Event.Table.GetPKValues(re.After.raw)
	var columnArgs []interface{}
	diff := re.rowDiff()
	suffix = ""
	set := ""
	for k, v := range diff {
		set = set + suffix + "`" + k + "`=?"
		columnArgs = append(columnArgs, v)
		suffix = " , "
	}
	columnArgs = append(columnArgs, args...)
	query := fmt.Sprintf(UpdateSQL, schema, table, set, where)
	var result sql.Result
	if result, err = db.Exec(query, columnArgs...); err != nil {
		return -1, err
	}
	return result.RowsAffected()
}

func (re *RowEvent) doInsert(name, schema, table string) (int64, error) {
	db, err := databaselib.Get(re.app, name)
	if err != nil {
		return -1, fmt.Errorf("database %s not initialized", name)
	}
	var columns []string
	var params []string
	for _, col := range re.Event.Table.Columns {
		columns = append(columns, "`"+col.Name+"`")
		params = append(params, "?")
	}
	query := fmt.Sprintf(InsertSQL, schema, table, strings.Join(columns, ","), strings.Join(params, ","))
	result, err := db.Exec(query, re.After.raw...)
	if err != nil {
		return -1, err
	}
	return result.LastInsertId()
}

func affectRows(result sql.Result) int64 {
	affected, rowErr := result.RowsAffected()
	if rowErr != nil {
		log.Infof("get result affected rows error:", rowErr)
		affected = -1
	}
	return affected
}

func NewRowEvent(app *sandbox.Applet, e *canal.RowsEvent, before []interface{}, after []interface{}) *RowEvent {
	return &RowEvent{
		app:    app,
		Event:  e,
		Before: &Row{raw: before, rowsEvent: e},
		After:  &Row{raw: after, rowsEvent: e},
	}
}
