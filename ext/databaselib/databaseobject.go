package databaselib

import (
	"bufio"
	"database/sql"
	"errors"
	"fmt"
	"github.com/cookieY/sqlx"
	"github.com/d5/tengo/v2"
	log "github.com/sirupsen/logrus"
	"io"
	"io/ioutil"
	"lightbox/ext/util"
	"lightbox/sandbox"
	"os"
	"strings"
	"time"
)

type Database struct {
	app  *sandbox.Applet
	name string
	db   *sqlx.DB
}

//snippet:name=database.close;prefix=close;body=close();desc=close database connection;
func (d *Database) Close(args ...tengo.Object) (tengo.Object, error) {
	return nil, Close(d.app, d.name)
}

//snippet:name=database.exec_file;prefix=exec_file;body=exec_file($1);desc=exec sql file(multiline);
func (d *Database) ExecFile(sql string) error {
	sqlFile, err := os.Open(sql)
	if err != nil {
		return err
	}
	defer func(sqlFile *os.File) {
		err := sqlFile.Close()
		if err != nil {
			log.Errorf("close file %s error:%s", sql, err)
		}
	}(sqlFile)
	reader := bufio.NewReader(sqlFile)
	var currentQuery []byte
	for {
		line, isPrefix, err := reader.ReadLine()
		if err != nil {
			if err == io.EOF {
				break
			} else {
				return err
			}
		}
		sline := string(line)
		if !strings.HasPrefix(sline, "--") {
			if !isPrefix {
				currentQuery = append(currentQuery, '\n')
			}
			currentQuery = append(currentQuery, line...)
		}
		if !isPrefix && strings.HasSuffix(sline, ";") {
			// run sql here
			result, err := d.db.Exec(string(currentQuery[0 : len(currentQuery)-1]))
			if err != nil {
				return err
			} else {
				rows, _ := result.RowsAffected()
				lastId, _ := result.LastInsertId()
				log.Infof("exec %s success,rows:%d,lastId:%d", currentQuery, rows, lastId)
			}
			currentQuery = []byte{}
		}
	}
	return nil
}

//snippet:name=database.exec(sql);prefix=exec;body=exec(${1:sql});
//snippet:name=database.exec;prefix=exec;body=exec(${1:sql},${2:params});
//snippet:name=database.exec;prefix=exec;body=exec(${1:sql},{${2:map}});
func (d *Database) Exec(args ...tengo.Object) (tengo.Object, error) {
	if len(args) < 1 {
		return nil, errors.New("query function need least 1 argument(query)")
	}
	var (
		query      string
		ok         bool
		result     sql.Result
		err        error
		lastId     int64
		rowsAffect int64
	)
	if query, ok = tengo.ToString(args[0]); !ok {
		return nil, errors.New("query argument is not a string")
	}

	if strings.HasPrefix(query, "@") {
		var buf []byte
		buf, err = ioutil.ReadFile(query[1:])
		query = string(buf)
	}

	if len(args) == 1 {
		result, err = d.db.Exec(query)
	} else {
		switch arg := args[1].(type) {
		case *tengo.Array:
			valueArg := util.ToSlice[any](arg.Value)
			query, valueArg, err = sqlx.In(query, valueArg...)
			if err != nil {
				return util.Error(err), nil
			}
			result, err = d.db.Exec(query, valueArg...)
		case *tengo.Map:
			var mapArg map[string]interface{}
			mapArg = util.ToMap[any](arg.Value)
			if err != nil {
				return util.Error(err), nil
			}
			result, err = d.db.NamedExec(query, mapArg)
		case *tengo.ImmutableMap:
			var mapArg map[string]interface{}
			mapArg = util.ToMap[any](arg.Value)
			result, err = d.db.NamedExec(query, mapArg)
		default:
			valueArg := util.ToSlice[any](args[1:])
			query, valueArg, err = sqlx.In(query, valueArg...)
			if err != nil {
				return util.Error(err), nil
			}
			result, err = d.db.Exec(query, valueArg...)
		}
	}
	if err != nil {
		return util.Error(err), nil
	}
	lastId, err = result.LastInsertId()
	if err != nil {
		return util.Error(err), nil
	}
	rowsAffect, err = result.RowsAffected()
	if err != nil {
		return util.Error(err), nil
	}
	return &tengo.Map{Value: map[string]tengo.Object{
		"lastInsertId": &tengo.Int{Value: lastId},
		"rowsAffected": &tengo.Int{Value: rowsAffect},
	}}, nil
}

