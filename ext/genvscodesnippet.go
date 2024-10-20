//go:build ignore
// +build ignore

package main

import (
	"encoding/json"
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"io/fs"
	"io/ioutil"
	"lightbox/ext"
	"lightbox/sandbox"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

type codeSnippet struct {
	Prefix      []string `json:"prefix"`
	Body        string   `json:"body"`
	Description string   `json:"description,omitempty"`
}

func main() {
	snippet := map[string]*codeSnippet{}
	extra := map[string]*codeSnippet{}
	if _, err := os.Stat("extra.json"); err == nil {
		if data, err := ioutil.ReadFile("extra.json"); err == nil {
			if err = json.Unmarshal(data, &extra); err != nil {
				fmt.Println("load extra code snippet error")
			}
		}
	}
	scanSnippet(extra)
	functions := tengo.GetAllBuiltinFunctions()
	for _, f := range functions {
		snippet[f.Name] = &codeSnippet{
			Prefix:      []string{f.Name},
			Body:        f.Name + "($1)",
			Description: f.Name,
		}
	}
	for n, m := range stdlib.BuiltinModules {
		snippet[fmt.Sprintf("%s:=import(\"%s\")", n, n)] = &codeSnippet{
			Prefix:      []string{n},
			Body:        fmt.Sprintf("%s:=import(\"%s\")", n, n),
			Description: fmt.Sprintf("import module %s", n),
		}
		for k, f := range m {
			s := &codeSnippet{
				Prefix:      []string{k},
				Body:        k + "($1)",
				Description: fmt.Sprintf("%s in module %s", k, n),
			}
			if !f.CanCall() {
				s.Body = k
			}
			snippet[n+"."+k] = s
		}
	}
	for n, m := range ext.RegistryTable.GetRegistryMap() {
		snippet[fmt.Sprintf("%s:=import(\"%s\")", n, n)] = &codeSnippet{
			Prefix:      []string{n},
			Body:        fmt.Sprintf("%s:=import(\"%s\")", n, n),
			Description: fmt.Sprintf("import module %s", n),
		}
		app, _ := sandbox.NewWithDir("DEFAULT", ".")
		for k, f := range m.GetModule(app) {
			s := &codeSnippet{
				Prefix:      []string{n + "." + k, k},
				Body:        n + "." + k + "($1)",
				Description: fmt.Sprintf("%s in module %s", k, n),
			}
			if !f.CanCall() {
				s.Body = n + "." + k
			}
			snippet[n+"."+k] = s
		}
	}

	//do document merge

	for k, v := range extra {
		if r, ok := snippet[k]; ok {
			//do merge action
			if v.Body != "" {
				r.Body = v.Body
			}
			if v.Prefix != nil && len(v.Prefix) > 0 {
				r.Prefix = v.Prefix
			}
			if v.Description != "" {
				r.Description = v.Description
			}
		} else {
			snippet[k] = v
		}
	}

	if data, err := json.MarshalIndent(snippet, "", "  "); err == nil {
		ioutil.WriteFile("tengo.json", data, 0644)
	} else {
		fmt.Println(err)
		os.Exit(-1)
	}
}

func scanSnippet(others map[string]*codeSnippet) {
	re := regexp.MustCompile("snippet:((name=(?P<name>.*?);)|(prefix=(?P<prefix>.*?);)|(body=(?P<body>.*?);)|(desc=(?P<desc>.*);))*")
	_ = filepath.Walk(".", func(path string, file fs.FileInfo, err error) error {
		if !file.IsDir() && !strings.HasSuffix(file.Name(), "_test.go") && (strings.HasSuffix(file.Name(), ".go") ||
			strings.HasSuffix(file.Name(), ".tengo")) {
			fmt.Println("scan snippet for:", path, file.Name())
			if data, readErr := ioutil.ReadFile(path); readErr == nil {
				result := re.FindAllStringSubmatch(string(data), -1)
				names := re.SubexpNames()
				for _, r := range result {
					name := ""
					snippet := &codeSnippet{}
					for idx, n := range names {

						switch n {
						case "name":
							name = r[idx]
						case "prefix":
							snippet.Prefix = strings.Split(r[idx], ",")
						case "body":
							snippet.Body = r[idx]
						case "desc":
							snippet.Description = r[idx]
						}
					}
					if name != "" {
						others[name] = snippet
					}

				}

			} else {
				fmt.Println("read file ", file.Name(), readErr)
			}
		}
		return nil
	})

}
