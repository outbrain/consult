package misc

import (
	"fmt"
	"reflect"
	"strings"
)

func PrettyPrint(data interface{}) {
	for _, s := range flatten(data) {
		fmt.Println(s)
	}
}

func flatten(o interface{}) []string {
	switch x := o.(type) {
	case map[string]string:
		res := make([]string, 0)
		for k, v := range x {
			res = append(res, k+"\t"+v)
		}
		return res
	case []string:
		return []string{strings.Join(x, ",")}
	case string:
		return []string{x}
	default:
		return flattenByType(x)
	}
}

func flattenByType(o interface{}) []string {
	res := make([]string, 0)
	x := reflect.ValueOf(o)
	t := x.Type()
	switch x.Kind() {
	case reflect.Map:
		for _, k := range x.MapKeys() {
			for _, v := range flatten(x.MapIndex(k).Interface()) {
				res = append(res, fmt.Sprintf("%v", k)+"\t"+v)
			}
		}
	case reflect.Array, reflect.Slice:
		for i := 0; i < x.Len(); i++ {
			res = append(res, flatten(x.Index(i).Interface())...)
		}
	case reflect.Struct:
		for i := 0; i < x.NumField(); i++ {
			f := x.Field(i)
			if f.CanInterface() {
				for _, v := range flatten(f.Interface()) {
					res = append(res, t.Field(i).Name+"\t"+v)
				}
			}
		}
	case reflect.Ptr:
		val := x.Elem()
		if val.IsValid() {
			return flattenByType(val.Interface())
		}
	default:
		return []string{fmt.Sprintf("%v", o)}
	}
	return res
}
