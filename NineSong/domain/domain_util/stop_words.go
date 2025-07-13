package domain_util

import (
	"bufio"
	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain/domain_resource"
	"io/fs"
	"log"
	"regexp"
	"strings"
	"unicode/utf8"
)

func LoadCombinedStopWords() (map[string]bool, error) {
	combined := make(map[string]bool)

	// 添加基础英文停用词（SEO标准列表），全部转为小写
	engStopwords := []string{"you", "me", "my", "it", "to", "on", "in", "that", "and",
		"be", "no", "don", "up", "we", "oh", "yeah", "na", "la", "your", "is", "are",
		"was", "were", "its", "this", "these", "those", "there", "here", "studio",
		"album", "track", "song", "music", "audio", "file", "mp3", "wav", "flac"}
	for _, word := range engStopwords {
		combined[strings.ToLower(word)] = true
	}

	// 添加音乐元数据术语 (中文通常不区分大小写，但保持一致性)
	musicTerms := []string{"作曲", "作词", "编曲", "混音", "制作", "演唱", "作曲人", "作词人",
		"录音室", "发行", "专辑", "单曲", "音乐人", "乐队", "乐团", "演唱会", "音乐", "歌曲",
		"音频", "音质", "音效", "音乐风格", "流派", "节奏", "旋律", "和声", "歌词", "曲目",
		"音轨", "音频文件", "音乐版权", "音乐制作人", "音乐发行", "音乐平台", "数字音乐"}
	for _, term := range musicTerms {
		combined[strings.ToLower(term)] = true
	}

	// 从嵌入的文件系统加载和合并词表
	err := fs.WalkDir(domain_resource.StopWordsFS, ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			content, readErr := domain_resource.StopWordsFS.ReadFile(path)
			if readErr != nil {
				log.Printf("警告: 无法读取嵌入的停用词文件: %s (%v)", path, readErr)
				return nil // 继续处理下一个文件
			}
			words, parseErr := loadStopWordsFromContent(string(content))
			if parseErr != nil {
				log.Printf("警告: 解析停用词文件失败: %s (%v)", path, parseErr)
				return nil // 继续处理下一个文件
			}
			for word := range words {
				combined[word] = true // loadStopWordsFromContent 已经处理了小写转换
			}
		}
		return nil
	})

	if err != nil {
		log.Printf("错误: 遍历嵌入的停用词目录失败: %v", err)
	}

	return combined, err
}

func loadStopWordsFromContent(content string) (map[string]bool, error) {
	result := make(map[string]bool)
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		if word != "" && !strings.HasPrefix(word, "//") {
			// 关键：所有停用词都以小写形式存储
			result[strings.ToLower(word)] = true
		}
	}
	return result, scanner.Err()
}

func shouldSkipWord(word string, stopWords map[string]bool) bool {
	// 1. 长度过滤
	if utf8.RuneCountInString(word) < 2 {
		return true
	}
	// 2. 停用词匹配
	if stopWords[word] {
		return true
	}
	// 3. 特殊字符检测（新增实现）
	if ContainsInvalidChars(word) {
		return true
	}
	// 4. 英文变体处理
	baseForm := getBaseForm(word)
	return stopWords[baseForm]
}

// 辅助函数：获取词干原型
func getBaseForm(word string) string {
	// 简单处理英文简写
	if strings.HasSuffix(word, "'s") || strings.HasSuffix(word, "'t") {
		return word[:len(word)-2]
	}
	return word
}

func isEnglishWord(s string) bool {
	for _, r := range s {
		if r < 128 {
			return true
		}
	}
	return false
}

// ContainsInvalidChars 新增：字符级过滤函数（基于亚马逊SKU禁用字符规范[9](@ref)）
func ContainsInvalidChars(word string) bool {
	// 匹配所有控制字符（\n, \r, \t等）和特殊Unicode字符
	// Go regexp不支持\uXXXX，需要使用\x{XXXX}格式
	invalidRegex := regexp.MustCompile(`[\x00-\x1F\x7F\x{200B}-\x{200F}\x{2028}\x{2029}\x{FEFF}]`)
	return invalidRegex.MatchString(word)
}
