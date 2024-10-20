package maillib

import (
	"crypto/tls"
	"fmt"
	"gopkg.in/gomail.v2"
	"net/mail"
	"testing"
)

func TestSendHtmlMail(t *testing.T) {
	client := gomail.NewDialer("", 465, "", "")
	client.TLSConfig = &tls.Config{InsecureSkipVerify: true}
	sender, err := client.Dial()
	if err != nil {
		t.Error("dial error", err)
		return
	}
	defer sender.Close()
	msg := gomail.NewMessage()
	msg.SetHeaders(map[string][]string{
		"From":    {""},
		"To":      {""},
		"Subject": {"测试邮件"},
	})
	msg.SetBody("text/html", "<a href=\"www.baidu.com\">baidu</a>")
	msg.Attach("mail.go")
	err = gomail.Send(sender, msg)
	if err != nil {
		t.Error("send mail error", err)
		return
	}
}
func TestAddress(t *testing.T) {
	a, _ := mail.ParseAddress(".com")
	fmt.Println(a.Name)
	addr := &mail.Address{Address: "", Name: ""}
	fmt.Println(addr.String())
}
func TestSMTPClient_SendText(t *testing.T) {
	addr := &mail.Address{Address: "", Name: ""}
	client, err := NewSMTPClient("", 465, "m", "")
	if err != nil {
		t.Error("create client error")
		return
	}
	err = client.SendText(map[string][]string{
		"To":       {"aa@qq.com", ""},
		"Subject":  {"简单的文本测试"},
		"From":     {addr.String()},
		"Reply-To": {"xx@xx.com"},
	}, "简单文本邮件测试")
	if err != nil {
		t.Error("send mail error", err)
		return
	}

}
