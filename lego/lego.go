package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/parser"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
	"io"
	"io/fs"
	"io/ioutil"
	"lightbox/env"
	"lightbox/ext"
	"lightbox/ext/modman"
	"lightbox/ext/transpile"
	"lightbox/ext/vfs"
	"lightbox/kvstore"
	"lightbox/sandbox"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"syscall"
)

const (
	sourceFileExt = ".tengo"
	replPrompt    = ">> "
)

type environMap map[string]string

func (receiver *environMap) Set(s string) error {
	idx := strings.Index(s, "=")
	if idx > 0 {
		env.Set(s[0:idx], s[idx+1:])
	}
	return nil
}
func (receiver *environMap) String() string {
	return ""
}

var (
	compileOutput string
	showHelp      bool
	showVersion   bool
	resolvePath   bool // TODO Remove this flag at version 3
	major         = 1
	minor         = 3
	libPath       = ""
	workDir       = ""
	showEnv       bool
	keepDir       bool
	profileName   string
	showMod       bool
	trans         bool
	//start log settings
	logFormat     string
	logLevel      string
	logFile       string
	logStdErr     bool
	logMaxAge     int
	logMaxSize    int
	logLocaltime  bool
	logMaxBackups int
	logCompress   bool
	//end log settings
	envs = environMap{}
	app  *sandbox.Applet
	eval bool
)

func init() {
	flag.Var(&envs, "D", "set environments")
	flag.BoolVar(&showHelp, "help", false, "Show help")
	flag.StringVar(&compileOutput, "o", "", "compile or transpile output file")
	flag.BoolVar(&showVersion, "version", false, "Show version")
	flag.BoolVar(&resolvePath, "resolve", false,
		"Resolve relative import paths")
	flag.StringVar(&libPath, "lib", "", "lib path")
	flag.StringVar(&workDir, "work_dir", "", "work directory")
	flag.BoolVar(&showEnv, "runtime", false, "show runtime environment")
	flag.BoolVar(&keepDir, "k", false, "don't change work dir")
	flag.StringVar(&profileName, "profile", "", "profile)")
	flag.BoolVar(&showMod, "m", false, "show all modules")
	flag.BoolVar(&logStdErr, "log_stderr", false, "also log to std err(default:false)")
	flag.StringVar(&logLevel, "log_level", "debug", "log level(error/warn/info/debug/trace)(default: debug)")
	flag.StringVar(&logFile, "log_file", "", "log file name(default: ./lego.log)")
	flag.IntVar(&logMaxAge, "log_max_age", 0, "maximum number of days to retain old log files(default 0,not remove old logs)")
	flag.IntVar(&logMaxSize, "log_max_size", 100, "log file max size(MB,default:100)")
	flag.IntVar(&logMaxBackups, "log_max_backups", 0, "maximum number of old log files to retain. The default is to retain all old log files )")
	flag.BoolVar(&logLocaltime, "log_localtime", false, "set log time to localtime(default:false)")
	flag.BoolVar(&logCompress, "log_compress", false, "compress log(default:false)")
	flag.StringVar(&logFormat, "log_format", "text", "log format,default text")
	flag.BoolVar(&trans, "t", false, "transpile source file")
	flag.BoolVar(&eval, "e", false, "eval input string")
}

func cleanup() {
	log.Info("do clean up work")
	if app != nil {
		app.Shutdown("sys exit")
	}
	kvstore.Shutdown()
}
func startup() {
	flag.Parse()
	//initialize logger
	if logFile == "" {
		if flag.NArg() > 0 && !trans && !eval {
			logFile = filepath.Join("log", filepath.Base(flag.Arg(0)+".log"))
		} else {
			logFile = "lego.log"
		}
	}
	logger := &lumberjack.Logger{
		Filename:   logFile,
		MaxSize:    logMaxSize,
		MaxAge:     logMaxAge,
		MaxBackups: logMaxBackups,
		LocalTime:  logLocaltime,
		Compress:   logCompress,
	}
	if logStdErr {
		log.SetOutput(io.MultiWriter(os.Stderr, logger))
	} else {
		log.SetOutput(logger)
	}

	lv, err := log.ParseLevel(logLevel)
	if err != nil {
		panic(err)
	}
	log.SetLevel(lv)
	logFormat = strings.ToLower(logFormat)
	var logFormatter log.Formatter
	switch logFormat {
	case "json":
		logFormatter = &log.JSONFormatter{}
	default:
		logFormatter = &log.TextFormatter{}
	}
	log.SetFormatter(logFormatter)
	enableLogger()
	//end
	if profileName != "" {
		env.Set(env.Profile, profileName)
	}
	sigCh := make(chan os.Signal)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGQUIT)
	go func() {
		for {
			sig := <-sigCh
			//do cleanup job
			switch sig {
			case syscall.SIGINT, syscall.SIGKILL, syscall.SIGABRT, syscall.SIGQUIT:
				cleanup()
				os.Exit(0)
			}
		}
	}()
	startHttpServer()
}

