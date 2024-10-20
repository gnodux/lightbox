package canallib

import "github.com/go-mysql-org/go-mysql/canal"

type HandlerConfig struct {
	OnRow          string `json:"on_row" yaml:"on_row" toml:"on_row"`
	OnDDL          string `json:"on_ddl" yaml:"on_ddl" toml:"on_ddl"`
	OnTableChanged string `json:"on_table_changed" yaml:"on_table_changed" toml:"on_table_changed"`
	OnRotate       string `json:"on_rotate" yaml:"on_rotate" toml:"on_rotate"`
	IgnoreError    bool   `json:"ignore_error" yaml:"ignore_error" toml:"ignore_error""`
}

type CanalConfig struct {
	*canal.Config
	Handler HandlerConfig
}
