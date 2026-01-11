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

func SupportedInputFormat(format string) (ImgFormat, bool) {
	imgFormat, ok := supportedInputFormats[ImgFormat(format)]
	return imgFormat, ok
}

func SupportedOutputFormat(format string) (ImgFormat, bool) {
	imgFormat, ok := supportedOutputFormats[ImgFormat(format)]
	return imgFormat, ok
}
