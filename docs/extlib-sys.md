# Module - "sys"

```golang
sys := import ("sys")
```

## Functions

- `sys.prop(name)=>string`: 获取系统配置，例如：`sys.prop("os.name")`
- `sys.env(name)=>string`: 获取系统环境变量，例如：`sys.env("HOME")`
- `sys.config()=>map[string]object`: 获取系统配置，例如：`sys.config()`

