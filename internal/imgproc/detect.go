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

func MimeType(data []byte) (string, error) {

	_, ext, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return "", err
	}

	mimeType, ok := supportedFormats[ImgFormat(ext)]
	if !ok {
		return "", fmt.Errorf("image fomat %s: %w", ext, errs.ErrUnsupportedImageFormat)
	}

	return mimeType, nil
}