//Query
//snippet:name=query(sql);prefix=query;body=query(${1:sql});desc=query and return a row array;
//snippet:name=query;prefix=query;body=query(${1:sql},[${2:params}]);desc=query and return row array;
//snippet:name=query;prefix=query;body=query(${1:sql},{${2:map}});desc=query and return row array;
func (d *Database) Query(args ...tengo.Object) (tengo.Object, error) {
	rows, err := d.doQuery(args)
	defer func(rows *sqlx.Rows) {
		if rows == nil {
			return
		}
		err = rows.Close()
		if err != nil {
			log.Error("close rows error:", err)
		}
	}(rows)
	if err != nil {
		return util.Error(err), nil
	}
	var columnTypes []*sql.ColumnType
	if columnTypes, err = rows.ColumnTypes(); err != nil {
		return util.Error(err), nil
	}
	var results []interface{}
	for rows.Next() {
		result, scanErr := rows.SliceScan()
		if scanErr != nil {
			return util.Error(scanErr), nil
		}
		sliceProcess(columnTypes, result)
		results = append(results, result)
	}
	return tengo.FromInterface(results)
}
func (d *Database) QueryRows(args ...tengo.Object) (tengo.Object, error) {
	rows, err := d.doQuery(args)
	defer func(rows *sqlx.Rows) {
		if rows == nil {
			return
		}
		err = rows.Close()
		if err != nil {
			log.Error("close rows error:", err)
		}
	}(rows)
	var columns []string
	columns, err = rows.Columns()
	if err != nil {
		return nil, err
	}
	if err != nil {
		return util.Error(err), nil
	}
	var columnTypes []*sql.ColumnType
	if columnTypes, err = rows.ColumnTypes(); err != nil {
		return util.Error(err), nil
	}
	var results []interface{}
	for rows.Next() {
		result, scanErr := rows.SliceScan()
		if scanErr != nil {
			return util.Error(scanErr), nil
		}
		sliceProcess(columnTypes, result)
		results = append(results, result)
	}
	var cols = make([]tengo.Object, len(columns))
	for idx, c := range columns {
		cols[idx] = &tengo.String{Value: c}
	}
	return tengo.FromInterface(map[string]interface{}{
		"columns": cols,
		"rows":    results,
	})
}

//snippet:name=database.query_first;prefix=query;body=query_first(${1:sql});desc=query and return first row(array);
//snippet:name=database.query_first;prefix=query;body=query_first(${1:sql},[${2:params}]);desc=query and return first row;
//snippet:name=database.query_first;prefix=query;body=query_first(${1:sql},{${2:map}});desc=query and return first row;
func (d *Database) QueryFirst(args ...tengo.Object) (tengo.Object, error) {
	ret, err := d.Query(args...)
	if err == nil {
		switch val := ret.(type) {
		case *tengo.Array:
			if len(val.Value) == 0 {
				ret = tengo.UndefinedValue
			} else {
				ret = val.Value[0]
			}
			return ret, nil
		default:
			return ret, err
		}
	} else {
		return util.Error(err), nil
	}
}

//snippet:name=database.select;prefix=select;body=select(${1:sql});desc=query and return map array(No case);
//snippet:name=database.select;prefix=select;body=select(${1:sql},[${2:params}]);desc=query map array(No case);
//snippet:name=database.select;prefix=select;body=select(${1:sql},{${2:map}});desc=query and map array(No case);
func (d *Database) Select(args ...tengo.Object) (ret tengo.Object, err error) {
	return d.queryMap(false, args...)
}

