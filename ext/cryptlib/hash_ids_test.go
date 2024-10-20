package cryptlib

import (
	"encoding/binary"
	"fmt"
	"github.com/d5/tengo/v2"
	uuid "github.com/satori/go.uuid"
	"github.com/speps/go-hashids"
	"lightbox/sandbox"
	"strconv"
	"testing"
	"time"
)

func IdToNumbers(id string) []int64 {
	var buf string
	var numbers []int64
	for _, c := range id {
		buf += strconv.Itoa(int(c))
		if len(buf) >= 12 {
			b, _ := strconv.ParseInt(buf, 10, 0)
			numbers = append(numbers, b)
			buf = ""
		}
	}
	if len(buf) > 0 {
		b, _ := strconv.ParseInt(buf, 10, 0)
		numbers = append(numbers, b)
	}
	return numbers
}

func Code(id string, expires int64) (string, error) {
	data := hashids.NewData()
	data.Salt = "e3ea1cd0-00af-425a-af0f-9dd848d39df7"
	h, err := hashids.NewWithData(data)
	if err != nil {
		return "", err
	}
	numbers := IdToNumbers(id)
	numbers = append(numbers, time.Now().Unix(), expires)
	return h.EncodeInt64(numbers)
}

func Test1(t *testing.T) {
	fmt.Println(Code("12232123", 3600))
}

func TestSeq(t *testing.T) {
	n := "0123456789abcdef"
	for _, c := range n {
		fmt.Println(c)
	}
}

func TestUUIDToNumbers(t *testing.T) {
	uuidStr := "6051fb4fecb646a39607888935a0b22c"
	numbers := IdToNumbers(uuidStr)
	numbers = append(numbers, time.Now().Unix())
	numbers = append(numbers, time.Now().Add(3*24*time.Hour).Unix())
	fmt.Println(numbers)
	data := hashids.NewData()
	data.Salt = "e3ea1cd0-00af-425a-af0f-9dd848d39df7"
	h, err := hashids.NewWithData(data)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	code, err := h.EncodeInt64(numbers)
	if err == nil {
		fmt.Println(code)
	}
}

func TestDecodeHashIds(t *testing.T) {
	code := "7Q3b8Xda4i0KwqKOHD3aB69FM5E"
	id := "14909203039"
	fmt.Println(IdToNumbers(id))

	//for _, c := range id {
	//	fmt.Println(c, ":", int(c))
	//}
	//uid, err := uuidlib.FromString(id)
	//uidBytes := uid.Bytes()
	//n1 := binary.BigEndian.Uint64(uidBytes[0:8])
	//n2 := binary.BigEndian.Uint64(uidBytes[8:])
	//fmt.Println(n1, n2)
	//if err == nil {
	//	fmt.Println(uid)
	//}
	data := hashids.NewData()
	data.Salt = "e3ea1cd0-00af-425a-af0f-9dd848d39df7"
	h, err := hashids.NewWithData(data)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}

	fmt.Println(h.DecodeInt64(code))
}

func TestHashIds(t *testing.T) {
	data := hashids.NewData()
	data.Salt = "e3ea1cd0-00af-425a-af0f-9dd848d39df7"
	//myuuid := "6051fb4f-ecb6-46a3-9607-888935a0b22c"
	myuuid, err := uuid.FromString("6051fb4f-ecb6-46a3-9607-888935a0b22c")
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	buf := myuuid.Bytes()
	n1 := binary.BigEndian.Uint64(buf[0:8])
	n2 := binary.BigEndian.Uint64(buf[8:])
	//n3 := binary.BigEndian.Uint64(buf[12:])
	fmt.Println(n1, n2)
	h, err := hashids.NewWithData(data)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}

	encodeInt64, err := h.EncodeInt64([]int64{int64(n1), int64(n2), time.Now().Unix(), time.Now().Add(24 * time.Hour).Unix()})
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	fmt.Println(encodeInt64)
	withError, err := h.DecodeInt64WithError(encodeInt64)
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	id := uuid.NewV4()
	fmt.Println(len(id.Bytes()))
	fmt.Println(id)
	fmt.Println(withError)
	fmt.Println(time.Unix(withError[1], 0), time.Unix(withError[2], 0))
}

func TestFuncEntry(t *testing.T) {
	app, _ := sandbox.NewWithDir("DEFAULT", ".")
	fmt.Println(Entry.Name(), Entry.AllNames(), Entry.GetModule(app))
}

func TestHashID_int(t *testing.T) {
	data := hashids.NewData()
	data.Salt = "e3ea1cd0-00af-425a-af0f-9dd848d39df7"
	h, err := hashids.NewWithData(data)
	if err != nil {
		fmt.Println(err)
		t.FailNow()
	}
	c, _ := h.Encode([]int{1, 2, 3, 5, 6})
	fmt.Println(c)
}

func UserName(app *sandbox.Applet, args ...tengo.Object) (tengo.Object, error) {
	return tengo.FromInterface("my user name")
}
