package maillib

import (
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"gopkg.in/gomail.v2"
	"io"
	"net/mail"
	"os"
)

const (
	HeaderFrom     = "From"
	headerFrom     = "from"
	HeaderTo       = "To"
	headerTo       = "to"
	HeaderCc       = "Cc"
	headerCc       = "cc"
	HeaderBcc      = "Bcc"
	headerBcc      = "bcc"
	HeaderReplyTo  = "Reply-To"
	headerReplyTo  = "reply-to"
	headerReplyTo1 = "reply_to"
	headerReplyTo2 = "reply"
	EMailSubject   = "Subject"
	emailSubject   = "subject"
)

var (
	defaultEncoding     = gomail.SetEncoding(gomail.Base64)
	defaultCharset      = gomail.SetCharset("UTF-8")
	defaultPartEncoding = gomail.SetPartEncoding(gomail.Base64)
)

// newSMTP
//
//snippet:name=smtp.dial;prefix=dial;body=dial(${1:host},${2:port},${3:username},${4:password});
func newSMTP(host string, port int, username string, password string) (tengo.Object, error) {
	client, err := NewSMTPClient(host, port, username, password)
	if err != nil {
		return tengo.FromInterface(err)
	}
	return &tengo.ImmutableMap{Value: map[string]tengo.Object{
		"send":  &tengo.UserFunction{Name: "send", Value: client.send},
		"close": &tengo.UserFunction{Name: "close", Value: stdlib.FuncARE(client.Close)},
	}}, nil
}

func FuncASISSRM(fn func(string, int, string, string) (tengo.Object, error)) tengo.CallableFunc {
	return func(args ...tengo.Object) (tengo.Object, error) {
		if len(args) != 4 {
			return nil, tengo.ErrWrongNumArguments
		}
		s1, ok := tengo.ToString(args[0])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "arg0",
				Expected: "string(compatible)",
				Found:    args[0].TypeName(),
			}
		}
		i2, ok := tengo.ToInt(args[1])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "arg1",
				Expected: "int(compatible)",
				Found:    args[1].TypeName(),
			}
		}

		s3, ok := tengo.ToString(args[2])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "arg3",
				Expected: "string(compatible)",
				Found:    args[2].TypeName(),
			}
		}
		s4, ok := tengo.ToString(args[3])
		if !ok {
			return nil, tengo.ErrInvalidArgumentType{
				Name:     "arg4",
				Expected: "string(compatible)",
				Found:    args[3].TypeName(),
			}
		}

		return fn(s1, i2, s3, s4)
	}
}

type smtpClient struct {
	sender gomail.SendCloser
}

// Send header & contentType ,see also: https://datatracker.ietf.org/doc/html/rfc4021
func (s *smtpClient) Send(header map[string][]string, contentType string, body string, attachments ...string) error {
	msg := gomail.NewMessage(defaultCharset, defaultEncoding)
	msg.SetHeaders(header)
	msg.SetBody(contentType, body, defaultPartEncoding)
	if contentType == "" {
		contentType = "text/html"
	}
	for _, att := range attachments {
		if _, err := os.Stat(att); err != nil {
			return err
		}
		msg.Attach(att)
	}

	return gomail.Send(s.sender, msg)
}

func (s *smtpClient) SendHtml(header map[string][]string, body string, attachments ...string) error {
	return s.Send(header, "text/html", body, attachments...)
}

