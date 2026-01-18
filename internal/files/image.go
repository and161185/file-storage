package files

import (
	"bytes"
	"file-storage/internal/errs"
	"file-storage/internal/imgproc"
	"fmt"

	"github.com/disintegration/imaging"
)

// ProcessImage converts image data to requested format and size
func ProcessImage(b []byte, targetExt string, targetWidth int, targetHeight int) ([]byte, *ImageInfo, error) {
	targetFormat, ok := imgproc.SupportedOutputFormat(targetExt)
	if !ok {
		return nil, nil, fmt.Errorf("unsupported target image format %s: %w", targetExt, errs.ErrUnsupportedImageFormat)
	}

	format, width, height, err := imgproc.ImageConfig(b)
	if err != nil {
		return nil, nil, fmt.Errorf("config observing error: %w", err)
	}

	if width <= 0 || height <= 0 {
		return nil, nil, fmt.Errorf("invalid image dimensions: %w", errs.ErrInvalidImage)
	}

	if targetWidth <= 0 || targetHeight <= 0 {
		return nil, nil, fmt.Errorf("invalid target image dimensions: %w", errs.ErrInvalidImage)
	}

	multiplierW := float64(targetWidth) / float64(width)
	multiplierH := float64(targetHeight) / float64(height)
	multiplier := min(multiplierW, multiplierH)

	if format == targetFormat && multiplier >= 1 {
		imageInfo := ImageInfo{
			Format: format,
			Width:  width,
			Height: height,
		}

		return b, &imageInfo, nil
	}

	imagingFormat, err := imgproc.ImagingOutputFormat(targetFormat)
	if err != nil {
		return nil, nil, fmt.Errorf("output format error: %w", err)
	}

	reader := bytes.NewReader(b)
	img, err := imaging.Decode(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("decode image error: %w", err)
	}

	if multiplier < 1 {
		img = imgproc.Resize(img, multiplier)
	}

	result, err := imgproc.Encode(img, imagingFormat)
	if err != nil {
		return nil, nil, fmt.Errorf("encode image error: %w", err)
	}

	imageInfo := ImageInfo{
		Format: targetFormat,
		Width:  img.Bounds().Dx(),
		Height: img.Bounds().Dy(),
	}

	return result, &imageInfo, nil
}