// 初始化工作目录（如果脚本文件存在
// initWorkDir
func initWorkDir(scriptFile string) error {
	if keepDir {
		return nil
	}
	if workDir == "" {
		if scriptFile != "" {
			baseDir := filepath.Dir(scriptFile)
			err := os.Chdir(baseDir)
			if err != nil {
				return fmt.Errorf("change work directory to script dir error:%s", err)
			}
		}
	} else {
		err := os.Chdir(workDir)
		if err != nil {
			return fmt.Errorf("change work directory to script dir error:%s", err)
		}
	}
	return nil
}

func printEnv() {
	fmt.Printf(`version         : %d.%d(%s)
runtime         : %s
install    path : %s
global  library : %s
private library : %s
`, major, minor, ext.BuildVersion, runtime.Version(), getInstallPath(), getDefaultLibPath(), getPrivateLibPath())
}

func main() {
	startup()
	defer cleanup()
	if showHelp {
		doHelp()
		os.Exit(2)
	} else if showVersion {
		fmt.Printf("%d.%d (build:%s)\n", major, minor, ext.BuildVersion)
		os.Exit(2)
	} else if showEnv {
		printEnv()
		os.Exit(2)
	}
	var (
		modules tengo.ModuleGetter
		err     error
	)

	inputFile := flag.Arg(0)
	if fi, err := os.Lstat(inputFile); err == nil {
		if (fi.Mode() & os.ModeSymlink) == os.ModeSymlink {
			if inputFile, err = filepath.EvalSymlinks(inputFile); err != nil {
				log.Error("eval symlinks error ", err)
			}
		}
	}
	if err = initWorkDir(inputFile); err != nil {
		log.Error("initialize work directory error:", err)
		os.Exit(1)
	}

	d, _ := filepath.Abs(".")
	app, err = sandbox.NewWithDir("DEFAULT", d)
	//继承自系统的环境变量
	for k, v := range env.All() {
		app.Context.Set(k, v)
	}
	if err != nil {
		log.Error("create sandbox error:", err)
		os.Exit(1)
	}
	modules = getAllModules(app)
	log.Info("module initialized")
	if showMod {
		showModule(app, modules)
		os.Exit(0)
	}
	enableProf()

	if inputFile == "" {
		// REPL
		RunREPL(app, modules, os.Stdin, os.Stdout)
		return
	}
	//transpile file from source files
	if trans {
		log.Info("transpile source file")
		for _, f := range flag.Args() {
			inputData, err := os.ReadFile(f)
			if err != nil {
				fmt.Println("read input file error:", err)
				os.Exit(1)
			}
			outputFile := compileOutput
			if outputFile == "" {
				outputFile = f + ".tengo"
			}
			fmt.Println("transpile source file:", inputFile, " to ", outputFile)
			if tb, err := app.Transpile(inputData); err == nil {
				err = ioutil.WriteFile(outputFile, tb, 0644)
				if err != nil {
					fmt.Println("write file error:", err)
				}
			} else {
				fmt.Println("transpile file error:", err)
			}
		}
		os.Exit(0)
	}
	var inputData []byte
	if eval {
		inputData = []byte(inputFile)
		inputFile = "main.tengo"
	} else {
		inputFile, err = filepath.Abs(inputFile)
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, "Error file path: %s\n", err)
			os.Exit(1)
		}
		//使用base转换文件路径，在之前的initworkdir中已经将目录配置为脚本所在目录
		inputData, err = os.ReadFile(filepath.Base(inputFile))
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr,
				"Error reading input file: %s\n", err.Error())
			os.Exit(1)
		}
	}

	//编译
	if compileOutput != "" {
		err := CompileOnly(modules, inputData, inputFile,
			compileOutput)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}

	if filepath.Ext(inputFile) == sourceFileExt {
		//运行脚本
		err := CompileAndRun(app, modules, inputData, inputFile)
		if err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	} else {
		//运行已编译脚本，但实际用处不大，注入的UserFunction 不能decode
		if err := RunCompiled(modules, inputData); err != nil {
			_, _ = fmt.Fprintln(os.Stderr, err.Error())
			os.Exit(1)
		}
	}
}

func showModule(app *sandbox.Applet, modules tengo.ModuleGetter) {
	for k, m := range ext.RegistryTable.GetRegistryMap() {
		fmt.Println(k, ":")
		mm := m.GetModule(app)
		for n, v := range mm {
			fmt.Println("\t", n, ":", v)
		}
	}
}

