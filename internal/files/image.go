package files

import (
	"bytes"
	"file-storage/internal/errs"
	"file-storage/internal/imgproc"
	"fmt"

	"github.com/disintegration/imaging"
)

func ProcessImage(b []byte, wantExt string, wantWidth int, wantHeight int) ([]byte, *ImageInfo, error) {
	wantFormat, ok := imgproc.SupportedOutputFormat(wantExt)
	if !ok {
		return nil, nil, fmt.Errorf("unsupported want image format %s: %w", wantExt, errs.ErrUnsupportedImageFormat)
	}

	format, width, height, err := imgproc.ImageConfig(b)
	if err != nil {
		return nil, nil, fmt.Errorf("config observing error: %w", err)
	}

	if width <= 0 || height <= 0 {
		return nil, nil, fmt.Errorf("invalid image dimentions: %w", errs.ErrInvalidImage)
	}

	multiplierW := float64(wantWidth) / float64(width)
	multiplierH := float64(wantHeight) / float64(height)
	multiplier := min(multiplierW, multiplierH)

	if format == wantFormat && multiplier >= 1 {
		imageInfo := ImageInfo{
			Format: format,
			Width:  width,
			Height: height,
		}

		return b, &imageInfo, nil
	}

	reader := bytes.NewReader(b)
	img, err := imaging.Decode(reader)
	if err != nil {
		return nil, nil, fmt.Errorf("decode image error: %w", err)
	}

	if multiplier < 1 {
		img = imgproc.Resize(img, multiplier)
	}

	imagingFormat, err := imgproc.ImagingOutputFormat(wantFormat)
	if err != nil {
		return nil, nil, fmt.Errorf("output format error: %w", err)
	}

	result, err := imgproc.Encode(img, imagingFormat)
	if err != nil {
		return nil, nil, fmt.Errorf("encode image error: %w", err)
	}

	imageInfo := ImageInfo{
		Format: wantFormat,
		Width:  img.Bounds().Dx(),
		Height: img.Bounds().Dy(),
	}

	return result, &imageInfo, nil
}
