package misc

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

var SEPARATOR = "\t"

func StructToString(o interface{}) string {
	x := reflect.ValueOf(o)
	if x.Kind() == reflect.Ptr {
		ref := x.Elem()
		if ref.IsValid() {
			return StructToString(ref.Interface())
		} else {
			panic("Null pointer")
		}
	}

	res := make([]string, x.NumField())

	for i := 0; i < x.NumField(); i++ {
		f := x.Field(i)
		switch f.Kind() {
		case reflect.Array, reflect.Slice:
			s := make([]string, f.Len())
			for j := 0; j < f.Len(); j++ {
				s[j] = fmt.Sprintf("%v", f.Index(j))
			}
			res[i] = sanitize(strings.Join(s, ","))
		case reflect.Struct:
			res[i] = StructToString(f.Interface())
		case reflect.Ptr:
			res[i] = StructToString(f.Elem().Interface())
		default:
			res[i] = sanitize(fmt.Sprintf("%v", f.Interface()))
		}
	}
	return strings.Join(res, "\t")
}

func sanitize(s string) string {
	return strconv.Quote(s)
}

func StructHeaderLine(o interface{}) string {
	fieldNames := make([]string, 0)
	for _, s := range structHeaderLineHelper(make([]string, 0), o) {
		fieldNames = append(fieldNames, strings.Join(s, "."))
	}
	return strings.Join(fieldNames, SEPARATOR)
}

func structHeaderLineHelper(pfx []string, o interface{}) [][]string {
	t := reflect.TypeOf(o)
	v := reflect.ValueOf(o)
	res := make([][]string, 0, t.NumField())

	for i := 0; i < t.NumField(); i++ {
		ft := t.Field(i)
		fieldNames := append(pfx, ft.Name)
		switch t.Field(i).Type.Kind() {
		case reflect.Struct:
			for _, s := range structHeaderLineHelper(fieldNames, v.Field(i).Interface()) {
				res = append(res, s)
			}
		case reflect.Ptr:
			var fv interface{}
			if ref := v.Field(i).Elem(); ref.IsValid() && ref.CanInterface() {
				fv = ref.Interface()
			} else {
				fv = reflect.New(ft.Type.Elem()).Elem().Interface()
			}
			for _, s := range structHeaderLineHelper(fieldNames, fv) {
				if len(s) > 0 {
					res = append(res, s)
				}
			}
		default:
			res = append(res, fieldNames)
		}
	}
	return res
}

func JoinWithSep(items ...string) string {
	return strings.Join(items, SEPARATOR)
}
