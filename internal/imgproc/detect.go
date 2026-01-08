package imgproc

import (
	"bytes"
	"file-storage/internal/errs"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func ImageConfig(data []byte) (Extension string, Width int, Height int, Error error) {

	imgCfg, ext, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", 0, 0, err
	}

	_, ok := supportedFormats[ImgFormat(ext)]
	if !ok {
		return "", 0, 0, fmt.Errorf("image format %s: %w", ext, errs.ErrUnsupportedImageFormat)
	}

	return ext, imgCfg.Width, imgCfg.Height, nil
}

func SupportedFormat(ext string) bool {
	_, ok := supportedFormats[ImgFormat(ext)]
	return ok
}
