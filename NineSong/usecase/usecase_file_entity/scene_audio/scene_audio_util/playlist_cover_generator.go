package usercase_audio_util

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"

	"github.com/EdlinOrg/prominentcolor"
	"github.com/disintegration/imaging"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// PlaylistCoverGenerator 播放列表封面生成器
type PlaylistCoverGenerator struct {
	coverBasePath string
}

// NewPlaylistCoverGenerator 创建播放列表封面生成器
func NewPlaylistCoverGenerator(coverBasePath string) *PlaylistCoverGenerator {
	return &PlaylistCoverGenerator{
		coverBasePath: coverBasePath,
	}
}

// GeneratePlaylistCover 生成播放列表封面图片（总是强制重新生成）
// firstTrackCoverPath: 第一首歌的封面图片路径
// playlistID: 播放列表ID
// 返回生成的封面图片路径
func (g *PlaylistCoverGenerator) GeneratePlaylistCover(
	ctx context.Context,
	firstTrackCoverPath string,
	playlistID primitive.ObjectID,
) (string, error) {
	// 如果第一首歌没有封面，返回空字符串
	if firstTrackCoverPath == "" {
		return "", nil
	}

	// 创建播放列表封面目录
	playlistCoverDir := filepath.Join(g.coverBasePath, "playlist")
	if err := os.MkdirAll(playlistCoverDir, 0755); err != nil {
		return "", fmt.Errorf("创建播放列表封面目录失败: %w", err)
	}

	// 生成封面图片路径
	coverFileName := fmt.Sprintf("playlist_%s.jpg", playlistID.Hex())
	coverPath := filepath.Join(playlistCoverDir, coverFileName)

	// 读取第一首歌的封面图片
	srcImg, err := imaging.Open(firstTrackCoverPath)
	if err != nil {
		return "", fmt.Errorf("打开封面图片失败: %w", err)
	}

	// 压缩图片到合适大小（500x500）以便颜色提取
	resizedImg := imaging.Resize(srcImg, 500, 500, imaging.Lanczos)

	// 提取主要颜色
	colors, err := prominentcolor.Kmeans(resizedImg)
	if err != nil {
		return "", fmt.Errorf("提取颜色失败: %w", err)
	}

	// 如果没有提取到颜色，使用默认颜色
	if len(colors) == 0 {
		colors = []prominentcolor.ColorItem{
			{Color: prominentcolor.ColorRGB{R: 100, G: 100, B: 100}},
			{Color: prominentcolor.ColorRGB{R: 50, G: 50, B: 50}},
		}
	}

	// 生成渐变封面图片（800x800）
	coverImg := g.generateGradientCover(colors, 800, 800)

	// 保存封面图片（JPEG质量80，优化压缩）
	if err := g.saveCoverImageOptimized(coverImg, coverPath); err != nil {
		return "", fmt.Errorf("保存封面图片失败: %w", err)
	}

	// 返回相对路径（相对于coverBasePath）
	relativePath := filepath.Join("playlist", coverFileName)
	return relativePath, nil
}

