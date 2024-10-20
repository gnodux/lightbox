package modman

import (
	"fmt"
	"path/filepath"
	"testing"
)

func TestImportScriptFromUrl(t *testing.T) {
	imp := ImportUrl("https://www.baidu.com")
	fmt.Println(imp)
}

func TestInt64(t *testing.T) {
	var n1 uint64 = 1234
	var n2 int64 = 1234
	t.Log(int64(n1) == n2)
}

func TestPath(t *testing.T) {
	part := filepath.SplitList("http/jwt/oauth.tengo")
	fmt.Println(part)
}
