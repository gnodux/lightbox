package xlslib

import (
	"fmt"
	"github.com/d5/tengo/v2"
	"github.com/d5/tengo/v2/stdlib/json"
	"github.com/xuri/excelize/v2"
	"lightbox/ext/util"
)

const ExcelTitleMap = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

func FormatTitle(num int) string {
	n := uint64(num)
	var chars []byte
	for {
		u := (n - 1) / 26
		c := ExcelTitleMap[n-u*26-1]
		chars = append([]byte{c}, chars...)
		n = u
		if u == 0 {
			break
		}
	}
	return string(chars)
}

type XlsFile struct {
	*excelize.File
	tengo.ObjectImpl
	currentSheet string
	currentRow   int
	methodMap    map[string]*tengo.UserFunction
}

func (x *XlsFile) init() {
	x.methodMap = map[string]*tengo.UserFunction{
		"new_sheet": {
			Name: "add_sheet",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) != 1 {
					return util.Error(tengo.ErrWrongNumArguments), nil
				}
				sheetName, ok := tengo.ToString(args[0])
				if !ok {
					return util.Error(&tengo.ErrInvalidArgumentType{Name: "sheetName", Expected: "string", Found: args[0].TypeName()}), nil
				}
				num := x.NewSheet(sheetName)
				x.currentSheet = sheetName
				return tengo.FromInterface(num)
			},
		},
		"save": {
			Name: "save",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 1 {
					return util.Error(tengo.ErrWrongNumArguments), nil
				}
				fileName, ok := tengo.ToString(args[0])
				if !ok {
					return util.Error(&tengo.ErrInvalidArgumentType{Name: "fileName", Expected: "string", Found: args[0].TypeName()}), nil
				}
				err := x.SaveAs(fileName)
				if err != nil {
					return util.Error(err), nil
				}
				return nil, nil
			},
		},
		"set_cell_value": {
			Name:  "set_cell_value",
			Value: x.setCellValue,
		},
		"get_cell_value": {
			Name:  "get_cell_value",
			Value: x.getCellValue,
		},
		"set_cell": {
			Name:  "set_cell",
			Value: x.setCell,
		},
		"get_rows": {
			Name: "get_rows",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) != 1 {
					return util.Error(tengo.ErrWrongNumArguments), nil
				}
				sheet, ok := tengo.ToString(args[0])
				if !ok {
					if num, ok := tengo.ToInt(args[0]); !ok {
						return util.Error(&tengo.ErrInvalidArgumentType{
							Name:     "sheet name",
							Expected: "string/number",
							Found:    args[0].TypeName(),
						}), nil
					} else {
						sheet = x.GetSheetName(num)
					}
				}
				var rows [][]string
				rows, err = x.GetRows(sheet)
				if err != nil {
					return util.Error(err), nil
				}
				var results []tengo.Object
				for _, row := range rows {
					var rowObj []tengo.Object
					for _, col := range row {
						rowObj = append(rowObj, &tengo.String{Value: col})
					}
					results = append(results, &tengo.ImmutableArray{Value: rowObj})
				}
				return &tengo.ImmutableArray{Value: results}, nil
			},
		},
		"active_sheet": {
			Name: "active_sheet",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) == 0 {
					return &tengo.Int{Value: int64(x.GetActiveSheetIndex())}, nil
				}
				n, ok := tengo.ToInt(args[0])
				if !ok {
					return tengo.UndefinedValue, nil
				}
				x.SetActiveSheet(n)
				x.currentSheet = x.GetSheetName(n)
				x.currentRow = 0
				return nil, nil
			},
		},
		"append": {
			Name: "append",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) == 0 {
					return nil, nil
				}
				x.currentRow += 1
				var data = args[0]
				if len(args) > 1 {
					data = &tengo.ImmutableArray{Value: args}
				}
				switch value := data.(type) {
				case *tengo.Array:
					for idx, v := range value.Value {
						var realValue interface{}
						switch v.(type) {
						case *tengo.Map, *tengo.ImmutableMap, *tengo.Array, *tengo.ImmutableArray:
							if realValue, err = json.Encode(v); err != nil {
								return util.Error(err), nil
							}
						default:
							realValue = tengo.ToInterface(v)
						}
						err = x.SetCellValue(x.currentSheet, fmt.Sprintf("%s%d", FormatTitle(idx+1), x.currentRow), realValue)
						if err != nil {
							return util.Error(err), nil
						}
					}
				case *tengo.ImmutableArray:
					for idx, v := range value.Value {
						var realValue interface{}
						switch v.(type) {
						case *tengo.Map, *tengo.ImmutableMap, *tengo.Array, *tengo.ImmutableArray:
							if realValue, err = json.Encode(v); err != nil {
								return util.Error(err), nil
							}
						default:
							realValue = tengo.ToInterface(v)
						}
						err = x.SetCellValue(x.currentSheet, fmt.Sprintf("%s%d", FormatTitle(idx+1), x.currentRow), realValue)
						if err != nil {
							return util.Error(err), nil
						}
					}
				default:
					return nil, fmt.Errorf("unsupport type:%s", args[1].TypeName())
				}
				return
			},
		},
		"sheets": {
			Name: "sheets",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				sheets := x.GetSheetList()
				var sheetObjs []tengo.Object
				for _, s := range sheets {
					sheetObjs = append(sheetObjs, &tengo.String{Value: s})
				}
				return &tengo.ImmutableArray{Value: sheetObjs}, nil
			},
		},
		"new_style": {
			Name: "new_style",
			Value: func(args ...tengo.Object) (tengo.Object, error) {
				if len(args) != 1 {
					return util.Error(tengo.ErrWrongNumArguments), nil
				}
				switch v := args[0].(type) {
				case *tengo.String:
					style, err := x.NewStyle(v.Value)
					if err != nil {
						return nil, err
					}
					return tengo.FromInterface(style)
				case *tengo.Bytes:
					style, err := x.NewStyle(v.Value)
					if err != nil {
						return nil, err
					}
					return tengo.FromInterface(style)
				case *tengo.Map, *tengo.ImmutableMap:
					encode, err := json.Encode(v)
					if err != nil {
						return nil, err
					}
					style, err := x.NewStyle(encode)
					if err != nil {
						return nil, err
					}
					return tengo.FromInterface(style)
				}
				return tengo.FromInterface(-1)
			},
		},
		"set_style": {
			Name: "set_style",
			Value: func(args ...tengo.Object) (ret tengo.Object, err error) {
				if len(args) != 4 {
					ret = util.Error(err)
					return
				}
				sheet, ok := tengo.ToString(args[0])
				if !ok {
					ret = util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "sheet",
						Expected: "string",
						Found:    args[0].TypeName(),
					})
				}
				start, ok := tengo.ToString(args[1])
				if !ok {
					ret = util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "start",
						Expected: "string",
						Found:    args[1].TypeName(),
					})
				}
				end, ok := tengo.ToString(args[2])
				if !ok {
					ret = util.Error(&tengo.ErrInvalidArgumentType{
						Name:     "end",
						Expected: "string",
						Found:    args[2].TypeName(),
					})
				}
				switch style := args[3].(type) {
				case *tengo.Int:
					err = x.SetCellStyle(sheet, start, end, int(style.Value))
					return
				case *tengo.String, *tengo.Bytes:
					slice, _ := tengo.ToByteSlice(style)
					var styleId int
					styleId, err = x.NewStyle(slice)
					if err != nil {
						return nil, err
					}
					err = x.SetCellStyle(sheet, start, end, styleId)
				case *tengo.Map, *tengo.ImmutableMap:
					var (
						data    []byte
						styleId int
					)

					data, err = json.Encode(style)
					if err != nil {
						return
					}
					styleId, err = x.NewStyle(data)
					if err != nil {
						return
					}
					err = x.SetCellStyle(sheet, start, end, styleId)
				}
				return
			},
		},
	}
}
func (x *XlsFile) setCell(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 2 {
		return util.Error(tengo.ErrWrongNumArguments), nil
	}
	sheet, ok := tengo.ToString(args[0])
	if !ok {
		return util.Error(&tengo.ErrInvalidArgumentType{
			Name:     "sheet name",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}
	switch m := args[1].(type) {
	case *tengo.Map:
		for k, v := range m.Value {
			if err := x.SetCellValue(sheet, k, tengo.ToInterface(v)); err != nil {
				return util.Error(err), nil
			}
		}
	case *tengo.ImmutableMap:
		for k, v := range m.Value {
			if err := x.SetCellValue(sheet, k, tengo.ToInterface(v)); err != nil {
				return util.Error(err), nil
			}
		}
	}
	return nil, nil
}
func (x *XlsFile) setCellValue(args ...tengo.Object) (tengo.Object, error) {
	if len(args) != 3 {
		return util.Error(tengo.ErrWrongNumArguments), nil
	}
	sheet, ok := tengo.ToString(args[0])
	if !ok {
		return util.Error(&tengo.ErrInvalidArgumentType{
			Name:     "sheet name",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}
	axis, ok := tengo.ToString(args[1])
	if !ok {
		return util.Error(&tengo.ErrInvalidArgumentType{
			Name:     "axis",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}
	intf := tengo.ToInterface(args[2])
	err := x.SetCellValue(sheet, axis, intf)

	if err != nil {
		return util.Error(err), nil
	}
	return nil, nil
}
func (x *XlsFile) TypeName() string {
	return "xls-file"
}

func (x *XlsFile) IndexGet(key tengo.Object) (tengo.Object, error) {
	s, _ := tengo.ToString(key)
	if x.methodMap != nil {
		if m, ok := x.methodMap[s]; ok {
			return m, nil
		}
	}
	return util.Error(fmt.Errorf("method/property %s not found", s)), nil
}

func (x *XlsFile) getCellValue(args ...tengo.Object) (ret tengo.Object, err error) {
	if len(args) < 2 {
		return util.Error(tengo.ErrWrongNumArguments), nil
	}
	sheet, ok := tengo.ToString(args[0])
	if !ok {
		return util.Error(&tengo.ErrInvalidArgumentType{
			Name:     "sheet name",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}
	axis, ok := tengo.ToString(args[1])
	if !ok {
		return util.Error(&tengo.ErrInvalidArgumentType{
			Name:     "axis",
			Expected: "string",
			Found:    args[0].TypeName(),
		}), nil
	}
	var value string
	value, err = x.File.GetCellValue(sheet, axis)
	if err != nil {
		return util.Error(err), nil
	}
	return tengo.FromInterface(value)
}

func (x *XlsFile) String() string {
	return x.File.Path
}

func open(fileName string) (tengo.Object, error) {
	file, err := excelize.OpenFile(fileName)
	if err != nil {
		return nil, err
	}
	xlsf := &XlsFile{File: file}
	xlsf.init()
	return xlsf, nil
}

func newXls() tengo.Object {
	file := excelize.NewFile()
	xlsFile := &XlsFile{File: file}
	xlsFile.init()
	return xlsFile
}