// CompileOnly compiles the source code and writes the compiled binary into
// outputFile.
func CompileOnly(
	modules tengo.ModuleGetter,
	data []byte,
	inputFile, outputFile string,
) (err error) {
	data, _ = transpile.Transpile(data)
	symbolTable, _ := preCompile()
	bytecode, err := compileSrc(modules, symbolTable, data, inputFile)
	if err != nil {
		return
	}
	if outputFile == "" {
		outputFile = basename(inputFile) + ".out"
	}

	out, err := os.OpenFile(outputFile, os.O_CREATE|os.O_WRONLY, os.ModePerm)
	if err != nil {
		return
	}
	defer func() {
		if err != nil {
			_ = out.Close()
		} else {
			err = out.Close()
		}
	}()

	err = bytecode.Encode(out)
	if err != nil {
		return
	}
	fmt.Println(outputFile)
	return
}

func preCompile() (*tengo.SymbolTable, []tengo.Object) {

	symbolTable := tengo.NewSymbolTable()
	globals := make([]tengo.Object, tengo.GlobalsSize)
	for idx, fn := range tengo.GetAllBuiltinFunctions() {
		symbolTable.DefineBuiltin(idx, fn.Name)
	}
	return symbolTable, globals
}

// CompileAndRun compiles the source code and executes it.
func CompileAndRun(
	app *sandbox.Applet,
	modules tengo.ModuleGetter,
	data []byte,
	inputFile string,
) (err error) {
	log.Info("compile and run script", inputFile)
	if _, err := app.Run(data, nil, inputFile); err != nil {
		fmt.Println("run script ", inputFile, " error :", err)
	}
	return
}

// RunCompiled reads the compiled binary from file and executes it.
func RunCompiled(modules tengo.ModuleGetter, data []byte) (err error) {
	log.Info("run compiled script")
	bytecode := &tengo.Bytecode{}
	//特殊处理一下，兼容ModuleMap
	var mm *tengo.ModuleMap
	if m, ok := modules.(*modman.Composite); !ok {
		mm = m.GetModuleMap()
	}
	err = bytecode.Decode(bytes.NewReader(data), mm)
	if err != nil {
		return
	}
	machine := tengo.NewVM(bytecode, nil, -1)
	err = machine.Run()
	return
}

// RunREPL starts REPL.
func RunREPL(app *sandbox.Applet, modules tengo.ModuleGetter, in io.Reader, out io.Writer) {
	log.Info("run REPL mode")
	app.Initialize()
	stdin := bufio.NewScanner(in)
	fileSet := parser.NewFileSet()
	symbolTable, globals := preCompile()

	// embed println function
	symbol := symbolTable.Define("__repl_println__")
	globals[symbol.Index] = &tengo.UserFunction{
		Name: "println",
		Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
			var printArgs []interface{}
			for _, arg := range args {
				if _, isUndefined := arg.(*tengo.Undefined); isUndefined {
					printArgs = append(printArgs, "<undefined>")
				} else {
					s, _ := tengo.ToString(arg)
					printArgs = append(printArgs, s)
				}
			}
			printArgs = append(printArgs, "\n")
			_, _ = fmt.Print(printArgs...)
			return
		},
	}

	var constants []tengo.Object
	for {
		if !eval {
			_, _ = fmt.Fprint(out, replPrompt)
		}
		scanned := stdin.Scan()
		if !scanned {
			return
		}

		line := stdin.Text()

		tcode, err := app.Transpile([]byte(line))
		if bytes.Compare(tcode, []byte(line)) != 0 {
			log.Info("transpile:", string(tcode))
		}
		if err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			continue
		}
		srcFile := fileSet.AddFile("repl", -1, len(tcode))
		p := parser.NewParser(srcFile, tcode, nil)
		file, err := p.ParseFile()
		if err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			continue
		}

		file = addPrints(file)
		c := tengo.NewCompiler(srcFile, symbolTable, constants, modules, nil)
		if err := c.Compile(file); err != nil {
			_, _ = fmt.Fprintln(out, err.Error())
			continue
		}

		bytecode := c.Bytecode()
		machine := tengo.NewVM(bytecode, globals, -1)
		if err := machine.Run(); err != nil {
			if !eval {
				_, _ = fmt.Fprintln(out, err.Error())
			}
			continue
		}
		constants = bytecode.Constants
	}
}

func compileSrc(
	modules tengo.ModuleGetter,
	symbolTable *tengo.SymbolTable,
	src []byte,
	inputFile string,
) (*tengo.Bytecode, error) {
	fileSet := parser.NewFileSet()
	srcFile := fileSet.AddFile(filepath.Base(inputFile), -1, len(src))

	p := parser.NewParser(srcFile, src, nil)
	file, err := p.ParseFile()
	if err != nil {
		return nil, err
	}

	c := tengo.NewCompiler(srcFile, symbolTable, nil, modules, nil)

	//c.EnableFileImport(true)
	//if resolvePath {
	//	c.SetImportDir(filepath.Dir(inputFile))
	//}

	if err := c.Compile(file); err != nil {
		return nil, err
	}

	bytecode := c.Bytecode()
	bytecode.RemoveDuplicates()
	return bytecode, nil
}

