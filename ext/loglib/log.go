package loglib

import (
	"errors"
	"github.com/d5/tengo/v2"
	log "github.com/sirupsen/logrus"
	"lightbox/ext/util"
	"lightbox/sandbox"
)

const (
	sandboxName = "sandbox"
)

var module = map[string]sandbox.UserFunction{

	//Trace
	"trace": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).Trace)(args...)
	},
	"trace_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Trace)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	"traceln": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).Traceln)(args...)
	},
	"traceln_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Traceln)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	"tracef": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncASIs(log.WithField(sandboxName, sandbox.Name).Tracef)(args...)
	},
	"tracef_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncASIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Tracef)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},

	//Deebug
	"debug": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).Debug)(args...)
	},
	"debug_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Debug)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	"debugln": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).Debugln)(args...)
	},
	"debugln_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Debugln)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	"debugf": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncASIs(log.WithField(sandboxName, sandbox.Name).Debugf)(args...)
	},

	"debugf_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncASIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Debugf)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},

	//Info
	"info": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).Info)(args...)
	},
	"info_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Info)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	"infoln": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).Infoln)(args...)
	},
	"infoln_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Infoln)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},

	//snippet:name=log.infof;prefix=log.infof;body=log.infof(${1:format},${2:values});
	"infof": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncASIs(log.WithField(sandboxName, sandbox.Name).Infof)(args...)
	},
	"infof_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncASIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Infof)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	//Error
	"err": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).Error)(args...)
	},
	"err_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Error)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	"errln": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).Errorln)(args...)
	},
	"errln_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Errorln)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	//snippet:name=log.errorf;prefix=log.errorf;body=log.errorf(${1:format},${2:arguments});
	"errf": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncASIs(log.WithField(sandboxName, sandbox.Name).Errorf)(args...)
	},
	"errf_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncASIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Errorf)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	//warning

	//snippet:name=log.warn;prefix=log.warn;body=log.warnf($1,$2);
	"warn": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).Warn)(args...)
	},
	"warn_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Warn)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	//snippet:name=log.warnln;prefix=log.warnln;body=log.warnf($1,$2);
	"warnln": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).Warnln)(args...)
	},
	"warnln_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncAIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Warnln)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
	//snippet:name=log.warnf;prefix=log.warnf;body=log.warnf($1,$2);
	"warnf": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		return util.FuncASIs(log.WithField(sandboxName, sandbox.Name).Warnf)(args...)
	},
	"warnf_with": func(sandbox *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
		if len(args) < 1 {
			return nil, tengo.ErrWrongNumArguments
		}
		m := tengo.ToInterface(args[0])
		if mm, ok := m.(map[string]interface{}); ok {
			return util.FuncASIs(log.WithField(sandboxName, sandbox.Name).WithFields(mm).Warnf)(args[1:]...)
		} else {
			return nil, errors.New("first argument must be a map")
		}
	},
}

var Entry = sandbox.NewRegistry("log", nil, module)