// generateGradientCover 基于提取的颜色生成渐变封面
func (g *PlaylistCoverGenerator) generateGradientCover(
	colors []prominentcolor.ColorItem,
	width, height int,
) image.Image {
	// 创建新图片
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// 获取主要颜色（最多3个）
	numColors := len(colors)
	if numColors > 3 {
		numColors = 3
	}

	if numColors == 0 {
		// 默认灰色渐变
		g.drawGradient(img, color.RGBA{R: 100, G: 100, B: 100, A: 255}, color.RGBA{R: 50, G: 50, B: 50, A: 255})
		return img
	}

	// 根据颜色数量生成不同的渐变效果
	switch numColors {
	case 1:
		// 单色渐变（从亮到暗）
		c := colors[0].Color
		// prominentcolor.ColorRGB 的字段是 uint32，需要转换
		r := uint32(c.R)
		green := uint32(c.G)
		blue := uint32(c.B)
		lightColor := color.RGBA{
			R: uint8(min(255, int(r+30))),
			G: uint8(min(255, int(green+30))),
			B: uint8(min(255, int(blue+30))),
			A: 255,
		}
		darkColor := color.RGBA{
			R: uint8(max(0, int(r-30))),
			G: uint8(max(0, int(green-30))),
			B: uint8(max(0, int(blue-30))),
			A: 255,
		}
		g.drawGradient(img, lightColor, darkColor)
	case 2:
		// 双色渐变（从左到右）
		c1 := color.RGBA{R: uint8(colors[0].Color.R), G: uint8(colors[0].Color.G), B: uint8(colors[0].Color.B), A: 255}
		c2 := color.RGBA{R: uint8(colors[1].Color.R), G: uint8(colors[1].Color.G), B: uint8(colors[1].Color.B), A: 255}
		g.drawGradient(img, c1, c2)
	default:
		// 三色渐变（对角线）
		c1 := color.RGBA{R: uint8(colors[0].Color.R), G: uint8(colors[0].Color.G), B: uint8(colors[0].Color.B), A: 255}
		c2 := color.RGBA{R: uint8(colors[1].Color.R), G: uint8(colors[1].Color.G), B: uint8(colors[1].Color.B), A: 255}
		c3 := color.RGBA{R: uint8(colors[2].Color.R), G: uint8(colors[2].Color.G), B: uint8(colors[2].Color.B), A: 255}
		g.drawTripleGradient(img, c1, c2, c3)
	}

	return img
}

// drawGradient 绘制线性渐变
func (g *PlaylistCoverGenerator) drawGradient(
	img *image.RGBA,
	color1, color2 color.RGBA,
) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	for y := 0; y < height; y++ {
		// 计算当前行的颜色插值
		t := float64(y) / float64(height)
		r := uint8(float64(color1.R)*(1-t) + float64(color2.R)*t)
		green := uint8(float64(color1.G)*(1-t) + float64(color2.G)*t)
		b := uint8(float64(color1.B)*(1-t) + float64(color2.B)*t)

		c := color.RGBA{R: r, G: green, B: b, A: 255}
		for x := 0; x < width; x++ {
			img.Set(x, y, c)
		}
	}
}

// drawTripleGradient 绘制三色对角线渐变
func (g *PlaylistCoverGenerator) drawTripleGradient(
	img *image.RGBA,
	color1, color2, color3 color.RGBA,
) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// 计算对角线位置
			diag := float64(x+y) / float64(width+height)

			var c color.RGBA
			if diag < 0.5 {
				// 前两色之间插值
				t := diag * 2
				c = color.RGBA{
					R: uint8(float64(color1.R)*(1-t) + float64(color2.R)*t),
					G: uint8(float64(color1.G)*(1-t) + float64(color2.G)*t),
					B: uint8(float64(color1.B)*(1-t) + float64(color2.B)*t),
					A: 255,
				}
			} else {
				// 后两色之间插值
				t := (diag - 0.5) * 2
				c = color.RGBA{
					R: uint8(float64(color2.R)*(1-t) + float64(color3.R)*t),
					G: uint8(float64(color2.G)*(1-t) + float64(color3.G)*t),
					B: uint8(float64(color2.B)*(1-t) + float64(color3.B)*t),
					A: 255,
				}
			}
			img.Set(x, y, c)
		}
	}
}

// saveCoverImage 保存封面图片
func (g *PlaylistCoverGenerator) saveCoverImage(img image.Image, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	// 使用JPEG格式保存，质量80%（平衡清晰度和文件大小）
	return jpeg.Encode(file, img, &jpeg.Options{Quality: 80})
}

// saveCoverImageOptimized 保存封面图片（优化压缩，与 saveCoverImage 相同）
func (g *PlaylistCoverGenerator) saveCoverImageOptimized(img image.Image, path string) error {
	return g.saveCoverImage(img, path)
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max 返回两个整数中的较大值
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
