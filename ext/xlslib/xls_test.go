package xlslib

import (
	"fmt"
	"github.com/xuri/excelize/v2"
	"testing"
)

func TestXlsFile(t *testing.T) {
	f := excelize.NewFile()
	f.Sheet.Range(func(key, value interface{}) bool {
		fmt.Println(key, ":", value)
		return true
	})

	f.Close()
}

func TestExcelHead(t *testing.T) {
	for i := 1; i < 1000; i++ {
		fmt.Println(FormatTitle(i))
	}
}

func TestEncrypt(t *testing.T){

}

