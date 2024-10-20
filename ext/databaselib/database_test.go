package databaselib

import (
	"fmt"
	"github.com/cookieY/sqlx"
	"github.com/d5/tengo/v2"
	"lightbox/ext/transpile"
	"lightbox/sandbox"
	"testing"
)

func TestOpen(t *testing.T) {
	var (
		db  *sqlx.DB
		err error
	)
	app, _ := sandbox.NewWithDir("DEFAULT", ".")
	db, err = GetOrOpen(app, "database1", "mysql", "xxtest:xxtest@tcp(127.0.0.1:3306)/metadata?parseTime=true")
	if err != nil {
		t.Log("open database:", db, err)
		t.FailNow()
	}

	type Field struct {
		Id   int64  `db:"meta_id""`
		Name string `db:"name"`
	}

	var results []*Field
	if err = db.Select(&results, "select meta_id,name from x_field"); err != nil {
		t.Log(err)
		t.FailNow()
	}
	t.Log(results)

}

//func TestDb(t *testing.T) {
//	db, err := GetOrOpen("database1", "mysql", "xxtest:xxtest@tcp(127.0.0.1:3306)/metadata?parseTime=false")
//	db.Exec("update x_field set label=? and tr_label=? where meta_id=?")
//}

func TestMySqlMap(t *testing.T) {
	var (
		db  *sqlx.DB
		err error
	)
	app, _ := sandbox.NewWithDir("DEFAULT", ".")
	db, err = GetOrOpen(app, "database1", "mysql", "xxtest:xxtest@tcp(127.0.0.1:3306)/metadata?parseTime=true")
	if err != nil {
		t.Log("open database:", db, err)
		t.FailNow()
	}
	stmt, err := db.Preparex("select * from x_field")
	if err != nil {
		t.Fatal(err)
	}
	rows, err := stmt.Queryx()
	if err != nil {
		t.Fatal(err)
	}
	var results []map[string]interface{}
	colTypes, _ := rows.ColumnTypes()
	for _, colType := range colTypes {
		t.Log(colType.Name(), colType.ScanType(), colType.DatabaseTypeName())
	}

	for rows.Next() {
		r := map[string]interface{}{}
		rows.MapScan(r)
		results = append(results, r)
	}
	t.Log(results)
}

func TestMapper(t *testing.T) {
	db, err := sqlx.Open("mysql", "xxtest:xxtest@tcp(127.0.0.1:3306)/metadata?parseTime=true")
	if err == nil {
		fmt.Println(db.DriverName())
	}
}

func TestSQLServer(t *testing.T) {
	app, _ := sandbox.NewWithDir("DEFAULT", ".")
	db, ok := GetOrOpen(app, "testdb1", "mssql", "server=127.0.0.1;port=1433;database=s1;user id=sa;password=xxxtest")
	//db.Exec("insert into company values(1,'n1')")
	//db.Exec("insert into company values(2,'n2')")
	rows, err := db.Queryx("select * from company")
	if err != nil {
		t.Errorf("query error %v", err)
	}
	for rows.Next() {
		r := map[string]interface{}{}
		rows.MapScan(r)
		t.Log(r)
	}
	t.Log(db, ok)
}
func TestIn(t *testing.T) {
	db, err := sqlx.Open("mysql", "xxtest:xxtest@tcp(127.0.0.1:3306)/metadata?parseTime=true")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(db)
	in, i, err := sqlx.In("select * from x_field where tenant_id in (?) and label like ?", []interface{}{100102, 100103}, "%org%")
	if err != nil {
		return
	}
	fmt.Println(in, i, err)

}
func BenchmarkExtLib(b *testing.B) {

	for i := 0; i < b.N; i++ {
		r := compiled.Clone()
		r.Set("name", i)
		r.Run()
	}
}

var compiled *tengo.Compiled

func TestMain(m *testing.M) {
	m.Run()
	//	script := []byte(`
	//fib := func(x) {
	//	if x == 0 {
	//		return 0
	//	} else if x == 1 {
	//		return 1
	//	}
	//	return fib(x-1) + fib(x-2)
	//}
	//fmt:=import("fmt")
	//fmt.println("name")
	//fib(35)
	//`)
	//	part := tengo.NewScript(script)
	//	modules := stdlib.GetAllModuleMap(stdlib.AllNames()...)
	//	modules.AddMap(GetAllModuleMap())
	//	part.SetImports(modules)
	//	var err error
	//	compiled, err = part.Compile()
	//	if err != nil {
	//		panic(err)
	//	}
	//	m.RunFile()
	//	os.Exit(0)
}

func TestSlice(t *testing.T) {
	app, _ := sandbox.NewWithDir("DEFAULT", ".")
	db, err := GetOrOpen(app, "database1", "mysql", "xxtest:xxtest@tcp(127.0.0.1:3306)/metadata?parseTime=true")
	if err != nil {
		t.Error(err)
		return
	}
	p, err := db.Preparex("select * from x_field where tenant_id = ?")
	if err != nil {
		t.Error(err)
		return
	}
	var arg int64 = 120200
	rows, err := p.Queryx(arg)
	if err != nil {
		t.Error(err)
		return
	}

	defer rows.Close()
}

func TestTranspile(t *testing.T) {
	code := `
db[abc].query=>select * from
opt_user where user1 in (1,2,3);


user_id:= 10001223
n:=times.today()
name:="12332"
db[test].exec=>
	select * from opt_user
	where 
	user_id=#user_id 
	and create_time> #n 
;

result:=db[test].exec=>delete from opt_user where user_id=#userId;
result=db[test].selectx=>select * from user where username like '%${name}%';
`

	transpile.G.Add(sqlDialect)
	r, _ := transpile.Transpile([]byte(code))
	fmt.Println(string(r))
}
