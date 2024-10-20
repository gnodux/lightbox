package chanlib

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"github.com/d5/tengo/v2"
	"math/rand"
	"testing"
	"time"
)

func TestNewChannel(t *testing.T) {
	opt := ChanOption{Name: "1", Async: true, BufferSize: 12}
	ch, err := NewChannel(opt)
	if err == nil {
		fmt.Println(ch)
	}
	ch, err = NewChannel(opt)
	if err == nil {
		fmt.Println(ch)
	}
	ch.Write(&tengo.String{Value: "Test"})
	fmt.Println(ch.Read())
}

func TestSeed(t *testing.T) {
	rand.Seed(time.Now().UnixNano())
	rand.Intn(9999999)
	h := md5.Sum([]byte("测试密码"))
	fmt.Printf("%x\n", h)
	target := make([]byte, len(h)*2)
	n := hex.Encode(target, h[:])
	fmt.Println(n)
	fmt.Println(string(target[:n]))
}