//snippet:name=database.select_first;prefix=select_first;body=select_first(${1:sql});desc=query and return map array(No case);
//snippet:name=database.select_first;prefix=select_first;body=select_first(${1:sql},[${2:params}]);desc=query map array(No case);
//snippet:name=database.select_first;prefix=select_first;body=select_first(${1:sql},{${2:map}});desc=query and map array(No case);
func (d *Database) SelectFirst(args ...tengo.Object) (tengo.Object, error) {
	ret, err := d.queryMap(false, args...)
	if err == nil {
		switch val := ret.(type) {
		case *tengo.Array:
			if len(val.Value) == 0 {
				ret = tengo.UndefinedValue
			} else {
				ret = val.Value[0]
			}
			return ret, nil
		default:
			return ret, err
		}
	} else {
		return util.Error(err), nil
	}
}

//snippet:name=database.selectx;prefix=selectx;body=select(${1:sql});desc=query and return map array(Camel case);
//snippet:name=database.selectx;prefix=selectx;body=select(${1:sql},[${2:params}]);desc=query map array(Camel case);
//snippet:name=database.selectx;prefix=selectx;body=select(${1:sql},{${2:map}});desc=query and map array(Camel case);
func (d *Database) Selectx(args ...tengo.Object) (ret tengo.Object, err error) {
	return d.queryMap(true, args...)
}

//snippet:name=database.selectx_first;prefix=selectx_first;body=select_first(${1:sql});desc=query and return map (Camel case);
//snippet:name=database.selectx_first;prefix=selectx_first;body=select_first(${1:sql},[${2:params}]);desc=query map (Camel case);
//snippet:name=database.selectx_first;prefix=selectx_first;body=select_first(${1:sql},{${2:map}});desc=query and map (Camel case)
func (d *Database) SelectxFirst(args ...tengo.Object) (tengo.Object, error) {
	ret, err := d.queryMap(true, args...)
	if err == nil {
		switch val := ret.(type) {
		case *tengo.Array:
			if len(val.Value) == 0 {
				ret = tengo.UndefinedValue
			} else {
				ret = val.Value[0]
			}
			return ret, nil
		default:
			return ret, err
		}
	} else {
		return util.Error(err), nil
	}
}

//snippet:name=database.set_max_open;prefix=set_max_open;body=set_max_open(${1:number});desc=set max open connection;
func (d *Database) SetMaxOpen(num int) string {
	d.db.SetMaxOpenConns(num)
	return "OK"
}

//snippet:name=database.set_max_idle;prefix=set_max_idle;body=set_max_idle(${1:number});desc=set max idle connection;
func (d *Database) SetMaxIdle(num int) string {
	d.db.SetMaxIdleConns(num)
	return "OK"
}

//snippet:name=database.set_max_idle_time;prefix=set_max_idle_time;body=set_max_idle_time(${1:duration});desc=set max idle time;
func (d *Database) SetMaxIdleTime(dur string) error {
	idleTime, err := time.ParseDuration(dur)
	if err != nil {
		return err
	}
	d.db.SetConnMaxIdleTime(idleTime)
	return nil
}

