package utils

import (
	"errors"
	"reflect"
)

func StructToMap(obj interface{}) map[string]interface{} {
	result := make(map[string]interface{})

	v := reflect.ValueOf(obj)
	// 如果是指针，获取指向的值
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		// 处理嵌套结构体（匿名字段）
		if field.Anonymous {
			// 递归处理嵌套结构体
			nestedMap := StructToMap(fieldValue.Interface())
			for k, v := range nestedMap {
				result[k] = v
			}
			continue
		}

		// 获取 JSON tag，如果不存在则使用字段名
		jsonTag := field.Tag.Get("json")
		if jsonTag == "" {
			jsonTag = field.Name
		}

		// 忽略 "-" 标签
		if jsonTag == "-" {
			continue
		}

		// 获取字段值
		if fieldValue.CanInterface() {
			result[jsonTag] = fieldValue.Interface()
		}
	}

	return result
}

func srcFilter(src interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(src)
	if v.Type().Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return reflect.Zero(v.Type()), errors.New("src type error: not a struct or a pointer to struct")
	}
	return v, nil
}

func dstFilter(src interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(src)
	if v.Type().Kind() != reflect.Ptr {
		return reflect.Zero(v.Type()), errors.New("src type error: not a pointer to struct")
	}
	if v.Elem().Kind() != reflect.Struct {
		return reflect.Zero(v.Type()), errors.New("src type error: not point to struct")
	}
	return v, nil
}
