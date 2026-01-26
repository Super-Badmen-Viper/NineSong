package core

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlaylistMetadata 播放列表元数据
type PlaylistMetadata struct {
	ID          primitive.ObjectID `bson:"_id"`
	Name        string             `bson:"name"`
	Comment     string             `bson:"comment"`
	Duration    float64            `bson:"duration"`
	SongCount   float64            `bson:"song_count"`
	CreatedAt   time.Time          `bson:"created_at"`
	UpdatedAt   time.Time          `bson:"updated_at"`
	Path        string             `bson:"path"`
	Sync        bool               `bson:"sync"`
	Size        int                `bson:"size"`
	Rules       string             `bson:"rules"`
	EvaluatedAt time.Time          `bson:"evaluated_at"`
}

// Validate 验证播放列表元数据
func (p *PlaylistMetadata) Validate() error {
	// 可以添加验证逻辑
	return nil
}

// SetTimestamps 设置时间戳
func (p *PlaylistMetadata) SetTimestamps() {
	now := time.Now()
	if p.CreatedAt.IsZero() {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
}
