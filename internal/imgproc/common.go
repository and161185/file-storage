package imgproc

type ImgFormat string

const (
	ImgFormatBMP  ImgFormat = "bmp"
	ImgFormatJPG  ImgFormat = "jpg"
	ImgFormatPNG  ImgFormat = "png"
	ImgFormatGIF  ImgFormat = "gif"
	ImgFormatTIFF ImgFormat = "tiff"
	ImgFormatWEBP ImgFormat = "webp"
)

var supportedFormats = map[ImgFormat]ImgFormat{
	ImgFormatBMP:  ImgFormatBMP,
	ImgFormatJPG:  ImgFormatJPG,
	ImgFormatPNG:  ImgFormatPNG,
	ImgFormatGIF:  ImgFormatGIF,
	ImgFormatTIFF: ImgFormatTIFF,
	ImgFormatWEBP: ImgFormatWEBP,
}
