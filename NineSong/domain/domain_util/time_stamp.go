package domain_util

import "time"

// IsTimestampFolder 检查是否为时间戳格式目录（YYYYMMDDHHMMSS）
func IsTimestampFolder(name string) bool {
	if len(name) != 14 { // 20060102150405 长度固定为14
		return false
	}
	_, err := time.Parse("20060102150405", name)
	return err == nil
}
