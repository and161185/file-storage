package imgproc

import (
	"bytes"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"

	_ "golang.org/x/image/bmp"
	_ "golang.org/x/image/tiff"
	_ "golang.org/x/image/webp"
)

func IsImage(data []byte) bool {

	_, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return false
	}

	return true
}
