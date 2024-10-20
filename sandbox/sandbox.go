package sandbox

import (
	"github.com/d5/tengo/v2"
)

type UserFunction func(sandbox *Applet, args ...tengo.Object) (tengo.Object, error)

func NewUserFunction(sandbox *Applet, fn UserFunction) tengo.CallableFunc {
	return func(args ...tengo.Object) (ret tengo.Object, err error) {
		return fn(sandbox, args...)
	}
}
