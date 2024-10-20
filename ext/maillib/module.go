package maillib

import (
	"github.com/d5/tengo/v2"
	"lightbox/sandbox"
)

var module = map[string]tengo.Object{
	"dial":      &tengo.UserFunction{Name: "dial", Value: FuncASISSRM(newSMTP)},
	"from":      &tengo.String{Value: headerFrom},
	"to":        &tengo.String{Value: headerTo},
	"cc":        &tengo.String{Value: headerCc},
	"bcc":       &tengo.String{Value: headerBcc},
	"replay_to": &tengo.String{Value: headerReplyTo1},
	"subject":   &tengo.String{Value: emailSubject},
	"body":      &tengo.String{Value: "body"},
	"attach":    &tengo.String{Value: "attach"},
	"embed":     &tengo.String{Value: "embed"},
}

var SMTPEntry = sandbox.NewRegistry("smtp", module, nil)
