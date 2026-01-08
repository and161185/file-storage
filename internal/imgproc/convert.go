package imgproc

import (
	"bytes"
	"image"

	"github.com/disintegration/imaging"
)

func Resize(img image.Image, multiplier float64) image.Image {

	newWidth := int(float64(img.Bounds().Dx()) * multiplier)
	newHeight := int(float64(img.Bounds().Dy()) * multiplier)
	return imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)
}

func Encode(img image.Image, format imaging.Format) ([]byte, error) {
	buf := new(bytes.Buffer)
	err := imaging.Encode(buf, img, format)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
