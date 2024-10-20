package transpile

import (
	"fmt"
	"regexp"
	"strings"
)

func ShebangRemove(input []byte) ([]byte, error) {
	if len(input) > 1 && string(input[:2]) == "#!" {
		copy(input, "//")
	}
	return input, nil
}

var impR = regexp.MustCompile("(?ms)^\\s*import\\s*\\(.*?\\)")

//ImportOptimize optimize import like import(sys,enum) to sys:=import("sys")\nenum:=import("enum")
// import(enum.each) will transpile to each:=import("enum").each
func ImportOptimize(src []byte) ([]byte, error) {
	return impR.ReplaceAllFunc(src, func(buf []byte) []byte {
		s := string(buf)
		imps := strings.Split(s[strings.Index(s, "(")+1:strings.Index(s, ")")], ",")
		var ret []byte
		prefix := ""
		for _, imp := range imps {
			n := strings.Trim(imp, " \"\n\r\t ")
			ret = append(ret, []byte(prefix)...)
			if strings.Contains(n, ".") {
				ret = append(ret, []byte(fmt.Sprintf("%s:=import(\"%s\")%s",
					n[strings.LastIndex(n, ".")+1:],
					n[0:strings.Index(n, ".")],
					n[strings.Index(n, "."):]))...)
			} else {
				ret = append(ret, []byte(fmt.Sprintf("%s:=import(\"%s\")", n, n))...)
			}
			if strings.Contains(imp, "\n") {
				ret = append(ret, '\n')
			} else {
				prefix = ";"
			}

		}
		//ret = append(ret, '\n')
		return ret
	}), nil
}
