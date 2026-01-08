package imgproc

import (
	"file-storage/internal/errs"

	"github.com/disintegration/imaging"
)

type ImgFormat string

const (
	ImgFormatBMP  ImgFormat = "bmp"
	ImgFormatJPG  ImgFormat = "jpg"
	ImgFormatJPEG ImgFormat = "jpeg"
	ImgFormatPNG  ImgFormat = "png"
	ImgFormatGIF  ImgFormat = "gif"
	ImgFormatTIFF ImgFormat = "tiff"
	ImgFormatWEBP ImgFormat = "webp"
)

var supportedInputFormats = map[ImgFormat]ImgFormat{
	ImgFormatBMP:  ImgFormatBMP,
	ImgFormatJPG:  ImgFormatJPEG,
	ImgFormatJPEG: ImgFormatJPEG,
	ImgFormatPNG:  ImgFormatPNG,
	ImgFormatGIF:  ImgFormatGIF,
	ImgFormatTIFF: ImgFormatTIFF,
	ImgFormatWEBP: ImgFormatWEBP,
}

var supportedOutputFormats = map[ImgFormat]ImgFormat{
	ImgFormatBMP:  ImgFormatBMP,
	ImgFormatJPG:  ImgFormatJPEG,
	ImgFormatJPEG: ImgFormatJPEG,
	ImgFormatPNG:  ImgFormatPNG,
	ImgFormatGIF:  ImgFormatGIF,
}

func ImagingOutputFormat(imf ImgFormat) (imaging.Format, error) {
	switch imf {
	case ImgFormatBMP:
		return imaging.BMP, nil
	case ImgFormatJPEG:
		return imaging.JPEG, nil
	case ImgFormatGIF:
		return imaging.GIF, nil
	case ImgFormatPNG:
		return imaging.PNG, nil
	default:
		return 0, errs.ErrUnsupportedImageFormat
	}
}
