package env

import (
	"fmt"
	"github.com/cookieY/sqlx"
	_ "github.com/go-sql-driver/mysql"
	"testing"
)

func TestNewCache(t *testing.T) {
	c := NewCache(func(opt string) (string, error) {
		return opt + ":?opt", nil
	})
	r, err := c.Get("a", "opt1")
	fmt.Println(r, err)
}
func TestNewCache2(T *testing.T) {
	c := NewCache(func(opt string) (*sqlx.DB, error) {
		if opt == "" {
			return nil, nil
		}
		return sqlx.Connect("mysql", opt)
	})

	db, err := c.Get("local", "xxtest:xxtest@tcp(127.0.0.1:3306)/metadata")
	fmt.Println(db, err)
	db, err = c.Get("local", "")
	fmt.Println(db, err)
}
