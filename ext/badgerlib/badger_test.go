package badgerlib

import (
	"encoding/json"
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"lightbox/kvstore"
	"os"
	"strings"
	"testing"
	"time"
)

func TestBadgerTTL(t *testing.T) {
	db, err := badger.Open(badger.DefaultOptions("testdata/managed"))
	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		if !db.IsClosed() {
			db.Close()
		}
	}()

	err = db.Update(func(txn *badger.Txn) error {
		entry := badger.NewEntry([]byte("k1"), []byte("value")).WithTTL(1 * time.Second).WithMeta(1)
		return txn.SetEntry(entry)
	})
	if err != nil {
		t.Fatal(err)
	}
	err = db.View(func(txn *badger.Txn) error {
		entry, err := txn.Get([]byte("k1"))
		if err != nil {
			return err
		}
		fmt.Println(entry.Version())
		return nil
	})
}

//
//func TestOpenDefaultBadgerDB(t *testing.T) {
//	ac, err := OpenDefaultBadgerDB()
//	bytes, err := ac.Get("k1")
//	if err != nil {
//		t.Error(err == badger.ErrKeyNotFound)
//		return
//	}
//	fmt.Println("data", string(bytes))
//}
func TestKVStore(t *testing.T) {
	kv, err := kvstore.OpenDefault()
	if err != nil {
		return
	}
	for i := 0; i < (1 << 30); i++ {
		err = kv.Update(func(txn *badger.Txn) error {
			return txn.Set([]byte(fmt.Sprintf("KEY_%d", i)), []byte(strings.Repeat("abc", 1000)))
		})
		if err != nil {
			t.Log(err)
			return
		}
	}
}

type Options struct {
	ValueLogSize int64
	MemTableSize int64
}

var DefaultOption = Options{
	ValueLogSize: 12,
	MemTableSize: 12,
}

func GetOption() Options {
	opt := DefaultOption
	opt.ValueLogSize = 100
	return opt
}

func TestOptions(t *testing.T) {
	fmt.Println(DefaultOption)
	fmt.Println(GetOption())
	fmt.Println(DefaultOption)
	DefaultOption.ValueLogSize = 15
	fmt.Println(DefaultOption)
}

type Animal struct {
	Name   string `json:"name"`
	Status string `json:"status"`
}
type Mammal struct {
	Animal
	MaxAge int
}
type Dog struct {
	Mammal
	Sound string
}

func TestJSONInherited(t *testing.T) {
	dog := &Dog{
		Sound: "wangwang",
	}
	dog.MaxAge = 12
	dog.Name = "dog"
	dog.Status = "available"
	encoder := json.NewEncoder(os.Stdout)
	encoder.Encode(dog)
}

func TestSearch(t *testing.T) {
	b, e := badger.Open(badger.DefaultOptions("").WithInMemory(true))
	if e != nil {
		t.Fatal(e)
	}

	b.Update(func(txn *badger.Txn) error {
		for i := 0; i < 100; i++ {
			txn.Set([]byte(fmt.Sprintf("k%d", i)), []byte(fmt.Sprintf("v%d", i)))
		}
		return nil
	})
	b.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.PrefetchValues = false
		it := txn.NewIterator(opt)
		defer it.Close()
		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()
			item.Value(func(val []byte) error {
				fmt.Println(string(val))
				return nil
			})
		}
		return nil
	})
}
