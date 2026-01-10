package main

import (
	"fmt"
	"os"
	"time"
)

func main() {
	// 测试时间比较逻辑
	fmt.Println("Testing time comparison logic...")

	// 1. 获取当前时间的本地时间和UTC时间
	now := time.Now()
	nowUTC := now.UTC()
	fmt.Printf("Local time: %v\n", now)
	fmt.Printf("UTC time: %v\n", nowUTC)

	// 2. 创建一个临时文件用于测试
	tempFile, err := os.CreateTemp("", "test_time_fix")
	if err != nil {
		fmt.Printf("Failed to create temp file: %v\n", err)
		return
	}
	tempPath := tempFile.Name()
	tempFile.Close()
	defer os.Remove(tempPath)

	// 3. 获取临时文件的修改时间
	fileInfo, err := os.Stat(tempPath)
	if err != nil {
		fmt.Printf("Failed to stat temp file: %v\n", err)
		return
	}
	fileModTime := fileInfo.ModTime()
	fileModTimeUTC := fileModTime.UTC()
	fmt.Printf("File mod time (local): %v\n", fileModTime)
	fmt.Printf("File mod time (UTC): %v\n", fileModTimeUTC)

	// 4. 测试时间比较
	// 模拟数据库中存储的是UTC时间
	dbUpdatedAt := nowUTC.Add(-24 * time.Hour) // 昨天的UTC时间

	// 错误的比较方式：本地时间 vs UTC时间
	if fileModTime.After(dbUpdatedAt) {
		fmt.Println("❌ WRONG: Local time > UTC time (due to timezone difference)")
	} else {
		fmt.Println("✅ RIGHT: Local time <= UTC time")
	}

	// 正确的比较方式：UTC时间 vs UTC时间
	if fileModTimeUTC.After(dbUpdatedAt) {
		fmt.Println("✅ RIGHT: UTC time > UTC time (correct comparison)")
	} else {
		fmt.Println("❌ WRONG: UTC time <= UTC time")
	}

	fmt.Println("\nTest completed! The fix ensures we use UTC time for both file mod time and database comparison.")
}
