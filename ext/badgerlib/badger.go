package badgerlib

import (
	"bytes"
	"errors"
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib"
	"github.com/dgraph-io/badger/v3"
	"lightbox/ext/util"
	"lightbox/kvstore"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type StoreItem struct {
	Key   []byte
	Value []byte
	Meta  byte
}

type Query struct {
	Offset int
	Limit  int
}

// badgerClient badger client封装
// todo: 1. list_keys/list_keys_with_prefix/list_keys_with_key_word，应该都有配套的count方法
// todo: 2. search 统一入口
type badgerClient struct {
	*badger.DB
	tengo.ObjectImpl
	methodIndex map[string]tengo.Object
	sync.Once
	Name string
	Dir  string
}

func (b *badgerClient) IndexGet(key tengo.Object) (tengo.Object, error) {
	b.Do(b.init)
	propKey, ok := tengo.ToString(key)
	if !ok {
		return nil, fmt.Errorf("index %v not a string/stringer", key)
	}
	if b.methodIndex != nil {
		if m, ok := b.methodIndex[propKey]; ok {
			return m, nil
		}
	}
	return nil, fmt.Errorf("%s not exists", key)
}

func (b *badgerClient) TypeName() string {
	return "badger-client"
}
func (b *badgerClient) String() string {
	return fmt.Sprintf("badger_client<name:%s dir:%s>", b.Name, b.Dir)
}

func (b *badgerClient) SearchKey(matchFn func(filter, value []byte) bool, offset, limit int) ([]*StoreItem, error) {
	//todo: 没有想清楚怎么写，待定
	var items []*StoreItem
	b.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		opt.PrefetchValues = false
		opt.PrefetchSize = limit
		iterator := txn.NewIterator(opt)
		for iterator.Rewind(); iterator.Valid(); iterator.Next() {
			//itm := iterator.Item()
			//key := itm.Key()

		}
		return nil
	})
	return items, nil
}

// SearchPrefix snippet:name=badger.search;prefix=search;body=search(${1:prefix});
func (b *badgerClient) SearchPrefix(prefix string) (map[string][]byte, error) {
	values := map[string][]byte{}
	err := b.View(func(txn *badger.Txn) error {
		opt := badger.DefaultIteratorOptions
		p := []byte(prefix)
		opt.Prefix = p
		iterator := txn.NewIterator(opt)
		for iterator.Seek(p); iterator.ValidForPrefix(p); iterator.Next() {
			itm := iterator.Item()
			err := itm.Value(func(val []byte) error {
				values[string(itm.Key())] = val
				return nil
			})
			if err != nil {
				return err
			}
		}
		defer iterator.Close()
		return nil
	})
	return values, err
}

// Get
//
//snippet:name=badger.get;prefix=get;body=get(${1:key});
func (b *badgerClient) Get(key string) ([]byte, error) {
	var (
		data []byte
		err  error
	)
	err = b.View(func(txn *badger.Txn) error {
		var itm *badger.Item
		if itm, err = txn.Get([]byte(key)); err != nil {
			if err == badger.ErrKeyNotFound {
				return nil
			} else {
				return err
			}
		}
		err = itm.Value(func(val []byte) error {
			data = val
			return nil
		})
		return nil
	})
	return data, err
}

// BatchGet
//
//snippet:name=badger.get([string,string,...]);prefix=get;body=get([$1]);
func (b *badgerClient) BatchGet(keys []string) (map[string][]byte, error) {
	var (
		err  error
		data = map[string][]byte{}
	)
	err = b.View(func(txn *badger.Txn) error {

		for _, key := range keys {
			var itm *badger.Item
			itm, err := txn.Get([]byte(key))
			if err != nil {
				if err != badger.ErrKeyNotFound {
					return err
				} else {
					continue
				}

			}
			err = itm.Value(func(val []byte) error {
				data[key] = val
				return nil
			})
			if err != nil {
				return err
			}
		}
		return nil
	})
	return data, err
}

// Set
//
//snippet:name=badger.set;prefix=set;body=set({$1});
func (b *badgerClient) Set(data map[string][]byte) error {
	return b.Update(func(txn *badger.Txn) error {
		for k, v := range data {
			entry := badger.NewEntry([]byte(k), v)
			err := txn.SetEntry(entry)
			if err != nil {
				return err
			}
		}
		return nil
	})
}

