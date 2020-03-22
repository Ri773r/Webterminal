package utils

import (
	"math/rand"
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
