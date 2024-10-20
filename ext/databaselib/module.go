package databaselib

import (
	"errors"
	"github.com/cookieY/sqlx"
	"github.com/d5/tengo/v2"
	log "github.com/sirupsen/logrus"
	"lightbox/env"
	"lightbox/sandbox"
	"regexp"
	"strings"
)

const DBCache = "__database__cache__"

func AppModule(app *sandbox.Applet) map[string]tengo.Object {
	app.Context.Set(DBCache, env.NewCache(func(option dbOpt) (*sqlx.DB, error) {
		if option.Driver == "" || option.DSN == "" {
			return nil, errors.New("driver or dsn is empty")
		}
		return sqlx.Connect(option.Driver, option.DSN)
	}))
	return map[string]tengo.Object{
		"open": &tengo.UserFunction{Value: sandbox.NewUserFunction(app, open)},
	}
}

//sql 方言正则表达式
var sqlRe = regexp.MustCompile(`(?s)db\[(?P<name>\w+)].(?P<func>\w+)=>(?P<sql>.*?);`)

//
//var (
//	dbStart    = "database.open(\""
//	dbEnd      = "\")."
//	dbStartSQL = "(\""
//	dbEndSQL   = "\""
//	paramStart = ","
//)

var sqlVarRe = regexp.MustCompile(`(?m)((#|\$)\w+)|((#|\$)\{.*?\})`)

func sqlDialect(src []byte) ([]byte, error) {
	newSrc := sqlRe.ReplaceAllStringFunc(string(src), func(stmt string) string {
		matches := sqlRe.FindStringSubmatch(stmt)
		//0:all,1:db name;2:func;3:sql
		var params []string
		sql := strings.ReplaceAll(strings.TrimSpace(matches[3]), "\n", "\\n")
		sql = sqlVarRe.ReplaceAllStringFunc(sql, func(s string) string {
			switch s[0] {
			case '#':
				params = append(params, strings.Trim(s, "#{} "))
				return "?"
			case '$':
				return `"+string(` + strings.Trim(s, "${} ") + `)+"`
			}
			return s
		})
		//if sql[len(sql)-1] == '"' {
		//	sql = sql[0:(len(sql) - 2)]
		//}
		var ret string
		if len(params) > 0 {
			ret = strings.Join([]string{`sys.must(database.open("`, matches[1], `")).`, matches[2], `("`, sql, `",`, strings.Join(params, ","), ")"}, "")
		} else {
			ret = strings.Join([]string{`sys.must(database.open("`, matches[1], `")).`, matches[2], `("`, sql, `")`}, "")
		}
		log.Tracef("transpile SQL `[%s] to [%s]", src, ret)
		return ret
	})

	return ([]byte)(newSrc), nil
}

var Entry = sandbox.NewRegistry("database",
	nil,
	map[string]sandbox.UserFunction{
		"open": open,
	}).WithTranspiler(sqlDialect).WithHook(sandbox.NewHook(sandbox.SigInitialized, func(app *sandbox.Applet) error {
	c := env.NewCache(func(option dbOpt) (*sqlx.DB, error) {
		if option.Driver == "" || option.DSN == "" {
			return nil, errors.New("driver or dsn is empty")
		}
		return sqlx.Connect(option.Driver, option.DSN)
	})
	app.Context.Set(DBCache, c)
	app.WithHook(sandbox.NewHook(sandbox.SigStop, func(applet *sandbox.Applet) error {
		c.EvictWith(func(name string, db *sqlx.DB) {
			log.WithField("sandbox", app.Name).Info("auto close database ", name)
			if db != nil {
				if err := db.Close(); err != nil {
					log.WithField("sandbox", app.Name).Error("close database error:", err)
				}

			}
		})
		app.Context.Delete(DBCache)
		return nil
	}))
	return nil
}))

//
//.WithAppInitializer(
//	func(app *sandbox.Applet) {
//		c := env.NewCache(func(option dbOpt) (*sqlx.DB, error) {
//			if option.Driver == "" || option.DSN == "" {
//				return nil, errors.New("driver or dsn is empty")
//			}
//			return sqlx.Connect(option.Driver, option.DSN)
//		})
//		app.Context.Set(DBCache, c)
//		app.WithHook(func(applet *sandbox.Applet, signal sandbox.Signal) {
//			c.EvictWith(func(name string, db *sqlx.DB) {
//				log.WithField("sandbox", app.Name).Info("auto close database ", name)
//				if db != nil {
//					if err := db.Close(); err != nil {
//						log.WithField("sandbox", app.Name).Error("close database error:", err)
//					}
//
//				}
//			})
//		}, sandbox.SigStop)
//	})
