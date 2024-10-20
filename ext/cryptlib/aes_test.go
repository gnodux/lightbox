package cryptlib

import (
	"fmt"
	"reflect"
	"testing"
)

func TestAesEncrypt(t *testing.T) {
	type args struct {
		origData []byte
		key      []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				origData: []byte("this is a key"),
				key:      []byte("sizesizesizesize"),
			},
			want: []byte{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var (
				encrypted []byte
				decrypted []byte
				err       error
			)
			encrypted, err = AesEncrypt(tt.args.origData, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("AesEncrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			decrypted, err = AesDecrypt(encrypted, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("AesDecrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(decrypted, tt.args.origData) {
				fmt.Println(decrypted)
				t.Errorf("AesEncrypt() got = %v, want %v", decrypted, tt.want)
			}
		})
	}
}
func TestKeyPadding(t *testing.T) {
	start := uint8('0')
	raw := []byte{}
	for i := 0; i < 64; i++ {
		raw = append(raw, uint8(start+uint8(i)))
		fmt.Println(string(KeyPadding(raw)), len(KeyPadding(raw)))
	}
}
