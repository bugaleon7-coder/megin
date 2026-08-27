package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strconv"
	"time"
)

func GetInt64(ctx context.Context, store BasicStore, key string) (int64, error) {
	raw, err := store.Get(ctx, key)
	if err != nil {
		return 0, err
	}
	value, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, err
	}
	return value, nil
}

func SetStruct[T any](store BasicStore, key string, value T, ttl time.Duration) error {
	if store == nil {
		return errors.New("cache store is nil")
	}
	if localStore, ok := store.(*LocalStore); ok {
		return localStore.setValue(key, value, ttl)
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return store.Set(context.Background(), key, string(raw), ttl)
}

func SetStructList[T any](store BasicStore, key string, values []T, ttl time.Duration) error {
	if store == nil {
		return errors.New("cache store is nil")
	}
	if localStore, ok := store.(*LocalStore); ok {
		return localStore.setValue(key, values, ttl)
	}
	raw, err := json.Marshal(values)
	if err != nil {
		return err
	}
	return store.Set(context.Background(), key, string(raw), ttl)
}

func SetString(store BasicStore, key string, value string, ttl time.Duration) error {
	if store == nil {
		return errors.New("cache store is nil")
	}
	return store.Set(context.Background(), key, value, ttl)
}

func GetStruct[T any](store BasicStore, key string) (*T, error) {
	if store == nil {
		return nil, errors.New("cache store is nil")
	}
	if localStore, ok := store.(*LocalStore); ok {
		return getLocalStruct[T](localStore, key)
	}
	value, err := store.Get(context.Background(), key)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			return nil, nil
		}
		return nil, err
	}
	var out T
	if err = decodeInto(value, &out); err != nil {
		return nil, err
	}
	return &out, nil
}

func GetStructList[T any](store BasicStore, key string) ([]T, error) {
	if store == nil {
		return nil, errors.New("cache store is nil")
	}
	if localStore, ok := store.(*LocalStore); ok {
		return getLocalStructList[T](localStore, key)
	}
	value, err := store.Get(context.Background(), key)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			return nil, nil
		}
		return nil, err
	}
	var out []T
	if err = decodeInto(value, &out); err != nil {
		return nil, err
	}
	return out, nil
}

func GetString(store BasicStore, key string) (string, error) {
	if store == nil {
		return "", errors.New("cache store is nil")
	}
	value, err := store.Get(context.Background(), key)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			return "", nil
		}
		return "", err
	}
	return value, nil
}

func DeleteKey(store BasicStore, key string) bool {
	if store == nil {
		return false
	}
	return store.Delete(context.Background(), key) == nil
}

func decodeInto(raw string, target any) error {
	return json.Unmarshal([]byte(raw), target)
}

func getLocalStruct[T any](store *LocalStore, key string) (*T, error) {
	value, err := store.getValue(key)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			return nil, nil
		}
		return nil, err
	}

	if typed, ok := value.(T); ok {
		out := typed
		return &out, nil
	}

	if typedPtr, ok := value.(*T); ok {
		if typedPtr == nil {
			return nil, nil
		}
		out := *typedPtr
		return &out, nil
	}

	var out T
	if err = assignStoredValue(&out, value); err != nil {
		return nil, err
	}
	return &out, nil
}

func getLocalStructList[T any](store *LocalStore, key string) ([]T, error) {
	value, err := store.getValue(key)
	if err != nil {
		if errors.Is(err, ErrCacheMiss) {
			return nil, nil
		}
		return nil, err
	}

	if typed, ok := value.([]T); ok {
		return typed, nil
	}

	if typedPtr, ok := value.(*[]T); ok {
		if typedPtr == nil {
			return nil, nil
		}
		return *typedPtr, nil
	}

	var out []T
	if err = assignStoredValue(&out, value); err != nil {
		return nil, err
	}
	return out, nil
}

func assignStoredValue(target any, value any) error {
	if target == nil {
		return errors.New("cache target is nil")
	}

	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Ptr || targetValue.IsNil() {
		return errors.New("cache target must be a non-nil pointer")
	}

	return assignReflectValue(targetValue.Elem(), reflect.ValueOf(value))
}

func assignReflectValue(target reflect.Value, source reflect.Value) error {
	if !source.IsValid() {
		target.SetZero()
		return nil
	}

	if source.Kind() == reflect.Ptr {
		if source.IsNil() {
			target.SetZero()
			return nil
		}
		return assignReflectValue(target, source.Elem())
	}

	if target.Kind() == reflect.Ptr {
		value := reflect.New(target.Type().Elem())
		if err := assignReflectValue(value.Elem(), source); err != nil {
			return err
		}
		target.Set(value)
		return nil
	}

	if source.Type().AssignableTo(target.Type()) {
		target.Set(source)
		return nil
	}

	if source.Type().ConvertibleTo(target.Type()) {
		target.Set(source.Convert(target.Type()))
		return nil
	}

	return fmt.Errorf("cached value type %s cannot be assigned to %s", source.Type(), target.Type())
}
