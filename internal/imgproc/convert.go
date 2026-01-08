package imgproc

import (
	"bytes"

	"github.com/disintegration/imaging"
)

func Convert(b []byte, srcExt string, srcWidth, srcHeight int, dstExt string, dstWidth, dstHeight int) ([]byte, error) {

	multiplierW := float64(dstWidth) / float64(srcWidth)
	multiplierH := float64(dstHeight) / float64(srcWidth)
	multiplier := min(multiplierW, multiplierH)

	if srcExt == dstExt && multiplier >= 1 {
		return b, nil
	}

	reader := bytes.NewReader(b)

	img, err := imaging.Decode(reader)
	if err != nil {
		return nil, err
	}

	if multiplier < 1 {
		newWidth := int(float64(srcWidth) * multiplier)
		newHeight := int(float64(srcHeight) * multiplier)
		img = imaging.Resize(img, newWidth, newHeight, imaging.BSpline)
	}

	format, err := imaging.FormatFromExtension(dstExt)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	err = imaging.Encode(buf, img, format)
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}
