package transpile

import (
	"bytes"
	"fmt"
	"lightbox/ext/util"
)

type TransFunc func(src []byte) ([]byte, error)

type Transpiler interface {
	Transpile(src []byte) ([]byte, error)
}

type Group []TransFunc

func (g Group) Transpile(src []byte) ([]byte, error) {
	if bytes.Index(src, []byte(noTranspile)) >= 0 {
		return src, nil
	}
	var err error
	for _, c := range g {
		if src, err = c(src); err != nil {
			break
		}
	}
	return src, err
}

//func (g *Group) Add(transpiler ...TransFunc) {
//	*g = append(*g, transpiler...)
//}
//
//func (g *Group) AddReplace(pattern, repl string) error {
//	if transpiler, err := NewReplace(pattern, repl); err == nil {
//		g.Add(transpiler)
//
//	} else {
//		return err
//	}
//	return nil
//}
//func (g *Group) AddReplaceFunc(pattern string, repl func([]byte) []byte) error {
//	if t, err := NewReplaceFunc(pattern, repl); err == nil {
//		g.Add(t)
//	} else {
//		return err
//	}
//	return nil
//}

var (
	noTranspile = fmt.Sprintf(util.LegoAnnotation, "transpile", "ignore")
)

func Must(transFunc TransFunc, err error) TransFunc {
	if err != nil {
		panic(err)
	}
	return transFunc
}

var G Group

var Transpile TransFunc

func init() {
	//G.Add(ShebangRemove, ImportOptimize)
	//G.AddReplace(`(?m)\((.*?)\)=>`, `func($1)`)
	//G.AddReplace(`(?m)func\s*(\S+)\s*\(`, `$1:=func(`)
	G = append(G, ShebangRemove, ImportOptimize, Must(NewReplace(`(?m)\((.*?)\)=>`, `func($1)`)),
		Must(NewReplace(`(?m)func\s*(\S+)\s*\(`, `$1:=func(`)))
	Transpile = G.Transpile
}
