package scene_audio_route_models

// FileInfo 音频文件详细信息模型
type FileInfo struct {
	ID          string  `json:"id"`
	FileName    string  `json:"file_name"`
	FileSize    int64   `json:"file_size"`
	FileType    string  `json:"file_type"`
	MimeType    string  `json:"mime_type"`
	Format      string  `json:"format"`
	Duration    float64 `json:"duration"`    // 音频时长，单位秒
	Bitrate     int     `json:"bitrate"`     // 比特率，单位kbps
	SampleRate  int     `json:"sample_rate"` // 采样率，单位Hz
	Channels    int     `json:"channels"`    // 声道数
	Encoding    string  `json:"encoding"`    // 编码方式
	Artist      string  `json:"artist"`      // 艺术家
	Title       string  `json:"title"`       // 标题
	Album       string  `json:"album"`       // 专辑
	Year        int     `json:"year"`        // 年份
	Genre       string  `json:"genre"`       // 流派
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

// ToFileInfoResponse 将数据库模型转换为路由响应模型
func ToFileInfoResponse(dbModel interface{}) *FileInfo {
	// 这里应该根据实际的数据库模型进行转换
	// 由于目前还没有具体的数据库模型定义，这里只是一个框架
	return &FileInfo{}
}