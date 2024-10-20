//go:build ignore
// +build ignore

package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"time"
)

var tengoModFileRE = regexp.MustCompile(`^srcmod_(\w+).tengo$`)

func main() {
	v, _ := os.Create("version.go")
	v.WriteString(fmt.Sprintf(`
package ext
const BuildVersion="%s"
`, time.Now().Format("20060102150405")))
	modules := make(map[string]string)
	// enumerate all tengo module files
	srcDir := "./sourcemod/"
	files, err := ioutil.ReadDir(srcDir)
	if err != nil {
		log.Fatal(err)
	}
	for _, file := range files {
		m := tengoModFileRE.FindStringSubmatch(file.Name())
		if m != nil {
			modName := m[1]

			src, err := ioutil.ReadFile(filepath.Join(srcDir, file.Name()))
			if err != nil {
				log.Fatalf("file '%s' read error: %s",
					file.Name(), err.Error())
			}

			modules[modName] = string(src)
		}
	}

	var out bytes.Buffer
	out.WriteString(`// Code generated using gensrcmods.go; DO NOT EDIT.

package ext

// SourceModules are source type standard library modules.
var SourceModules = map[string]string{` + "\n")
	for modName, modSrc := range modules {
		out.WriteString("\t\"" + modName + "\": " +
			strconv.Quote(modSrc) + ",\n")
	}
	out.WriteString("}\n")

	const target = "source_modules.go"
	if err := ioutil.WriteFile(target, out.Bytes(), 0644); err != nil {
		log.Fatal(err)
	}
}
