package misc

import (
	"fmt"
	"reflect"
	"strings"
)

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
			res[i] = strings.Join(s, ",")
		default:
			res[i] = fmt.Sprintf("%v", f.Interface())
		}
	}
	return strings.Join(res, "\t")
}

func StructHeaderLine(o interface{}) string {
	t := reflect.TypeOf(o)
	res := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		res[i] = t.Field(i).Name
	}
	return strings.Join(res, "\t")
}
