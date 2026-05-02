package imgproc

import (
	"bytes"
	"file-storage/internal/errs"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"strings"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

// ImageConfig detects image format and dimensions from raw file data.
func ImageConfig(data []byte) (format ImgFormat, w, h int, err error) {

	var emptyImgFormat ImgFormat

	imgCfg, ext, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return emptyImgFormat, 0, 0, fmt.Errorf("decode image error: %w: %v", errs.ErrInvalidImage, err)
	}

	ext = strings.ToLower(ext)
	format, ok := SupportedInputFormat(ext)
	if !ok {
		return emptyImgFormat, 0, 0, fmt.Errorf("unsupported image format %s: %w", ext, errs.ErrUnsupportedImageFormat)
	}

	return format, imgCfg.Width, imgCfg.Height, nil
}

// SupportedInputFormat reports whether the provided format is accepted as an upload input format.
func SupportedInputFormat(format string) (ImgFormat, bool) {
	imgFormat, ok := supportedInputFormats[ImgFormat(format)]
	return imgFormat, ok
}

// SupportedOutputFormat reports whether the provided format can be used for stored image output.
func SupportedOutputFormat(format string) (ImgFormat, bool) {
	imgFormat, ok := supportedOutputFormats[ImgFormat(format)]
	return imgFormat, ok
}
