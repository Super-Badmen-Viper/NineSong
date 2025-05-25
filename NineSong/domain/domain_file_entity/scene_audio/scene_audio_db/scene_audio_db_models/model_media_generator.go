package scene_audio_db_models

import "go.mongodb.org/mongo-driver/bson/primitive"

type MediaGeneratorMetadata struct {
	ID            primitive.ObjectID `bson:"_id"`
	GeneratorType string             `bson:"generator_type"` // 音频生成的类型（如正弦波、方波等）
	Frequency     float64            `bson:"frequency"`      // 基础频率（Hz）
	Waveform      string             `bson:"waveform"`       // 波形类型（如正弦波、三角波等）

	// 合成参数
	Harmonics    int     `bson:"harmonics"`     // 谐波数量[9](@ref)
	PhaseNoise   float64 `bson:"phase_noise"`   // 相位噪声强度[14](@ref)
	FormantRatio float64 `bson:"formant_ratio"` // 共振峰比例[10](@ref)

	// 生成模型参数
	GANLatentVector []float64 `bson:"gan_latent"` // 生成对抗网络潜向量[14](@ref)
	VAEKLDivergence float64   `bson:"vae_kld"`    // 变分自编码器KL散度[14](@ref)
}