// Delete
//
//snippet:name=badger.del;prefix=del;body=del([${1:strings}]);
func (b *badgerClient) Delete(keys []string) error {
	return b.Update(func(txn *badger.Txn) error {
		for _, k := range keys {
			if e := txn.Delete([]byte(k)); e != nil {
				return e
			}
		}
		return nil
	})
}

func (b *badgerClient) init() {
	b.methodIndex = map[string]tengo.Object{
		"list_keys_with_prefix": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) != 1 && len(args) != 3 {
					return tengo.FromInterface("arguments require(prefix string/bytes,[offset int=0,limit int=100])")
				}
				var (
					prefix []byte
					keys   []interface{}
					offset int = 0
					limit  int = 100
					ok     bool
				)

				prefix, ok = tengo.ToByteSlice(args[0])

				if len(args) == 3 {
					offset, ok = tengo.ToInt(args[1])
					if !ok {
						return tengo.FromInterface(errors.New("offset require a number"))
					}
					limit, ok = tengo.ToInt(args[2])
					if !ok {
						return tengo.FromInterface(errors.New("limit require a number"))
					}
				}
				err = b.View(func(txn *badger.Txn) error {
					opt := badger.DefaultIteratorOptions
					opt.PrefetchValues = false
					opt.Prefix = prefix
					it := txn.NewIterator(opt)
					defer it.Close()
					count := 0
					for it.Seek(prefix); it.ValidForPrefix(prefix); it.Next() {
						count += 1
						if count <= offset {
							continue
						}
						if count > limit {
							break
						}
						var key []byte = make([]byte, it.Item().KeySize())
						it.Item().KeyCopy(key)
						keys = append(keys, key)
					}
					return nil
				})
				if err != nil {
					return tengo.FromInterface(err)
				} else {
					return tengo.FromInterface(keys)
				}
			}},
		//snippet:name=badger.list_keys;prefix=list_keys;body=list_keys();desc=list keys(offset=0,limit=100);
		//snippet:name=badger.list_keys;prefix=list_keys;body=list_keys({$1:offset},{$2:limit});
		"list_keys": &tengo.UserFunction{
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				var (
					keys   []interface{}
					offset int = 0
					limit  int = 100
					ok     bool
				)
				if len(args) == 2 {
					offset, ok = tengo.ToInt(args[0])
					if !ok {
						return tengo.FromInterface(errors.New("offset require a number"))
					}
					limit, ok = tengo.ToInt(args[1])
					if !ok {
						return tengo.FromInterface(errors.New("limit require a number"))
					}
				}
				limit = offset + limit
				err = b.View(func(txn *badger.Txn) error {
					opt := badger.DefaultIteratorOptions
					opt.PrefetchValues = false
					it := txn.NewIterator(opt)
					defer it.Close()
					count := 0
					for it.Rewind(); it.Valid(); it.Next() {
						count += 1
						if count <= offset {
							continue
						}
						if count > limit {
							break
						}
						var key []byte = make([]byte, it.Item().KeySize())
						it.Item().KeyCopy(key)
						keys = append(keys, key)
					}
					return nil
				})
				if err != nil {
					return tengo.FromInterface(err)
				} else {
					return tengo.FromInterface(keys)
				}
			},
		},
		"del": &tengo.UserFunction{
			Name: "del",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				var keys []string
				for _, arg := range args {
					if k, ok := tengo.ToString(arg); !ok {
						return nil, &tengo.ErrInvalidArgumentType{Name: "key", Expected: "string", Found: arg.TypeName()}
					} else {
						keys = append(keys, k)
					}
				}
				if err := b.Delete(keys); err != nil {
					return tengo.FromInterface(err)
				}
				return nil, nil
			}},
		"backup": &tengo.UserFunction{
			Value: stdlib.FuncASI64RE(func(backupName string, since int64) error {
				if backupName == "" {
					return errors.New("backup file name is empty")
				}
				fi, err := os.Stat(backupName)
				if err == nil {
					if fi.IsDir() {
						backupName = filepath.Join(backupName, time.Now().Format("20060102150504.bbk"))
					} else {
						return fmt.Errorf("backup file %s already exists", fi.Name())
					}
				}
				create, err := os.Create(backupName)
				if err != nil {
					return err
				}
				_, err = b.Backup(create, uint64(since))
				return err
			}),
		},
		"get": &tengo.UserFunction{
			Name: "get",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				var keys []string
				for _, k := range args {
					switch kk := k.(type) {
					case *tengo.Array:
						for _, itm := range kk.Value {
							if v, ok := tengo.ToString(itm); ok {
								keys = append(keys, v)
							}
						}
					default:
						if v, ok := tengo.ToString(kk); ok {
							keys = append(keys, v)
						}
					}
				}
				if keys == nil || len(keys) == 0 {
					return &tengo.Array{Value: []tengo.Object{}}, nil
				}
				if len(keys) == 1 {
					bytes, err := b.Get(keys[0])
					if err != nil {
						return tengo.FromInterface(err)
					}
					return tengo.FromInterface(bytes)
				} else {
					data := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
					rowData, err := b.BatchGet(keys)
					if err != nil {
						return tengo.FromInterface(err)
					}
					for k, v := range rowData {
						data.Value[k] = &tengo.Bytes{Value: v}
					}
					return data, nil
				}
			}},

		//snippet:name=badger.set;prefix=set;body=set($1)
		"set": &tengo.UserFunction{
			Name: "set",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) == 0 {
					return tengo.FromInterface(tengo.ErrWrongNumArguments)
				}
				data := map[string][]byte{}
				if len(args) == 1 {
					//1个参数且参数是map，批量存储所有的Value
					kvs, ok := args[0].(*tengo.Map)
					if !ok {
						return tengo.FromInterface(&tengo.ErrInvalidArgumentType{Name: "set", Expected: "Map", Found: args[0].TypeName()})
					}
					for k, v := range kvs.Value {
						val, ok := tengo.ToByteSlice(v)
						if !ok {
							return tengo.FromInterface(errors.New("invalidate value type " + v.TypeName()))
						}
						data[k] = val
					}
					if err := b.Set(data); err != nil {
						return tengo.FromInterface(err)
					}
				} else if len(args) == 2 {
					//两个参数，就是Key/value
					kk, ok := tengo.ToString(args[0])
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{Name: "key", Expected: "string/bytes", Found: args[0].TypeName()}
					}
					vv, ok := tengo.ToByteSlice(args[1])
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{Name: "value", Expected: "string/bytes", Found: args[1].TypeName()}
					}
					data[kk] = vv
					if err := b.Set(data); err != nil {
						return tengo.FromInterface(err)
					}

				} else if len(args) == 3 {
					//3个参数，key/value/ttl(相对时间)
					kk, ok := tengo.ToString(args[0])
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{Name: "key", Expected: "string/bytes", Found: args[0].TypeName()}
					}
					vv, ok := tengo.ToByteSlice(args[1])
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{Name: "value", Expected: "string/bytes", Found: args[1].TypeName()}
					}
					ttl, ok := tengo.ToInt64(args[2])
					if !ok {
						return nil, tengo.ErrInvalidArgumentType{Name: "ttl", Expected: "int", Found: args[2].TypeName()}
					}
					return nil, b.DB.Update(func(txn *badger.Txn) error {
						entry := badger.NewEntry([]byte(kk), vv).WithTTL(time.Duration(ttl))
						return txn.SetEntry(entry)
					})
				} else {
					return nil, tengo.ErrWrongNumArguments
				}

				return nil, nil
			},
		},
		"search": &tengo.UserFunction{
			Name: "search",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) <= 1 {
					return nil, tengo.ErrWrongNumArguments
				}
				var v []byte
				arg := tengo.ToInterface(args[0])
				switch a := arg.(type) {
				case string:
					v = []byte(a)
				case []byte:
					v = a
				}

				result := map[string]tengo.Object{}
				err = b.DB.View(func(txn *badger.Txn) error {
					opt := badger.DefaultIteratorOptions
					opt.PrefetchValues = false
					it := txn.NewIterator(opt)
					defer it.Close()
					for it.Rewind(); it.Valid(); it.Next() {
						item := it.Item()
						if bytes.Contains(item.Key(), v) {
							result[string(item.Key())] = &tengo.Bytes{Value: item.Key()}
						}
					}
					return nil
				})
				if err != nil {
					return util.Error(err), nil
				}
				return &tengo.ImmutableMap{Value: result}, nil
			},
		},
		"search_prefix": &tengo.UserFunction{
			Name: "search_prefix",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) == 0 {
					return nil, tengo.ErrWrongNumArguments
				}
				var (
					prefix string
					ok     bool
				)
				if prefix, ok = tengo.ToString(args[0]); !ok {
					return nil, tengo.ErrInvalidArgumentType{Name: "prefix", Expected: "string", Found: args[0].TypeName()}
				}
				values, err := b.SearchPrefix(prefix)
				if err != nil {
					return tengo.FromInterface(err)
				}
				data := &tengo.ImmutableMap{Value: map[string]tengo.Object{}}
				if err != nil {
					return tengo.FromInterface(err)
				}
				for k, v := range values {
					data.Value[k] = &tengo.Bytes{Value: v}
				}
				return data, nil

			}},
	}
}

