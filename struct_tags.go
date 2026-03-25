package maml

import (
	"reflect"
	"strings"
	"sync"
)

type fieldInfo struct {
	name      string // Go field name
	mamlName  string // key name in MAML
	omitEmpty bool
	ignore    bool
	index     []int // field index for reflect
}

type structInfo struct {
	fields []fieldInfo
}

var structCache sync.Map // map[reflect.Type]*structInfo

func getStructInfo(t reflect.Type) *structInfo {
	if cached, ok := structCache.Load(t); ok {
		return cached.(*structInfo)
	}
	info := buildStructInfo(t, nil)
	structCache.Store(t, info)
	return info
}

func buildStructInfo(t reflect.Type, parentIndex []int) *structInfo {
	info := &structInfo{}
	for i := 0; i < t.NumField(); i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}

		index := make([]int, len(parentIndex)+1)
		copy(index, parentIndex)
		index[len(parentIndex)] = i

		// Handle embedded structs
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			embedded := buildStructInfo(f.Type, index)
			info.fields = append(info.fields, embedded.fields...)
			continue
		}

		fi := fieldInfo{
			name:  f.Name,
			index: index,
		}

		tag := f.Tag.Get("maml")
		if tag == "" {
			tag = f.Tag.Get("json")
		}

		if tag == "-" {
			fi.ignore = true
			fi.mamlName = f.Name
			info.fields = append(info.fields, fi)
			continue
		}

		if tag != "" {
			parts := strings.Split(tag, ",")
			if parts[0] != "" {
				fi.mamlName = parts[0]
			} else {
				fi.mamlName = strings.ToLower(f.Name[:1]) + f.Name[1:]
			}
			for _, opt := range parts[1:] {
				if opt == "omitempty" {
					fi.omitEmpty = true
				}
			}
		} else {
			fi.mamlName = strings.ToLower(f.Name[:1]) + f.Name[1:]
		}

		info.fields = append(info.fields, fi)
	}
	return info
}
