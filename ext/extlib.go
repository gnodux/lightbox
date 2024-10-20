package ext

//go:generate go run gensrcmods.go
//go:generate go run genvscodesnippet.go
import (
	"github.com/d5/tengo/v2/stdlib"
	"lightbox/ext/amqplib"
	"lightbox/ext/badgerlib"
	"lightbox/ext/canallib"
	"lightbox/ext/cronlib"
	"lightbox/ext/cryptlib"
	"lightbox/ext/databaselib"
	"lightbox/ext/envlib"
	"lightbox/ext/helplib"
	"lightbox/ext/httplib"
	"lightbox/ext/loglib"
	"lightbox/ext/maillib"
	"lightbox/ext/osslib"
	"lightbox/ext/pathlib"
	"lightbox/ext/redislib"
	"lightbox/ext/syslib"
	"lightbox/ext/tpllib"
	"lightbox/ext/uuidlib"
	"lightbox/ext/xlslib"
	"lightbox/sandbox"
)

//var importDir = flag.String("import", "", "set import directories")

//snippet:name=import;prefix=import;body=import($1,$2);

//RegistryTable 所有模块的入口,包括扩展模块及标准模块

var RegistryTable = sandbox.NewRegistryTable(
	syslib.Entry,
	pathlib.PathEntry,
	pathlib.FilePathEntry,
	databaselib.Entry,
	httplib.Entry,
	maillib.SMTPEntry,
	badgerlib.Entry,
	canallib.Entry,
	cronlib.Entry,
	tpllib.Entry,
	osslib.Entry,
	loglib.Entry,
	envlib.Entry,
	xlslib.Entry,
	amqplib.Entry,
	redislib.Entry,
	cryptlib.Entry,
	helplib.Entry,
	uuidlib.Entry,
).WithSourceModule(SourceModules).WithSourceModule(stdlib.SourceModules).WithModule(stdlib.BuiltinModules)
