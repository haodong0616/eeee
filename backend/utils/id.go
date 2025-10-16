package utils

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

// GenerateObjectID 生成MongoDB风格的24位16进制字符串ID
func GenerateObjectID() string {
	timestamp := time.Now().Unix()
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	
	// 4字节时间戳 + 8字节随机数 = 12字节 = 24位16进制字符
	id := fmt.Sprintf("%08x%s", timestamp, hex.EncodeToString(randomBytes))
	return id
}

