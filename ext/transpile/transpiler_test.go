package transpile

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"regexp"
	"testing"
)

func TestGroup_Transpile(t *testing.T) {
	code := `#!/usr/local/bin/lego
sys:=import("sys")
import(enum,sys,database)
import(
	canal,
	amqp,
	redis
)
f1:=()=>{
	fmt.println(1)
}
f2:=(a,b,c)=>{
	fmt.println(a,b,c)
}


import(enum)
enum.each(items,(idx,v)=>{
})
`
	result, _ := G.Transpile([]byte(code))
	fmt.Println(string(result))
}

func TestScriptTranspile(t *testing.T) {
	script := tengo.NewScript([]byte(`
fmt:=import("fmt")
fmt.println(src)
dest:=src+"imb"
`))
	script.SetImports(stdlib.GetModuleMap(stdlib.AllModuleNames()...))
	script.WithConstants([]tengo.Object{&tengo.Int{Value: 1}})
	script.Add("src", "source1")
	ret, err := script.Run()
	fmt.Println(ret.GetAll())
	fmt.Println(ret, err)
}
func TestMultiGroup(t *testing.T) {
	code := `#!/usr/local/bin/lego
import(fmt,http,sys,os,crypt.md5)

passwd:=md5.encode("123456")

for{
    fmt.println(http.post({url:"http://192.168.120.249:8080/deviceManage/getDeviceSN",data:{pass:passwd}}))
}

`
	result, _ := G.Transpile([]byte(code))
	fmt.Println(string(result))
}

func TestMultiLineReplace(t *testing.T) {
	code := `
db[abc].query: select * from
opt_user where user1 in (1,2,3);


user_id:= 10001223
n:=times.today()

db[test].exec:select * from opt_user where user_id=$user_id and create_time> $n ;

result:=db[test].exec:delete from opt_user where user_id=userId;
`

	r := regexp.MustCompile(`(?s)db\[(?P<name>\w+)].(?P<func>\w+):(?P<sql>.*?);`)
	fmt.Println(r.SubexpNames())
	tcode := r.ReplaceAllStringFunc(code, func(s string) string {
		vars := r.FindStringSubmatch(s)
		s = "sys.must(database.open(\"" + vars[1] + "\"))." + vars[2] + "(`" + vars[3] + "`)"
		return s
	})
	println(tcode)
}