func doHelp() {
	fmt.Println(`Usage:
lego [flags] {input-file}
	Flags:
		-o        compile output file
		-version  show version
Examples:
	lego
		Start Tengo REPL
	lego myapp.tengo
		Compile and run source file (myapp.tengo)
		Source file must have .tengo extension
	lego -o myapp myapp.tengo
		Compile source file (myapp.tengo) into bytecode file (myapp)
	lego myapp
		RunFile bytecode file (myapp)`)
}

func addPrints(file *parser.File) *parser.File {
	var stmts []parser.Stmt
	for _, s := range file.Stmts {
		switch s := s.(type) {
		case *parser.ExprStmt:
			stmts = append(stmts, &parser.ExprStmt{
				Expr: &parser.CallExpr{
					Func: &parser.Ident{Name: "__repl_println__"},
					Args: []parser.Expr{s.Expr},
				},
			})
		case *parser.AssignStmt:
			stmts = append(stmts, s)

			stmts = append(stmts, &parser.ExprStmt{
				Expr: &parser.CallExpr{
					Func: &parser.Ident{
						Name: "__repl_println__",
					},
					Args: s.LHS,
				},
			})
		default:
			stmts = append(stmts, s)
		}
	}
	return &parser.File{
		InputFile: file.InputFile,
		Stmts:     stmts,
	}
}

func basename(s string) string {
	s = filepath.Base(s)
	n := strings.LastIndexByte(s, '.')
	if n > 0 {
		return s[:n]
	}
	return s
}

func getPrivateLibPath() string {
	if libPath == "" {
		libPath, _ = filepath.Abs("./lib")
	}
	return libPath
}

// 直接ZIP读取效率比较低，直接切换到文件系统
func getAllModules(app *sandbox.Applet) tengo.ModuleGetter {
	module, transpiler, hooks := ext.RegistryTable.GetAll(app, ext.RegistryTable.AllNames()...)
	privatePath, err := filepath.Abs(".")
	if err != nil {
		panic(err)
	}
	newModule := modman.NewModule(module, modman.NewFSImporter(os.DirFS(privatePath), app.Context, app, sourceFileExt))
	for _, p := range filepath.SplitList(getPrivateLibPath()) {
		if p != "" {
			log.Infof("add private lib path: %s", p)
			newModule.AddImporter(modman.NewFSImporter(os.DirFS(p), app.Context, app, sourceFileExt))
			//scan zip files
			zipFiles, _ := fs.Glob(os.DirFS(p), "*.zip")
			log.Info("found zip", zipFiles)
			for _, zf := range zipFiles {
				newModule.AddImporter(modman.NewZipImporter(filepath.Join(p, zf), app.Context, app, sourceFileExt))
			}
		}
	}
	defaultLib := getDefaultLibPath()
	//全局模块优先级较低(方便覆盖或局部使用新版本)
	//默认添加全局路径= 可执行文件路径+lib
	if defaultLib != "" {
		log.Infof("global lib path : %s ", defaultLib)
		for _, p := range filepath.SplitList(defaultLib) {
			newModule.AddImporter(modman.NewFSImporter(os.DirFS(p), app.Context, transpiler, sourceFileExt))
			//scan zip files
			zipFiles, _ := fs.Glob(os.DirFS(p), "*.zip")
			if len(zipFiles) != 0 {
				log.Info("found zip", zipFiles)

				for _, zf := range zipFiles {
					newModule.AddImporter(modman.NewFSImporter(vfs.NewZipFS(filepath.Join(p, zf)), app.Context, transpiler, sourceFileExt))
				}
			}
		}
	}
	app.WithModule(newModule).WithTranspiler(transpiler...).WithHook(hooks...)
	return newModule
}

func getInstallPath() string {
	ex, err := os.Executable()
	if err != nil {
		return ""
	}
	_, err = os.Stat(ex)
	if err == nil {
		if ex, err = filepath.EvalSymlinks(ex); err != nil {
			return ""
		}
	} else {
		return ""
	}
	return filepath.Dir(ex)
}
func getDefaultLibPath() string {
	envLib := os.Getenv("TENGO_LIB")
	if envLib == "" {
		envLib = os.Getenv("LEGO_LIB")
	}
	if envLib == "" {
		envLib, _ = filepath.Abs(filepath.Join(getInstallPath(), "lib/"))
	}
	return envLib
}
