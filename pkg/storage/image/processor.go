package image

import (
	"context"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/gif"
	"image/jpeg"
	"image/png"
	"io"
	"strings"

	"api-core/pkg/storage/interfaces"

	"github.com/disintegration/imaging"
)

// ImageProcessor implementation cho xử lý ảnh
type ImageProcessor struct {
	quality int // JPEG quality (1-100)
}

// NewImageProcessor tạo instance mới của ImageProcessor
func NewImageProcessor(quality int) *ImageProcessor {
	if quality <= 0 || quality > 100 {
		quality = 90 // Default quality
	}

	return &ImageProcessor{
		quality: quality,
	}
}

// Resize ảnh
func (p *ImageProcessor) Resize(ctx context.Context, reader io.Reader, width, height int) (io.Reader, error) {
	// Decode image
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Resize image
	resized := imaging.Resize(img, width, height, imaging.Lanczos)

	// Encode back to reader
	return p.encodeImage(resized, format)
}

// Crop ảnh
func (p *ImageProcessor) Crop(ctx context.Context, reader io.Reader, x, y, width, height int) (io.Reader, error) {
	// Decode image
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Create crop rectangle
	bounds := img.Bounds()
	cropRect := image.Rect(x, y, x+width, y+height)

	// Ensure crop rectangle is within image bounds
	if cropRect.Min.X < bounds.Min.X {
		cropRect.Min.X = bounds.Min.X
	}
	if cropRect.Min.Y < bounds.Min.Y {
		cropRect.Min.Y = bounds.Min.Y
	}
	if cropRect.Max.X > bounds.Max.X {
		cropRect.Max.X = bounds.Max.X
	}
	if cropRect.Max.Y > bounds.Max.Y {
		cropRect.Max.Y = bounds.Max.Y
	}

	// Crop image
	cropped := imaging.Crop(img, cropRect)

	// Encode back to reader
	return p.encodeImage(cropped, format)
}

// AddWatermark thêm watermark
func (p *ImageProcessor) AddWatermark(ctx context.Context, reader io.Reader, watermarkPath string, position string) (io.Reader, error) {
	// Decode main image
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode main image: %w", err)
	}

	// Load watermark image
	watermark, err := imaging.Open(watermarkPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load watermark: %w", err)
	}

	// Resize watermark to fit (max 20% of main image)
	mainBounds := img.Bounds()
	maxWatermarkSize := mainBounds.Dx() / 5
	if watermark.Bounds().Dx() > maxWatermarkSize {
		watermark = imaging.Resize(watermark, maxWatermarkSize, 0, imaging.Lanczos)
	}

	// Calculate watermark position
	var watermarkPos image.Point
	switch strings.ToLower(position) {
	case "top-left":
		watermarkPos = image.Point{X: 10, Y: 10}
	case "top-right":
		watermarkPos = image.Point{X: mainBounds.Dx() - watermark.Bounds().Dx() - 10, Y: 10}
	case "bottom-left":
		watermarkPos = image.Point{X: 10, Y: mainBounds.Dy() - watermark.Bounds().Dy() - 10}
	case "bottom-right":
		watermarkPos = image.Point{X: mainBounds.Dx() - watermark.Bounds().Dx() - 10, Y: mainBounds.Dy() - watermark.Bounds().Dy() - 10}
	case "center":
		watermarkPos = image.Point{
			X: (mainBounds.Dx() - watermark.Bounds().Dx()) / 2,
			Y: (mainBounds.Dy() - watermark.Bounds().Dy()) / 2,
		}
	default:
		watermarkPos = image.Point{X: 10, Y: 10} // Default to top-left
	}

	// Create new image with watermark
	result := imaging.Clone(img)
	result = imaging.Overlay(result, watermark, watermarkPos, 0.7) // 70% opacity

	// Encode back to reader
	return p.encodeImage(result, format)
}

// Convert format
func (p *ImageProcessor) Convert(ctx context.Context, reader io.Reader, format string) (io.Reader, error) {
	// Decode image
	img, _, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	// Encode to new format
	return p.encodeImage(img, format)
}

// GetInfo lấy thông tin ảnh
func (p *ImageProcessor) GetInfo(ctx context.Context, reader io.Reader) (*interfaces.ImageInfo, error) {
	// Decode image
	img, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}

	bounds := img.Bounds()

	// Get color model
	var colorModel string
	switch img.ColorModel() {
	case color.RGBAModel:
		colorModel = "RGBA"
	case color.RGBA64Model:
		colorModel = "RGBA64"
	case color.NRGBAModel:
		colorModel = "NRGBA"
	case color.NRGBA64Model:
		colorModel = "NRGBA64"
	case color.AlphaModel:
		colorModel = "Alpha"
	case color.Alpha16Model:
		colorModel = "Alpha16"
	case color.GrayModel:
		colorModel = "Gray"
	case color.Gray16Model:
		colorModel = "Gray16"
	default:
		colorModel = "Unknown"
	}

	return &interfaces.ImageInfo{
		Width:      bounds.Dx(),
		Height:     bounds.Dy(),
		Format:     strings.ToUpper(format),
		ColorModel: colorModel,
		Size:       0, // Size will be calculated when encoding
	}, nil
}

// encodeImage encode image to reader
func (p *ImageProcessor) encodeImage(img image.Image, format string) (io.Reader, error) {
	// Create pipe for streaming
	reader, writer := io.Pipe()

	go func() {
		defer writer.Close()

		switch strings.ToLower(format) {
		case "jpeg", "jpg":
			err := jpeg.Encode(writer, img, &jpeg.Options{Quality: p.quality})
			if err != nil {
				writer.CloseWithError(err)
				return
			}
		case "png":
			err := png.Encode(writer, img)
			if err != nil {
				writer.CloseWithError(err)
				return
			}
		case "gif":
			// Convert to palette for GIF
			paletted := image.NewPaletted(img.Bounds(), nil)
			draw.Draw(paletted, paletted.Bounds(), img, img.Bounds().Min, draw.Src)

			err := gif.Encode(writer, paletted, nil)
			if err != nil {
				writer.CloseWithError(err)
				return
			}
		default:
			// Default to JPEG
			err := jpeg.Encode(writer, img, &jpeg.Options{Quality: p.quality})
			if err != nil {
				writer.CloseWithError(err)
				return
			}
		}
	}()

	return reader, nil
}

// ValidateImage kiểm tra file có phải ảnh hợp lệ không
func (p *ImageProcessor) ValidateImage(reader io.Reader) error {
	_, _, err := image.Decode(reader)
	return err
}

// GetImageFormat lấy format của ảnh
func (p *ImageProcessor) GetImageFormat(reader io.Reader) (string, error) {
	_, format, err := image.Decode(reader)
	return format, err
}