// send
//
//snippet:name=smtp.send;prefix=send;body=send({ from:$1, to:$2, cc:$3, bcc:$4, reply_to:$5,subject:$6,body:$7});
func (s *smtpClient) send(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 1 {
		return nil, tengo.ErrWrongNumArguments
	}
	aMap, ok := args[0].(*tengo.Map)
	if !ok {
		return nil, tengo.ErrInvalidArgumentType{Name: "send", Expected: "Map"}
	}
	msg := gomail.NewMessage(defaultEncoding, defaultCharset)
	header := map[string][]string{}

	if subRaw, ok := aMap.Value[emailSubject]; ok {
		if subject, ok := tengo.ToString(subRaw); ok {
			header[EMailSubject] = []string{subject}
		}
	}

	if toRaw, ok := aMap.Value[headerTo]; ok {
		addrs, err := extractAddress(toRaw, msg)
		if err != nil {
			return tengo.FromInterface(err)
		}
		header[HeaderTo] = addrs
	} else {
		return tengo.FromInterface(errors.New("send email require to field"))
	}

	//抄送字段，非必填
	if ccRaw, ok := aMap.Value[headerCc]; ok {
		if addrs, err := extractAddress(ccRaw, msg); err != nil {
			return tengo.FromInterface(err)
		} else {
			header[HeaderCc] = addrs
		}
	}
	//密送字段，非必填
	if bccRaw, ok := aMap.Value[headerBcc]; ok {
		if addrs, err := extractAddress(bccRaw, msg); err != nil {
			return tengo.FromInterface(err)
		} else {
			header[HeaderBcc] = addrs
		}
	}
	//回复到，非必填
	replyRaw, ok := aMap.Value[headerReplyTo]
	if !ok {
		replyRaw, ok = aMap.Value[headerReplyTo1]
	}
	if !ok {
		replyRaw, ok = aMap.Value[headerReplyTo2]
	}
	if ok {
		if addrs, err := extractAddress(replyRaw, msg); err != nil {
			return tengo.FromInterface(err)
		} else {
			header[HeaderReplyTo] = addrs
		}
	}
	//发件人
	if fromRaw, ok := aMap.Value[headerFrom]; ok {
		if addrs, err := extractAddress(fromRaw, msg); err != nil {
			return tengo.FromInterface(err)
		} else {
			header[HeaderFrom] = addrs
		}
	}
	msg.SetHeaders(header)
	var (
		body        string
		contentType string
		attachList  *tengo.Array
		embedList   *tengo.Array
		err         error
	)
	if body, err = getString(aMap, "body"); err != nil {
		return tengo.FromInterface(err)
	}
	contentType, _ = getString(aMap, "contentType")
	if contentType == "" {
		contentType, _ = getString(aMap, "content-type")
	}
	if contentType == "" {
		contentType = "text/html"
	}
	msg.SetBody(contentType, body, defaultPartEncoding)
	if attachList, err = getObjects(aMap, "attach"); err != nil && attachList != nil && attachList.Value != nil && len(attachList.Value) > 0 {
		for _, f := range attachList.Value {
			switch ff := f.(type) {
			case *tengo.Map:
				name, err := getString(ff, "name")
				if err != nil {
					return tengo.FromInterface(fmt.Errorf("require file name"))
				}
				payload, ok := ff.Value["body"]
				if !ok {
					return tengo.FromInterface(fmt.Errorf("require body filed"))
				}
				switch p := payload.(type) {
				case *tengo.String:
					msg.Attach(p.Value, gomail.Rename(name))
				case *tengo.Bytes:
					msg.Attach(name, gomail.SetCopyFunc(func(writer io.Writer) error {
						_, werr := writer.Write(p.Value)
						return werr
					}))
				default:
					return tengo.FromInterface(fmt.Errorf("unknown attach type:%s", payload.TypeName()))
				}
			case *tengo.String:
				msg.Attach(ff.Value)
			}
		}

	}

	if embedList, err = getObjects(aMap, "embed"); err != nil && embedList != nil && embedList.Value != nil && len(embedList.Value) > 0 {
		for _, f := range embedList.Value {
			switch ff := f.(type) {
			case *tengo.Map:
				name, err := getString(ff, "name")
				if err != nil {
					return tengo.FromInterface(fmt.Errorf("require file name"))
				}
				payload, ok := ff.Value["body"]
				if !ok {
					return tengo.FromInterface(fmt.Errorf("require body filed"))
				}
				switch p := payload.(type) {
				case *tengo.String:
					msg.Embed(p.Value, gomail.Rename(name))
				case *tengo.Bytes:
					msg.Embed(name, gomail.SetCopyFunc(func(writer io.Writer) error {
						_, werr := writer.Write(p.Value)
						return werr
					}))
				default:
					return tengo.FromInterface(fmt.Errorf("unknown attach type:%s", payload.TypeName()))
				}
			case *tengo.String:
				msg.Embed(ff.Value)
			default:
				return tengo.FromInterface(fmt.Errorf("unknown embed type:%s", f.TypeName()))
			}
		}
	}
	err = gomail.Send(s.sender, msg)
	if err != nil {
		return tengo.FromInterface(err)
	}

	return nil, nil
}
func getObjects(m *tengo.Map, key string) (*tengo.Array, error) {
	r, ok := m.Value[key]
	if !ok {
		return nil, fmt.Errorf("%s not found", key)
	}
	switch v := r.(type) {
	case *tengo.Array:
		return v, nil
	default:
		return &tengo.Array{Value: []tengo.Object{v}}, nil
	}

}
func getStrings(m *tengo.Map, key string) ([]string, error) {
	r, ok := m.Value[key]
	if !ok {
		return nil, fmt.Errorf("%s not found", key)
	}
	var strs []string
	switch t := r.(type) {
	case *tengo.Array:
		for _, v := range t.Value {
			vv, ok := tengo.ToString(v)
			if !ok {
				return nil, errors.New("convert [to] field to string error")
			}
			strs = append(strs, vv)

		}
	default:
		vv, ok := tengo.ToString(t)
		if !ok {
			return nil, errors.New("convert [to] field to string error")
		}
		strs = append(strs, vv)
	}
	return strs, nil
}
func getString(m *tengo.Map, key string) (string, error) {
	if r, ok := m.Value[key]; ok {
		var v string
		if v, ok = tengo.ToString(r); !ok {
			return "", fmt.Errorf("%s must be a string", key)
		} else {
			return v, nil
		}
	} else {
		return "", fmt.Errorf("key %s not found", key)
	}
}

func extractAddress(toRaw tengo.Object, msg *gomail.Message) ([]string, error) {
	var addrs []string
	switch t := toRaw.(type) {
	case *tengo.Array:
		for _, v := range t.Value {
			vv, ok := tengo.ToString(v)
			if !ok {
				return nil, errors.New("convert [to] field to string error")
			}
			if v, err := mail.ParseAddress(vv); err == nil {
				addrs = append(addrs, msg.FormatAddress(v.Address, v.Name))
			} else {
				return nil, err
			}
		}
	default:
		vv, ok := tengo.ToString(t)
		if !ok {
			return nil, errors.New("convert [to] field to string error")
		}
		if v, err := mail.ParseAddress(vv); err == nil {
			addrs = append(addrs, msg.FormatAddress(v.Address, v.Name))
		} else {
			return nil, err
		}
	}
	return addrs, nil
}

func (s *smtpClient) SendText(header map[string][]string, body string, attachments ...string) error {
	return s.Send(header, "text/plain", body, attachments...)
}

func (s *smtpClient) Close() error {
	return s.sender.Close()
}
func NewSMTPClient(host string, port int, account string, password string) (*smtpClient, error) {
	dialer := gomail.NewDialer(host, port, account, password)
	dialer.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	if sender, err := dialer.Dial(); err == nil {
		return &smtpClient{sender: sender}, nil
	} else {
		return nil, err
	}
}
