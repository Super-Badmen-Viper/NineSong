package scene_audio_route_interface

// ProgressCallback 进度回调函数类型
// 参数: progress (0-100), status (当前状态), msg (详细信息)
type ProgressCallback func(progress float64, status string, msg string)