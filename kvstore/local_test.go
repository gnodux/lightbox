package kvstore

import (
	"fmt"
	"github.com/dgraph-io/badger/v3"
	"os"
	"testing"
)

func TestOpenDefault2(t *testing.T) {
	for i := 0; i < 100; i++ {
		if db, err := OpenDefault(); err == nil {
			db.View(func(txn *badger.Txn) error {
				opt := badger.DefaultIteratorOptions
				opt.Prefix = []byte("k:1")
				opt.PrefetchValues = false
				iterator := txn.NewIterator(opt)
				defer iterator.Close()
				for iterator.Rewind(); iterator.Valid(); iterator.Next() {
					itm := iterator.Item()
					itm.Value(func(val []byte) error {
						fmt.Println(string(val))
						return nil
					})

				}
				return nil
			})
		} else {
			t.Error("open default db error", err)
		}
	}
}

func TestOpen(t *testing.T) {
	opt1 := DefaultOptions("n1")
	Open("n1", opt1)
	opt2 := DefaultOptions("n2")
	opt2 = opt2.WithMemTableSize(10 * 1024 * 1024)
	Open("n2", opt2)
}

func TestOpenDefault(t *testing.T) {
	db, err := OpenDefault()
	if err != nil {
		t.Error(err)
		t.FailNow()
	}
	defer func() {
		err := CloseDefault()
		if err != nil {
			t.Error("close error", err)
		}
	}()
	//db.Update(func(txn *badger.Txn) error {
	//	for i := 0; i < 10; i++ {
	//		for j := 0; j < 10; j++ {
	//			txn.Set([]byte(fmt.Sprintf("k:%d:%d", i, j)), []byte(fmt.Sprintf("v(%d,%d)", i, j)))
	//		}
	//	}
	//	return nil
	//})
	err = db.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.PrefetchValues = true
		opt.Prefix = []byte("k")
		iterator := txn.NewIterator(opt)
		defer iterator.Close()
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			fmt.Println(iterator.Item())
		}
		return nil
	})

	f, err := os.Create("./backup01.back")
	if err != nil {
		panic(err)
	}
	defer func(f *os.File) {
		err := f.Close()
		if err != nil {

		}
	}(f)
	db.Backup(f, 0)
	if err != nil {
		t.Error("database operation error", err)
		return
	}
}
