package modman

import (
	"fmt"
	"lightbox/env"
	"testing"
)

type getter func(string) (any, bool)

func Test_getter(t *testing.T) {
	var g getter

	fmt.Println(g("profile"))
}

func Test_parseEnv(t *testing.T) {
	env.Set("profile", "test")
	type args struct {
		src string
	}
	tests := []struct {
		name    string
		args    args
		want    string
		wantErr bool
	}{
		{
			name: "test1",
			args: args{
				src: "{profile}",
			},
			want: "test",
		},
		{
			name: "test2",
			args: args{
				src: "db_{profile}.tengo",
			},
			want: "db_test.tengo",
		}, {
			name: "test3",
			args: args{
				src: "db_{profile}_{tenant}.tengo",
			},
			want: "db_test_!tenant.tengo",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := env.Parse(tt.args.src, func(a any) (string, bool) {
				return env.Get[string](a)
			})
			if (err != nil) != tt.wantErr {
				t.Errorf("parseEnv() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("parseEnv() got = %v, want %v", got, tt.want)
			}
		})
	}
}
