package cryptlib

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func Test_sumFileMD5(t *testing.T) {
	const other = "2de94b672434c7c29c77b4d0f574648f"
	b, err := sumFileMD5("./public.pem")
	if err == nil {
		fmt.Printf("%x\n", b)
	} else {
		fmt.Println(err)
	}
	assert.Equal(t, other, fmt.Sprintf("%x", b))
}

func TestSha1(t *testing.T) {

}
