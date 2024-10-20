package cryptlib

import (
	"io/ioutil"
	"reflect"
	"testing"
)

func TestRsaEncrypt(t *testing.T) {
	pubKey, _ := ioutil.ReadFile("public2.pem")
	privKey, _ := ioutil.ReadFile("private2.pem")
	type args struct {
		origData   []byte
		publicKey  []byte
		privateKey []byte
	}
	tests := []struct {
		name    string
		args    args
		want    []byte
		wantErr bool
	}{
		{
			name: "key1",
			args: args{
				origData:   []byte("this is a orginal data"),
				publicKey:  pubKey,
				privateKey: privKey,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := RsaEncrypt(tt.args.origData, tt.args.publicKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("RsaEncrypt() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			ret, err := RsaDecrypt(got, tt.args.privateKey)
			if (err != nil) != tt.wantErr {
				t.Errorf("RsaDecrypt() error = %v,want Err %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(tt.args.origData, ret) {
				t.Errorf("RsaEncrypt() got = %v, want %v", got, ret)
			}
		})
	}
}

func TestCerts(t *testing.T) {

}
