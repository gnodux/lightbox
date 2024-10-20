package contract

import (
	"fmt"
	"testing"
)

func TestRequire(t *testing.T) {
	var a interface{}
	err := Require(NotNil(a, "a value is nil"))
	if err != nil {
		fmt.Println(err)
	}
}
