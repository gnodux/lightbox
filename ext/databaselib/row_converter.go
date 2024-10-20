package databaselib

import (
	"database/sql"
	"encoding/json"
	"github.com/golang/protobuf/protoc-gen-go/generator"
	log "github.com/sirupsen/logrus"
)

func bytesToString(input interface{}) interface{} {
	if input != nil {
		if b, ok := input.([]byte); ok {
			return string(b)
		}
	}
	return input
}
func bytesToMap(input interface{}) interface{} {
	if input != nil {
		if data, ok := input.([]byte); ok {
			m := map[string]interface{}{}
			if err := json.Unmarshal(data, &m); err == nil {
				return m
			} else {
				log.Info("unmarshal map error:", err)
			}
		}
	}
	return input
}

var columnHandler = map[string]func(interface{}) interface{}{
	"VARCHAR": bytesToString,
	"TEXT":    bytesToString,
	"JSON":    bytesToMap,
}

func sliceProcess(columnTypes []*sql.ColumnType, row []interface{}) {
	for idx, colType := range columnTypes {
		if h, ok := columnHandler[colType.DatabaseTypeName()]; ok {
			row[idx] = h(row[idx])
		}
	}
}
func mapRowProcess(columnTypes map[string]*sql.ColumnType, row map[string]interface{}, camelCase bool) map[string]interface{} {
	if camelCase {
		return mapRowCamelProcess(columnTypes, row)
	}
	for _, ct := range columnTypes {
		if h, ok := columnHandler[ct.DatabaseTypeName()]; ok {
			row[ct.Name()] = h(row[ct.Name()])
		}
	}
	return row
}
func mapRowCamelProcess(columnTypes map[string]*sql.ColumnType, row map[string]interface{}) map[string]interface{} {
	newRow := make(map[string]interface{}, len(row))
	for _, ct := range columnTypes {
		if h, ok := columnHandler[ct.DatabaseTypeName()]; ok {
			newRow[generator.CamelCase(ct.Name())] = h(row[ct.Name()])
		} else {
			newRow[generator.CamelCase(ct.Name())] = row[ct.Name()]
		}
	}
	return newRow
}
