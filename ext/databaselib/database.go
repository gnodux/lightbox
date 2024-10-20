package databaselib

import (
	"encoding/base64"
	"errors"
	"fmt"
	_ "github.com/bmizerany/pq"
	"github.com/cookieY/sqlx"
	"github.com/d5/tengo/v2"
	_ "github.com/denisenkom/go-mssqldb"
	_ "github.com/go-sql-driver/mysql"
	"lightbox/env"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

type dbOpt struct {
	Driver string
	DSN    string
}

func Get(app *sandbox.Applet, name string) (*sqlx.DB, error) {
	if c, ok := app.Context.Get(DBCache); ok {
		if cache, ok := c.(*env.GroupCache[*sqlx.DB, dbOpt]); ok {
			return cache.Get(name, dbOpt{})
		}
	}
	return nil, fmt.Errorf("database %s not exists", name)
}

func GetOrOpen(app *sandbox.Applet, name string, driver string, dsn string) (*sqlx.DB, error) {
	if c, ok := app.Context.Get(DBCache); ok {
		if cache, ok := c.(*env.GroupCache[*sqlx.DB, dbOpt]); ok {
			return cache.Get(name, dbOpt{Driver: driver, DSN: dsn})
		}
	}
	return nil, fmt.Errorf("database %s not exists", name)

}
func Open(app *sandbox.Applet, driver string, dsn string) (*sqlx.DB, error) {
	return GetOrOpen(app, base64.StdEncoding.EncodeToString([]byte(dsn)), driver, dsn)
}

func Close(app *sandbox.Applet, name string) error {
	if db, err := Get(app, name); err == nil && db != nil {
		return db.Close()
	}
	return nil
}

//open
//snippet:name=database.open(name);prefix=open;body=open(${1:name});
//snippet:name=database.open;prefix=open;body=open(${1:name},${2:driver},${3:dsn});
//snippet:name=database.open(driver,dsn);prefix=open;body=open(${1:driver},${2:dsn});
func open(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 {
		return nil, errors.New("open database new least 1 argument")
	}
	// 一个参数，获取已经存在的链接
	if len(args) == 1 {
		switch a := args[0].(type) {
		case *tengo.String:
			if db, err := Get(app, a.Value); err == nil {
				return newDatabaseObject(app, db, a.Value), nil
			} else {
				return util.Error(fmt.Errorf("database %s not exists : %s", a.Value, err)), nil
			}
		default:
			return util.Error(fmt.Errorf("unknown open argument:%v", args)), nil
		}
	}
	//两个参数，打开链接(仍然会缓存)
	if len(args) == 2 {
		a, b := args[0].(*tengo.String), args[1].(*tengo.String)
		if a != nil && b != nil {
			if db, err := Open(app, a.Value, b.Value); err == nil {
				return newDatabaseObject(app, db, b.Value), nil
			} else {
				return util.Error(err), nil
			}
		}
	}
	//三个参数
	if len(args) == 3 {
		a, b, c := args[0].(*tengo.String), args[1].(*tengo.String), args[2].(*tengo.String)
		if a != nil && b != nil && c != nil {
			if db, err := GetOrOpen(app, a.Value, b.Value, c.Value); err == nil {
				return newDatabaseObject(app, db, a.Value), nil
			} else {
				return util.Error(err), nil
			}
		}
	}

	return util.Error(fmt.Errorf("unknown argument %v", args)), nil
}
