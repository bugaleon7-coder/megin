package utils

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"log"
	"math/rand"
	"os"
	"runtime/debug"
)

func IsFileExist(filePath string) bool {
	stat, err := os.Stat(filePath)
	if os.IsNotExist(err) {
		return false
	}

	return stat.Mode().IsRegular()
}

func IsDirExist(dirPath string) bool {
	stat, err := os.Stat(dirPath)
	return err == nil && stat.IsDir()
}

func Md5(data []byte) (string, error) {
	hash := md5.New()
	_, err := hash.Write(data)
	if err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func SafeGoroutine(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic in goroutine:", r)
				debug.PrintStack()
			}
		}()
		fn()
	}()
}

func SafeGo(ctx context.Context, fn func(ctx context.Context)) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic in goroutine:", r, string(debug.Stack()))
			}
		}()
		fn(ctx)
	}()
}

// 随机数字字符串
func RandomNumbers(length int) string {
	const digits = "0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = digits[rand.Intn(len(digits))]
	}
	return string(b)
}

func MapToStruct(m map[string]any, result any) error {
	data, err := json.Marshal(m)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, result)
}
