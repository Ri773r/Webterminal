package utils

import (
	"math/rand"
	"os"
	"time"
)

// GetRandomString Get random string of specified length
func GetRandomString(length int) string {
	str := "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	rand.Seed(time.Now().UnixNano())
	bytes := []byte(str)
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = bytes[rand.Intn(len(bytes))]
	}
	return string(result)
}

// IsExist Determine if the file exists
func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil || os.IsExist(err)
}