func newBadgerClient(db *badger.DB, name string, options badger.Options) *badgerClient {
	client := &badgerClient{
		DB:   db,
		Name: name,
		Dir:  options.Dir,
	}
	//client.init()
	return client
}
func wrapBadgerClient(db *badger.DB, name string) *badgerClient {
	return &badgerClient{
		DB:   db,
		Name: name,
		Dir:  name,
	}
}

func openBadger(args ...tengo.Object) (tengo.Object, error) {
	if len(args) == 0 || len(args) > 2 {
		return tengo.FromInterface(errors.New("arguments (name string,[name string/op option])"))
	}
	//无参数，打开默认的KV存储，路径在当前路径的data目录
	//if len(args) == 0 {
	//	if db, err := OpenDefaultBadgerDB(); err == nil {
	//		return db, nil
	//	} else {
	//		return tengo.FromInterface(err)
	//	}
	//}
	//打开一个指定名称的存储(之前已经打开过的,直接创建新的风险比较高,没想好怎么修改)
	if len(args) == 1 {
		name, ok := tengo.ToString(args[0])
		if !ok {
			return tengo.FromInterface(errors.New("name must be a  string"))
		}
		if db, err := OpenBadgerDB(name, kvstore.DefaultOptions(name)); err == nil {
			return db, nil
		} else {
			return tengo.FromInterface(err)
		}
	}
	//打开指定目录，指定名称的kv存储
	if len(args) == 2 {
		name, ok := tengo.ToString(args[0])
		if !ok {
			return tengo.FromInterface(errors.New("name must be a  string"))
		}
		var opt *badger.Options
		arg1 := tengo.ToInterface(args[1])

		switch v := arg1.(type) {
		case string:
			o := kvstore.DefaultOptions(v)
			opt = &o
		case *util.ReflectProxy:
			opt, ok = v.Self.(*badger.Options)
			if !ok {
				return util.Error(&tengo.ErrInvalidArgumentType{Name: "option", Expected: "badger.Option", Found: fmt.Sprintf("%v", v.Self)}), nil
			}
		case map[string]interface{}:
			o := badger.DefaultOptions(name)
			if err := util.StructFromObject(args[1], &o); err != nil {
				return util.Error(err), nil
			}
			o.ValueDir = o.Dir
			opt = &o
		default:
			return util.Error(&tengo.ErrInvalidArgumentType{Name: "option", Expected: "badger.Option", Found: fmt.Sprintf("%v", v)}), nil
		}
		if db, err := OpenBadgerDB(name, *opt); err == nil {
			return db, nil
		} else {
			return tengo.FromInterface(err)
		}
	}

	return tengo.FromInterface(errors.New("unknown arguments count"))
}

func OpenBadgerDB(name string, options badger.Options) (*badgerClient, error) {
	var (
		db  *badger.DB
		err error
	)

	if db, err = kvstore.Open(name, options); err != nil {
		return nil, err
	} else {
		return newBadgerClient(db, name, options), nil
	}

}

//func OpenDefaultBadgerDB() (*badgerClient, error) {
//	return OpenBadgerDB("DEFAULT", kvstore.DefaultOptions("kv-data"))
//}
