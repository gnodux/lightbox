package ext

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	tjson "github.com/d5/tengo/v2/stdlib/json"
	"lightbox/sandbox"
	"path/filepath"
	"testing"
)

func TestVM(t *testing.T) {

}

func TestBadgerAccess_Get(t *testing.T) {

	src := []byte(`
fmt:=import("fmt")
kv:=import("badger")
t:=kv.open("c1","kv-test-data")
t.set({"ns1:k1":"value1","ns1:k2":"value2"})
v:=t.get("ns1:k1","ns1:k2")
database:=import("database")
conn:=database.open("conn","mysql","xxtest:xxtest@tcp(127.0.0.1:3306)/metadata")
if is_error(v){
	fmt.println("error:",v)
}else{
	fmt.println(v)
}
export func(){
	times:=import("times")
	return "here"+times.now()
}
`)

	app, _ := sandbox.NewWithDir("DEFAULT", ".")
	modules := RegistryTable.GetModuleMap(app, RegistryTable.AllNames()...)
	app.WithModule(modules)
	c, err2 := app.Run(src, nil, "main.tengo")
	if err2 != nil {
		panic(err2)
	}
	vars := c.GetAll()
	for _, v := range vars {
		fmt.Println(v.Name(), v.Value())
	}
	if err2 != nil {
		t.Error(err2)
		return
	}
}

func TestGet(t *testing.T) {
	tests := []struct {
		name   string
		script string
	}{
		{
			name: "test_local_db",
			script: `
main:=func(){
	database:=import("database")
	fmt:=import("fmt")
	d1:=database.open("db1","mysql","xxtest:xxtest@tcp(127.0.0.1:3306)/metadata")
	data:=d1.select("select * from x_field")
	d1.close()
	return data
}
ex:=main()
`}, {
			name: "test_kv",
			script: `
ex:=undefined
badger:=import("badger")
fmt:=import("fmt")
kv:=badger.open()
if is_error(kv){
	ex=kv
}
//ex=kv.set({"ns2:k1":"vvvvvvv"})
if !is_error(ex){
	ex=kv.get("ns2:k1")
	fmt.println(ex)
	ex=kv.del("ns2:k1","ns2:k2")
	fmt.println(ex)
}

`,
		}, {
			name: "test_db1",
			script: `
database:=import("database")
fmt:=import("fmt")
db:=database.open("mysql","xxtest:xxtest@tcp(127.0.0.1:3306)/metadata")
data:=db.query("select * from x_object")
ex:=undefined
if is_error(data){
	ex=data
}
`,
		}, {
			name: "test_http",
			script: `http:=import("http")
os:=import("os")
fmt:=import("fmt")
response:=http.get("https://www.baidu.com")
//fmt.println(response)
f:=os.create("baidu.html")
if !is_error(f){
	f.write(response.body)
	f.close()
}
`,
		}, {
			name: "text template",
			script: `
fmt:=import("fmt")
tpl:=import("tpl")
t:=tpl.text("Hello,{{.name.Value}}")
fmt.println(t.render({name:"gnodux<gnodux@gmail.com>>"}))
`,
		},
		{
			name: "html template",
			script: `
fmt:=import("fmt")
tpl:=import("tpl")
t:=tpl.html("<div>Hello,{{.name.Value -}}\n</div>")
fmt.println(t.render({name:"gnodux<gnodux@gmail.com>>"}))
`,
		}, {
			name: "html template from file",
			script: `
fmt:=import("fmt")
tpl:=import("tpl")
t:=tpl.html("@test.tpl")
fmt.println(t.render({name:"gnodux<gnodux@gmail.com>>"}))
`,
		}, {
			name: "dynamic import",
			script: `
fmt:=import("fmt")
sys:=import("sys")
sys.must(fmt.println("abc"))


`,
		}, {
			name: "http api",
			script: `
http:=import("http")
s:=http.server({
	addr:":8099"
})
s.handle_static("/static/","./")
s.handle_script_dir("/api/","./controller")
//s.serve()
`,
		}, {
			name: "badger options",
			script: `
badger:=import("badger")
opt:=badger.option("./store01")
fmt:=import("fmt")
fmt.println(opt)
//opt.MemTableSize=102345
fmt.println(opt)
b:=badger.open("test1",opt)
fmt.println(b.set("a","b"))
`,
		},
	}
	app, _ := sandbox.NewWithDir("DEFAULT", ".")
	app.WithModule(RegistryTable.GetModuleMap(app, RegistryTable.AllNames()...))
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, e := app.Run([]byte(tt.script), nil, tt.name)
			if e != nil {
				t.Errorf("run script %s error:%v", tt.name, e)
			} else {
				if c.IsDefined("ex") {
					v := c.Get("ex")
					if v.Value() != nil && v.ValueType() == "error" {
						t.Error(v.Value())
						t.FailNow()
					}
				}
			}
		})
	}
}

func TestBytes(t *testing.T) {
	var buffer = []byte{82, 20, 47, 190, 227, 238, 4, 3, 100, 21, 64}
	id, minor, major := decodeBLE(buffer)
	fmt.Println(id, major, minor)
}

func decodeBLE(buffer []byte) (string, uint16, uint16) {
	id := hex.EncodeToString(buffer[0:6])
	minor := binary.BigEndian.Uint16(buffer[6:8])
	major := binary.BigEndian.Uint16(buffer[8:10])
	return id, minor, major
}
func TestPathDir(t *testing.T) {
	fmt.Println(filepath.Dir("hyperv/counter.tengo"))
}

func TestCallback(t *testing.T) {
	var f *tengo.Variable
	s1 := tengo.NewScript([]byte(`
callback:=func(){
	import("fmt").println("123")
}
`))
	s1.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	s2 := tengo.NewScript([]byte(`
fmt:=import("fmt")
fmt.println(f)
`))
	s2.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	r, err := s1.Run()
	if err != nil {
		fmt.Println(err)
	}
	f = r.Get("callback")
	s2.Add("f", f.Value())
	r, err = s2.Run()
	if err != nil {
		fmt.Println(err)
	}
}

func TestScriptModule(t *testing.T) {

}

func TestJSON(t *testing.T) {
	m := &tengo.ImmutableArray{Value: []tengo.Object{
		&tengo.Int{Value: 123},
		&tengo.String{Value: `{"zh_cn":"组织机构"}`},
		&tengo.String{Value: `{"zh_cn":"组织机构"}`},
		&tengo.String{Value: "name_is"},
	}}
	buf, err := tjson.Encode(m)
	if err == nil {
		fmt.Println(string(buf))
	}
}