//snippet:name=database.set_max_life_time;prefix=set_max_life_time;body=set_max_life(${1:duration});desc=set max lifetime;
func (d *Database) SetMaxLifeTime(dur string) error {
	lifeTime, err := time.ParseDuration(dur)
	if err != nil {
		return err
	}
	d.db.SetConnMaxLifetime(lifeTime)
	return nil
}
func (d *Database) doQuery(args []tengo.Object) (*sqlx.Rows, error) {
	if len(args) < 1 {
		return nil, errors.New("query function need least 1 argument(sql)")
	}
	var (
		query string
		ok    bool
		err   error
		rows  *sqlx.Rows
	)
	if query, ok = tengo.ToString(args[0]); !ok {
		return nil, errors.New("query must be a string")
	}
	//如果字符串带@开头，则从文件读取
	if strings.HasPrefix(query, "@") && !strings.Contains(query, "\n") {
		var buf []byte
		buf, err = ioutil.ReadFile(query[1:])
		if err != nil {
			return nil, err
		}
		query = string(buf)
	}
	if len(args) == 2 {
		switch arg := args[1].(type) {
		case *tengo.ImmutableArray:
			//参数第一个是数组，忽略后面的
			var (
				stmt      *sqlx.Stmt
				valueArgs []interface{}
			)
			valueArgs = util.ToSlice[any](arg.Value)
			query, valueArgs, err = sqlx.In(query, valueArgs...)
			if err != nil {
				return nil, err
			}
			if stmt, err = d.db.Preparex(query); err != nil {
				return nil, err
			}
			if rows, err = stmt.Queryx(valueArgs...); err != nil {
				return nil, err
			}
		case *tengo.Array:
			//参数第一个是数组，忽略后面的
			var (
				stmt      *sqlx.Stmt
				valueArgs []interface{}
			)
			valueArgs = util.ToSlice[any](arg.Value)
			query, valueArgs, err = sqlx.In(query, valueArgs...)
			if err != nil {
				return nil, err
			}
			if stmt, err = d.db.Preparex(query); err != nil {
				return nil, err
			}
			if rows, err = stmt.Queryx(valueArgs...); err != nil {
				return nil, err
			}
		case *tengo.Map:
			// 参数是map，忽略后面的
			var (
				nameStmt *sqlx.NamedStmt
				argMap   map[string]interface{}
			)
			if nameStmt, err = d.db.PrepareNamed(query); err != nil {
				return nil, err
			}
			argMap = util.ToMap[any](arg.Value)
			if rows, err = nameStmt.Queryx(argMap); err != nil {
				return nil, err
			}
		case *tengo.ImmutableMap:
			var (
				nameStmt *sqlx.NamedStmt
				argMap   map[string]interface{}
			)
			if nameStmt, err = d.db.PrepareNamed(query); err != nil {
				return nil, err
			}
			argMap = util.ToMap[any](arg.Value)
			if rows, err = nameStmt.Queryx(argMap); err != nil {
				return nil, err
			}
		default:
			//只有一个参数的情况
			var (
				stmt *sqlx.Stmt
			)
			if stmt, err = d.db.Preparex(query); err != nil {
				return nil, err
			}
			argSingle := tengo.ToInterface(args[1])
			if rows, err = stmt.Queryx(argSingle); err != nil {
				return nil, err
			}

		}
	} else if len(args) > 2 {
		//开放参数
		var (
			stmt      *sqlx.Stmt
			valueArgs []interface{}
		)
		valueArgs = util.ToSlice[any](args[1:])
		query, valueArgs, err = sqlx.In(query, valueArgs...)
		if err != nil {
			return nil, err
		}
		if stmt, err = d.db.Preparex(query); err != nil {
			return nil, err
		}
		if rows, err = stmt.Queryx(valueArgs...); err != nil {
			return nil, err
		}
	} else {
		//没有参数
		var stmt *sqlx.Stmt
		if stmt, err = d.db.Preparex(query); err != nil {
			return nil, err
		}
		if rows, err = stmt.Queryx(); err != nil {
			return nil, err
		}
	}
	return rows, err
}

func (d *Database) String() string {
	return fmt.Sprintf("connection:{name:%s,driver:%s,state:%v}", d.name, d.db.DriverName(), d.db.Stats())
}
func (d *Database) queryMap(camelCase bool, args ...tengo.Object) (tengo.Object, error) {
	rows, err := d.doQuery(args)

	defer func(rows *sqlx.Rows) {
		if rows != nil {
			err = rows.Close()
			if err != nil {
				log.Error("close rows error:", err)
			}
		}
	}(rows)
	if err != nil {
		return util.Error(err), nil
	}

	var columnTypes []*sql.ColumnType
	if columnTypes, err = rows.ColumnTypes(); err != nil {
		return util.Error(err), nil
	}
	columns, _ := rows.Columns()
	cm := map[string]*sql.ColumnType{}
	for _, ctype := range columnTypes {
		cm[ctype.Name()] = ctype
	}
	var results []interface{}
	for rows.Next() {
		row := make(map[string]interface{}, len(columns))
		scanErr := rows.MapScan(row)
		if scanErr != nil {
			return util.Error(scanErr), nil
		}
		row = mapRowProcess(cm, row, camelCase)
		results = append(results, row)
	}

	return tengo.FromInterface(results)
}

func NewDatabase(app *sandbox.Applet, db *sqlx.DB, name string) *Database {
	return &Database{app: app, db: db, name: name}
}

func newDatabaseObject(app *sandbox.Applet, db *sqlx.DB, name string) tengo.Object {
	return util.NewReflectProxy(NewDatabase(app, db, name))
}
